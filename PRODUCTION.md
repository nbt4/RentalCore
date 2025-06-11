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

### 3. Create Admin User
```bash
# Create the first admin user for production
./create-production-user.sh
```

### 4. Start Application

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
- **Login**: `http://your-server:8080/login`
- **User Management**: `http://your-server:8080/users`
- **Default Port**: 8080
- **Logs**: `logs/production.log` or `journalctl -u jobscanner`

## User Management

### Initial Setup
1. Create admin user: `./create-production-user.sh`
2. Start application and log in
3. Access user management at `/users`

### Managing Users
- **View Users**: `/users` - Lists all users with their status
- **Create User**: `/user-management/new` - Add new users with roles
- **Edit User**: `/user-management/:id/edit` - Modify user details  
- **User Details**: `/user-management/:id/view` - View user information

### URL Structure
The user management uses a dedicated `/user-management/` path to avoid routing conflicts:
- Main list: `http://your-server:8080/users`
- Create new: `http://your-server:8080/user-management/new`
- Edit user: `http://your-server:8080/user-management/123/edit`
- View user: `http://your-server:8080/user-management/123/view`

### User Roles & Permissions
- All authenticated users can access the application
- User management is available to logged-in users
- Consider implementing role-based permissions for production use

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