package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"go-barcode-webapp/internal/config"
	"go-barcode-webapp/internal/repository"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config.production.direct.json")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Connect to database
	db, err := repository.NewDatabase(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Create job repository
	jobRepo := repository.NewJobRepository(db)

	// Get job data like the handler does
	job, err := jobRepo.GetByID(1004)
	if err != nil {
		log.Fatal("Failed to get job:", err)
	}

	assignedDevices, err := jobRepo.GetJobDevices(1004)
	if err != nil {
		log.Fatal("Failed to get job devices:", err)
	}

	// Test template data
	data := map[string]interface{}{
		"title":           "Scanning Job #1004",
		"job":             job,
		"assignedDevices": assignedDevices,
	}

	// Try to render a simple test template
	tmplText := `Job ID: {{.job.JobID}}
Customer: {{.job.Customer.GetDisplayName}}
Description: {{if .job.Description}}{{.job.Description}}{{else}}No description{{end}}
Assigned devices: {{len .assignedDevices}}`

	tmpl, err := template.New("test").Parse(tmplText)
	if err != nil {
		log.Fatal("Template parse error:", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		log.Fatal("Template execution error:", err)
	}

	fmt.Println("✅ Template test successful:")
	fmt.Println(buf.String())
}