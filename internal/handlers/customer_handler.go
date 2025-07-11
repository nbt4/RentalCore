package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"go-barcode-webapp/internal/models"
	"go-barcode-webapp/internal/repository"

	"github.com/gin-gonic/gin"
)

type CustomerHandler struct {
	customerRepo *repository.CustomerRepository
}

func NewCustomerHandler(customerRepo *repository.CustomerRepository) *CustomerHandler {
	return &CustomerHandler{customerRepo: customerRepo}
}

func (h *CustomerHandler) ListCustomers(c *gin.Context) {
	user, _ := GetCurrentUser(c)
	
	params := &models.FilterParams{}
	if err := c.ShouldBindQuery(params); err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": err.Error(), "user": user})
		return
	}

	// DEBUG: Log all query parameters
	fmt.Printf("DEBUG Customer Handler: All query params: %+v\n", c.Request.URL.Query())
	
	// Manual parameter extraction to ensure search works
	searchParam := c.Query("search")
	fmt.Printf("DEBUG Customer Handler: Raw search parameter: '%s'\n", searchParam)
	if searchParam != "" {
		params.SearchTerm = searchParam
		fmt.Printf("DEBUG Customer Handler: Search parameter SET to: '%s'\n", searchParam)
	}
	
	// DEBUG: Log params after binding
	fmt.Printf("DEBUG Customer Handler: Final params: SearchTerm='%s'\n", params.SearchTerm)

	customers, err := h.customerRepo.List(params)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error(), "user": user})
		return
	}

	fmt.Printf("DEBUG: Found %d customers with search term '%s'\n", len(customers), params.SearchTerm)

	c.HTML(http.StatusOK, "customers.html", gin.H{
		"title":     "Customers",
		"customers": customers,
		"params":    params,
		"user":      user,
	})
}

func (h *CustomerHandler) NewCustomerForm(c *gin.Context) {
	user, _ := GetCurrentUser(c)
	
	c.HTML(http.StatusOK, "customer_form.html", gin.H{
		"title":    "New Customer",
		"customer": &models.Customer{},
		"user":     user,
	})
}

func (h *CustomerHandler) CreateCustomer(c *gin.Context) {
	companyName := c.PostForm("companyname")
	firstName := c.PostForm("firstname")
	lastName := c.PostForm("lastname")
	email := c.PostForm("email")
	phoneNumber := c.PostForm("phonenumber")
	street := c.PostForm("street")
	houseNumber := c.PostForm("housenumber")
	zip := c.PostForm("zip")
	city := c.PostForm("city")
	federalState := c.PostForm("federalstate")
	country := c.PostForm("country")
	customerType := c.PostForm("customertype")
	notes := c.PostForm("notes")
	
	customer := models.Customer{
		CompanyName:  &companyName,
		FirstName:    &firstName,
		LastName:     &lastName,
		Email:        &email,
		PhoneNumber:  &phoneNumber,
		Street:       &street,
		HouseNumber:  &houseNumber,
		ZIP:          &zip,
		City:         &city,
		FederalState: &federalState,
		Country:      &country,
		CustomerType: &customerType,
		Notes:        &notes,
	}

	if err := h.customerRepo.Create(&customer); err != nil {
		user, _ := GetCurrentUser(c)
		c.HTML(http.StatusInternalServerError, "customer_form.html", gin.H{
			"title":    "New Customer",
			"customer": &customer,
			"error":    err.Error(),
			"user":     user,
		})
		return
	}

	c.Redirect(http.StatusFound, "/customers")
}

func (h *CustomerHandler) GetCustomer(c *gin.Context) {
	user, _ := GetCurrentUser(c)
	
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid customer ID", "user": user})
		return
	}

	customer, err := h.customerRepo.GetByID(uint(id))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Customer not found", "user": user})
		return
	}

	c.HTML(http.StatusOK, "customer_detail.html", gin.H{
		"customer": customer,
		"user":     user,
	})
}

func (h *CustomerHandler) EditCustomerForm(c *gin.Context) {
	user, _ := GetCurrentUser(c)
	
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid customer ID", "user": user})
		return
	}

	customer, err := h.customerRepo.GetByID(uint(id))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Customer not found", "user": user})
		return
	}

	c.HTML(http.StatusOK, "customer_form.html", gin.H{
		"title":    "Edit Customer",
		"customer": customer,
		"user":     user,
	})
}

func (h *CustomerHandler) UpdateCustomer(c *gin.Context) {
	user, _ := GetCurrentUser(c)
	
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid customer ID", "user": user})
		return
	}

	companyName := c.PostForm("companyname")
	firstName := c.PostForm("firstname")
	lastName := c.PostForm("lastname")
	email := c.PostForm("email")
	phoneNumber := c.PostForm("phonenumber")
	street := c.PostForm("street")
	houseNumber := c.PostForm("housenumber")
	zip := c.PostForm("zip")
	city := c.PostForm("city")
	federalState := c.PostForm("federalstate")
	country := c.PostForm("country")
	customerType := c.PostForm("customertype")
	notes := c.PostForm("notes")
	
	customer := models.Customer{
		CustomerID:   uint(id),
		CompanyName:  &companyName,
		FirstName:    &firstName,
		LastName:     &lastName,
		Email:        &email,
		PhoneNumber:  &phoneNumber,
		Street:       &street,
		HouseNumber:  &houseNumber,
		ZIP:          &zip,
		City:         &city,
		FederalState: &federalState,
		Country:      &country,
		CustomerType: &customerType,
		Notes:        &notes,
	}

	if err := h.customerRepo.Update(&customer); err != nil {
		c.HTML(http.StatusInternalServerError, "customer_form.html", gin.H{
			"title":    "Edit Customer",
			"customer": &customer,
			"error":    err.Error(),
			"user":     user,
		})
		return
	}

	c.Redirect(http.StatusFound, "/customers")
}

func (h *CustomerHandler) DeleteCustomer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	if err := h.customerRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Customer deleted successfully"})
}

// API handlers
func (h *CustomerHandler) ListCustomersAPI(c *gin.Context) {
	params := &models.FilterParams{}
	if err := c.ShouldBindQuery(params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customers, err := h.customerRepo.List(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"customers": customers})
}

func (h *CustomerHandler) CreateCustomerAPI(c *gin.Context) {
	var customer models.Customer
	if err := c.ShouldBindJSON(&customer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.customerRepo.Create(&customer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, customer)
}

func (h *CustomerHandler) GetCustomerAPI(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	customer, err := h.customerRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	c.JSON(http.StatusOK, customer)
}

func (h *CustomerHandler) UpdateCustomerAPI(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	var customer models.Customer
	if err := c.ShouldBindJSON(&customer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customer.CustomerID = uint(id)
	if err := h.customerRepo.Update(&customer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, customer)
}

func (h *CustomerHandler) DeleteCustomerAPI(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	if err := h.customerRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Customer deleted successfully"})
}