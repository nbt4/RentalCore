package handlers

import (
	"log"
	"net/http"

	"go-barcode-webapp/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type EmailTemplateHandler struct {
	db *gorm.DB
}

func NewEmailTemplateHandler(db *gorm.DB) *EmailTemplateHandler {
	return &EmailTemplateHandler{
		db: db,
	}
}

// ListEmailTemplates displays all email templates - simplified version without DB operations
func (h *EmailTemplateHandler) ListEmailTemplates(c *gin.Context) {
	user, exists := GetCurrentUser(c)
	if !exists {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	log.Printf("📧 EMAIL TEMPLATES ROUTE CALLED - URL: %s", c.Request.URL.Path)
	
	// Temporarily use empty templates list to avoid DB migration issues
	log.Printf("📧 DEBUG: Using empty templates list to avoid REFERENCES permission issues")
	
	var templates []models.EmailTemplate
	// Empty list for now until DB permissions are resolved
	
	log.Printf("📧 DEBUG: Found %d templates (empty for testing)", len(templates))
	
	// Render template
	c.HTML(http.StatusOK, "email_templates_list.html", gin.H{
		"title":     "E-Mail Settings",
		"templates": templates,
		"user":      user,
	})
}

// Placeholder methods to prevent compilation errors - to be implemented later
func (h *EmailTemplateHandler) NewEmailTemplateForm(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}

func (h *EmailTemplateHandler) CreateEmailTemplate(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}

func (h *EmailTemplateHandler) GetEmailTemplate(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}

func (h *EmailTemplateHandler) EditEmailTemplateForm(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}

func (h *EmailTemplateHandler) UpdateEmailTemplate(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}

func (h *EmailTemplateHandler) DeleteEmailTemplate(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}

func (h *EmailTemplateHandler) PreviewEmailTemplate(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}

func (h *EmailTemplateHandler) SetDefaultEmailTemplate(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}

func (h *EmailTemplateHandler) TestEmailTemplate(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}