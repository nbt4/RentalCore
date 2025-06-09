package main

import (
	"html/template"
	"log"
	"os"
)

func main() {
	// Load templates the same way the main app does with functions
	tmpl := template.New("").Funcs(template.FuncMap{
		"deref": func(p *uint) uint {
			if p != nil {
				return *p
			}
			return 0
		},
		"derefString": func(p *string) string {
			if p != nil {
				return *p
			}
			return ""
		},
	})
	
	tmpl, err := tmpl.ParseGlob("web/templates/*")
	if err != nil {
		log.Fatal("Error parsing templates:", err)
	}

	// List all templates
	log.Println("Available templates:")
	for _, t := range tmpl.Templates() {
		log.Printf("  - %s", t.Name())
	}

	// Test data similar to what the handler passes
	data := map[string]interface{}{
		"title": "Scanning Job #1004",
		"job": map[string]interface{}{
			"JobID": 1004,
			"Description": "Test job",
		},
		"assignedDevices": []interface{}{},
	}

	// Try to execute scan_job.html
	log.Println("\nTrying to execute scan_job.html...")
	err = tmpl.ExecuteTemplate(os.Stdout, "scan_job.html", data)
	if err != nil {
		log.Fatal("Error executing scan_job.html:", err)
	}
}