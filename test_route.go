package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	
	scan := r.Group("/scan")
	{
		scan.GET("/select", func(c *gin.Context) {
			c.String(200, "This is SELECT page")
		})
		scan.GET("/:jobId", func(c *gin.Context) {
			jobId := c.Param("jobId")
			c.String(200, "This is JOB page, jobId: %s", jobId)
		})
	}
	
	fmt.Println("Test server running on :9001")
	fmt.Println("Test URLs:")
	fmt.Println("  http://localhost:9001/scan/select")
	fmt.Println("  http://localhost:9001/scan/1004")
	
	r.Run(":9001")
}