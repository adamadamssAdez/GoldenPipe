package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/goldenpipe/microservice/internal/kubevirt"
	"github.com/goldenpipe/microservice/internal/storage"
	"github.com/goldenpipe/microservice/pkg/types"
)

// Handler handles HTTP requests for the GoldenPipe API
type Handler struct {
	vmManager      *kubevirt.Manager
	storageManager *storage.Manager
	logger         *logrus.Logger
}

// NewHandler creates a new API handler
func NewHandler(vmManager *kubevirt.Manager, storageManager *storage.Manager, logger *logrus.Logger) *Handler {
	return &Handler{
		vmManager:      vmManager,
		storageManager: storageManager,
		logger:         logger,
	}
}

// auditLog logs security-relevant events for compliance
func (h *Handler) auditLog(event, action, sourceIP, details string) {
	h.logger.WithFields(logrus.Fields{
		"audit_event": event,
		"action":      action,
		"source_ip":   sourceIP,
		"details":     details,
	}).Info("Audit log entry")
}

// SetupRoutes configures all API routes
func (h *Handler) SetupRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		// Image management endpoints
		v1.POST("/images", h.CreateImage)
		v1.GET("/images", h.ListImages)
		v1.GET("/images/:name", h.GetImage)
		v1.GET("/images/:name/status", h.GetImageStatus)
		v1.DELETE("/images/:name", h.DeleteImage)

		// VM management endpoints
		v1.GET("/vms", h.ListVMs)
		v1.GET("/vms/:name", h.GetVM)
		v1.DELETE("/vms/:name", h.DeleteVM)

		// Health and metrics
		v1.GET("/health", h.Health)
		v1.GET("/metrics", h.Metrics)
	}
}

// CreateImage handles POST /api/v1/images
func (h *Handler) CreateImage(c *gin.Context) {
	var req types.CreateImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.auditLog("image_creation_failed", "validation_error", c.ClientIP(), err.Error())
		h.logger.WithError(err).Error("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Audit log the image creation request
	h.auditLog("image_creation_started", "user_request", c.ClientIP(), req.Name)

	h.logger.WithFields(logrus.Fields{
		"name":    req.Name,
		"os_type": req.OSType,
		"iso_url": req.BaseISOURL,
	}).Info("Creating golden image")

	// Validate request
	if err := req.Validate(); err != nil {
		h.auditLog("image_creation_failed", "validation_error", c.ClientIP(), err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create the image
	image, err := h.vmManager.CreateGoldenImage(c.Request.Context(), &req)
	if err != nil {
		h.auditLog("image_creation_failed", "system_error", c.ClientIP(), err.Error())
		h.logger.WithError(err).Error("Failed to create golden image")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.auditLog("image_creation_success", "user_request", c.ClientIP(), image.Name)
	c.JSON(http.StatusAccepted, gin.H{
		"message": "Golden image creation started",
		"image":   image,
	})
}

// ListImages handles GET /api/v1/images
func (h *Handler) ListImages(c *gin.Context) {
	images, err := h.storageManager.ListImages()
	if err != nil {
		h.logger.WithError(err).Error("Failed to list images")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"images": images,
		"count":  len(images),
	})
}

// GetImage handles GET /api/v1/images/:name
func (h *Handler) GetImage(c *gin.Context) {
	name := c.Param("name")

	image, err := h.storageManager.GetImage(name)
	if err != nil {
		h.logger.WithError(err).WithField("name", name).Error("Failed to get image")
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	c.JSON(http.StatusOK, image)
}

// GetImageStatus handles GET /api/v1/images/:name/status
func (h *Handler) GetImageStatus(c *gin.Context) {
	name := c.Param("name")

	status, err := h.vmManager.GetImageStatus(c.Request.Context(), name)
	if err != nil {
		h.logger.WithError(err).WithField("name", name).Error("Failed to get image status")
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	c.JSON(http.StatusOK, status)
}

// DeleteImage handles DELETE /api/v1/images/:name
func (h *Handler) DeleteImage(c *gin.Context) {
	name := c.Param("name")

	h.logger.WithField("name", name).Info("Deleting golden image")

	err := h.vmManager.DeleteImage(c.Request.Context(), name)
	if err != nil {
		h.logger.WithError(err).WithField("name", name).Error("Failed to delete image")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Image deletion started"})
}

// ListVMs handles GET /api/v1/vms
func (h *Handler) ListVMs(c *gin.Context) {
	vms, err := h.vmManager.ListVMs(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to list VMs")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"vms":   vms,
		"count": len(vms),
	})
}

// GetVM handles GET /api/v1/vms/:name
func (h *Handler) GetVM(c *gin.Context) {
	name := c.Param("name")

	vm, err := h.vmManager.GetVM(c.Request.Context(), name)
	if err != nil {
		h.logger.WithError(err).WithField("name", name).Error("Failed to get VM")
		c.JSON(http.StatusNotFound, gin.H{"error": "VM not found"})
		return
	}

	c.JSON(http.StatusOK, vm)
}

// DeleteVM handles DELETE /api/v1/vms/:name
func (h *Handler) DeleteVM(c *gin.Context) {
	name := c.Param("name")

	h.logger.WithField("name", name).Info("Deleting VM")

	err := h.vmManager.DeleteVM(c.Request.Context(), name)
	if err != nil {
		h.logger.WithError(err).WithField("name", name).Error("Failed to delete VM")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "VM deletion started"})
}

// Health handles GET /api/v1/health
func (h *Handler) Health(c *gin.Context) {
	// Check if we can connect to Kubernetes
	healthy, err := h.vmManager.HealthCheck(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Health check failed")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}

	if healthy {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy"})
	}
}

// Metrics handles GET /api/v1/metrics
func (h *Handler) Metrics(c *gin.Context) {
	metrics, err := h.vmManager.GetMetrics(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get metrics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}
