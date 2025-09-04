#!/bin/bash
set -e

echo "Preparing system for golden image creation..."

# Update package lists
apt-get update

# Install additional tools for image preparation
apt-get install -y \
    qemu-utils \
    cloud-guest-utils \
    zerofree \
    e2fsprogs

# Clean package cache
apt-get clean
apt-get autoremove -y

# Clear logs
echo "Clearing system logs..."
find /var/log -type f -name "*.log" -exec truncate -s 0 {} \;
find /var/log -type f -name "*.1" -delete
find /var/log -type f -name "*.gz" -delete
find /var/log -type f -name "*.old" -delete

# Clear temporary files
echo "Clearing temporary files..."
rm -rf /tmp/*
rm -rf /var/tmp/*
rm -rf /var/cache/apt/archives/*

# Clear bash history
echo "Clearing bash history..."
history -c
rm -f /root/.bash_history
rm -f /home/*/.bash_history 2>/dev/null || true

# Clear cloud-init data
echo "Clearing cloud-init data..."
cloud-init clean --logs

# Clear machine ID
echo "Clearing machine ID..."
rm -f /etc/machine-id
rm -f /var/lib/dbus/machine-id

# Clear network configuration
echo "Clearing network configuration..."
rm -f /etc/netplan/*.yaml
rm -f /etc/network/interfaces.d/*

# Clear SSH host keys
echo "Clearing SSH host keys..."
rm -f /etc/ssh/ssh_host_*

# Clear user data
echo "Clearing user data..."
rm -rf /var/lib/cloud/instances/*

# Clear systemd journal
echo "Clearing systemd journal..."
journalctl --vacuum-time=1s

# Clear package manager cache
echo "Clearing package manager cache..."
apt-get clean
apt-get autoclean

# Clear pip cache (if exists)
rm -rf /root/.cache/pip 2>/dev/null || true
rm -rf /home/*/.cache/pip 2>/dev/null || true

# Clear npm cache (if exists)
rm -rf /root/.npm 2>/dev/null || true
rm -rf /home/*/.npm 2>/dev/null || true

# Clear Docker data (if Docker is installed)
if command -v docker &> /dev/null; then
    echo "Clearing Docker data..."
    docker system prune -af 2>/dev/null || true
    rm -rf /var/lib/docker/tmp/* 2>/dev/null || true
fi

# Clear any swap files
echo "Clearing swap files..."
swapoff -a 2>/dev/null || true
rm -f /swapfile 2>/dev/null || true

# Clear cron jobs
echo "Clearing cron jobs..."
rm -f /var/spool/cron/crontabs/* 2>/dev/null || true
rm -f /etc/cron.d/* 2>/dev/null || true

# Clear mail spool
echo "Clearing mail spool..."
rm -rf /var/mail/* 2>/dev/null || true
rm -rf /var/spool/mail/* 2>/dev/null || true

# Clear audit logs
echo "Clearing audit logs..."
rm -f /var/log/audit/* 2>/dev/null || true

# Clear wtmp and utmp
echo "Clearing login records..."
> /var/log/wtmp
> /var/log/utmp
> /var/log/lastlog

# Clear yum/dnf cache (for RHEL-based systems)
if command -v yum &> /dev/null; then
    yum clean all 2>/dev/null || true
fi
if command -v dnf &> /dev/null; then
    dnf clean all 2>/dev/null || true
fi

# Create new machine ID
echo "Creating new machine ID..."
systemd-machine-id-setup

# Zero out free space to reduce image size
echo "Zeroing out free space..."
if command -v zerofree &> /dev/null; then
    # Find the root filesystem
    ROOT_DEV=$(df / | tail -1 | awk '{print $1}')
    if [[ "$ROOT_DEV" =~ ^/dev/ ]]; then
        echo "Zeroing free space on $ROOT_DEV..."
        # Unmount if possible and zero free space
        umount /tmp 2>/dev/null || true
        zerofree -v "$ROOT_DEV" 2>/dev/null || echo "Could not zero free space on $ROOT_DEV"
    fi
fi

# Create completion marker
echo "Creating completion marker..."
touch /tmp/golden-image-ready

echo "Golden image preparation completed successfully!"
echo "System is ready for imaging."
echo "Timestamp: $(date)"
