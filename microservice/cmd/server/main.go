package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/goldenpipe/microservice/internal/api"
	"github.com/goldenpipe/microservice/internal/kubevirt"
	"github.com/goldenpipe/microservice/internal/storage"
	"github.com/goldenpipe/microservice/pkg/config"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "goldenpipe",
		Short: "GoldenPipe - VM Golden Image Automation Microservice",
		Long:  "A Kubernetes-native microservice that automates the creation of custom VM golden images for both Linux and Windows.",
		Run:   runServer,
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.goldenpipe.yaml)")

	// Configuration flags
	rootCmd.Flags().String("kubeconfig", "", "path to kubeconfig file")
	rootCmd.Flags().String("namespace", "goldenpipe-system", "Kubernetes namespace")
	rootCmd.Flags().String("storage-class", "rook-ceph-block", "Storage class for persistent volumes")
	rootCmd.Flags().Int("port", 8080, "API server port")
	rootCmd.Flags().String("log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.Flags().Int("max-concurrent-vms", 5, "Maximum number of concurrent VM operations")

	// Bind flags to viper
	viper.BindPFlags(rootCmd.Flags())
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".goldenpipe")
	}

	viper.AutomaticEnv()
	viper.ReadInConfig()
}

func runServer(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg := config.Load()

	// Setup logging
	logger := logrus.New()
	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		logger.Fatal("Invalid log level:", err)
	}
	logger.SetLevel(level)

	logger.WithFields(logrus.Fields{
		"namespace": cfg.Namespace,
		"port":      cfg.Port,
		"log_level": cfg.LogLevel,
	}).Info("Starting GoldenPipe server")

	// Initialize Kubernetes client
	kubeClient, err := kubevirt.NewClient(cfg.KubeconfigPath)
	if err != nil {
		logger.Fatal("Failed to create Kubernetes client:", err)
	}

	// Initialize storage manager
	storageManager, err := storage.NewManager(kubeClient.KubeClient, cfg.StorageClass, cfg.Namespace)
	if err != nil {
		logger.Fatal("Failed to create storage manager:", err)
	}

	// Initialize KubeVirt manager
	vmManager, err := kubevirt.NewManager(kubeClient, storageManager, cfg.Namespace)
	if err != nil {
		logger.Fatal("Failed to create KubeVirt manager:", err)
	}

	// Setup Gin router
	if cfg.LogLevel != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Initialize API handlers
	apiHandler := api.NewHandler(vmManager, storageManager, logger)
	apiHandler.SetupRoutes(router)

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.WithField("port", cfg.Port).Info("Starting HTTP server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server:", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown:", err)
	}

	logger.Info("Server exited")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
