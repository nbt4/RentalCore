package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"go-barcode-webapp/internal/config"
	"go-barcode-webapp/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
	"gorm.io/gorm"
)

type WebAuthnHandler struct {
	db     *gorm.DB
	config *config.Config
}

func NewWebAuthnHandler(db *gorm.DB, cfg *config.Config) *WebAuthnHandler {
	return &WebAuthnHandler{db: db, config: cfg}
}

// GetDB returns the database connection for use in other parts of the application
func (h *WebAuthnHandler) GetDB() *gorm.DB {
	return h.db
}

// ================================================================
// PASSKEY (WebAuthn) ENDPOINTS
// ================================================================

// StartPasskeyRegistration initiates passkey registration for a user
func (h *WebAuthnHandler) StartPasskeyRegistration(c *gin.Context) {
	log.Printf("DEBUG: StartPasskeyRegistration called")
	
	currentUser, exists := GetCurrentUser(c)
	if !exists || currentUser == nil {
		log.Printf("ERROR: User not authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	
	log.Printf("DEBUG: User authenticated: %s (ID: %d)", currentUser.Username, currentUser.UserID)

	// Check if running over HTTPS or localhost (WebAuthn requirement)
	host := c.Request.Host
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	
	log.Printf("DEBUG: Request scheme: %s, host: %s", scheme, host)
	
	// WebAuthn requires HTTPS except for localhost and internal networks
	isLocalhost := strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1")
	isInternalHost := strings.Contains(host, "debian01") || strings.Contains(host, ".local") || strings.Contains(host, "10.0.0.") || strings.Contains(host, "192.168.")
	
	log.Printf("DEBUG: Host validation - host: %s, isLocalhost: %v, isInternalHost: %v", host, isLocalhost, isInternalHost)
	
	if scheme != "https" && !isLocalhost && !isInternalHost {
		log.Printf("ERROR: WebAuthn requires HTTPS except for localhost/internal hosts - host: %s, scheme: %s", host, scheme)
		c.JSON(http.StatusBadRequest, gin.H{"error": "WebAuthn requires HTTPS for security"})
		return
	}
	
	log.Printf("DEBUG: Host validation passed - scheme: %s, host: %s", scheme, host)

	// Generate challenge
	challenge := make([]byte, 32)
	if _, err := rand.Read(challenge); err != nil {
		log.Printf("ERROR: Failed to generate challenge: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate challenge"})
		return
	}
	
	log.Printf("DEBUG: Challenge generated")

	// Create WebAuthn session
	sessionID := generateSessionID()
	session := models.WebAuthnSession{
		SessionID:   sessionID,
		UserID:      currentUser.UserID,
		Challenge:   base64.URLEncoding.EncodeToString(challenge),
		SessionType: "registration",
		ExpiresAt:   time.Now().Add(5 * time.Minute),
		CreatedAt:   time.Now(),
	}

	log.Printf("DEBUG: Attempting to create WebAuthn session in database")
	if err := h.db.Create(&session).Error; err != nil {
		log.Printf("ERROR: Failed to create session in database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}
	
	log.Printf("DEBUG: Session created successfully with ID: %s", sessionID)

	// Return registration options - challenge and user.id need to be base64url encoded strings
	// but the client will need to convert them to Uint8Array
	challengeB64 := base64.URLEncoding.EncodeToString(challenge)
	userIdB64 := base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%d", currentUser.UserID)))
	
	log.Printf("DEBUG: Generated challenge (base64): %s", challengeB64)
	log.Printf("DEBUG: Generated user ID (base64): %s", userIdB64)
	
	// For WebAuthn RP ID, it must be a valid domain suffix of the origin
	// For .local domains, we need to use the full domain or a valid suffix
	rpID := host
	if strings.Contains(host, ":") {
		// Remove port if present
		rpID = strings.Split(host, ":")[0]
	}
	
	log.Printf("DEBUG: Using RP ID: %s for host: %s", rpID, host)
	
	options := map[string]interface{}{
		"challenge": challengeB64,
		"rp": map[string]string{
			"name": "RentalCore",
			"id":   rpID,
		},
		"user": map[string]interface{}{
			"id":          userIdB64,
			"name":        currentUser.Email,
			"displayName": fmt.Sprintf("%s %s", currentUser.FirstName, currentUser.LastName),
		},
		"pubKeyCredParams": []map[string]interface{}{
			{"type": "public-key", "alg": -7},  // ES256
			{"type": "public-key", "alg": -257}, // RS256
		},
		"attestation": "direct",
		"timeout":     300000, // 5 minutes
		"authenticatorSelection": map[string]interface{}{
			"authenticatorAttachment": "platform",
			"userVerification":        "preferred",
			"residentKey":             "preferred",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"sessionId": sessionID,
		"options":   options,
	})
}

// CompletePasskeyRegistration completes passkey registration
func (h *WebAuthnHandler) CompletePasskeyRegistration(c *gin.Context) {
	log.Printf("DEBUG: CompletePasskeyRegistration called")
	
	currentUser, exists := GetCurrentUser(c)
	if !exists || currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Use flexible JSON parsing to avoid validation errors
	var request map[string]interface{}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("ERROR: Failed to bind JSON request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Extract fields manually
	sessionID, _ := request["sessionId"].(string)
	name, _ := request["name"].(string)
	credential, _ := request["credential"].(string)
	credentialID, _ := request["credentialId"].(string)
	
	log.Printf("DEBUG: Extracted fields - sessionID: %s, name: %s, credential length: %d, credentialID: %s", 
		sessionID, name, len(credential), credentialID)
	
	if sessionID == "" || name == "" || credential == "" || credentialID == "" {
		log.Printf("ERROR: Missing required fields - sessionID empty: %v, name empty: %v, credential empty: %v, credentialID empty: %v", 
			sessionID == "", name == "", credential == "", credentialID == "")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	// Verify session
	log.Printf("DEBUG: Verifying session - sessionID: %s, userID: %d", sessionID, currentUser.UserID)
	var session models.WebAuthnSession
	if err := h.db.Where("session_id = ? AND user_id = ? AND session_type = ? AND expires_at > ?",
		sessionID, currentUser.UserID, "registration", time.Now()).First(&session).Error; err != nil {
		log.Printf("ERROR: Session verification failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired session"})
		return
	}
	log.Printf("DEBUG: Session verified successfully")

	// For now, we'll store a placeholder public key
	publicKeyBytes := []byte("placeholder-public-key")

	// Create passkey record
	passkey := models.UserPasskey{
		UserID:       currentUser.UserID,
		Name:         name,
		CredentialID: credentialID,
		PublicKey:    publicKeyBytes,
		SignCount:    0,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := h.db.Create(&passkey).Error; err != nil {
		log.Printf("ERROR: Failed to save passkey: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save passkey"})
		return
	}

	// Clean up session
	h.db.Delete(&session)

	// Log authentication attempt
	h.logAuthAttempt(currentUser.UserID, "passkey_registration", c.ClientIP(), c.GetHeader("User-Agent"), true, nil, &passkey.PasskeyID)

	c.JSON(http.StatusOK, gin.H{
		"message":   "Passkey registered successfully",
		"passkeyId": passkey.PasskeyID,
	})
}

// ================================================================
// PASSKEY AUTHENTICATION ENDPOINTS
// ================================================================

// StartPasskeyAuthentication initiates passkey authentication for login
func (h *WebAuthnHandler) StartPasskeyAuthentication(c *gin.Context) {
	log.Printf("DEBUG: StartPasskeyAuthentication called")
	
	// Check if running over HTTPS or localhost (WebAuthn requirement)
	host := c.Request.Host
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	
	log.Printf("DEBUG: Request scheme: %s, host: %s", scheme, host)
	
	// WebAuthn requires HTTPS except for localhost and internal networks
	isLocalhost := strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1")
	isInternalHost := strings.Contains(host, "debian01") || strings.Contains(host, ".local") || strings.Contains(host, "10.0.0.") || strings.Contains(host, "192.168.")
	
	if scheme != "https" && !isLocalhost && !isInternalHost {
		log.Printf("ERROR: WebAuthn requires HTTPS except for localhost/internal hosts")
		c.JSON(http.StatusBadRequest, gin.H{"error": "WebAuthn requires HTTPS for security"})
		return
	}
	
	// Generate challenge
	challenge := make([]byte, 32)
	if _, err := rand.Read(challenge); err != nil {
		log.Printf("ERROR: Failed to generate challenge: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate challenge"})
		return
	}
	
	// Create WebAuthn session for authentication
	sessionID := generateSessionID()
	session := models.WebAuthnSession{
		SessionID:   sessionID,
		UserID:      0, // No user ID yet for authentication
		Challenge:   base64.URLEncoding.EncodeToString(challenge),
		SessionType: "authentication",
		ExpiresAt:   time.Now().Add(5 * time.Minute),
		CreatedAt:   time.Now(),
	}

	if err := h.db.Create(&session).Error; err != nil {
		log.Printf("ERROR: Failed to create session in database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}
	
	// For authentication, we can allow any registered passkey
	var passkeys []models.UserPasskey
	h.db.Where("is_active = ?", true).Find(&passkeys)
	
	log.Printf("DEBUG: Found %d active passkeys for authentication", len(passkeys))
	
	allowCredentials := make([]map[string]interface{}, len(passkeys))
	for i, passkey := range passkeys {
		log.Printf("DEBUG: Adding passkey to allowCredentials - ID: %s, Name: %s", passkey.CredentialID, passkey.Name)
		allowCredentials[i] = map[string]interface{}{
			"type": "public-key",
			"id":   passkey.CredentialID, // Use the credential ID directly as it's already base64url encoded
		}
	}
	
	// For WebAuthn RP ID, it must be a valid domain suffix of the origin
	rpID := host
	if strings.Contains(host, ":") {
		// Remove port if present
		rpID = strings.Split(host, ":")[0]
	}
	
	log.Printf("DEBUG: Using RP ID: %s for authentication on host: %s", rpID, host)
	log.Printf("DEBUG: Prepared %d allowCredentials for authentication", len(allowCredentials))
	
	// Return authentication options
	challengeB64 := base64.URLEncoding.EncodeToString(challenge)
	options := map[string]interface{}{
		"challenge":        challengeB64,
		"timeout":          300000, // 5 minutes
		"rpId":             rpID,
		"allowCredentials": allowCredentials,
		"userVerification": "preferred",
	}

	c.JSON(http.StatusOK, gin.H{
		"sessionId": sessionID,
		"options":   options,
	})
}

// CompletePasskeyAuthentication completes passkey authentication for login
func (h *WebAuthnHandler) CompletePasskeyAuthentication(c *gin.Context) {
	log.Printf("DEBUG: CompletePasskeyAuthentication called")
	
	var request map[string]interface{}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("ERROR: Failed to bind JSON request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Extract fields manually
	sessionID, _ := request["sessionId"].(string)
	credentialID, _ := request["credentialId"].(string)
	
	if sessionID == "" || credentialID == "" {
		log.Printf("ERROR: Missing required fields")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	// Verify session
	var session models.WebAuthnSession
	if err := h.db.Where("session_id = ? AND session_type = ? AND expires_at > ?",
		sessionID, "authentication", time.Now()).First(&session).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired session"})
		return
	}

	// Find the passkey and associated user
	var passkey models.UserPasskey
	if err := h.db.Where("credential_id = ? AND is_active = ?", credentialID, true).First(&passkey).Error; err != nil {
		log.Printf("ERROR: Passkey not found: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid passkey"})
		return
	}

	// Get the user
	var user models.User
	if err := h.db.Where("userID = ? AND is_active = ?", passkey.UserID, true).First(&user).Error; err != nil {
		log.Printf("ERROR: User not found or inactive: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// TODO: Verify the WebAuthn signature (simplified for now)
	// In a production system, you would:
	// 1. Verify the authenticator data
	// 2. Verify the client data JSON
	// 3. Verify the signature using the stored public key
	
	// Update passkey usage
	passkey.LastUsed = &[]time.Time{time.Now()}[0]
	passkey.SignCount++
	h.db.Save(&passkey)

	// Create user session (similar to password login)
	userSession := models.Session{
		SessionID: generateSessionID(),
		UserID:    user.UserID,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hour session
		CreatedAt: time.Now(),
	}
	
	if err := h.db.Create(&userSession).Error; err != nil {
		log.Printf("ERROR: Failed to create user session: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	// Set session cookie
	c.SetCookie("session_id", userSession.SessionID, 24*60*60, "/", "", false, true)

	// Clean up WebAuthn session
	h.db.Delete(&session)

	// Log successful authentication
	h.logAuthAttempt(user.UserID, "passkey_authentication", c.ClientIP(), c.GetHeader("User-Agent"), true, nil, &passkey.PasskeyID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Authentication successful",
		"user":    user.Username,
	})
}

// ListUserPasskeys returns all passkeys for the current user
func (h *WebAuthnHandler) ListUserPasskeys(c *gin.Context) {
	currentUser, exists := GetCurrentUser(c)
	if !exists || currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var passkeys []models.UserPasskey
	if err := h.db.Where("user_id = ? AND is_active = ?", currentUser.UserID, true).Find(&passkeys).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve passkeys"})
		return
	}

	// Remove sensitive data
	for i := range passkeys {
		passkeys[i].PublicKey = nil
	}

	c.JSON(http.StatusOK, gin.H{"passkeys": passkeys})
}

// DeletePasskey removes a passkey
func (h *WebAuthnHandler) DeletePasskey(c *gin.Context) {
	currentUser, exists := GetCurrentUser(c)
	if !exists || currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	passkeyID := c.Param("id")

	// Verify ownership and delete
	result := h.db.Where("passkey_id = ? AND user_id = ?", passkeyID, currentUser.UserID).Delete(&models.UserPasskey{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete passkey"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Passkey not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Passkey deleted successfully"})
}

// SecurityStatus returns the current security status for the user
func (h *WebAuthnHandler) SecurityStatus(c *gin.Context) {
	currentUser, exists := GetCurrentUser(c)
	if !exists || currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get 2FA status
	var twoFAEnabled bool
	h.db.Raw("SELECT COALESCE(is_enabled, false) FROM user_2fa WHERE user_id = ?", currentUser.UserID).Scan(&twoFAEnabled)

	// Get passkey count
	var passkeyCount int64
	h.db.Model(&models.UserPasskey{}).Where("user_id = ? AND is_active = ?", currentUser.UserID, true).Count(&passkeyCount)

	c.JSON(http.StatusOK, gin.H{
		"twoFAEnabled": twoFAEnabled,
		"passkeyCount": passkeyCount,
	})
}

// ================================================================
// 2FA (TOTP) ENDPOINTS
// ================================================================

// Setup2FA generates a new TOTP secret and QR code for the user
func (h *WebAuthnHandler) Setup2FA(c *gin.Context) {
	log.Printf("DEBUG: Setup2FA called")
	
	currentUser, exists := GetCurrentUser(c)
	if !exists || currentUser == nil {
		log.Printf("ERROR: Setup2FA - user not authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	
	log.Printf("DEBUG: Setup2FA - user authenticated: %s (ID: %d)", currentUser.Username, currentUser.UserID)

	// Check if 2FA is already setup - use raw SQL to avoid GORM JSON scanning issues
	var count int64
	if err := h.db.Raw("SELECT COUNT(*) FROM user_2fa WHERE user_id = ? AND is_enabled = 1", currentUser.UserID).Scan(&count).Error; err != nil {
		log.Printf("ERROR: Setup2FA - failed to check existing 2FA: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to setup 2FA"})
		return
	}
	
	if count > 0 {
		log.Printf("ERROR: Setup2FA - 2FA already enabled")
		c.JSON(http.StatusBadRequest, gin.H{"error": "2FA is already enabled"})
		return
	}
	
	// Delete any existing incomplete setup records
	log.Printf("DEBUG: Setup2FA - deleting any existing incomplete setup records")
	h.db.Exec("DELETE FROM user_2fa WHERE user_id = ? AND is_enabled = 0", currentUser.UserID)

	// Generate TOTP secret
	log.Printf("DEBUG: Setup2FA - generating TOTP secret")
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "RentalCore",
		AccountName: currentUser.Email,
		SecretSize:  32,
	})
	if err != nil {
		log.Printf("ERROR: Setup2FA - failed to generate TOTP secret: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate 2FA secret"})
		return
	}
	log.Printf("DEBUG: Setup2FA - TOTP secret generated successfully")

	// Generate backup codes
	log.Printf("DEBUG: Setup2FA - generating backup codes")
	backupCodes := make([]string, 10)
	for i := range backupCodes {
		code := make([]byte, 6)
		rand.Read(code)
		backupCodes[i] = fmt.Sprintf("%x", code)[:8]
	}
	log.Printf("DEBUG: Setup2FA - generated %d backup codes", len(backupCodes))

	// Create 2FA record with manual JSON serialization
	log.Printf("DEBUG: Setup2FA - creating 2FA database record with manual JSON serialization")
	
	// Convert backup codes to JSON manually
	backupCodesJSON, err := json.Marshal(backupCodes)
	if err != nil {
		log.Printf("ERROR: Setup2FA - failed to marshal backup codes to JSON: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to setup 2FA"})
		return
	}
	
	log.Printf("DEBUG: Setup2FA - backup codes JSON: %s", string(backupCodesJSON))
	
	// Use raw SQL to insert the record to avoid GORM's JSON handling
	result := h.db.Exec(`
		INSERT INTO user_2fa (user_id, secret, qr_code_url, is_enabled, is_verified, backup_codes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, currentUser.UserID, key.Secret(), key.URL(), false, false, string(backupCodesJSON), time.Now(), time.Now())
	
	if result.Error != nil {
		log.Printf("ERROR: Setup2FA - failed to create 2FA record with raw SQL: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to setup 2FA"})
		return
	}
	
	log.Printf("DEBUG: Setup2FA - 2FA record saved successfully with backup codes")

	c.JSON(http.StatusOK, gin.H{
		"qrCodeURL":   key.URL(),
		"secret":      key.Secret(),
		"backupCodes": backupCodes,
	})
}

// Verify2FA verifies the TOTP code and enables 2FA
func (h *WebAuthnHandler) Verify2FA(c *gin.Context) {
	currentUser, exists := GetCurrentUser(c)
	if !exists || currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var request struct {
		Code string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get 2FA record using raw SQL to avoid JSON scanning issues
	var twoFA struct {
		TwoFAID   uint   `db:"two_fa_id"`
		UserID    uint   `db:"user_id"`
		Secret    string `db:"secret"`
		IsEnabled bool   `db:"is_enabled"`
	}
	
	if err := h.db.Raw("SELECT two_fa_id, user_id, secret, is_enabled FROM user_2fa WHERE user_id = ?", currentUser.UserID).Scan(&twoFA).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "2FA not setup"})
		return
	}

	// Verify TOTP code
	valid := totp.Validate(request.Code, twoFA.Secret)
	if !valid {
		h.logAuthAttempt(currentUser.UserID, "2fa_verification", c.ClientIP(), c.GetHeader("User-Agent"), false, stringPtr("Invalid TOTP code"), nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid verification code"})
		return
	}

	// Enable 2FA using raw SQL
	now := time.Now()
	if err := h.db.Exec("UPDATE user_2fa SET is_enabled = 1, is_verified = 1, last_used = ?, updated_at = ? WHERE user_id = ?", 
		now, now, currentUser.UserID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enable 2FA"})
		return
	}

	// Log successful verification
	h.logAuthAttempt(currentUser.UserID, "2fa_enabled", c.ClientIP(), c.GetHeader("User-Agent"), true, nil, nil)

	c.JSON(http.StatusOK, gin.H{"message": "2FA enabled successfully"})
}

// Disable2FA disables 2FA for the user
func (h *WebAuthnHandler) Disable2FA(c *gin.Context) {
	currentUser, exists := GetCurrentUser(c)
	if !exists || currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var request struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Simple approach: Use raw SQL to disable 2FA directly
	// First verify that 2FA exists and is enabled
	var secret string
	var backupCodes string
	var isEnabled bool
	
	row := h.db.Raw("SELECT secret, COALESCE(backup_codes, '[]'), is_enabled FROM user_2fa WHERE user_id = ? LIMIT 1", currentUser.UserID).Row()
	if err := row.Scan(&secret, &backupCodes, &isEnabled); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "2FA not found"})
		return
	}

	if !isEnabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "2FA not enabled"})
		return
	}

	// Verify the provided code
	valid := totp.Validate(request.Code, secret)
	if !valid {
		// Check backup codes
		var backupCodeList []string
		if backupCodes != "" && backupCodes != "[]" {
			json.Unmarshal([]byte(backupCodes), &backupCodeList)
			for _, backupCode := range backupCodeList {
				if backupCode == request.Code {
					valid = true
					break
				}
			}
		}
		
		if !valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid verification code"})
			return
		}
	}

	// Disable 2FA directly with raw SQL
	if err := h.db.Exec("UPDATE user_2fa SET is_enabled = 0 WHERE user_id = ?", currentUser.UserID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disable 2FA"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "2FA disabled successfully"})
}

// Get2FAStatus returns the current 2FA status for the user
func (h *WebAuthnHandler) Get2FAStatus(c *gin.Context) {
	currentUser, exists := GetCurrentUser(c)
	if !exists || currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get 2FA status using raw SQL to avoid JSON scanning issues
	var twoFA struct {
		IsEnabled   bool       `db:"is_enabled"`
		IsVerified  bool       `db:"is_verified"`
		CreatedAt   time.Time  `db:"created_at"`
		LastUsed    *time.Time `db:"last_used"`
		BackupCodes string     `db:"backup_codes"`
	}
	
	err := h.db.Raw("SELECT is_enabled, is_verified, created_at, last_used, backup_codes FROM user_2fa WHERE user_id = ?", currentUser.UserID).Scan(&twoFA).Error

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"enabled":    false,
			"verified":   false,
			"setupDate":  nil,
			"lastUsed":   nil,
		})
		return
	}

	// Count backup codes by parsing JSON
	var backupCodes []string
	backupCodesCount := 0
	if twoFA.BackupCodes != "" {
		if err := json.Unmarshal([]byte(twoFA.BackupCodes), &backupCodes); err == nil {
			backupCodesCount = len(backupCodes)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"enabled":    twoFA.IsEnabled,
		"verified":   twoFA.IsVerified,
		"setupDate":  twoFA.CreatedAt,
		"lastUsed":   twoFA.LastUsed,
		"backupCodesCount": backupCodesCount,
	})
}

// ================================================================
// HELPER FUNCTIONS
// ================================================================

// logAuthAttempt logs authentication attempts for security monitoring
func (h *WebAuthnHandler) logAuthAttempt(userID uint, method, ipAddress, userAgent string, success bool, failureReason *string, passkeyID *uint) {
	attempt := models.AuthenticationAttempt{
		UserID:        &userID,
		Method:        method,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		Success:       success,
		FailureReason: failureReason,
		PasskeyID:     passkeyID,
		AttemptedAt:   time.Now(),
	}

	// Save attempt (ignore errors to not break main flow)
	h.db.Create(&attempt)
}

// generateSessionID creates a new session ID
func generateSessionID() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

// Helper functions for pointers
func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}