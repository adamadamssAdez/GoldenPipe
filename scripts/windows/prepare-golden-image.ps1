# PowerShell script to prepare Windows system for golden image creation
# This script should be run as Administrator

Write-Host "Preparing Windows system for golden image creation..." -ForegroundColor Green

# Check if running as Administrator
if (-NOT ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
    Write-Error "This script must be run as Administrator"
    exit 1
}

# Function to clear Windows event logs
function Clear-EventLogs {
    Write-Host "Clearing Windows event logs..." -ForegroundColor Yellow
    try {
        Get-WinEvent -ListLog * | Where-Object {$_.RecordCount -gt 0} | ForEach-Object {
            try {
                [System.Diagnostics.Eventing.Reader.EventLogSession]::GlobalSession.ClearLog($_.LogName)
                Write-Host "Cleared log: $($_.LogName)" -ForegroundColor Green
            } catch {
                Write-Warning "Could not clear log: $($_.LogName)"
            }
        }
    } catch {
        Write-Warning "Error clearing event logs: $_"
    }
}

# Function to clear temporary files
function Clear-TempFiles {
    Write-Host "Clearing temporary files..." -ForegroundColor Yellow
    
    $tempPaths = @(
        "$env:TEMP\*",
        "$env:TMP\*",
        "C:\Windows\Temp\*",
        "C:\Windows\Prefetch\*",
        "C:\Windows\SoftwareDistribution\Download\*",
        "$env:LOCALAPPDATA\Temp\*"
    )
    
    foreach ($path in $tempPaths) {
        try {
            if (Test-Path $path) {
                Remove-Item $path -Recurse -Force -ErrorAction SilentlyContinue
                Write-Host "Cleared: $path" -ForegroundColor Green
            }
        } catch {
            Write-Warning "Could not clear: $path"
        }
    }
}

# Function to clear browser data
function Clear-BrowserData {
    Write-Host "Clearing browser data..." -ForegroundColor Yellow
    
    $browserPaths = @(
        "$env:LOCALAPPDATA\Google\Chrome\User Data\Default\Cache",
        "$env:LOCALAPPDATA\Google\Chrome\User Data\Default\Code Cache",
        "$env:LOCALAPPDATA\Microsoft\Edge\User Data\Default\Cache",
        "$env:LOCALAPPDATA\Microsoft\Edge\User Data\Default\Code Cache",
        "$env:APPDATA\Mozilla\Firefox\Profiles\*\cache2"
    )
    
    foreach ($path in $browserPaths) {
        try {
            if (Test-Path $path) {
                Remove-Item $path -Recurse -Force -ErrorAction SilentlyContinue
                Write-Host "Cleared browser cache: $path" -ForegroundColor Green
            }
        } catch {
            Write-Warning "Could not clear browser cache: $path"
        }
    }
}

# Function to clear Windows Update cache
function Clear-WindowsUpdateCache {
    Write-Host "Clearing Windows Update cache..." -ForegroundColor Yellow
    try {
        Stop-Service -Name "wuauserv" -Force -ErrorAction SilentlyContinue
        Remove-Item "C:\Windows\SoftwareDistribution\Download\*" -Recurse -Force -ErrorAction SilentlyContinue
        Start-Service -Name "wuauserv" -ErrorAction SilentlyContinue
        Write-Host "Windows Update cache cleared" -ForegroundColor Green
    } catch {
        Write-Warning "Could not clear Windows Update cache: $_"
    }
}

# Function to clear IIS logs (if IIS is installed)
function Clear-IISLogs {
    Write-Host "Clearing IIS logs..." -ForegroundColor Yellow
    try {
        if (Get-WindowsFeature -Name IIS-WebServer -ErrorAction SilentlyContinue | Where-Object {$_.InstallState -eq "Installed"}) {
            Remove-Item "C:\inetpub\logs\LogFiles\*" -Recurse -Force -ErrorAction SilentlyContinue
            Write-Host "IIS logs cleared" -ForegroundColor Green
        }
    } catch {
        Write-Warning "Could not clear IIS logs: $_"
    }
}

# Function to clear PowerShell history
function Clear-PowerShellHistory {
    Write-Host "Clearing PowerShell history..." -ForegroundColor Yellow
    try {
        Remove-Item (Get-PSReadlineOption).HistorySavePath -Force -ErrorAction SilentlyContinue
        Write-Host "PowerShell history cleared" -ForegroundColor Green
    } catch {
        Write-Warning "Could not clear PowerShell history: $_"
    }
}

# Function to clear Windows Defender logs
function Clear-DefenderLogs {
    Write-Host "Clearing Windows Defender logs..." -ForegroundColor Yellow
    try {
        Remove-Item "C:\ProgramData\Microsoft\Windows Defender\Support\*" -Recurse -Force -ErrorAction SilentlyContinue
        Write-Host "Windows Defender logs cleared" -ForegroundColor Green
    } catch {
        Write-Warning "Could not clear Windows Defender logs: $_"
    }
}

# Function to clear Windows Search index
function Clear-SearchIndex {
    Write-Host "Clearing Windows Search index..." -ForegroundColor Yellow
    try {
        Stop-Service -Name "WSearch" -Force -ErrorAction SilentlyContinue
        Remove-Item "C:\ProgramData\Microsoft\Search\Data\*" -Recurse -Force -ErrorAction SilentlyContinue
        Start-Service -Name "WSearch" -ErrorAction SilentlyContinue
        Write-Host "Windows Search index cleared" -ForegroundColor Green
    } catch {
        Write-Warning "Could not clear Windows Search index: $_"
    }
}

# Function to clear Windows thumbnail cache
function Clear-ThumbnailCache {
    Write-Host "Clearing thumbnail cache..." -ForegroundColor Yellow
    try {
        Remove-Item "$env:LOCALAPPDATA\Microsoft\Windows\Explorer\thumbcache_*.db" -Force -ErrorAction SilentlyContinue
        Write-Host "Thumbnail cache cleared" -ForegroundColor Green
    } catch {
        Write-Warning "Could not clear thumbnail cache: $_"
    }
}

# Function to clear Windows font cache
function Clear-FontCache {
    Write-Host "Clearing font cache..." -ForegroundColor Yellow
    try {
        Stop-Service -Name "FontCache" -Force -ErrorAction SilentlyContinue
        Remove-Item "$env:LOCALAPPDATA\Microsoft\Windows\Fonts\*" -Force -ErrorAction SilentlyContinue
        Start-Service -Name "FontCache" -ErrorAction SilentlyContinue
        Write-Host "Font cache cleared" -ForegroundColor Green
    } catch {
        Write-Warning "Could not clear font cache: $_"
    }
}

# Function to run disk cleanup
function Run-DiskCleanup {
    Write-Host "Running disk cleanup..." -ForegroundColor Yellow
    try {
        # Run cleanmgr with various cleanup options
        $cleanupScript = @"
@echo off
cleanmgr /sagerun:1
"@
        $scriptPath = "$env:TEMP\cleanup.bat"
        Set-Content -Path $scriptPath -Value $cleanupScript -Force
        Start-Process -FilePath $scriptPath -Wait -WindowStyle Hidden
        Remove-Item $scriptPath -Force
        Write-Host "Disk cleanup completed" -ForegroundColor Green
    } catch {
        Write-Warning "Could not run disk cleanup: $_"
    }
}

# Function to clear Windows Store cache
function Clear-StoreCache {
    Write-Host "Clearing Windows Store cache..." -ForegroundColor Yellow
    try {
        Get-AppxPackage -AllUsers | ForEach-Object {
            try {
                Remove-Item "$($_.InstallLocation)\*" -Recurse -Force -ErrorAction SilentlyContinue
            } catch {
                # Ignore errors for protected packages
            }
        }
        Write-Host "Windows Store cache cleared" -ForegroundColor Green
    } catch {
        Write-Warning "Could not clear Windows Store cache: $_"
    }
}

# Main execution
Write-Host "Starting Windows golden image preparation..." -ForegroundColor Green
Write-Host "Timestamp: $(Get-Date)" -ForegroundColor Cyan

# Clear various caches and logs
Clear-TempFiles
Clear-EventLogs
Clear-BrowserData
Clear-WindowsUpdateCache
Clear-IISLogs
Clear-PowerShellHistory
Clear-DefenderLogs
Clear-SearchIndex
Clear-ThumbnailCache
Clear-FontCache
Clear-StoreCache

# Run disk cleanup
Run-DiskCleanup

# Clear Windows Defender quarantine
Write-Host "Clearing Windows Defender quarantine..." -ForegroundColor Yellow
try {
    Remove-Item "C:\ProgramData\Microsoft\Windows Defender\Quarantine\*" -Recurse -Force -ErrorAction SilentlyContinue
    Write-Host "Windows Defender quarantine cleared" -ForegroundColor Green
} catch {
    Write-Warning "Could not clear Windows Defender quarantine: $_"
}

# Clear Windows Error Reporting
Write-Host "Clearing Windows Error Reporting..." -ForegroundColor Yellow
try {
    Remove-Item "C:\ProgramData\Microsoft\Windows\WER\*" -Recurse -Force -ErrorAction SilentlyContinue
    Write-Host "Windows Error Reporting cleared" -ForegroundColor Green
} catch {
    Write-Warning "Could not clear Windows Error Reporting: $_"
}

# Clear Windows Performance Logs
Write-Host "Clearing Windows Performance Logs..." -ForegroundColor Yellow
try {
    Remove-Item "C:\PerfLogs\*" -Recurse -Force -ErrorAction SilentlyContinue
    Write-Host "Windows Performance Logs cleared" -ForegroundColor Green
} catch {
    Write-Warning "Could not clear Windows Performance Logs: $_"
}

# Clear Windows Installer cache
Write-Host "Clearing Windows Installer cache..." -ForegroundColor Yellow
try {
    Remove-Item "C:\Windows\Installer\$PatchCache$\*" -Recurse -Force -ErrorAction SilentlyContinue
    Write-Host "Windows Installer cache cleared" -ForegroundColor Green
} catch {
    Write-Warning "Could not clear Windows Installer cache: $_"
}

# Clear Windows Driver Store
Write-Host "Clearing Windows Driver Store..." -ForegroundColor Yellow
try {
    # This requires elevated privileges and may take time
    Start-Process -FilePath "pnputil.exe" -ArgumentList "/delete-driver", "oem*.inf", "/uninstall", "/force" -Wait -WindowStyle Hidden -ErrorAction SilentlyContinue
    Write-Host "Windows Driver Store cleared" -ForegroundColor Green
} catch {
    Write-Warning "Could not clear Windows Driver Store: $_"
}

# Create completion marker
Write-Host "Creating completion marker..." -ForegroundColor Yellow
$completionMarker = "C:\golden-image-ready.txt"
Set-Content -Path $completionMarker -Value "Golden image preparation completed at $(Get-Date)" -Force

Write-Host "`nWindows golden image preparation completed successfully!" -ForegroundColor Green
Write-Host "System is ready for imaging." -ForegroundColor Green
Write-Host "Completion timestamp: $(Get-Date)" -ForegroundColor Cyan
Write-Host "`nNote: You may want to run sysprep before creating the final image." -ForegroundColor Yellow
