#!/bin/bash
set -e

echo "Installing FedRAMP-compliant security tools..."

# Update package lists
apt-get update

# Install security tools
apt-get install -y \
    fail2ban \
    aide \
    rkhunter \
    clamav \
    chkrootkit \
    lynis \
    auditd \
    apparmor \
    apparmor-utils \
    ufw \
    iptables-persistent \
    logwatch \
    aide-common

# Configure fail2ban
cat > /etc/fail2ban/jail.local << 'EOF'
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 3

[sshd]
enabled = true
port = ssh
logpath = /var/log/auth.log
maxretry = 3
bantime = 3600
EOF

systemctl enable fail2ban
systemctl start fail2ban

# Configure AIDE (Advanced Intrusion Detection Environment)
aideinit
mv /var/lib/aide/aide.db.new /var/lib/aide/aide.db

# Configure auditd
cat > /etc/audit/auditd.conf << 'EOF'
log_file = /var/log/audit/audit.log
log_format = RAW
flush = INCREMENTAL_ASYNC
freq = 50
num_logs = 5
priority_boost = 4
name_format = HOSTNAME
max_log_file = 6
max_log_file_action = ROTATE
space_left = 75
space_left_action = SYSLOG
action_mail_acct = root
admin_space_left = 50
admin_space_left_action = SUSPEND
disk_full_action = SUSPEND
disk_error_action = SUSPEND
EOF

systemctl enable auditd
systemctl start auditd

# Configure AppArmor
systemctl enable apparmor
systemctl start apparmor

# Configure UFW (Uncomplicated Firewall)
ufw --force enable
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw allow 80/tcp
ufw allow 443/tcp

# Configure logwatch
cat > /etc/logwatch/conf/logwatch.conf << 'EOF'
LogDir = /var/log
TmpDir = /var/cache/logwatch
MailTo = root
MailFrom = Logwatch
Print = No
Save = /var/cache/logwatch/logwatch
Range = yesterday
Detail = Med
Service = All
Format = text
Encode = none
EOF

echo "Security tools installation completed successfully!"
