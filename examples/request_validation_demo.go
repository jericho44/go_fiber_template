package main

import (
	"fmt"
	"log"
	"strings"

	"go-fiber-template/internal/config"
	"go-fiber-template/internal/middleware"
	"go-fiber-template/internal/utils"

	"github.com/gofiber/fiber/v2"
)

// RequestValidationDemo demonstrates the request validation middleware functionality
func main() {
	fmt.Println("🔒 Request Validation Middleware Demo")
	fmt.Println("=====================================")

	// Create a test configuration
	cfg := &config.Config{
		Request: config.RequestConfig{
			MaxBodySize:        1024, // 1KB for demo
			MaxFileSize:        2048, // 2KB for demo
			MaxMultipartSize:   4096, // 4KB for demo
			AllowedMimeTypes:   []string{"text/plain", "image/jpeg", "image/png"},
			EnableSanitization: true,
		},
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Create request validation middleware
	requestValidation := middleware.NewRequestValidationMiddleware(cfg)

	// Demo endpoints

	// 1. Request size limit demo
	app.Post("/demo/size-limit", requestValidation.RequestSizeLimit(), func(c *fiber.Ctx) error {
		body := c.Body()
		return c.JSON(fiber.Map{
			"message":   "Request accepted",
			"body_size": len(body),
			"max_size":  cfg.Request.MaxBodySize,
		})
	})

	// 2. File upload size limit demo
	app.Post("/demo/file-upload", requestValidation.FileUploadSizeLimit(), func(c *fiber.Ctx) error {
		form, err := c.MultipartForm()
		if err != nil {
			return utils.BadRequestResponse(c, "Failed to parse form", nil)
		}

		files := make([]fiber.Map, 0)
		for fieldName, fileHeaders := range form.File {
			for _, fileHeader := range fileHeaders {
				files = append(files, fiber.Map{
					"field":    fieldName,
					"filename": fileHeader.Filename,
					"size":     fileHeader.Size,
				})
			}
		}

		return c.JSON(fiber.Map{
			"message":       "Files uploaded successfully",
			"files":         files,
			"max_file_size": cfg.Request.MaxFileSize,
		})
	})

	// 3. Content type validation demo
	app.Post("/demo/content-type", requestValidation.ContentTypeValidation(), func(c *fiber.Ctx) error {
		form, err := c.MultipartForm()
		if err != nil {
			return utils.BadRequestResponse(c, "Failed to parse form", nil)
		}

		return c.JSON(fiber.Map{
			"message":            "Content types validated successfully",
			"allowed_mime_types": cfg.Request.AllowedMimeTypes,
			"files_count":        len(form.File),
		})
	})

	// 4. Request sanitization demo
	app.Get("/demo/sanitization", requestValidation.RequestSanitization(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message":              "Query parameters sanitized",
			"sanitized_params":     c.Queries(),
			"sanitization_enabled": cfg.Request.EnableSanitization,
		})
	})

	// 5. Combined validation demo
	app.Post("/demo/combined", requestValidation.CombinedRequestValidation(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "All validations passed successfully",
			"config": fiber.Map{
				"max_body_size":        cfg.Request.MaxBodySize,
				"max_file_size":        cfg.Request.MaxFileSize,
				"max_multipart_size":   cfg.Request.MaxMultipartSize,
				"allowed_mime_types":   cfg.Request.AllowedMimeTypes,
				"sanitization_enabled": cfg.Request.EnableSanitization,
			},
		})
	})

	// Demo information endpoint
	app.Get("/demo", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"title":       "Request Validation Middleware Demo",
			"description": "This demo showcases the request validation middleware features",
			"endpoints": fiber.Map{
				"size_limit":   "POST /demo/size-limit - Test request body size limits",
				"file_upload":  "POST /demo/file-upload - Test file upload size limits",
				"content_type": "POST /demo/content-type - Test MIME type validation",
				"sanitization": "GET /demo/sanitization?param=<script>alert('xss')</script> - Test query sanitization",
				"combined":     "POST /demo/combined - Test all validations together",
			},
			"test_scenarios": fiber.Map{
				"large_request": "Send a request larger than 1KB to /demo/size-limit",
				"large_file":    "Upload a file larger than 2KB to /demo/file-upload",
				"invalid_mime":  "Upload a file with unsupported MIME type to /demo/content-type",
				"xss_attempt":   "Send XSS payload in query params to /demo/sanitization",
			},
			"configuration": fiber.Map{
				"max_body_size":        cfg.Request.MaxBodySize,
				"max_file_size":        cfg.Request.MaxFileSize,
				"max_multipart_size":   cfg.Request.MaxMultipartSize,
				"allowed_mime_types":   cfg.Request.AllowedMimeTypes,
				"sanitization_enabled": cfg.Request.EnableSanitization,
			},
		})
	})

	// Test utilities
	fmt.Println("\n📋 Demo Endpoints:")
	fmt.Println("• GET  /demo - Demo information and test scenarios")
	fmt.Println("• POST /demo/size-limit - Test request body size limits")
	fmt.Println("• POST /demo/file-upload - Test file upload size limits")
	fmt.Println("• POST /demo/content-type - Test MIME type validation")
	fmt.Println("• GET  /demo/sanitization - Test query parameter sanitization")
	fmt.Println("• POST /demo/combined - Test all validations together")

	fmt.Println("\n🧪 Test Scenarios:")
	fmt.Println("1. Large Request Test:")
	fmt.Printf("   curl -X POST http://localhost:3000/demo/size-limit -d '%s'\n", strings.Repeat("a", 2000))

	fmt.Println("\n2. XSS Sanitization Test:")
	fmt.Println("   curl 'http://localhost:3000/demo/sanitization?test=<script>alert(\"xss\")</script>'")

	fmt.Println("\n3. File Upload Test:")
	fmt.Println("   curl -X POST -F 'file=@large_file.txt' http://localhost:3000/demo/file-upload")

	fmt.Println("\n🚀 Starting demo server on :3000...")
	log.Fatal(app.Listen(":3000"))
}
