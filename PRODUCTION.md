# JobScanner Pro - Production Deployment Guide

## Quick Start

### 1. Set Environment Variables
```bash
export DB_HOST=tsunami-events.de
export DB_NAME=TS-Lager
export DB_USER=root
export DB_PASSWORD=your_secure_password
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
- ✅ **Secure Config**: Database credentials via environment variables
- ✅ **Logging**: Structured logging to `logs/production.log`
- ✅ **Auto-restart**: Systemd service with automatic restart on failure
- ✅ **Security**: Restricted file access and no new privileges

## Application Access

- **URL**: `http://your-server:8080`
- **Default Port**: 8080
- **Logs**: `logs/production.log` or `journalctl -u jobscanner`

## Configuration

- **Production Config**: `config.production.json` (template with env vars)
- **Runtime Config**: `config.runtime.json` (generated at startup)
- **Environment**: `.env.production.example` (copy and customize)

## Security Notes

- Never commit database passwords to git
- Use environment variables for all sensitive data
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

# Verify environment variables
sudo systemctl show jobscanner -p Environment

# Test manually
sudo -u www-data ./start-production.sh
```

### Database connection issues
- Verify database host is reachable
- Check database credentials
- Ensure database exists and user has permissions
- Test connection manually: `mysql -h $DB_HOST -u $DB_USER -p $DB_NAME`

## Backup & Maintenance

```bash
# Backup database
mysqldump -h $DB_HOST -u $DB_USER -p $DB_NAME > backup_$(date +%Y%m%d).sql

# Update application
git pull
go build -o server ./cmd/server
sudo systemctl restart jobscanner

# View application logs
tail -f logs/production.log
```