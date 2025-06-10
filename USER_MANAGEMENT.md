# TS Jobscanner User Management

This document explains how to manage users in the TS Jobscanner authentication system.

## Overview

The TS Jobscanner includes a comprehensive user management system with:
- Encrypted password storage using bcrypt (cost factor 14)
- Session-based authentication with 24-hour expiration
- Secure HTTP-only cookies
- Complete route protection (all pages require login)
- Web-based user management interface
- Command-line user creation utility

## User Management Methods

There are two ways to manage users in TS Jobscanner:

### 1. Web Interface (Recommended)

Once you have at least one user account, you can manage users through the web interface:

1. **Access User Management:**
   - Login to the application
   - Navigate to "Users" in the top menu
   - Or go directly to: `http://localhost:8080/users`

2. **Create New Users:**
   - Click "Create New User" button
   - Fill in the required fields:
     - Username (required, unique)
     - Email (required, unique)
     - Password (required)
     - First Name (optional)
     - Last Name (optional)
     - Active status (checkbox, default: checked)

3. **Edit Existing Users:**
   - Click the edit icon next to any user
   - Update any fields (leave password blank to keep current password)
   - Toggle active status to enable/disable user login

4. **View User Details:**
   - Click the view icon to see detailed user information
   - Shows creation date, last login, and all user details

5. **Delete Users:**
   - Click the delete icon next to any user
   - Confirm deletion (cannot be undone)
   - Note: You cannot delete your own account

### 2. Command-Line Utility

A command-line user creation utility is provided: `create_user.go`

#### Usage Examples

**Create a user with all details:**
```bash
go run create_user.go -username=admin -email=admin@company.com -password=secure123 -firstname=Admin -lastname=User
```

**Create a basic user:**
```bash
go run create_user.go -username=john -email=john@company.com -password=password123
```

**Using a different config file:**
```bash
go run create_user.go -config=config.production.json -username=admin -email=admin@company.com -password=secure123
```

**Command-line parameters:**
- `-username` (required): Unique username for login
- `-email` (required): Unique email address
- `-password` (required): Password for the user
- `-firstname` (optional): User's first name
- `-lastname` (optional): User's last name
- `-config` (optional): Configuration file path (default: config.json)

## Getting Started - First User Creation

Since you need to be logged in to access the web interface, you must create your first user via command line:

1. **Navigate to the application directory:**
   ```bash
   cd /opt/dev/go-barcode-webapp
   ```

2. **Create your first admin user:**
   ```bash
   go run create_user.go -username=admin -email=admin@company.com -password=your_secure_password -firstname=Admin -lastname=User
   ```

3. **Start the application:**
   ```bash
   go run cmd/server/main.go
   # OR if you have a compiled binary:
   ./server
   ```

4. **Access the application:**
   - Open http://localhost:8080
   - You'll be redirected to the login page
   - Enter your admin credentials
   - Navigate to "Users" to manage additional users

## User Account Features

### User Properties
- **Username**: Unique identifier for login (cannot be changed after creation)
- **Email**: Must be unique across all users
- **Password**: Securely hashed using bcrypt
- **First Name & Last Name**: Optional display names
- **Active Status**: Controls whether user can log in
- **Timestamps**: Created date, last update, last login

### Security Features
1. **Password Hashing**: All passwords are hashed using bcrypt with cost factor 14
2. **Session Management**: Secure session cookies with HTTP-only flag
3. **Route Protection**: All application routes require authentication
4. **Session Expiration**: Sessions expire after 24 hours
5. **Secure Logout**: Sessions are properly cleaned up on logout
6. **Duplicate Prevention**: Usernames and emails must be unique

### User States
- **Active**: User can log in and access the system
- **Inactive**: User account exists but cannot log in

## Database Schema

The authentication system uses two tables:

### users table
```sql
CREATE TABLE users (
  userID bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  username varchar(191) NOT NULL UNIQUE,
  email varchar(191) NOT NULL UNIQUE,
  password_hash longtext NOT NULL,
  first_name longtext,
  last_name longtext,
  is_active tinyint(1) DEFAULT '1',
  created_at datetime(3) DEFAULT NULL,
  updated_at datetime(3) DEFAULT NULL,
  last_login datetime(3) DEFAULT NULL,
  PRIMARY KEY (userID)
);
```

### sessions table
```sql
CREATE TABLE sessions (
  session_id varchar(191) NOT NULL,
  user_id bigint UNSIGNED NOT NULL,
  expires_at datetime(3) NOT NULL,
  created_at datetime(3) DEFAULT NULL,
  PRIMARY KEY (session_id)
);
```

## Production Deployment

For production deployment:

1. **Create production users:**
   ```bash
   go run create_user.go -config=config.production.json -username=admin -email=admin@yourdomain.com -password=secure_password
   ```

2. **Security considerations:**
   - Use strong passwords for all users
   - Regularly review and remove unused accounts
   - Monitor user login activity
   - Consider implementing additional security measures:
     - HTTPS enforcement
     - Rate limiting on login attempts
     - Regular password rotation policies

3. **Backup considerations:**
   - Include `users` and `sessions` tables in database backups
   - Sessions can be cleared during maintenance (users will need to re-login)
   - Test user account restoration in disaster recovery procedures

## Troubleshooting

### Common Login Issues

**"Invalid username or password" error:**
- Verify the username exists and is spelled correctly
- Check if the user account is active (not disabled)
- Ensure the password is correct
- Check the Users page to confirm account status

**Redirect loops or session issues:**
- Clear browser cookies and try again
- Check if the sessions table exists in the database
- Verify system time is correct (affects session expiration)
- Restart the application to clear any corrupted session data

### User Management Issues

**Cannot create user - "User already exists":**
- Check if username or email is already in use
- Usernames and emails must be unique across all users

**Cannot access user management:**
- Ensure you're logged in with a valid account
- Check if your user account is still active
- Verify the /users route is properly configured

### Database Issues

**Database connection errors:**
- Verify database configuration in your config file
- Ensure the database server is running
- Check database connection permissions
- Confirm the database exists and is accessible

**Missing tables:**
- The application automatically creates user tables on startup
- If tables are missing, restart the application
- Check database migration logs for errors

## API Access

The authentication system protects all API endpoints. For programmatic access:
- All API calls require a valid session cookie
- Session cookies are automatically included when using the same browser
- For automated systems, consider implementing API tokens (future enhancement)

## Security Best Practices

1. **Password Policies:**
   - Use strong passwords with mixed characters
   - Avoid common passwords or dictionary words
   - Consider password complexity requirements

2. **Account Management:**
   - Regularly review user accounts and remove unused ones
   - Disable accounts immediately when users leave
   - Use descriptive usernames and proper names

3. **Session Security:**
   - Users should log out when finished
   - Clear browser data on shared computers
   - Monitor for unusual login patterns

4. **System Security:**
   - Keep the application updated
   - Use HTTPS in production
   - Regularly backup user data
   - Monitor application logs for security issues

## Future Enhancements

Planned improvements for the user management system:
- Role-based access control (admin vs regular users)
- Multi-factor authentication
- API token authentication
- Password reset functionality
- User audit logging
- Bulk user management operations