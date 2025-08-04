package middleware

import (
	"fmt"
	"html"
	"io"
	"mime"
	"net/http"
	"regexp"
	"strings"

	"go-fiber-template/internal/config"
	"go-fiber-template/internal/utils"

	"github.com/gofiber/fiber/v2"
)

// RequestValidationMiddleware handles request size limits and validation
type RequestValidationMiddleware struct {
	config *config.Config
}

// NewRequestValidationMiddleware creates a new request validation middleware
func NewRequestValidationMiddleware(cfg *config.Config) *RequestValidationMiddleware {
	return &RequestValidationMiddleware{
		config: cfg,
	}
}

// RequestSizeLimit creates middleware that enforces request body size limits
func (r *RequestValidationMiddleware) RequestSizeLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get content length from header
		contentLength := int64(c.Request().Header.ContentLength())

		// Check if content length exceeds maximum body size
		if contentLength > r.config.Request.MaxBodySize {
			return utils.BadRequestResponse(c,
				fmt.Sprintf("Request body too large. Maximum allowed size is %d bytes", r.config.Request.MaxBodySize),
				map[string]string{
					"max_size":      fmt.Sprintf("%d", r.config.Request.MaxBodySize),
					"received_size": fmt.Sprintf("%d", contentLength),
				})
		}

		// For multipart forms, check against multipart size limit
		contentType := c.Get("Content-Type")
		if strings.HasPrefix(contentType, "multipart/") {
			if contentLength > r.config.Request.MaxMultipartSize {
				return utils.BadRequestResponse(c,
					fmt.Sprintf("Multipart request too large. Maximum allowed size is %d bytes", r.config.Request.MaxMultipartSize),
					map[string]string{
						"max_multipart_size": fmt.Sprintf("%d", r.config.Request.MaxMultipartSize),
						"received_size":      fmt.Sprintf("%d", contentLength),
					})
			}
		}

		return c.Next()
	}
}

// FileUploadSizeLimit creates middleware that enforces file upload size limits
func (r *RequestValidationMiddleware) FileUploadSizeLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Only check multipart forms
		contentType := c.Get("Content-Type")
		if !strings.HasPrefix(contentType, "multipart/") {
			return c.Next()
		}

		// Parse multipart form with size limit
		form, err := c.MultipartForm()
		if err != nil {
			return utils.BadRequestResponse(c, "Failed to parse multipart form", map[string]string{
				"error": err.Error(),
			})
		}

		// Check each file size
		for fieldName, files := range form.File {
			for _, fileHeader := range files {
				if fileHeader.Size > r.config.Request.MaxFileSize {
					return utils.BadRequestResponse(c,
						fmt.Sprintf("File '%s' in field '%s' is too large. Maximum allowed size is %d bytes",
							fileHeader.Filename, fieldName, r.config.Request.MaxFileSize),
						map[string]string{
							"field_name":    fieldName,
							"filename":      fileHeader.Filename,
							"file_size":     fmt.Sprintf("%d", fileHeader.Size),
							"max_file_size": fmt.Sprintf("%d", r.config.Request.MaxFileSize),
						})
				}
			}
		}

		return c.Next()
	}
}

// ContentTypeValidation creates middleware that validates content types for file uploads
func (r *RequestValidationMiddleware) ContentTypeValidation() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Only check multipart forms
		contentType := c.Get("Content-Type")
		if !strings.HasPrefix(contentType, "multipart/") {
			return c.Next()
		}

		// Parse multipart form
		form, err := c.MultipartForm()
		if err != nil {
			return utils.BadRequestResponse(c, "Failed to parse multipart form", map[string]string{
				"error": err.Error(),
			})
		}

		// Check each file's MIME type
		for fieldName, files := range form.File {
			for _, fileHeader := range files {
				// Open the file to detect its MIME type
				file, err := fileHeader.Open()
				if err != nil {
					return utils.InternalServerErrorResponse(c, "Failed to open uploaded file")
				}
				defer file.Close()

				// Read first 512 bytes to detect MIME type
				buffer := make([]byte, 512)
				n, err := file.Read(buffer)
				if err != nil && err != io.EOF {
					return utils.InternalServerErrorResponse(c, "Failed to read uploaded file")
				}

				// Detect MIME type
				detectedMimeType := http.DetectContentType(buffer[:n])

				// Also check the MIME type from the file header
				headerMimeType := fileHeader.Header.Get("Content-Type")

				// Use detected MIME type if header is not present or generic
				mimeTypeToCheck := detectedMimeType
				if headerMimeType != "" && headerMimeType != "application/octet-stream" {
					mimeTypeToCheck = headerMimeType
				}

				// Check if MIME type is allowed
				if !r.isMimeTypeAllowed(mimeTypeToCheck) {
					return utils.BadRequestResponse(c,
						fmt.Sprintf("File '%s' in field '%s' has unsupported MIME type '%s'",
							fileHeader.Filename, fieldName, mimeTypeToCheck),
						map[string]string{
							"field_name":         fieldName,
							"filename":           fileHeader.Filename,
							"detected_mime_type": detectedMimeType,
							"header_mime_type":   headerMimeType,
							"allowed_mime_types": strings.Join(r.config.Request.AllowedMimeTypes, ","),
						})
				}
			}
		}

		return c.Next()
	}
}

// RequestSanitization creates middleware that sanitizes request data
func (r *RequestValidationMiddleware) RequestSanitization() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip if sanitization is disabled
		if !r.config.Request.EnableSanitization {
			return c.Next()
		}

		// Sanitize query parameters
		r.sanitizeQueryParams(c)

		// Sanitize form data for non-multipart forms
		contentType := c.Get("Content-Type")
		if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
			r.sanitizeFormData(c)
		}

		// For JSON requests, we'll sanitize after parsing in the handler
		// This middleware focuses on URL and form data sanitization

		return c.Next()
	}
}

// CombinedRequestValidation creates a combined middleware that applies all request validations
func (r *RequestValidationMiddleware) CombinedRequestValidation() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Apply request size limit
		if err := r.RequestSizeLimit()(c); err != nil {
			return err
		}

		// Apply file upload size limit
		if err := r.FileUploadSizeLimit()(c); err != nil {
			return err
		}

		// Apply content type validation
		if err := r.ContentTypeValidation()(c); err != nil {
			return err
		}

		// Apply request sanitization
		if err := r.RequestSanitization()(c); err != nil {
			return err
		}

		return c.Next()
	}
}

// Helper methods

// isMimeTypeAllowed checks if a MIME type is in the allowed list
func (r *RequestValidationMiddleware) isMimeTypeAllowed(mimeType string) bool {
	// Extract base MIME type (remove parameters)
	baseMimeType, _, err := mime.ParseMediaType(mimeType)
	if err != nil {
		baseMimeType = mimeType
	}

	for _, allowedType := range r.config.Request.AllowedMimeTypes {
		if strings.EqualFold(baseMimeType, allowedType) {
			return true
		}
		// Support wildcard matching (e.g., "image/*")
		if strings.HasSuffix(allowedType, "/*") {
			prefix := strings.TrimSuffix(allowedType, "/*")
			if strings.HasPrefix(baseMimeType, prefix+"/") {
				return true
			}
		}
	}
	return false
}

// sanitizeQueryParams sanitizes query parameters
func (r *RequestValidationMiddleware) sanitizeQueryParams(c *fiber.Ctx) {
	// Get all query parameters
	queries := c.Queries()

	// Sanitize each parameter
	for key, value := range queries {
		sanitized := r.sanitizeString(value)
		if sanitized != value {
			// Update the query parameter with sanitized value
			c.Request().URI().QueryArgs().Set(key, sanitized)
		}
	}
}

// sanitizeFormData sanitizes form data
func (r *RequestValidationMiddleware) sanitizeFormData(c *fiber.Ctx) {
	// Sanitize each form field
	c.Request().PostArgs().VisitAll(func(key, value []byte) {
		sanitized := r.sanitizeString(string(value))
		if sanitized != string(value) {
			c.Request().PostArgs().Set(string(key), sanitized)
		}
	})
}

// sanitizeString sanitizes a string by removing/escaping potentially dangerous content
func (r *RequestValidationMiddleware) sanitizeString(input string) string {
	// HTML escape
	sanitized := html.EscapeString(input)

	// Remove null bytes
	sanitized = strings.ReplaceAll(sanitized, "\x00", "")

	// Remove or escape common injection patterns
	sanitized = r.removeScriptTags(sanitized)
	sanitized = r.removeSQLInjectionPatterns(sanitized)

	// Trim whitespace
	sanitized = strings.TrimSpace(sanitized)

	return sanitized
}

// removeScriptTags removes script tags and their content
func (r *RequestValidationMiddleware) removeScriptTags(input string) string {
	// Remove script tags (case insensitive)
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	return scriptRegex.ReplaceAllString(input, "")
}

// removeSQLInjectionPatterns removes common SQL injection patterns
func (r *RequestValidationMiddleware) removeSQLInjectionPatterns(input string) string {
	// Common SQL injection patterns
	patterns := []string{
		`(?i)\bunion\s+select\b`,
		`(?i)\bdrop\s+table\b`,
		`(?i)\bdelete\s+from\b`,
		`(?i)\binsert\s+into\b`,
		`(?i)\bupdate\s+set\b`,
		`(?i)--`,
		`(?i)/\*.*?\*/`,
		`(?i)\bor\s+1\s*=\s*1\b`,
		`(?i)\band\s+1\s*=\s*1\b`,
	}

	result := input
	for _, pattern := range patterns {
		regex := regexp.MustCompile(pattern)
		result = regex.ReplaceAllString(result, "")
	}

	return result
}

// Legacy functions for backward compatibility

// RequestSizeLimitMiddleware creates a simple request size limit middleware
func RequestSizeLimitMiddleware(maxSize int64) fiber.Handler {
	return func(c *fiber.Ctx) error {
		contentLength := int64(c.Request().Header.ContentLength())
		if contentLength > maxSize {
			return utils.BadRequestResponse(c,
				fmt.Sprintf("Request body too large. Maximum allowed size is %d bytes", maxSize),
				map[string]string{
					"max_size":      fmt.Sprintf("%d", maxSize),
					"received_size": fmt.Sprintf("%d", contentLength),
				})
		}
		return c.Next()
	}
}

// FileUploadLimitMiddleware creates a simple file upload limit middleware
func FileUploadLimitMiddleware(maxFileSize int64) fiber.Handler {
	return func(c *fiber.Ctx) error {
		contentType := c.Get("Content-Type")
		if !strings.HasPrefix(contentType, "multipart/") {
			return c.Next()
		}

		form, err := c.MultipartForm()
		if err != nil {
			return utils.BadRequestResponse(c, "Failed to parse multipart form", nil)
		}

		for _, files := range form.File {
			for _, fileHeader := range files {
				if fileHeader.Size > maxFileSize {
					return utils.BadRequestResponse(c,
						fmt.Sprintf("File too large. Maximum allowed size is %d bytes", maxFileSize),
						map[string]string{
							"filename":      fileHeader.Filename,
							"file_size":     fmt.Sprintf("%d", fileHeader.Size),
							"max_file_size": fmt.Sprintf("%d", maxFileSize),
						})
				}
			}
		}

		return c.Next()
	}
}

// ContentTypeValidationMiddleware creates a simple content type validation middleware
func ContentTypeValidationMiddleware(allowedTypes []string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		contentType := c.Get("Content-Type")
		if !strings.HasPrefix(contentType, "multipart/") {
			return c.Next()
		}

		form, err := c.MultipartForm()
		if err != nil {
			return utils.BadRequestResponse(c, "Failed to parse multipart form", nil)
		}

		for _, files := range form.File {
			for _, fileHeader := range files {
				file, err := fileHeader.Open()
				if err != nil {
					continue
				}
				defer file.Close()

				buffer := make([]byte, 512)
				n, _ := file.Read(buffer)
				detectedType := http.DetectContentType(buffer[:n])

				allowed := false
				for _, allowedType := range allowedTypes {
					if strings.HasPrefix(detectedType, allowedType) {
						allowed = true
						break
					}
				}

				if !allowed {
					return utils.BadRequestResponse(c,
						fmt.Sprintf("File type not allowed: %s", detectedType),
						map[string]string{
							"filename":      fileHeader.Filename,
							"detected_type": detectedType,
							"allowed_types": strings.Join(allowedTypes, ","),
						})
				}
			}
		}

		return c.Next()
	}
}
