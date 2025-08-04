package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// DeviceFingerprintGenerator generates device fingerprints for security
type DeviceFingerprintGenerator struct{}

// NewDeviceFingerprintGenerator creates a new device fingerprint generator
func NewDeviceFingerprintGenerator() *DeviceFingerprintGenerator {
	return &DeviceFingerprintGenerator{}
}

// GenerateFingerprint generates a device fingerprint from request headers
func (d *DeviceFingerprintGenerator) GenerateFingerprint(userAgent, acceptLanguage, acceptEncoding, acceptCharset string, ipAddress string) string {
	// Normalize inputs
	userAgent = strings.TrimSpace(userAgent)
	acceptLanguage = strings.TrimSpace(acceptLanguage)
	acceptEncoding = strings.TrimSpace(acceptEncoding)
	acceptCharset = strings.TrimSpace(acceptCharset)

	// Create fingerprint components
	components := []string{
		d.normalizeUserAgent(userAgent),
		d.normalizeAcceptLanguage(acceptLanguage),
		d.normalizeAcceptEncoding(acceptEncoding),
		d.normalizeAcceptCharset(acceptCharset),
		d.normalizeIPAddress(ipAddress),
	}

	// Combine components
	fingerprint := strings.Join(components, "|")

	// Hash the fingerprint for consistency and privacy
	hash := sha256.Sum256([]byte(fingerprint))
	return hex.EncodeToString(hash[:])
}

// GenerateFingerprintFromHeaders generates fingerprint from common HTTP headers
func (d *DeviceFingerprintGenerator) GenerateFingerprintFromHeaders(headers map[string]string, ipAddress string) string {
	userAgent := headers["User-Agent"]
	acceptLanguage := headers["Accept-Language"]
	acceptEncoding := headers["Accept-Encoding"]
	acceptCharset := headers["Accept-Charset"]

	return d.GenerateFingerprint(userAgent, acceptLanguage, acceptEncoding, acceptCharset, ipAddress)
}

// normalizeUserAgent extracts key information from user agent
func (d *DeviceFingerprintGenerator) normalizeUserAgent(userAgent string) string {
	if userAgent == "" {
		return "unknown"
	}

	ua := strings.ToLower(userAgent)

	// Extract browser information
	var browser string
	var os string

	// Browser detection
	if strings.Contains(ua, "chrome") && !strings.Contains(ua, "edg") {
		browser = "chrome"
	} else if strings.Contains(ua, "firefox") {
		browser = "firefox"
	} else if strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome") {
		browser = "safari"
	} else if strings.Contains(ua, "edg") {
		browser = "edge"
	} else if strings.Contains(ua, "opera") || strings.Contains(ua, "opr") {
		browser = "opera"
	} else {
		browser = "other"
	}

	// OS detection
	if strings.Contains(ua, "windows") {
		os = "windows"
	} else if strings.Contains(ua, "macintosh") || strings.Contains(ua, "mac os") {
		os = "macos"
	} else if strings.Contains(ua, "linux") {
		os = "linux"
	} else if strings.Contains(ua, "android") {
		os = "android"
	} else if strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") {
		os = "ios"
	} else {
		os = "other"
	}

	return fmt.Sprintf("%s_%s", browser, os)
}

// normalizeAcceptLanguage normalizes accept language header
func (d *DeviceFingerprintGenerator) normalizeAcceptLanguage(acceptLanguage string) string {
	if acceptLanguage == "" {
		return "unknown"
	}

	// Extract primary language (before comma or semicolon)
	lang := strings.ToLower(acceptLanguage)
	if idx := strings.Index(lang, ","); idx != -1 {
		lang = lang[:idx]
	}
	if idx := strings.Index(lang, ";"); idx != -1 {
		lang = lang[:idx]
	}

	// Extract just the language code (before dash)
	if idx := strings.Index(lang, "-"); idx != -1 {
		lang = lang[:idx]
	}

	return strings.TrimSpace(lang)
}

// normalizeAcceptEncoding normalizes accept encoding header
func (d *DeviceFingerprintGenerator) normalizeAcceptEncoding(acceptEncoding string) string {
	if acceptEncoding == "" {
		return "unknown"
	}

	encoding := strings.ToLower(acceptEncoding)

	// Check for common encodings
	var encodings []string
	if strings.Contains(encoding, "gzip") {
		encodings = append(encodings, "gzip")
	}
	if strings.Contains(encoding, "deflate") {
		encodings = append(encodings, "deflate")
	}
	if strings.Contains(encoding, "br") {
		encodings = append(encodings, "br")
	}

	if len(encodings) == 0 {
		return "none"
	}

	return strings.Join(encodings, "_")
}

// normalizeAcceptCharset normalizes accept charset header
func (d *DeviceFingerprintGenerator) normalizeAcceptCharset(acceptCharset string) string {
	if acceptCharset == "" {
		return "utf8" // Default assumption
	}

	charset := strings.ToLower(acceptCharset)

	if strings.Contains(charset, "utf-8") {
		return "utf8"
	}
	if strings.Contains(charset, "iso-8859") {
		return "iso"
	}

	return "other"
}

// normalizeIPAddress normalizes IP address for fingerprinting
func (d *DeviceFingerprintGenerator) normalizeIPAddress(ipAddress string) string {
	if ipAddress == "" {
		return "unknown"
	}

	// For IPv4, use the first 3 octets for privacy
	if strings.Contains(ipAddress, ".") && !strings.Contains(ipAddress, ":") {
		parts := strings.Split(ipAddress, ".")
		if len(parts) >= 3 {
			return fmt.Sprintf("%s.%s.%s.x", parts[0], parts[1], parts[2])
		}
	}

	// For IPv6, use the first 4 groups for privacy
	if strings.Contains(ipAddress, ":") {
		parts := strings.Split(ipAddress, ":")
		if len(parts) >= 4 {
			return fmt.Sprintf("%s:%s:%s:%s::x", parts[0], parts[1], parts[2], parts[3])
		}
	}

	return "unknown"
}

// ValidateFingerprint validates that a fingerprint is properly formatted
func (d *DeviceFingerprintGenerator) ValidateFingerprint(fingerprint string) bool {
	// Should be a 64-character hex string (SHA256)
	if len(fingerprint) != 64 {
		return false
	}

	// Check if all characters are valid hex
	for _, char := range fingerprint {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			return false
		}
	}

	return true
}

// ExtractFingerprintFromFiberContext extracts device fingerprint from Fiber context
func (d *DeviceFingerprintGenerator) ExtractFingerprintFromFiberContext(c interface{}) string {
	// This would be implemented based on your Fiber context
	// For now, return a placeholder
	return d.GenerateFingerprint("", "", "", "", "")
}
