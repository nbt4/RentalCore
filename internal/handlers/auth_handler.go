package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"go-barcode-webapp/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{db: db}
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
	session := models.Session{
		SessionID: sessionID,
		UserID:    user.UserID,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hour session
		CreatedAt: time.Now(),
	}

	if err := h.db.Create(&session).Error; err != nil {
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
	c.SetCookie("session_id", sessionID, 86400, "/", "", false, true)

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
		sessionID, err := c.Cookie("session_id")
		if err != nil || sessionID == "" {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		// Validate session
		var session models.Session
		if err := h.db.Preload("User").Where("session_id = ? AND expires_at > ?", sessionID, time.Now()).First(&session).Error; err != nil {
			c.SetCookie("session_id", "", -1, "/", "", false, true)
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		// Store user in context
		c.Set("user", session.User)
		c.Set("user_id", session.UserID)
		c.Next()
	}
}

// validateSession checks if a session is valid
func (h *AuthHandler) validateSession(sessionID string) bool {
	var session models.Session
	return h.db.Where("session_id = ? AND expires_at > ?", sessionID, time.Now()).First(&session).Error == nil
}

// generateSessionID creates a new session ID
func (h *AuthHandler) generateSessionID() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
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