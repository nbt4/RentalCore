# JobScanner Pro - Production Deployment Guide

## Quick Start

### 1. Configure Application
The application now uses a configuration file instead of environment variables. The configuration is stored in `config.production.direct.json`.

```bash
# Configuration is already set up for your environment
# Database: tsunami-events.de / TS-Lager
# Port: 8080
# No manual environment variables needed!
```

### 2. Deploy Application
```bash
# Build and configure for production
./deploy-production.sh
```

### 3. Start Application

#### Option A: Manual Start
```bash
./start-production.sh
```

#### Option B: Systemd Service (Recommended)
```bash
# Install service (run as root)
sudo ./deploy-production.sh

# Start service
sudo systemctl start jobscanner

# Check status
sudo systemctl status jobscanner

# View logs
sudo journalctl -u jobscanner -f
```

## Production Features

- ✅ **Release Mode**: Optimized performance with GIN_MODE=release
- ✅ **File-based Config**: Configuration stored in `config.production.direct.json`
- ✅ **Logging**: Structured logging to `logs/production.log`
- ✅ **Auto-restart**: Systemd service with automatic restart on failure
- ✅ **Security**: Restricted file access and no new privileges
- ✅ **No Environment Variables**: All configuration in JSON file

## Application Access

- **URL**: `http://your-server:8080`
- **Default Port**: 8080
- **Logs**: `logs/production.log` or `journalctl -u jobscanner`

## Configuration

- **Production Config**: `config.production.direct.json` (contains all settings)
- **Template Config**: `config.production.json` (legacy template with env vars)
- **Development Config**: `config.json` (development settings)

## Security Notes

- ⚠️ **Config File Security**: `config.production.direct.json` contains sensitive database credentials
- Keep the production config file secure with appropriate file permissions (600)
- Never commit production config files to git
- Run service as non-root user (www-data)
- Enable firewall and restrict access to port 8080
- Consider using HTTPS reverse proxy (nginx/apache)

## Monitoring

```bash
# Check service status
sudo systemctl status jobscanner

# View real-time logs
sudo journalctl -u jobscanner -f

# Check application health
curl http://localhost:8080/

# Database connection test
curl http://localhost:8080/jobs
```

## Troubleshooting

### Service won't start
```bash
# Check configuration
sudo journalctl -u jobscanner --no-pager

# Verify config file exists and is readable
ls -la config.production.direct.json

# Test manually
sudo -u www-data ./start-production.sh
```

### Database connection issues
- Verify database host is reachable
- Check database credentials in config.production.direct.json
- Ensure database exists and user has permissions
- Test connection manually: `mysql -h tsunami-events.de -u root -p TS-Lager`

## Backup & Maintenance

```bash
# Backup database
mysqldump -h tsunami-events.de -u root -p TS-Lager > backup_$(date +%Y%m%d).sql

# Update application
git pull
go build -o server ./cmd/server
sudo systemctl restart jobscanner

# View application logs
tail -f logs/production.log
```