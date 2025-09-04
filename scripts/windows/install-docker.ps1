# PowerShell script to install Docker Desktop on Windows
# This script should be run as Administrator

Write-Host "Installing Docker Desktop on Windows..." -ForegroundColor Green

# Check if running as Administrator
if (-NOT ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
    Write-Error "This script must be run as Administrator"
    exit 1
}

# Enable Windows features required for Docker
Write-Host "Enabling Windows features for Docker..." -ForegroundColor Yellow
Enable-WindowsOptionalFeature -Online -FeatureName Microsoft-Hyper-V-All -All -NoRestart
Enable-WindowsOptionalFeature -Online -FeatureName Containers -All -NoRestart
Enable-WindowsOptionalFeature -Online -FeatureName VirtualMachinePlatform -All -NoRestart

# Download Docker Desktop installer
$dockerInstallerUrl = "https://desktop.docker.com/win/main/amd64/Docker%20Desktop%20Installer.exe"
$dockerInstallerPath = "$env:TEMP\DockerDesktopInstaller.exe"

Write-Host "Downloading Docker Desktop installer..." -ForegroundColor Yellow
try {
    Invoke-WebRequest -Uri $dockerInstallerUrl -OutFile $dockerInstallerPath -UseBasicParsing
    Write-Host "Download completed successfully" -ForegroundColor Green
} catch {
    Write-Error "Failed to download Docker Desktop installer: $_"
    exit 1
}

# Install Docker Desktop silently
Write-Host "Installing Docker Desktop..." -ForegroundColor Yellow
try {
    Start-Process -FilePath $dockerInstallerPath -ArgumentList "install", "--quiet", "--accept-license" -Wait
    Write-Host "Docker Desktop installation completed" -ForegroundColor Green
} catch {
    Write-Error "Failed to install Docker Desktop: $_"
    exit 1
}

# Clean up installer
Remove-Item $dockerInstallerPath -Force

# Add Docker to PATH
$dockerPath = "${env:ProgramFiles}\Docker\Docker\resources\bin"
if (Test-Path $dockerPath) {
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "Machine")
    if ($currentPath -notlike "*$dockerPath*") {
        [Environment]::SetEnvironmentVariable("PATH", "$currentPath;$dockerPath", "Machine")
        Write-Host "Added Docker to system PATH" -ForegroundColor Green
    }
}

# Start Docker Desktop service
Write-Host "Starting Docker Desktop service..." -ForegroundColor Yellow
try {
    Start-Service "com.docker.service" -ErrorAction SilentlyContinue
    Write-Host "Docker service started successfully" -ForegroundColor Green
} catch {
    Write-Warning "Could not start Docker service automatically. Please start Docker Desktop manually."
}

# Wait for Docker to be ready
Write-Host "Waiting for Docker to be ready..." -ForegroundColor Yellow
$timeout = 300 # 5 minutes
$elapsed = 0
do {
    Start-Sleep -Seconds 10
    $elapsed += 10
    try {
        $dockerVersion = docker --version 2>$null
        if ($dockerVersion) {
            Write-Host "Docker is ready: $dockerVersion" -ForegroundColor Green
            break
        }
    } catch {
        # Docker not ready yet
    }
} while ($elapsed -lt $timeout)

if ($elapsed -ge $timeout) {
    Write-Warning "Docker did not become ready within the timeout period. Please check Docker Desktop manually."
}

Write-Host "Docker Desktop installation completed!" -ForegroundColor Green
Write-Host "Please restart your computer to complete the installation." -ForegroundColor Yellow
