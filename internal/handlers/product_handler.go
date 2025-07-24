package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"go-barcode-webapp/internal/models"
	"go-barcode-webapp/internal/repository"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	productRepo *repository.ProductRepository
}

func NewProductHandler(productRepo *repository.ProductRepository) *ProductHandler {
	return &ProductHandler{productRepo: productRepo}
}

// Web interface handlers
func (h *ProductHandler) ListProductsWeb(c *gin.Context) {
	startTime := time.Now()
	log.Printf("🚀 ProductHandler.ListProductsWeb() started")
	
	user, _ := GetCurrentUser(c)
	
	params := &models.FilterParams{}
	if err := c.ShouldBindQuery(params); err != nil {
		log.Printf("❌ Error binding query parameters: %v", err)
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/error?code=400&message=Bad Request&details=%s", err.Error()))
		return
	}
	
	// Handle search parameter
	searchParam := c.Query("search")
	if searchParam != "" {
		params.SearchTerm = searchParam
	}

	// Handle pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	
	limit := 20 // Products per page
	params.Limit = limit
	params.Offset = (page - 1) * limit
	params.Page = page

	viewType := c.DefaultQuery("view", "list") // Default to list view
	log.Printf("🐛 DEBUG: Product view requested: viewType='%s'", viewType)

	// Get products from database
	dbStart := time.Now()
	products, err := h.productRepo.List(params)
	dbTime := time.Since(dbStart)
	log.Printf("⏱️  Database query took: %v", dbTime)
	
	if err != nil {
		log.Printf("❌ Database error: %v", err)
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/error?code=500&message=Database Error&details=%s", err.Error()))
		return
	}
	
	// Get total product count for pagination (simplified for now)
	totalProducts := len(products) // This should be a proper count query
	totalPages := (totalProducts + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}

	templateStart := time.Now()
	SafeHTML(c, http.StatusOK, "products_standalone.html", gin.H{
		"title":         "Products",
		"products":      products,
		"params":        params,
		"user":          user,
		"viewType":      viewType,
		"currentPage":   "products",
		"pageNumber":    page,
		"hasNextPage":   page < totalPages,
		"totalPages":    totalPages,
		"totalProducts": totalProducts,
	})
	
	templateTime := time.Since(templateStart)
	totalTime := time.Since(startTime)
	log.Printf("⏱️  Template rendering took: %v", templateTime)
	log.Printf("🏁 ProductHandler.ListProductsWeb() completed in %v", totalTime)
}

func (h *ProductHandler) NewProductForm(c *gin.Context) {
	user, _ := GetCurrentUser(c)
	
	// Get categories for the form
	categories, err := h.productRepo.GetAllCategories()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error(), "user": user})
		return
	}

	c.HTML(http.StatusOK, "product_form.html", gin.H{
		"title":      "New Product",
		"product":    &models.Product{},
		"categories": categories,
		"user":       user,
	})
}

// API handlers (existing)
func (h *ProductHandler) ListProducts(c *gin.Context) {
	params := &models.FilterParams{}
	if err := c.ShouldBindQuery(params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	products, err := h.productRepo.List(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"products": products})
}

func (h *ProductHandler) GetProductAPI(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}
	
	product, err := h.productRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"product": product})
}

func (h *ProductHandler) CreateProductAPI(c *gin.Context) {
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product data"})
		return
	}

	if err := h.productRepo.Create(&product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"product": product})
}

func (h *ProductHandler) UpdateProductAPI(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product data"})
		return
	}

	product.ProductID = uint(id)
	if err := h.productRepo.Update(&product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"product": product})
}

func (h *ProductHandler) DeleteProductAPI(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	if err := h.productRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

func (h *ProductHandler) GetCategoriesAPI(c *gin.Context) {
	categories, err := h.productRepo.GetAllCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get categories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"categories": categories})
}