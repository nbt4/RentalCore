package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"io"
	"regexp"
)

func main() {
	fmt.Println("=== TESTING CASE MANAGEMENT PAGE ===")
	
	// Test the demo route first (no auth required)
	fmt.Println("\n1. Testing demo route (no auth):")
	resp, err := http.Get("http://localhost:9000/demo/case-management-real")
	if err != nil {
		log.Printf("Error accessing demo route: %v", err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		fmt.Printf("❌ Demo route failed with status: %d\n", resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Response: %s\n", string(body)[:200])
		return
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		return
	}
	
	html := string(body)
	
	// Check if devices are in the JavaScript array
	fmt.Println("\n2. Checking for device data in JavaScript:")
	
	// Look for the availableDevices array
	deviceArrayRegex := regexp.MustCompile(`const availableDevices = \[(.*?)\];`)
	matches := deviceArrayRegex.FindStringSubmatch(html)
	
	if len(matches) > 1 {
		deviceData := matches[1]
		if strings.TrimSpace(deviceData) == "" {
			fmt.Println("❌ FOUND THE PROBLEM: availableDevices array is EMPTY!")
			fmt.Println("   The array exists but contains no device data.")
		} else {
			fmt.Println("✅ Device data found in JavaScript array!")
			
			// Count devices by looking for deviceID entries
			deviceIdRegex := regexp.MustCompile(`deviceID:\s*"[^"]+`)
			deviceMatches := deviceIdRegex.FindAllString(deviceData, -1)
			fmt.Printf("   Found %d devices in JavaScript array\n", len(deviceMatches))
			
			// Show first few devices
			if len(deviceMatches) > 0 {
				fmt.Println("   Sample devices:")
				for i, match := range deviceMatches {
					if i >= 3 { break }
					fmt.Printf("     %d. %s\n", i+1, match)
				}
			}
		}
	} else {
		fmt.Println("❌ availableDevices JavaScript array NOT FOUND!")
		
		// Check if the template is rendering the devices section at all
		if strings.Contains(html, "{{range .devices}}") {
			fmt.Println("   Template contains device loop syntax (not compiled)")
		} else if strings.Contains(html, "const availableDevices") {
			fmt.Println("   Found availableDevices constant but couldn't parse it")
		} else {
			fmt.Println("   No device-related JavaScript found at all")
		}
	}
	
	// Check for error indicators
	fmt.Println("\n3. Checking for error indicators:")
	
	if strings.Contains(html, "No Available Devices") {
		fmt.Println("❌ Found 'No Available Devices' message")
	}
	
	if strings.Contains(html, "Failed to load devices") {
		fmt.Println("❌ Found 'Failed to load devices' error")
	}
	
	if strings.Contains(html, "error") || strings.Contains(html, "Error") {
		fmt.Println("⚠️  Page contains error-related text")
		// Extract error context
		errorRegex := regexp.MustCompile(`(?i)(error[^<\n]*|Error[^<\n]*)`)
		errorMatches := errorRegex.FindAllString(html, 3)
		for _, match := range errorMatches {
			fmt.Printf("   Error text: %s\n", strings.TrimSpace(match))
		}
	}
	
	// Check if the template variables are being passed correctly
	fmt.Println("\n4. Template variable analysis:")
	
	// Look for evidence that .devices is being iterated
	if strings.Contains(html, "deviceID:") {
		fmt.Println("✅ Found deviceID references in page")
	} else {
		fmt.Println("❌ No deviceID references found")
	}
	
	// Look for product/category data
	if strings.Contains(html, "product") && strings.Contains(html, "category") {
		fmt.Println("✅ Found product/category data structure")
	} else {
		fmt.Println("❌ No product/category structure found")
	}
	
	fmt.Printf("\n5. Response size: %d bytes\n", len(html))
	fmt.Printf("   Status code: %d\n", resp.StatusCode)
	
	// Save a snippet for manual inspection
	fmt.Println("\n6. JavaScript array section (first 500 chars):")
	arrayStart := strings.Index(html, "const availableDevices")
	if arrayStart != -1 {
		arrayEnd := strings.Index(html[arrayStart:], "];")
		if arrayEnd != -1 {
			arrayEnd += arrayStart + 2
			arraySection := html[arrayStart:arrayEnd]
			if len(arraySection) > 500 {
				arraySection = arraySection[:500] + "..."
			}
			fmt.Printf("   %s\n", arraySection)
		}
	}
	
	fmt.Println("\n=== TEST COMPLETE ===")
}