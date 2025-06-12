package main

import (
	"fmt"
	"log"
	"net/http"
	"io"
	"regexp"
	"strings"
)

func main() {
	fmt.Println("=== DEBUGGING FRONTEND DEVICE RENDERING ===")
	
	// Get the case management page
	resp, err := http.Get("http://localhost:9000/demo/case-management-real")
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		return
	}
	
	html := string(body)
	
	// Extract the complete availableDevices array
	fmt.Println("1. Extracting device data...")
	
	// Find the start and end of the devices array
	startPattern := "const availableDevices = ["
	startIndex := strings.Index(html, startPattern)
	if startIndex == -1 {
		fmt.Println("❌ Could not find availableDevices array")
		return
	}
	
	// Find the closing bracket and semicolon
	endPattern := "];"
	searchFrom := startIndex + len(startPattern)
	endIndex := strings.Index(html[searchFrom:], endPattern)
	if endIndex == -1 {
		fmt.Println("❌ Could not find end of availableDevices array")
		return
	}
	
	arrayContent := html[startIndex+len(startPattern) : searchFrom+endIndex]
	
	// Count devices by looking for deviceID entries
	deviceIdPattern := regexp.MustCompile(`deviceID:\s*"([^"]+)"`)
	deviceMatches := deviceIdPattern.FindAllStringSubmatch(arrayContent, -1)
	
	fmt.Printf("✅ Found %d devices in JavaScript array\n", len(deviceMatches))
	
	// Analyze the device data
	fmt.Println("\n2. Device data analysis:")
	statusPattern := regexp.MustCompile(`status:\s*"([^"]+)"`)
	statusMatches := statusPattern.FindAllStringSubmatch(arrayContent, -1)
	
	freeCount := 0
	for _, match := range statusMatches {
		if len(match) > 1 && match[1] == "free" {
			freeCount++
		}
	}
	
	fmt.Printf("   Devices with status='free': %d\n", freeCount)
	fmt.Printf("   Total devices in array: %d\n", len(deviceMatches))
	
	// Show sample devices
	fmt.Println("\n3. Sample devices:")
	for i, match := range deviceMatches {
		if i >= 5 { break }
		fmt.Printf("   %d. Device ID: %s\n", i+1, match[1])
	}
	
	// Check for product data
	fmt.Println("\n4. Product data check:")
	productNamePattern := regexp.MustCompile(`name:\s*"([^"]+)"`)
	productMatches := productNamePattern.FindAllStringSubmatch(arrayContent, -1)
	
	validProducts := 0
	for _, match := range productMatches {
		if len(match) > 1 && match[1] != "" {
			validProducts++
		}
	}
	
	fmt.Printf("   Products with names: %d\n", validProducts)
	
	// Check for category data
	categoryPattern := regexp.MustCompile(`category:\s*\{\s*name:\s*"([^"]+)"`)
	categoryMatches := categoryPattern.FindAllStringSubmatch(arrayContent, -1)
	
	fmt.Printf("   Categories found: %d\n", len(categoryMatches))
	
	categoryMap := make(map[string]int)
	for _, match := range categoryMatches {
		if len(match) > 1 {
			categoryMap[match[1]]++
		}
	}
	
	fmt.Println("   Category breakdown:")
	for cat, count := range categoryMap {
		fmt.Printf("     %s: %d devices\n", cat, count)
	}
	
	// Check JavaScript execution issues
	fmt.Println("\n5. Checking for JavaScript issues:")
	
	// Look for syntax errors or malformed JSON
	if strings.Count(arrayContent, "{") != strings.Count(arrayContent, "}") {
		fmt.Println("❌ Mismatched braces in device data")
	} else {
		fmt.Println("✅ Braces are balanced")
	}
	
	if strings.Count(arrayContent, "[") != strings.Count(arrayContent, "]") {
		fmt.Println("❌ Mismatched brackets in device data")
	} else {
		fmt.Println("✅ Brackets are balanced")
	}
	
	// Check for trailing commas that might break parsing
	if strings.Contains(arrayContent, ",\n            \n        ];") || strings.Contains(arrayContent, ",\n        ];") {
		fmt.Println("⚠️  Potential trailing comma issue found")
	}
	
	// Look for the loadDevicesByCategory function call
	fmt.Println("\n6. Checking function calls:")
	if strings.Contains(html, "loadDevicesByCategory()") {
		fmt.Println("✅ loadDevicesByCategory() function is called")
	} else {
		fmt.Println("❌ loadDevicesByCategory() function is NOT called")
	}
	
	// Check for the device-categories div
	if strings.Contains(html, `id="device-categories"`) {
		fmt.Println("✅ device-categories container exists")
	} else {
		fmt.Println("❌ device-categories container is missing")
	}
	
	// Save the device array to a file for inspection
	fmt.Printf("\n7. Device array length: %d characters\n", len(arrayContent))
	
	// Show a cleaned sample of the first device
	fmt.Println("\n8. First device structure:")
	firstDevicePattern := regexp.MustCompile(`\{\s*deviceID:\s*"[^"]+",[\s\S]*?\},`)
	firstMatch := firstDevicePattern.FindString(arrayContent)
	if firstMatch != "" {
		lines := strings.Split(firstMatch, "\n")
		for i, line := range lines {
			if i >= 15 { break } // Show first 15 lines
			fmt.Printf("   %s\n", strings.TrimSpace(line))
		}
	}
	
	fmt.Println("\n=== FRONTEND DEBUG COMPLETE ===")
}