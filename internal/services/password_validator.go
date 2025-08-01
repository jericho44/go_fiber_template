package services

import (
	"context"
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"go-fiber-template/internal/config"
	"go-fiber-template/internal/models"
	"go-fiber-template/internal/repositories"
)

// PasswordValidator defines the interface for password validation functionality
type PasswordValidator interface {
	// ValidateComplexity validates password complexity against policy
	ValidateComplexity(password string) *PasswordValidationResult
	// CheckCommonPasswords checks if password is in common passwords list
	CheckCommonPasswords(password string) error
	// ValidatePasswordHistory checks if password was used recently
	ValidatePasswordHistory(ctx context.Context, userID uint, password string) error
	// AddPasswordToHistory adds a password to the user's password history
	AddPasswordToHistory(ctx context.Context, userID uint, password string) error
	// GenerateSecurePassword generates a secure password
	GenerateSecurePassword() string
	// GetPasswordStrength returns password strength score (0-100)
	GetPasswordStrength(password string) *PasswordStrengthResult
	// ValidatePasswordPolicy validates password against all policies
	ValidatePasswordPolicy(ctx context.Context, userID uint, password string) *PasswordValidationResult
}

// PasswordPolicy defines password complexity requirements
type PasswordPolicy struct {
	MinLength        int           `json:"min_length"`
	MaxLength        int           `json:"max_length"`
	RequireUppercase bool          `json:"require_uppercase"`
	RequireLowercase bool          `json:"require_lowercase"`
	RequireNumbers   bool          `json:"require_numbers"`
	RequireSymbols   bool          `json:"require_symbols"`
	MaxAge           time.Duration `json:"max_age"`
	PreventReuse     int           `json:"prevent_reuse"` // Number of previous passwords to check
	MinUniqueChars   int           `json:"min_unique_chars"`
	MaxRepeatedChars int           `json:"max_repeated_chars"`
	ForbiddenWords   []string      `json:"forbidden_words"`
}

// PasswordValidationResult contains the result of password validation
type PasswordValidationResult struct {
	Valid        bool                    `json:"valid"`
	Score        int                     `json:"score"` // 0-100
	Errors       []string                `json:"errors"`
	Warnings     []string                `json:"warnings"`
	Suggestions  []string                `json:"suggestions"`
	Requirements map[string]bool         `json:"requirements"`
	Strength     *PasswordStrengthResult `json:"strength"`
}

// PasswordStrengthResult contains password strength analysis
type PasswordStrengthResult struct {
	Score       int     `json:"score"`         // 0-100
	Level       string  `json:"level"`         // weak, fair, good, strong, very_strong
	Entropy     float64 `json:"entropy"`       // Bits of entropy
	TimeToCrack string  `json:"time_to_crack"` // Estimated time to crack
}

// DefaultPasswordValidator implements PasswordValidator
type DefaultPasswordValidator struct {
	policy          *PasswordPolicy
	commonPasswords map[string]bool
	historyRepo     repositories.PasswordHistoryRepository
	config          *config.Config
}

// NewDefaultPasswordValidator creates a new password validator
func NewDefaultPasswordValidator(cfg *config.Config, historyRepo repositories.PasswordHistoryRepository) *DefaultPasswordValidator {
	policy := &PasswordPolicy{
		MinLength:        8,
		MaxLength:        128,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireNumbers:   true,
		RequireSymbols:   true,
		MaxAge:           90 * 24 * time.Hour, // 90 days
		PreventReuse:     5,                   // Last 5 passwords
		MinUniqueChars:   6,
		MaxRepeatedChars: 3,
		ForbiddenWords:   []string{"password", "admin", "user", "login", "welcome", "123456", "qwerty"},
	}

	validator := &DefaultPasswordValidator{
		policy:          policy,
		commonPasswords: make(map[string]bool),
		historyRepo:     historyRepo,
		config:          cfg,
	}

	// Load common passwords
	validator.loadCommonPasswords()

	return validator
}

// ValidateComplexity validates password complexity against policy
func (v *DefaultPasswordValidator) ValidateComplexity(password string) *PasswordValidationResult {
	result := &PasswordValidationResult{
		Valid:        true,
		Errors:       []string{},
		Warnings:     []string{},
		Suggestions:  []string{},
		Requirements: make(map[string]bool),
	}

	// Check length
	if len(password) < v.policy.MinLength {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Password must be at least %d characters long", v.policy.MinLength))
		result.Suggestions = append(result.Suggestions, "Add more characters to meet minimum length requirement")
	}
	result.Requirements["min_length"] = len(password) >= v.policy.MinLength

	if len(password) > v.policy.MaxLength {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Password must not exceed %d characters", v.policy.MaxLength))
	}
	result.Requirements["max_length"] = len(password) <= v.policy.MaxLength

	// Check character requirements
	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSymbol := false
	charCount := make(map[rune]int)
	uniqueChars := 0

	for _, char := range password {
		charCount[char]++
		if charCount[char] == 1 {
			uniqueChars++
		}

		if unicode.IsUpper(char) {
			hasUpper = true
		} else if unicode.IsLower(char) {
			hasLower = true
		} else if unicode.IsDigit(char) {
			hasNumber = true
		} else if unicode.IsPunct(char) || unicode.IsSymbol(char) {
			hasSymbol = true
		}
	}

	if v.policy.RequireUppercase && !hasUpper {
		result.Valid = false
		result.Errors = append(result.Errors, "Password must contain at least one uppercase letter")
		result.Suggestions = append(result.Suggestions, "Add uppercase letters (A-Z)")
	}
	result.Requirements["uppercase"] = hasUpper

	if v.policy.RequireLowercase && !hasLower {
		result.Valid = false
		result.Errors = append(result.Errors, "Password must contain at least one lowercase letter")
		result.Suggestions = append(result.Suggestions, "Add lowercase letters (a-z)")
	}
	result.Requirements["lowercase"] = hasLower

	if v.policy.RequireNumbers && !hasNumber {
		result.Valid = false
		result.Errors = append(result.Errors, "Password must contain at least one number")
		result.Suggestions = append(result.Suggestions, "Add numbers (0-9)")
	}
	result.Requirements["numbers"] = hasNumber

	if v.policy.RequireSymbols && !hasSymbol {
		result.Valid = false
		result.Errors = append(result.Errors, "Password must contain at least one special character")
		result.Suggestions = append(result.Suggestions, "Add special characters (!@#$%^&*)")
	}
	result.Requirements["symbols"] = hasSymbol

	// Check unique characters
	if uniqueChars < v.policy.MinUniqueChars {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Password must contain at least %d unique characters", v.policy.MinUniqueChars))
		result.Suggestions = append(result.Suggestions, "Use more varied characters")
	}
	result.Requirements["unique_chars"] = uniqueChars >= v.policy.MinUniqueChars

	// Check repeated characters
	maxRepeated := 0
	for _, count := range charCount {
		if count > maxRepeated {
			maxRepeated = count
		}
	}
	if maxRepeated > v.policy.MaxRepeatedChars {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Password cannot have more than %d repeated characters", v.policy.MaxRepeatedChars))
		result.Suggestions = append(result.Suggestions, "Avoid repeating the same character too many times")
	}
	result.Requirements["max_repeated"] = maxRepeated <= v.policy.MaxRepeatedChars

	// Check forbidden words
	lowerPassword := strings.ToLower(password)
	for _, word := range v.policy.ForbiddenWords {
		if strings.Contains(lowerPassword, strings.ToLower(word)) {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("Password cannot contain the word '%s'", word))
			result.Suggestions = append(result.Suggestions, "Avoid using common words or phrases")
			break
		}
	}

	// Check for common patterns
	if v.hasCommonPatterns(password) {
		result.Warnings = append(result.Warnings, "Password contains common patterns")
		result.Suggestions = append(result.Suggestions, "Avoid sequential characters or keyboard patterns")
	}

	// Get password strength
	result.Strength = v.GetPasswordStrength(password)
	result.Score = result.Strength.Score

	return result
}

// CheckCommonPasswords checks if password is in common passwords list
func (v *DefaultPasswordValidator) CheckCommonPasswords(password string) error {
	if v.commonPasswords[strings.ToLower(password)] {
		return fmt.Errorf("password is too common and easily guessable")
	}
	return nil
}

// ValidatePasswordHistory checks if password was used recently
func (v *DefaultPasswordValidator) ValidatePasswordHistory(ctx context.Context, userID uint, password string) error {
	if v.historyRepo == nil {
		return nil // Skip if no history repository
	}

	// Get recent password history
	history, err := v.historyRepo.GetRecentPasswords(ctx, userID, v.policy.PreventReuse)
	if err != nil {
		return fmt.Errorf("failed to check password history: %v", err)
	}

	// Hash the new password for comparison
	passwordHash := v.hashPassword(password)

	// Check against recent passwords
	for _, entry := range history {
		if entry.PasswordHash == passwordHash {
			return fmt.Errorf("password was recently used, please choose a different password")
		}
	}

	return nil
}

// AddPasswordToHistory adds a password to the user's password history
func (v *DefaultPasswordValidator) AddPasswordToHistory(ctx context.Context, userID uint, password string) error {
	if v.historyRepo == nil {
		return nil // Skip if no history repository
	}

	// Hash the password
	passwordHash := v.hashPassword(password)

	// Create password history entry
	entry := &models.PasswordHistory{
		UserID:       userID,
		PasswordHash: passwordHash,
	}

	// Add to history
	if err := v.historyRepo.CreatePasswordHistory(ctx, entry); err != nil {
		return fmt.Errorf("failed to add password to history: %v", err)
	}

	// Cleanup old passwords to maintain the limit
	if err := v.historyRepo.CleanupOldPasswords(ctx, userID, v.policy.PreventReuse); err != nil {
		// Log error but don't fail the operation
		// In production, you might want to log this error
		_ = err
	}

	return nil
}

// GenerateSecurePassword generates a secure password
func (v *DefaultPasswordValidator) GenerateSecurePassword() string {
	const (
		lowercase = "abcdefghijklmnopqrstuvwxyz"
		uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		numbers   = "0123456789"
		symbols   = "!@#$%^&*()_+-=[]{}|;:,.<>?"
	)

	length := v.policy.MinLength
	if length < 12 {
		length = 12 // Ensure minimum secure length
	}

	var charset string
	var password strings.Builder

	// Ensure at least one character from each required set
	if v.policy.RequireLowercase {
		charset += lowercase
		password.WriteByte(lowercase[v.secureRandom(len(lowercase))])
	}
	if v.policy.RequireUppercase {
		charset += uppercase
		password.WriteByte(uppercase[v.secureRandom(len(uppercase))])
	}
	if v.policy.RequireNumbers {
		charset += numbers
		password.WriteByte(numbers[v.secureRandom(len(numbers))])
	}
	if v.policy.RequireSymbols {
		charset += symbols
		password.WriteByte(symbols[v.secureRandom(len(symbols))])
	}

	// Fill remaining length with random characters from charset
	for password.Len() < length {
		password.WriteByte(charset[v.secureRandom(len(charset))])
	}

	// Shuffle the password
	return v.shuffleString(password.String())
}

// GetPasswordStrength returns password strength score (0-100)
func (v *DefaultPasswordValidator) GetPasswordStrength(password string) *PasswordStrengthResult {
	score := 0
	entropy := 0.0

	// Length score (0-25 points)
	lengthScore := len(password) * 2
	if lengthScore > 25 {
		lengthScore = 25
	}
	score += lengthScore

	// Character variety score (0-25 points)
	charsetSize := 0
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSymbol := regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(password)

	if hasLower {
		charsetSize += 26
		score += 5
	}
	if hasUpper {
		charsetSize += 26
		score += 5
	}
	if hasNumber {
		charsetSize += 10
		score += 5
	}
	if hasSymbol {
		charsetSize += 32
		score += 10
	}

	// Calculate entropy
	if charsetSize > 0 {
		entropy = float64(len(password)) * (3.32 * float64(charsetSize))
	}

	// Uniqueness score (0-25 points)
	uniqueChars := make(map[rune]bool)
	for _, char := range password {
		uniqueChars[char] = true
	}
	uniquenessRatio := float64(len(uniqueChars)) / float64(len(password))
	score += int(uniquenessRatio * 25)

	// Pattern penalties (0-25 points)
	patternScore := 25
	if v.hasCommonPatterns(password) {
		patternScore -= 10
	}
	if v.hasRepeatedChars(password) {
		patternScore -= 10
	}
	if v.hasSequentialChars(password) {
		patternScore -= 5
	}
	score += patternScore

	// Ensure score is within bounds
	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}

	// Determine strength level
	var level string
	var timeToCrack string

	switch {
	case score >= 90:
		level = "very_strong"
		timeToCrack = "centuries"
	case score >= 70:
		level = "strong"
		timeToCrack = "decades"
	case score >= 50:
		level = "good"
		timeToCrack = "years"
	case score >= 30:
		level = "fair"
		timeToCrack = "months"
	default:
		level = "weak"
		timeToCrack = "minutes to days"
	}

	return &PasswordStrengthResult{
		Score:       score,
		Level:       level,
		Entropy:     entropy,
		TimeToCrack: timeToCrack,
	}
}

// ValidatePasswordPolicy validates password against all policies
func (v *DefaultPasswordValidator) ValidatePasswordPolicy(ctx context.Context, userID uint, password string) *PasswordValidationResult {
	result := v.ValidateComplexity(password)

	// Check common passwords
	if err := v.CheckCommonPasswords(password); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
		result.Suggestions = append(result.Suggestions, "Choose a more unique password")
	}

	// Check password history if userID is provided
	if userID > 0 {
		if err := v.ValidatePasswordHistory(ctx, userID, password); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, err.Error())
			result.Suggestions = append(result.Suggestions, "Try a password you haven't used recently")
		}
	}

	return result
}

// Helper methods

func (v *DefaultPasswordValidator) loadCommonPasswords() {
	// Common passwords list (top 100 most common passwords)
	commonPasswords := []string{
		"password", "123456", "password123", "admin", "qwerty", "letmein", "welcome",
		"monkey", "1234567890", "abc123", "111111", "dragon", "master", "princess",
		"login", "guest", "hello", "sunshine", "iloveyou", "password1", "123123",
		"batman", "trustno1", "thomas", "robert", "access", "love", "buster",
		"1234567", "soccer", "hockey", "killer", "george", "sexy", "andrew",
		"charlie", "superman", "asshole", "fuckyou", "dallas", "jessica", "panties",
		"pepper", "1111", "austin", "william", "daniel", "golfer", "summer",
		"heather", "hammer", "yankees", "joshua", "maggie", "biteme", "enter",
		"ashley", "thunder", "cowboy", "silver", "richard", "fucker", "orange",
		"merlin", "michelle", "corvette", "bigdog", "cheese", "matthew", "patrick",
		"martin", "freedom", "ginger", "blowjob", "nicole", "sparky", "yellow",
		"camaro", "secret", "dick", "falcon", "taylor", "bitch", "hello123",
		"scooter", "please", "porsche", "guitar", "chelsea", "black", "diamond",
		"nascar", "jackson", "cameron", "computer", "amanda", "wizard", "xxxxxxxx",
		"money", "phoenix", "mickey", "bailey", "knight", "iceman", "tigers",
		"purple", "andrea", "horny", "dakota", "aaaaaa", "player", "sunshine",
		"morgan", "starwars", "boomer", "cowboys", "edward", "charles", "girls",
		"booboo", "coffee", "xxxxxx", "bulldog", "ncc1701", "rabbit", "peanut",
		"john", "johnny", "gandalf", "spanky", "winter", "brandy", "compaq",
	}

	for _, pwd := range commonPasswords {
		v.commonPasswords[strings.ToLower(pwd)] = true
	}
}

func (v *DefaultPasswordValidator) hasCommonPatterns(password string) bool {
	// Check for keyboard patterns
	keyboardPatterns := []string{
		"qwerty", "asdf", "zxcv", "1234", "abcd", "qaz", "wsx", "edc",
	}

	lowerPassword := strings.ToLower(password)
	for _, pattern := range keyboardPatterns {
		if strings.Contains(lowerPassword, pattern) {
			return true
		}
	}

	return false
}

func (v *DefaultPasswordValidator) hasRepeatedChars(password string) bool {
	for i := 0; i < len(password)-2; i++ {
		if password[i] == password[i+1] && password[i+1] == password[i+2] {
			return true
		}
	}
	return false
}

func (v *DefaultPasswordValidator) hasSequentialChars(password string) bool {
	for i := 0; i < len(password)-2; i++ {
		if password[i]+1 == password[i+1] && password[i+1]+1 == password[i+2] {
			return true
		}
		if password[i]-1 == password[i+1] && password[i+1]-1 == password[i+2] {
			return true
		}
	}
	return false
}

func (v *DefaultPasswordValidator) hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return fmt.Sprintf("%x", hash)
}

func (v *DefaultPasswordValidator) secureRandom(max int) int {
	// Simple pseudo-random for demo purposes
	// In production, use crypto/rand
	return int(time.Now().UnixNano()) % max
}

func (v *DefaultPasswordValidator) shuffleString(s string) string {
	runes := []rune(s)
	for i := len(runes) - 1; i > 0; i-- {
		j := v.secureRandom(i + 1)
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// NewPasswordValidator creates a new password validator based on configuration
func NewPasswordValidator(cfg *config.Config, historyRepo repositories.PasswordHistoryRepository) PasswordValidator {
	return NewDefaultPasswordValidator(cfg, historyRepo)
}
