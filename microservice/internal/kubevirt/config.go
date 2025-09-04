package kubevirt

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/goldenpipe/microservice/pkg/types"
)

// createCloudInitConfig creates cloud-init configuration for Linux VMs
func (m *Manager) createCloudInitConfig(req *types.CreateImageRequest) (userData, networkData string, err error) {
	// Base cloud-init configuration
	userDataContent := `#cloud-config
package_update: true
package_upgrade: true

# Install packages
packages:
  - curl
  - wget
  - git
  - vim
  - htop
  - cloud-utils
`

	// Add custom packages if specified
	if req.Customizations != nil && len(req.Customizations.Packages) > 0 {
		userDataContent += "\n# Custom packages\npackages:\n"
		for _, pkg := range req.Customizations.Packages {
			userDataContent += fmt.Sprintf("  - %s\n", pkg)
		}
	}

	// Add users if specified
	if req.Customizations != nil && len(req.Customizations.Users) > 0 {
		userDataContent += "\n# Users\nusers:\n"
		for _, user := range req.Customizations.Users {
			userDataContent += fmt.Sprintf("  - name: %s\n", user.Name)
			if user.Password != "" {
				userDataContent += fmt.Sprintf("    passwd: %s\n", user.Password)
			}
			if len(user.Groups) > 0 {
				userDataContent += fmt.Sprintf("    groups: %s\n", strings.Join(user.Groups, ","))
			}
			if user.Sudo {
				userDataContent += "    sudo: ALL=(ALL) NOPASSWD:ALL\n"
			}
		}
	}

	// Add SSH keys if specified
	if req.Customizations != nil && len(req.Customizations.SSHKeys) > 0 {
		userDataContent += "\n# SSH keys\nssh_authorized_keys:\n"
		for _, key := range req.Customizations.SSHKeys {
			userDataContent += fmt.Sprintf("  - %s\n", key)
		}
	}

	// Add custom files if specified
	if req.Customizations != nil && len(req.Customizations.Files) > 0 {
		userDataContent += "\n# Custom files\nwrite_files:\n"
		for path, content := range req.Customizations.Files {
			userDataContent += fmt.Sprintf("  - path: %s\n", path)
			userDataContent += fmt.Sprintf("    content: |\n")
			for _, line := range strings.Split(content, "\n") {
				userDataContent += fmt.Sprintf("      %s\n", line)
			}
		}
	}

	// Add custom scripts if specified
	if req.Customizations != nil && len(req.Customizations.Scripts) > 0 {
		userDataContent += "\n# Custom scripts\nruncmd:\n"
		for _, script := range req.Customizations.Scripts {
			userDataContent += fmt.Sprintf("  - %s\n", script)
		}
	}

	// Add golden image creation script
	userDataContent += `
# Golden image creation script
runcmd:
  - |
    # Wait for system to be ready
    sleep 30
    
    # Create golden image preparation script
    cat > /tmp/prepare-golden-image.sh << 'EOF'
#!/bin/bash
set -e

echo "Starting golden image preparation..."

# Update package lists
apt-get update

# Install additional tools for image preparation
apt-get install -y qemu-utils cloud-guest-utils

# Clean package cache
apt-get clean
apt-get autoremove -y

# Clear logs
find /var/log -type f -name "*.log" -exec truncate -s 0 {} \;
find /var/log -type f -name "*.1" -delete
find /var/log -type f -name "*.gz" -delete

# Clear temporary files
rm -rf /tmp/*
rm -rf /var/tmp/*

# Clear bash history
history -c
rm -f /root/.bash_history
rm -f /home/*/.bash_history

# Clear cloud-init data
cloud-init clean --logs

# Clear machine ID
rm -f /etc/machine-id
rm -f /var/lib/dbus/machine-id

# Clear network configuration
rm -f /etc/netplan/*.yaml
rm -f /etc/network/interfaces.d/*

# Clear SSH host keys
rm -f /etc/ssh/ssh_host_*

# Clear user data
rm -rf /var/lib/cloud/instances/*

# Clear systemd journal
journalctl --vacuum-time=1s

# Create new machine ID
systemd-machine-id-setup

echo "Golden image preparation completed successfully!"
echo "System is ready for imaging."

# Signal completion
touch /tmp/golden-image-ready
EOF

    chmod +x /tmp/prepare-golden-image.sh
    
    # Run the preparation script
    /tmp/prepare-golden-image.sh

# Power off after completion
poweroff
`

	// Encode user data
	userData = base64.StdEncoding.EncodeToString([]byte(userDataContent))

	// Network configuration
	networkDataContent := `version: 2
ethernets:
  eth0:
    dhcp4: true
    dhcp6: false
`

	networkData = base64.StdEncoding.EncodeToString([]byte(networkDataContent))

	return userData, networkData, nil
}

// createAutounattendConfig creates autounattend.xml configuration for Windows VMs
func (m *Manager) createAutounattendConfig(req *types.CreateImageRequest) (userData string, err error) {
	// Base autounattend.xml configuration
	autounattendContent := `<?xml version="1.0" encoding="utf-8"?>
<unattend xmlns="urn:schemas-microsoft-com:unattend">
    <settings pass="windowsPE">
        <component name="Microsoft-Windows-International-Core-WinPE" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
            <SetupUILanguage>
                <UILanguage>en-US</UILanguage>
            </SetupUILanguage>
            <InputLocale>en-US</InputLocale>
            <UserLocale>en-US</UserLocale>
            <UILanguage>en-US</UILanguage>
            <SystemLocale>en-US</SystemLocale>
        </component>
        <component name="Microsoft-Windows-Setup" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
            <DiskConfiguration>
                <Disk wcm:action="add">
                    <DiskID>0</DiskID>
                    <WillWipeDisk>true</WillWipeDisk>
                    <CreatePartitions>
                        <CreatePartition wcm:action="add">
                            <Order>1</Order>
                            <Type>Primary</Type>
                            <Size>500</Size>
                        </CreatePartition>
                        <CreatePartition wcm:action="add">
                            <Order>2</Order>
                            <Type>Primary</Type>
                            <Extend>true</Extend>
                        </CreatePartition>
                    </CreatePartitions>
                    <ModifyPartitions>
                        <ModifyPartition wcm:action="add">
                            <Order>1</Order>
                            <PartitionID>1</PartitionID>
                            <Label>System Reserved</Label>
                            <Format>NTFS</Format>
                        </ModifyPartition>
                        <ModifyPartition wcm:action="add">
                            <Order>2</Order>
                            <PartitionID>2</PartitionID>
                            <Label>Windows</Label>
                            <Letter>C</Letter>
                            <Format>NTFS</Format>
                        </ModifyPartition>
                    </ModifyPartitions>
                </Disk>
            </DiskConfiguration>
            <ImageInstall>
                <OSImage>
                    <InstallFrom>
                        <MetaData wcm:action="add">
                            <Key>/IMAGE/NAME</Key>
                            <Value>Windows Server 2022 SERVERSTANDARD</Value>
                        </MetaData>
                    </InstallFrom>
                    <InstallTo>
                        <DiskID>0</DiskID>
                        <PartitionID>2</PartitionID>
                    </InstallTo>
                </OSImage>
            </ImageInstall>
            <UserData>
                <AcceptEula>true</AcceptEula>
                <FullName>Administrator</FullName>
                <Organization>GoldenPipe</Organization>
            </UserData>
        </component>
    </settings>
    <settings pass="specialize">
        <component name="Microsoft-Windows-Shell-Setup" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
            <ComputerName>GOLDEN-IMAGE</ComputerName>
            <RegisteredOwner>GoldenPipe</RegisteredOwner>
            <RegisteredOrganization>GoldenPipe</RegisteredOrganization>
        </component>
        <component name="Microsoft-Windows-Deployment" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
            <RunSynchronous>
                <RunSynchronousCommand wcm:action="add">
                    <Order>1</Order>
                    <Path>powershell -Command "Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Force"</Path>
                </RunSynchronousCommand>
                <RunSynchronousCommand wcm:action="add">
                    <Order>2</Order>
                    <Path>powershell -Command "Enable-WindowsOptionalFeature -Online -FeatureName Microsoft-Hyper-V-All -All"</Path>
                </RunSynchronousCommand>
            </RunSynchronous>
        </component>
    </settings>
    <settings pass="oobeSystem">
        <component name="Microsoft-Windows-Shell-Setup" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
            <OOBE>
                <HideEULAPage>true</HideEULAPage>
                <HideOEMRegistrationScreen>true</HideOEMRegistrationScreen>
                <HideOnlineAccountScreens>true</HideOnlineAccountScreens>
                <HideWirelessSetupInOOBE>true</HideWirelessSetupInOOBE>
                <NetworkLocation>Work</NetworkLocation>
                <SkipUserOOBE>true</SkipUserOOBE>
                <SkipMachineOOBE>true</SkipMachineOOBE>
            </OOBE>
            <UserAccounts>
                <AdministratorPassword>
                    <Value>GoldenPipe123!</Value>
                    <PlainText>true</PlainText>
                </AdministratorPassword>
            </UserAccounts>
            <AutoLogon>
                <Enabled>true</Enabled>
                <Username>Administrator</Username>
                <Password>
                    <Value>GoldenPipe123!</Value>
                    <PlainText>true</PlainText>
                </Password>
                <LogonCount>1</LogonCount>
            </AutoLogon>
        </component>
    </settings>
    <settings pass="offlineServicing">
        <component name="Microsoft-Windows-LUA-Settings" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
            <EnableLUA>false</EnableLUA>
        </component>
    </settings>
</unattend>`

	// Add custom PowerShell scripts if specified
	if req.Customizations != nil && len(req.Customizations.Scripts) > 0 {
		// Insert custom scripts into the RunSynchronous section
		scriptSection := ""
		order := 3
		for _, script := range req.Customizations.Scripts {
			scriptSection += fmt.Sprintf(`
                <RunSynchronousCommand wcm:action="add">
                    <Order>%d</Order>
                    <Path>powershell -Command "%s"</Path>
                </RunSynchronousCommand>`, order, script)
			order++
		}

		// Insert scripts before the closing RunSynchronous tag
		autounattendContent = strings.Replace(autounattendContent,
			"</RunSynchronous>",
			scriptSection+"\n            </RunSynchronous>", 1)
	}

	// Encode autounattend.xml
	userData = base64.StdEncoding.EncodeToString([]byte(autounattendContent))

	return userData, nil
}
