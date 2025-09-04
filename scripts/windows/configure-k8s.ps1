# PowerShell script to configure Kubernetes tools on Windows
# This script should be run as Administrator

Write-Host "Configuring Kubernetes tools on Windows..." -ForegroundColor Green

# Check if running as Administrator
if (-NOT ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
    Write-Error "This script must be run as Administrator"
    exit 1
}

# Create tools directory
$toolsDir = "C:\k8s-tools"
if (-not (Test-Path $toolsDir)) {
    New-Item -ItemType Directory -Path $toolsDir -Force
    Write-Host "Created tools directory: $toolsDir" -ForegroundColor Green
}

# Install kubectl
Write-Host "Installing kubectl..." -ForegroundColor Yellow
$kubectlUrl = "https://dl.k8s.io/release/v1.28.0/bin/windows/amd64/kubectl.exe"
$kubectlPath = "$toolsDir\kubectl.exe"

try {
    Invoke-WebRequest -Uri $kubectlUrl -OutFile $kubectlPath -UseBasicParsing
    Write-Host "kubectl installed successfully" -ForegroundColor Green
} catch {
    Write-Error "Failed to install kubectl: $_"
    exit 1
}

# Install helm
Write-Host "Installing Helm..." -ForegroundColor Yellow
$helmUrl = "https://get.helm.sh/helm-v3.13.0-windows-amd64.zip"
$helmZipPath = "$env:TEMP\helm.zip"
$helmExtractPath = "$env:TEMP\helm"

try {
    Invoke-WebRequest -Uri $helmUrl -OutFile $helmZipPath -UseBasicParsing
    Expand-Archive -Path $helmZipPath -DestinationPath $helmExtractPath -Force
    Copy-Item "$helmExtractPath\windows-amd64\helm.exe" -Destination "$toolsDir\helm.exe" -Force
    Remove-Item $helmZipPath -Force
    Remove-Item $helmExtractPath -Recurse -Force
    Write-Host "Helm installed successfully" -ForegroundColor Green
} catch {
    Write-Error "Failed to install Helm: $_"
    exit 1
}

# Install k9s (optional but useful)
Write-Host "Installing k9s..." -ForegroundColor Yellow
$k9sUrl = "https://github.com/derailed/k9s/releases/download/v0.27.4/k9s_Windows_x86_64.tar.gz"
$k9sTarPath = "$env:TEMP\k9s.tar.gz"
$k9sExtractPath = "$env:TEMP\k9s"

try {
    Invoke-WebRequest -Uri $k9sUrl -OutFile $k9sTarPath -UseBasicParsing
    # Note: Windows doesn't have built-in tar.gz support, so we'll skip k9s for now
    # In a real scenario, you'd use 7-Zip or similar to extract
    Write-Host "k9s download completed (extraction requires additional tools)" -ForegroundColor Yellow
    Remove-Item $k9sTarPath -Force
} catch {
    Write-Warning "Failed to install k9s: $_"
}

# Add tools to PATH
$currentPath = [Environment]::GetEnvironmentVariable("PATH", "Machine")
if ($currentPath -notlike "*$toolsDir*") {
    [Environment]::SetEnvironmentVariable("PATH", "$currentPath;$toolsDir", "Machine")
    Write-Host "Added tools directory to system PATH" -ForegroundColor Green
}

# Create PowerShell profile with useful aliases
$profilePath = "$env:USERPROFILE\Documents\WindowsPowerShell\Microsoft.PowerShell_profile.ps1"
$profileDir = Split-Path $profilePath -Parent

if (-not (Test-Path $profileDir)) {
    New-Item -ItemType Directory -Path $profileDir -Force
}

$aliases = @"
# Kubernetes aliases
function k { kubectl `$args }
function kgp { kubectl get pods `$args }
function kgs { kubectl get services `$args }
function kgd { kubectl get deployments `$args }
function kgn { kubectl get nodes `$args }
function kdp { kubectl describe pod `$args }
function kds { kubectl describe service `$args }
function kdd { kubectl describe deployment `$args }
function kdn { kubectl describe node `$args }
function kaf { kubectl apply -f `$args }
function kdf { kubectl delete -f `$args }
function kex { kubectl exec -it `$args }
function kl { kubectl logs `$args }
function kpf { kubectl port-forward `$args }

# Helm aliases
function h { helm `$args }
function hi { helm install `$args }
function hu { helm uninstall `$args }
function hls { helm list `$args }
function hst { helm status `$args }

Write-Host "Kubernetes aliases loaded!" -ForegroundColor Green
"@

Add-Content -Path $profilePath -Value $aliases -Force
Write-Host "Created PowerShell profile with Kubernetes aliases" -ForegroundColor Green

# Create kubeconfig directory
$kubeconfigDir = "$env:USERPROFILE\.kube"
if (-not (Test-Path $kubeconfigDir)) {
    New-Item -ItemType Directory -Path $kubeconfigDir -Force
    Write-Host "Created kubeconfig directory: $kubeconfigDir" -ForegroundColor Green
}

# Create kubeconfig template
$kubeconfigTemplate = @"
apiVersion: v1
clusters:
- cluster:
    server: https://kubernetes.default.svc.cluster.local
  name: default-cluster
contexts:
- context:
    cluster: default-cluster
    user: default-user
  name: default-context
current-context: default-context
kind: Config
preferences: {}
users:
- name: default-user
  user:
    token: ""
"@

$kubeconfigPath = "$kubeconfigDir\config"
if (-not (Test-Path $kubeconfigPath)) {
    Set-Content -Path $kubeconfigPath -Value $kubeconfigTemplate -Force
    Write-Host "Created kubeconfig template" -ForegroundColor Green
}

# Verify installations
Write-Host "`nVerifying installations..." -ForegroundColor Yellow
try {
    $kubectlVersion = & "$toolsDir\kubectl.exe" version --client --short 2>$null
    Write-Host "kubectl: $kubectlVersion" -ForegroundColor Green
} catch {
    Write-Warning "kubectl verification failed"
}

try {
    $helmVersion = & "$toolsDir\helm.exe" version --short 2>$null
    Write-Host "helm: $helmVersion" -ForegroundColor Green
} catch {
    Write-Warning "Helm verification failed"
}

Write-Host "`nKubernetes tools configuration completed successfully!" -ForegroundColor Green
Write-Host "Please restart your PowerShell session to use the new aliases." -ForegroundColor Yellow
