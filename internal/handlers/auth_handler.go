package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go-barcode-webapp/internal/config"
	"go-barcode-webapp/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db     *gorm.DB
	config *config.Config
}

func NewAuthHandler(db *gorm.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{db: db, config: cfg}
}

// LoginForm displays the login page
func (h *AuthHandler) LoginForm(c *gin.Context) {
	// Check if user is already logged in
	if sessionID, err := c.Cookie("session_id"); err == nil && sessionID != "" {
		if h.validateSession(sessionID) {
			c.Redirect(http.StatusSeeOther, "/")
			return
		}
	}

	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "Login",
	})
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var loginData struct {
		Username string `form:"username" binding:"required"`
		Password string `form:"password" binding:"required"`
	}

	if err := c.ShouldBind(&loginData); err != nil {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"title": "Login",
			"error": "Please fill in all fields",
		})
		return
	}

	// Find user by username
	var user models.User
	if err := h.db.Where("username = ? AND is_active = ?", loginData.Username, true).First(&user).Error; err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"title": "Login",
			"error": "Invalid username or password",
		})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(loginData.Password)); err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"title": "Login",
			"error": "Invalid username or password",
		})
		return
	}

	// Create session
	sessionID := h.generateSessionID()
	sessionTimeout := time.Duration(h.config.Security.SessionTimeout) * time.Second
	session := models.Session{
		SessionID: sessionID,
		UserID:    user.UserID,
		ExpiresAt: time.Now().Add(sessionTimeout),
		CreatedAt: time.Now(),
	}

	fmt.Printf("DEBUG: Creating session for user %s (ID: %d)\n", user.Username, user.UserID)
	if err := h.db.Create(&session).Error; err != nil {
		fmt.Printf("DEBUG: Session creation failed: %v\n", err)
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{
			"title": "Login",
			"error": "Login failed. Please try again.",
		})
		return
	}

	// Update last login
	now := time.Now()
	user.LastLogin = &now
	h.db.Save(&user)

	// Set cookie
	c.SetCookie("session_id", sessionID, h.config.Security.SessionTimeout, "/", "", false, true)
	fmt.Printf("DEBUG: Login successful, session created: %s\n", sessionID)

	// Redirect to home
	c.Redirect(http.StatusSeeOther, "/")
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	if sessionID, err := c.Cookie("session_id"); err == nil {
		// Delete session from database
		h.db.Where("session_id = ?", sessionID).Delete(&models.Session{})
	}

	// Clear cookie
	c.SetCookie("session_id", "", -1, "/", "", false, true)

	// Redirect to login
	c.Redirect(http.StatusSeeOther, "/login")
}

// AuthMiddleware checks if user is authenticated
func (h *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Printf("DEBUG: AuthMiddleware called for URL: %s\n", c.Request.URL.Path)
		
		sessionID, err := c.Cookie("session_id")
		if err != nil || sessionID == "" {
			fmt.Printf("DEBUG: No session cookie found, redirecting to login\n")
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		fmt.Printf("DEBUG: Found session cookie: %s\n", sessionID)

		// Validate session
		var session models.Session
		if err := h.db.Where("session_id = ? AND expires_at > ?", sessionID, time.Now()).First(&session).Error; err != nil {
			fmt.Printf("DEBUG: Session validation failed: %v\n", err)
			// Clean up invalid session cookie
			c.SetCookie("session_id", "", -1, "/", "", false, true)
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		// Load the user and verify they are still active
		var user models.User
		if err := h.db.Where("userID = ? AND is_active = ?", session.UserID, true).First(&user).Error; err != nil {
			fmt.Printf("DEBUG: User not found or inactive for session: %v\n", err)
			// Delete the session since user is inactive/deleted
			h.db.Where("session_id = ?", sessionID).Delete(&models.Session{})
			c.SetCookie("session_id", "", -1, "/", "", false, true)
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		// Optional: Extend session on activity (sliding expiration)
		// Uncomment if you want sessions to extend on each request
		// sessionTimeout := time.Duration(h.config.Security.SessionTimeout) * time.Second
		// session.ExpiresAt = time.Now().Add(sessionTimeout)
		// h.db.Save(&session)

		fmt.Printf("DEBUG: Session valid for user: %s (ID: %d)\n", user.Username, user.UserID)

		// Store user in context
		c.Set("user", user)
		c.Set("userID", session.UserID)
		c.Next()
	}
}

// validateSession checks if a session is valid and the user is active
func (h *AuthHandler) validateSession(sessionID string) bool {
	var session models.Session
	if err := h.db.Where("session_id = ? AND expires_at > ?", sessionID, time.Now()).First(&session).Error; err != nil {
		return false
	}
	
	// Also check if the user is still active
	var user models.User
	return h.db.Where("userID = ? AND is_active = ?", session.UserID, true).First(&user).Error == nil
}

// generateSessionID creates a new session ID
func (h *AuthHandler) generateSessionID() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// CleanupExpiredSessions removes expired sessions from the database
func (h *AuthHandler) CleanupExpiredSessions() error {
	result := h.db.Where("expires_at < ?", time.Now()).Delete(&models.Session{})
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected > 0 {
		fmt.Printf("DEBUG: Cleaned up %d expired sessions\n", result.RowsAffected)
	}
	
	return nil
}

// StartSessionCleanup starts a background goroutine to periodically clean up expired sessions
func (h *AuthHandler) StartSessionCleanup() {
	go func() {
		ticker := time.NewTicker(30 * time.Minute) // Clean up every 30 minutes
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				if err := h.CleanupExpiredSessions(); err != nil {
					fmt.Printf("ERROR: Failed to cleanup expired sessions: %v\n", err)
				}
			}
		}
	}()
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CreateUser creates a new user (helper function for user management)
func (h *AuthHandler) CreateUser(username, email, password, firstName, lastName string) error {
	// Check if user already exists
	var existingUser models.User
	if err := h.db.Where("username = ? OR email = ?", username, email).First(&existingUser).Error; err == nil {
		return gorm.ErrDuplicatedKey
	}

	// Hash password
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return err
	}

	// Create user
	user := models.User{
		Username:     username,
		Email:        email,
		PasswordHash: hashedPassword,
		FirstName:    firstName,
		LastName:     lastName,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	return h.db.Create(&user).Error
}

// GetCurrentUser returns the current authenticated user
func GetCurrentUser(c *gin.Context) (*models.User, bool) {
	if user, exists := c.Get("user"); exists {
		if u, ok := user.(models.User); ok {
			return &u, true
		}
	}
	return nil, false
}

// User Management Web Interface Handlers

// ListUsers displays all users
func (h *AuthHandler) ListUsers(c *gin.Context) {
	fmt.Printf("DEBUG: ListUsers called - URL: %s\n", c.Request.URL.Path)
	
	var users []models.User
	if err := h.db.Order("created_at DESC").Find(&users).Error; err != nil {
		fmt.Printf("DEBUG: Database error: %v\n", err)
		currentUser, _ := GetCurrentUser(c)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error(), "user": currentUser})
		return
	}

	fmt.Printf("DEBUG: Found %d users\n", len(users))
	currentUser, exists := GetCurrentUser(c)
	fmt.Printf("DEBUG: Current user exists: %v, User: %+v\n", exists, currentUser)
	
	c.HTML(http.StatusOK, "users_list.html", gin.H{
		"title": "User Management",
		"users": users,
		"user":  currentUser,
	})
	fmt.Printf("DEBUG: ListUsers template rendered\n")
}

// NewUserForm displays the create user form
func (h *AuthHandler) NewUserForm(c *gin.Context) {
	// Debug: Let's see what's happening
	fmt.Printf("DEBUG: NewUserForm called - URL: %s\n", c.Request.URL.Path)
	
	currentUser, exists := GetCurrentUser(c)
	fmt.Printf("DEBUG: User exists: %v, User: %+v\n", exists, currentUser)
	
	if !exists || currentUser == nil {
		fmt.Printf("DEBUG: No user found, redirecting to login\n")
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}
	
	fmt.Printf("DEBUG: Rendering user_form.html template\n")
	c.HTML(http.StatusOK, "user_form.html", gin.H{
		"title":    "Create New User",
		"formUser": &models.User{},
		"user":     currentUser,
	})
	fmt.Printf("DEBUG: Template rendered successfully\n")
}

// CreateUserWeb handles user creation from web form
func (h *AuthHandler) CreateUserWeb(c *gin.Context) {
	username := c.PostForm("username")
	email := c.PostForm("email")
	password := c.PostForm("password")
	firstName := c.PostForm("first_name")
	lastName := c.PostForm("last_name")
	isActiveStr := c.PostForm("is_active")
	
	isActive := isActiveStr == "on" || isActiveStr == "true"

	if username == "" || email == "" || password == "" {
		currentUser, _ := GetCurrentUser(c)
		c.HTML(http.StatusBadRequest, "user_form.html", gin.H{
			"title": "Create New User",
			"formUser": &models.User{
				Username:  username,
				Email:     email,
				FirstName: firstName,
				LastName:  lastName,
				IsActive:  isActive,
			},
			"user":  currentUser,
			"error": "Username, email and password are required",
		})
		return
	}

	if err := h.CreateUser(username, email, password, firstName, lastName); err != nil {
		var errorMsg string
		if err == gorm.ErrDuplicatedKey {
			errorMsg = "User with this username or email already exists"
		} else {
			errorMsg = err.Error()
		}
		
		currentUser, _ := GetCurrentUser(c)
		c.HTML(http.StatusInternalServerError, "user_form.html", gin.H{
			"title": "Create New User",
			"formUser": &models.User{
				Username:  username,
				Email:     email,
				FirstName: firstName,
				LastName:  lastName,
				IsActive:  isActive,
			},
			"user":  currentUser,
			"error": errorMsg,
		})
		return
	}

	c.Redirect(http.StatusFound, "/users")
}

// GetUser displays user details
func (h *AuthHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	
	var user models.User
	if err := h.db.Where("userID = ?", userID).First(&user).Error; err != nil {
		currentUser, _ := GetCurrentUser(c)
	c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "User not found", "user": currentUser})
		return
	}

	currentUser, _ := GetCurrentUser(c)
	c.HTML(http.StatusOK, "user_detail.html", gin.H{
		"title":    "User Details",
		"viewUser": user,
		"user":     currentUser,
	})
}

// EditUserForm displays the edit user form
func (h *AuthHandler) EditUserForm(c *gin.Context) {
	userID := c.Param("id")
	
	var user models.User
	if err := h.db.Where("userID = ?", userID).First(&user).Error; err != nil {
		currentUser, _ := GetCurrentUser(c)
	c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "User not found", "user": currentUser})
		return
	}

	currentUser, _ := GetCurrentUser(c)
	c.HTML(http.StatusOK, "user_form.html", gin.H{
		"title":    "Edit User",
		"formUser": user,
		"user":     currentUser,
	})
}

// UpdateUser handles user updates
func (h *AuthHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	
	var user models.User
	if err := h.db.Where("userID = ?", userID).First(&user).Error; err != nil {
		currentUser, _ := GetCurrentUser(c)
	c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "User not found", "user": currentUser})
		return
	}

	username := c.PostForm("username")
	email := c.PostForm("email")
	password := c.PostForm("password")
	firstName := c.PostForm("first_name")
	lastName := c.PostForm("last_name")
	isActiveStr := c.PostForm("is_active")
	
	isActive := isActiveStr == "on" || isActiveStr == "true"

	if username == "" || email == "" {
		currentUser, _ := GetCurrentUser(c)
		c.HTML(http.StatusBadRequest, "user_form.html", gin.H{
			"title":    "Edit User",
			"formUser": user,
			"user":     currentUser,
			"error":    "Username and email are required",
		})
		return
	}

	// Check for duplicate username/email (excluding current user)
	var existingUser models.User
	if err := h.db.Where("(username = ? OR email = ?) AND userID != ?", username, email, userID).First(&existingUser).Error; err == nil {
		currentUser, _ := GetCurrentUser(c)
		c.HTML(http.StatusBadRequest, "user_form.html", gin.H{
			"title":    "Edit User",
			"formUser": user,
			"user":     currentUser,
			"error":    "User with this username or email already exists",
		})
		return
	}

	// Update user fields
	user.Username = username
	user.Email = email
	user.FirstName = firstName
	user.LastName = lastName
	user.IsActive = isActive
	user.UpdatedAt = time.Now()

	// Update password if provided
	if password != "" {
		hashedPassword, err := HashPassword(password)
		if err != nil {
			currentUser, _ := GetCurrentUser(c)
			c.HTML(http.StatusInternalServerError, "user_form.html", gin.H{
				"title":    "Edit User",
				"formUser": user,
				"user":     currentUser,
				"error":    "Failed to hash password",
			})
			return
		}
		user.PasswordHash = hashedPassword
	}

	if err := h.db.Save(&user).Error; err != nil {
		currentUser, _ := GetCurrentUser(c)
		c.HTML(http.StatusInternalServerError, "user_form.html", gin.H{
			"title":    "Edit User",
			"formUser": user,
			"user":     currentUser,
			"error":    err.Error(),
		})
		return
	}

	c.Redirect(http.StatusFound, "/users")
}

// DeleteUser handles user deletion
func (h *AuthHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	
	// Don't allow deleting the current user
	currentUser, exists := GetCurrentUser(c)
	if exists && currentUser.UserID == parseUserID(userID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete your own account"})
		return
	}

	if err := h.db.Where("userID = ?", userID).Delete(&models.User{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// Helper function to parse user ID
func parseUserID(userIDStr string) uint {
	if userIDStr == "" {
		return 0
	}
	
	// Convert string to uint
	if id, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
		return uint(id)
	}
	
	return 0
}