package utils

import "time"

// GetCurrentTimestamp returns the current timestamp in ISO 8601 format
func GetCurrentTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// FormatTimestamp formats a time.Time to ISO 8601 string
func FormatTimestamp(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

// ParseTimestamp parses an ISO 8601 timestamp string
func ParseTimestamp(timestamp string) (time.Time, error) {
	return time.Parse(time.RFC3339, timestamp)
}
