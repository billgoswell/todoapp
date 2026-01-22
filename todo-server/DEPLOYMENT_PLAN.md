# CommandLineTodo REST API Server - Proxmox Local Network Deployment Plan

## Overview

Deploy the CommandLineTodo REST API server (`/home/bill/personal/workspace/commandlinetodo-server`) on Proxmox VE for local network use only with automated backups and monitoring.

**Technology Stack:**
- Server: Go 1.21 + Gin framework
- Database: PostgreSQL 18
- Platform: 2 VMs on Proxmox VE
- OS: Ubuntu 24.04 LTS
- Network: Local network only (no internet exposure)

## Architecture: Simple 2-VM Setup

```
                    [Local Network Clients]
                    (CLI app, Mobile apps)
                           |
                    [App Server VM]
                    192.168.100.20
                    ┌─────────────────┐
                    │ Go API Server   │
                    │ PostgreSQL 15   │
                    │ Port 8080       │
                    └─────────────────┘
                           |
                    [Monitoring VM]
                    192.168.100.30
                    ┌─────────────────┐
                    │ Prometheus      │
                    │ Grafana         │
                    │ Backups         │
                    └─────────────────┘
```

### VM Specifications

| VM Name | Role | vCPU | RAM | Storage | IP Address |
|---------|------|------|-----|---------|------------|
| app-server | Application + Database | 4 | 8GB | 80GB | 192.168.100.20 |
| monitoring | Monitoring + Backups | 2 | 4GB | 100GB | 192.168.100.30 |

**Total Resources:** 6 vCPU, 12GB RAM, 180GB Storage

## Design Decisions

### 1. Single Server Architecture
- **All-in-one**: Go API server and PostgreSQL on same VM
- **No redundancy**: Simple deployment, manual recovery if needed
- **Local network only**: No SSL/TLS required, firewall protects from external access
- **Auto-restart**: SystemD handles service crashes

### 2. Backup Strategy
1. **Daily PostgreSQL Backups** - pg_dump at 02:00 AM, stored on monitoring VM
2. **Weekly VM Snapshots** - Proxmox snapshots for full system recovery
3. **30-day retention** - Automated cleanup of old backups

### 3. Monitoring
- **Prometheus**: Collects metrics from app server
- **Grafana**: Dashboards for PostgreSQL, system resources, API performance
- **Optional**: Simple alerting for disk space and service health

### 4. Network Security
- **Firewall**: iptables on app server allows only local network
- **No internet exposure**: Server not accessible from outside
- **No SSL**: Plain HTTP on port 8080 (internal network only)

## Implementation Steps

### Step 1: Create VMs in Proxmox (30 minutes)

**Create app-server VM:**
```bash
# In Proxmox UI or CLI
# VM ID: 201
# Name: app-server
# Memory: 8192 MB
# Cores: 4
# Disk: 80GB
# Network: Bridge to local network
# Static IP: 192.168.100.20/24
```

**Create monitoring VM:**
```bash
# VM ID: 202
# Name: monitoring
# Memory: 4096 MB
# Cores: 2
# Disk: 100GB
# Network: Bridge to local network
# Static IP: 192.168.100.30/24
```

**Install Ubuntu 24.04 LTS on both VMs**

### Step 2: Base OS Configuration (30 minutes)

**On both VMs:**

```bash
# Update system
apt update && apt upgrade -y

# Install essential tools
apt install -y vim curl wget git htop

# Configure timezone
timedatectl set-timezone America/New_York  # Adjust to your timezone

# Disable password authentication (use SSH keys)
sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
systemctl restart sshd

# Install node exporter for monitoring
apt install -y prometheus-node-exporter
```

**Configure Static IP with Netplan:**

Edit the Netplan configuration:
```bash
sudo nano /etc/netplan/00-installer-config.yaml
```

Replace contents with (app-server example):
```yaml
network:
  version: 2
  ethernets:
    ens18:
      dhcp4: false
      addresses:
        - 192.168.100.20/24
      routes:
        - to: 0.0.0.0/0
          via: 192.168.100.1
      nameservers:
        addresses: [8.8.8.8, 8.8.4.4]
```

For monitoring VM, change address to `192.168.100.30/24`

Apply and verify:
```bash
sudo netplan apply
ip addr show ens18
```

### Step 3: PostgreSQL Setup (1 hour)

**On app-server:**

1. Add PostgreSQL repository and install PostgreSQL 18:
```bash
# Add PostgreSQL repository
sudo sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list'

# Add the repository key
wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | sudo apt-key add -

# Update package list
apt update

# Install PostgreSQL 18
apt install -y postgresql-18 postgresql-contrib-18
```

2. Configure PostgreSQL - `/etc/postgresql/18/main/postgresql.conf`:
```conf
# Networking
listen_addresses = 'localhost'  # Only local connections

# Performance (tuned for 8GB RAM)
shared_buffers = 2GB
effective_cache_size = 6GB
maintenance_work_mem = 512MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
work_mem = 20MB

# Logging
logging_collector = on
log_directory = '/var/log/postgresql'
log_filename = 'postgresql-%Y-%m-%d.log'
log_rotation_age = 1d
log_connections = on
log_disconnections = on
log_duration = off
log_line_prefix = '%m [%p] %u@%d '
```

3. Create application database and user:
```bash
sudo -u postgres psql << EOF
CREATE USER commandlinetodo WITH PASSWORD 'your-secure-password-here';
CREATE DATABASE commandlinetodo OWNER commandlinetodo;
GRANT ALL PRIVILEGES ON DATABASE commandlinetodo TO commandlinetodo;
EOF
```

4. Verify connection:
```bash
sudo -u postgres psql -d commandlinetodo -c "\dt"
```

### Step 4: Deploy Application (1 hour)

**On app-server:**

1. Create application user and directories:
```bash
useradd --system --no-create-home --shell /bin/false commandlinetodo
mkdir -p /opt/commandlinetodo /var/log/commandlinetodo /etc/commandlinetodo
chown commandlinetodo:commandlinetodo /var/log/commandlinetodo
```

2. Build application (on your dev machine):
```bash
cd /home/bill/personal/workspace/commandlinetodo-server
GOOS=linux GOARCH=amd64 go build -o server -ldflags="-s -w" ./cmd/server/main.go
```

3. Copy binary to server:
```bash
scp server app-server:/tmp/
```

4. Install binary on app-server:
```bash
sudo cp /tmp/server /opt/commandlinetodo/
sudo chown commandlinetodo:commandlinetodo /opt/commandlinetodo/server
sudo chmod 755 /opt/commandlinetodo/server
```

5. Create environment file - `/etc/commandlinetodo/server.env`:
```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=commandlinetodo
DB_PASSWORD=your-secure-password-here
DB_NAME=commandlinetodo
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
GIN_MODE=release
```

Secure the file:
```bash
chmod 600 /etc/commandlinetodo/server.env
chown commandlinetodo:commandlinetodo /etc/commandlinetodo/server.env
```

6. Create systemd service - `/etc/systemd/system/commandlinetodo-api.service`:
```ini
[Unit]
Description=CommandLineTodo REST API Server
After=network.target postgresql.service
Requires=postgresql.service

[Service]
Type=simple
User=commandlinetodo
Group=commandlinetodo
WorkingDirectory=/opt/commandlinetodo
ExecStart=/opt/commandlinetodo/server
EnvironmentFile=/etc/commandlinetodo/server.env

# Auto-restart on failure
Restart=always
RestartSec=10
StartLimitInterval=300
StartLimitBurst=5

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log/commandlinetodo

# Resource limits
LimitNOFILE=65536

# Logging
StandardOutput=append:/var/log/commandlinetodo/server.log
StandardError=append:/var/log/commandlinetodo/error.log

[Install]
WantedBy=multi-user.target
```

7. Enable and start service:
```bash
systemctl daemon-reload
systemctl enable commandlinetodo-api
systemctl start commandlinetodo-api

# Check status
systemctl status commandlinetodo-api

# Check logs
journalctl -u commandlinetodo-api -f
```

8. Test the API:
```bash
# Health check (should return {"status":"ok"})
curl http://localhost:8080/api/v1/health

# Test from another machine on local network
curl http://192.168.100.20:8080/api/v1/health
```

### Step 5: Firewall Configuration (15 minutes)

**On app-server:**

Configure iptables to allow only local network access:

```bash
# Flush existing rules
iptables -F
iptables -X

# Default policies
iptables -P INPUT DROP
iptables -P FORWARD DROP
iptables -P OUTPUT ACCEPT

# Allow loopback
iptables -A INPUT -i lo -j ACCEPT

# Allow established connections
iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT

# SSH from local network only
iptables -A INPUT -p tcp --dport 22 -s 192.168.100.0/24 -j ACCEPT

# API from local network only
iptables -A INPUT -p tcp --dport 8080 -s 192.168.100.0/24 -j ACCEPT

# Prometheus metrics from monitoring VM
iptables -A INPUT -p tcp --dport 9100 -s 192.168.100.30 -j ACCEPT

# PostgreSQL exporter from monitoring VM (if installed)
iptables -A INPUT -p tcp --dport 9187 -s 192.168.100.30 -j ACCEPT

# Drop everything else
iptables -A INPUT -j DROP

# Save rules
apt install -y iptables-persistent
iptables-save > /etc/iptables/rules.v4
```

**On monitoring VM:**

```bash
# Similar setup but only allow SSH from local network
iptables -A INPUT -p tcp --dport 22 -s 192.168.100.0/24 -j ACCEPT
iptables -A INPUT -p tcp --dport 3000 -s 192.168.100.0/24 -j ACCEPT  # Grafana
iptables-save > /etc/iptables/rules.v4
```

### Step 6: Backup Configuration (30 minutes)

**On monitoring VM:**

1. Create backup directory:
```bash
mkdir -p /backup/postgresql
chmod 755 /backup
```

2. Set up SSH key for app-server to copy backups:
```bash
# On monitoring VM
ssh-keygen -t ed25519 -f /root/.ssh/backup_key -N ""

# Copy public key to app-server
ssh-copy-id -i /root/.ssh/backup_key.pub root@192.168.100.20
```

**On app-server:**

1. Create backup script - `/usr/local/bin/backup-postgres.sh`:
```bash
#!/bin/bash
set -e

BACKUP_DIR="/var/backups/postgresql"
REMOTE_HOST="192.168.100.30"
REMOTE_DIR="/backup/postgresql"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS=30

# Create local backup directory
mkdir -p $BACKUP_DIR

# Dump database
pg_dump -h localhost -U commandlinetodo -d commandlinetodo -Fc > \
  $BACKUP_DIR/commandlinetodo_$TIMESTAMP.dump

# Compress
gzip $BACKUP_DIR/commandlinetodo_$TIMESTAMP.dump

# Copy to monitoring server
rsync -az $BACKUP_DIR/commandlinetodo_$TIMESTAMP.dump.gz \
  $REMOTE_HOST:$REMOTE_DIR/

# Clean old local backups (keep last 7 days locally)
find $BACKUP_DIR -name "*.dump.gz" -mtime +7 -delete

# Clean old remote backups (monitoring server keeps 30 days)
ssh $REMOTE_HOST "find $REMOTE_DIR -name '*.dump.gz' -mtime +$RETENTION_DAYS -delete"

# Log completion
logger "PostgreSQL backup completed: commandlinetodo_$TIMESTAMP"
echo "Backup completed: commandlinetodo_$TIMESTAMP.dump.gz"
```

2. Make executable:
```bash
chmod +x /usr/local/bin/backup-postgres.sh
```

3. Test backup:
```bash
/usr/local/bin/backup-postgres.sh
```

4. Schedule daily backups:
```bash
# Add to crontab (runs at 2 AM daily)
echo "0 2 * * * /usr/local/bin/backup-postgres.sh >> /var/log/commandlinetodo/backup.log 2>&1" | crontab -

# View crontab
crontab -l
```

5. Configure log rotation - `/etc/logrotate.d/commandlinetodo`:
```
/var/log/commandlinetodo/*.log {
    daily
    missingok
    rotate 14
    compress
    delaycompress
    notifempty
    create 0640 commandlinetodo commandlinetodo
    sharedscripts
    postrotate
        systemctl reload commandlinetodo-api > /dev/null 2>&1 || true
    endscript
}
```

### Step 7: Monitoring Setup (1 hour)

**On monitoring VM:**

1. Install Prometheus:
```bash
apt install -y prometheus
```

2. Configure Prometheus - `/etc/prometheus/prometheus.yml`:
```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'app-server'
    static_configs:
      - targets: ['192.168.100.20:9100']
        labels:
          instance: 'app-server'

  - job_name: 'monitoring'
    static_configs:
      - targets: ['localhost:9100']
        labels:
          instance: 'monitoring'
```

3. Start Prometheus:
```bash
systemctl enable prometheus
systemctl start prometheus
```

4. Install Grafana:
```bash
wget -q -O - https://packages.grafana.com/gpg.key | apt-key add -
echo "deb https://packages.grafana.com/oss/deb stable main" | \
  tee /etc/apt/sources.list.d/grafana.list
apt update && apt install -y grafana

systemctl enable grafana-server
systemctl start grafana-server
```

5. Access Grafana:
   - URL: http://192.168.100.30:3000
   - Default credentials: admin / admin (change on first login)

6. Configure Grafana:
   - Add Prometheus data source (URL: http://localhost:9090)
   - Import dashboard for Node Exporter (Dashboard ID: 1860)

**Optional: Install PostgreSQL Exporter on app-server:**

```bash
# On app-server
wget https://github.com/prometheus-contrib/postgres_exporter/releases/download/v0.15.0/postgres_exporter-0.15.0.linux-amd64.tar.gz
tar xvf postgres_exporter-0.15.0.linux-amd64.tar.gz
sudo cp postgres_exporter-0.15.0.linux-amd64/postgres_exporter /usr/local/bin/

# Create systemd service
sudo cat > /etc/systemd/system/postgres_exporter.service << EOF
[Unit]
Description=PostgreSQL Exporter
After=network.target

[Service]
Type=simple
User=postgres
Environment="DATA_SOURCE_NAME=postgresql://commandlinetodo:your-secure-password-here@localhost:5432/commandlinetodo?sslmode=disable"
ExecStart=/usr/local/bin/postgres_exporter

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable postgres_exporter
sudo systemctl start postgres_exporter
```

**Note:** Replace `your-secure-password-here` with the actual password you set in Step 3.

### Step 8: Testing & Validation (30 minutes)

**Functional Tests:**

1. Test health endpoint from local network:
```bash
curl http://192.168.100.20:8080/api/v1/health
# Should return: {"status":"ok"}
```

2. Test with API key:
```bash
# First, create a test user and API key in the database
sudo -u postgres psql -d commandlinetodo << EOF
INSERT INTO users (api_key, created_at)
VALUES ('test-api-key-12345', EXTRACT(EPOCH FROM NOW())::BIGINT);
EOF

# Test authenticated endpoint
curl -H "Authorization: Bearer test-api-key-12345" \
  http://192.168.100.20:8080/api/v1/lists
```

3. Test sync endpoints:
```bash
# Pull sync
curl -X POST \
  -H "Authorization: Bearer test-api-key-12345" \
  -H "Content-Type: application/json" \
  -d '{"since": 0}' \
  http://192.168.100.20:8080/api/v1/sync/pull

# Should return: {"tasks":null,"lists":null}
```

4. Test with CLI app:
```bash
# On your development machine or any local network machine
export TODO_SYNC_ENABLED=true
export TODO_SYNC_SERVER_URL=http://192.168.100.20:8080
export TODO_SYNC_API_KEY=test-api-key-12345

./commandlinetodo
# Should sync successfully
```

5. Test mobile app connectivity:
   - Configure server URL: http://192.168.100.20:8080
   - Add API key: test-api-key-12345
   - Test connection (should see green status)

**Service Reliability Tests:**

1. Test auto-restart:
```bash
# Kill the process
sudo pkill -9 server

# Wait 10 seconds
sleep 10

# Check status (should be running)
systemctl status commandlinetodo-api
```

2. Test after reboot:
```bash
sudo reboot

# After reboot, check service started automatically
systemctl status commandlinetodo-api postgresql
```

**Backup Tests:**

1. Test manual backup:
```bash
/usr/local/bin/backup-postgres.sh

# Verify backup exists
ls -lh /var/backups/postgresql/
ssh 192.168.100.30 "ls -lh /backup/postgresql/"
```

2. Test backup restoration:
```bash
# On app-server
# Stop application
systemctl stop commandlinetodo-api

# Restore from backup (replace TIMESTAMP with actual backup timestamp)
gunzip -c /var/backups/postgresql/commandlinetodo_TIMESTAMP.dump.gz | \
  sudo -u postgres pg_restore -d commandlinetodo -c

# Start application
systemctl start commandlinetodo-api
```

**Monitoring Tests:**

1. Check Prometheus targets:
   - URL: http://192.168.100.30:9090/targets
   - All targets should be "UP"

2. Check Grafana dashboard:
   - URL: http://192.168.100.30:3000
   - View Node Exporter dashboard
   - Should see metrics from app-server

## Disaster Recovery Procedures

### Application Failure

**If API service crashes:**
```bash
# SystemD will automatically restart within 10 seconds
# Check logs to identify issue
journalctl -u commandlinetodo-api -n 100

# Manual restart if needed
systemctl restart commandlinetodo-api
```

### Database Corruption

**Restore from backup:**
```bash
# Stop application
systemctl stop commandlinetodo-api

# Drop and recreate database
sudo -u postgres psql << EOF
DROP DATABASE commandlinetodo;
CREATE DATABASE commandlinetodo OWNER commandlinetodo;
EOF

# Restore latest backup
LATEST_BACKUP=$(ls -t /var/backups/postgresql/*.dump.gz | head -1)
gunzip -c $LATEST_BACKUP | sudo -u postgres pg_restore -d commandlinetodo

# Start application
systemctl start commandlinetodo-api
```

### Complete VM Failure

**Restore from Proxmox snapshot:**
1. In Proxmox UI, select the VM
2. Go to Snapshots tab
3. Select most recent snapshot
4. Click "Rollback"
5. Start VM
6. Verify services started

**Or rebuild from backup:**
1. Create new VM with same specifications
2. Install Ubuntu 24.04 LTS
3. Follow Steps 2-4 to reinstall PostgreSQL and application
4. Restore database from backup on monitoring VM
5. Update IP address if changed

## Verification Checklist

Before production use:

- [ ] Both VMs created and accessible
- [ ] PostgreSQL installed and database created
- [ ] Application binary deployed and running
- [ ] Health endpoint responding
- [ ] Firewall rules in place (local network only)
- [ ] Backups running and stored on monitoring VM
- [ ] Monitoring setup (Prometheus + Grafana)
- [ ] Services auto-start after reboot
- [ ] CLI app can connect and sync
- [ ] Mobile app can connect and sync
- [ ] Backup restoration tested

## Critical Files

**On app-server (192.168.100.20):**

1. **`/etc/systemd/system/commandlinetodo-api.service`** - Service definition with auto-restart
2. **`/etc/commandlinetodo/server.env`** - Environment variables (database credentials)
3. **`/opt/commandlinetodo/server`** - Go application binary
4. **`/usr/local/bin/backup-postgres.sh`** - Automated backup script
5. **`/etc/iptables/rules.v4`** - Firewall rules (local network only)

**On monitoring VM (192.168.100.30):**

1. **`/etc/prometheus/prometheus.yml`** - Prometheus configuration
2. **`/backup/postgresql/`** - Backup storage directory

## Client Configuration

### CLI App

Set environment variables:
```bash
export TODO_SYNC_ENABLED=true
export TODO_SYNC_SERVER_URL=http://192.168.100.20:8080
export TODO_SYNC_API_KEY=your-api-key-here
```

Or add to `~/.bashrc` or `~/.zshrc` for persistence.

### Mobile App

In Settings screen:
- **Server URL**: http://192.168.100.20:8080
- **API Key**: (paste your API key)
- **Test Connection**: Should show "Connected" status

## Maintenance Tasks

### Daily
```bash
# Check service status
systemctl status commandlinetodo-api postgresql

# Check recent errors
journalctl -u commandlinetodo-api --since "24 hours ago" | grep -i error

# Verify backup completed
ls -lh /backup/postgresql/ | tail -5
```

### Weekly
- Review Grafana dashboards for performance trends
- Check disk space on both VMs
- Review application logs for any warnings

### Monthly
- Test backup restoration
- Update system packages:
```bash
apt update && apt upgrade -y
systemctl restart commandlinetodo-api
```
- Review and clean old logs

## Upgrading Application

**To deploy new version:**

1. Build new binary:
```bash
cd /home/bill/personal/workspace/commandlinetodo-server
GOOS=linux GOARCH=amd64 go build -o server -ldflags="-s -w" ./cmd/server/main.go
```

2. Create backup of current binary:
```bash
# On app-server
cp /opt/commandlinetodo/server /opt/commandlinetodo/server.backup.$(date +%Y%m%d)
```

3. Copy and install new binary:
```bash
scp server app-server:/tmp/
ssh app-server "sudo cp /tmp/server /opt/commandlinetodo/ && sudo chown commandlinetodo:commandlinetodo /opt/commandlinetodo/server && sudo systemctl restart commandlinetodo-api"
```

4. Verify:
```bash
curl http://192.168.100.20:8080/api/v1/health
```

5. If issues, rollback:
```bash
ssh app-server "sudo cp /opt/commandlinetodo/server.backup.YYYYMMDD /opt/commandlinetodo/server && sudo systemctl restart commandlinetodo-api"
```

## Summary

This deployment provides:

✅ **Simple Setup**: 2 VMs, easy to understand and maintain
✅ **Local Network Only**: Firewall-protected, no internet exposure
✅ **Automated Backups**: Daily PostgreSQL dumps + weekly VM snapshots
✅ **Auto-Recovery**: SystemD restarts service on crashes
✅ **Monitoring**: Prometheus + Grafana for visibility
✅ **Easy Maintenance**: Straightforward upgrade and recovery procedures

**Total Setup Time:** 4-5 hours
**Resources:** 6 vCPU, 12GB RAM, 180GB storage
**Network:** Local network only (192.168.100.0/24)
