# TS Jobscanner User Management

This document explains how to manage users in the TS Jobscanner authentication system.

## Overview

The TS Jobscanner now includes a secure login system with:
- Encrypted password storage using bcrypt (cost factor 14)
- Session-based authentication with 24-hour expiration
- Secure HTTP-only cookies
- Complete route protection (all pages require login)

## User Management Tool

A command-line user management utility is provided: `user_manager.go`

### Building the User Manager

```bash
go build -o user_manager user_manager.go
```

### Usage Examples

#### Interactive User Creation
```bash
./user_manager
```
This will prompt you for all user details interactively.

#### Command-line User Creation
```bash
./user_manager -username admin -email admin@tsunami-events.de -firstname Admin -lastname User
```
You'll be prompted for the password securely.

#### List All Users
```bash
./user_manager -list
```

#### Delete a User
```bash
./user_manager -delete username
```

#### Using Different Config File
```bash
./user_manager -config config.production.json -username admin -email admin@example.com
```

## Creating Your First Admin User

1. **Build the user manager:**
   ```bash
   cd /opt/dev/go-barcode-webapp
   go build -o user_manager user_manager.go
   ```

2. **Create an admin user:**
   ```bash
   ./user_manager -username admin -email admin@tsunami-events.de -firstname Admin -lastname User
   ```
   
3. **Enter a secure password when prompted**

4. **Start the application:**
   ```bash
   ./server
   ```

5. **Access the application:**
   - Open http://localhost:8080
   - You'll be redirected to the login page
   - Enter your admin credentials

## Database Schema

The authentication system creates two new tables:

### users table
- `userID` (Primary Key)
- `username` (Unique, Required)
- `email` (Unique, Required)
- `password_hash` (bcrypt hash, not plaintext)
- `first_name`
- `last_name`
- `is_active` (Boolean, default true)
- `created_at`
- `updated_at`
- `last_login`

### sessions table
- `session_id` (Primary Key)
- `user_id` (Foreign Key to users)
- `expires_at`
- `created_at`

## Security Features

1. **Password Hashing:** All passwords are hashed using bcrypt with cost factor 14
2. **Session Management:** Secure session cookies with HTTP-only flag
3. **Route Protection:** All application routes require authentication
4. **Session Expiration:** Sessions expire after 24 hours
5. **Secure Logout:** Sessions are properly cleaned up on logout

## Manual Database User Creation (Advanced)

If you prefer to create users directly in the database:

```sql
-- First, hash your password using bcrypt cost 14
-- You can use online bcrypt generators or the user_manager tool

INSERT INTO users (username, email, password_hash, first_name, last_name, is_active, created_at, updated_at) 
VALUES ('admin', 'admin@tsunami-events.de', '$2a$14$YourBcryptHashHere', 'Admin', 'User', true, NOW(), NOW());
```

**Note:** It's strongly recommended to use the user_manager tool instead of manual database insertion.

## Troubleshooting

### "Invalid username or password" error
- Check if the user exists: `./user_manager -list`
- Verify the user is active (is_active = true)
- Ensure you're entering the correct password

### Database connection issues
- Verify your database configuration in `config.json`
- Ensure the database is running and accessible
- Check database permissions

### Session issues
- Clear browser cookies if experiencing login loops
- Check if sessions table exists and is accessible
- Verify system time is correct (affects session expiration)

## Production Deployment

For production deployment:

1. Create a production config file with secure database credentials
2. Use the production config when creating users:
   ```bash
   ./user_manager -config config.production.json -username admin -email admin@yourdomain.com
   ```
3. Ensure proper database backups include the users and sessions tables
4. Consider implementing additional security measures like:
   - HTTPS enforcement
   - Rate limiting on login attempts
   - Multi-factor authentication (future enhancement)

## Backup and Recovery

When backing up your TS Jobscanner database, ensure you include:
- All existing tables (jobs, devices, customers, etc.)
- New authentication tables: `users` and `sessions`

To restore user access after a database restore:
1. Users table will be restored with all accounts
2. Sessions table can be cleared (users will need to log in again)
3. Verify user accounts work with the user_manager tool

## API Access

The authentication system protects API endpoints as well. For programmatic access:
- All API calls require a valid session cookie
- Consider implementing API tokens for automated systems (future enhancement)