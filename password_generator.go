package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
)

// PasswordGeneratorOptions defines options for password generation
type PasswordGeneratorOptions struct {
	Length           int  `json:"length"`
	IncludeLower     bool `json:"include_lowercase"`
	IncludeUpper     bool `json:"include_uppercase"`
	IncludeNumbers   bool `json:"include_numbers"`
	IncludeSpecial   bool `json:"include_special"`
	ExcludeAmbiguous bool `json:"exclude_ambiguous"`
	MinLower         int  `json:"min_lowercase,omitempty"`
	MinUpper         int  `json:"min_uppercase,omitempty"`
	MinNumbers       int  `json:"min_numbers,omitempty"`
	MinSpecial       int  `json:"min_special,omitempty"`
}

// Character sets for password generation
const (
	lowercaseChars = "abcdefghijklmnopqrstuvwxyz"
	uppercaseChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numberChars    = "0123456789"
	specialChars   = "!@#$%^&*()_+-=[]{}|;:,.<>?"
	ambiguousChars = "il1Lo0O" // Characters that look similar
)

// DefaultPasswordOptions returns default password generation options
func DefaultPasswordOptions() PasswordGeneratorOptions {
	return PasswordGeneratorOptions{
		Length:           16,
		IncludeLower:     true,
		IncludeUpper:     true,
		IncludeNumbers:   true,
		IncludeSpecial:   true,
		ExcludeAmbiguous: false,
	}
}

// GeneratedPassword contains the generated password and its analysis
type GeneratedPassword struct {
	Password     string                   `json:"password"`
	Length       int                      `json:"length"`
	Strength     PasswordStrength         `json:"strength"`
	StrengthText string                   `json:"strength_text"`
	Entropy      float64                  `json:"entropy"`
	Options      PasswordGeneratorOptions `json:"options"`
}

// GeneratePassword generates a cryptographically secure random password
func GeneratePassword(opts PasswordGeneratorOptions) (*GeneratedPassword, error) {
	// Validate options
	if err := validateOptions(opts); err != nil {
		return nil, err
	}

	// Build character set
	charset := buildCharset(opts)
	if len(charset) == 0 {
		return nil, errors.New("no character types selected")
	}

	// Generate password
	password, err := generateRandomString(charset, opts.Length)
	if err != nil {
		return nil, fmt.Errorf("failed to generate password: %w", err)
	}

	// Ensure minimum requirements are met
	if hasMinimumRequirements(opts) {
		password, err = ensureMinimumRequirements(password, opts)
		if err != nil {
			return nil, err
		}
	}

	// Calculate entropy
	entropy := calculateEntropy(len(charset), opts.Length)

	// Create result
	result := &GeneratedPassword{
		Password: password,
		Length:   len(password),
		Entropy:  entropy,
		Options:  opts,
	}

	// Analyze strength (without security checker as this is a new password)
	result.Strength = estimateStrength(password, entropy)
	result.StrengthText = result.Strength.String()

	return result, nil
}

// validateOptions validates password generation options
func validateOptions(opts PasswordGeneratorOptions) error {
	if opts.Length < 1 {
		return errors.New("password length must be at least 1")
	}

	if opts.Length > 256 {
		return errors.New("password length cannot exceed 256 characters")
	}

	if !opts.IncludeLower && !opts.IncludeUpper && !opts.IncludeNumbers && !opts.IncludeSpecial {
		return errors.New("at least one character type must be selected")
	}

	// Check minimum requirements don't exceed length
	minTotal := opts.MinLower + opts.MinUpper + opts.MinNumbers + opts.MinSpecial
	if minTotal > opts.Length {
		return fmt.Errorf(
			"minimum requirements (%d) exceed password length (%d)",
			minTotal,
			opts.Length,
		)
	}

	return nil
}

// buildCharset builds the character set based on options
func buildCharset(opts PasswordGeneratorOptions) string {
	charset := ""

	if opts.IncludeLower {
		chars := lowercaseChars
		if opts.ExcludeAmbiguous {
			chars = removeAmbiguous(chars)
		}
		charset += chars
	}

	if opts.IncludeUpper {
		chars := uppercaseChars
		if opts.ExcludeAmbiguous {
			chars = removeAmbiguous(chars)
		}
		charset += chars
	}

	if opts.IncludeNumbers {
		chars := numberChars
		if opts.ExcludeAmbiguous {
			chars = removeAmbiguous(chars)
		}
		charset += chars
	}

	if opts.IncludeSpecial {
		charset += specialChars
	}

	return charset
}

// removeAmbiguous removes ambiguous characters from a string
func removeAmbiguous(s string) string {
	result := ""
	for _, char := range s {
		isAmbiguous := false
		for _, amb := range ambiguousChars {
			if char == amb {
				isAmbiguous = true
				break
			}
		}
		if !isAmbiguous {
			result += string(char)
		}
	}
	return result
}

// generateRandomString generates a cryptographically secure random string
func generateRandomString(charset string, length int) (string, error) {
	if len(charset) == 0 {
		return "", errors.New("charset is empty")
	}

	result := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		randomIndex, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		result[i] = charset[randomIndex.Int64()]
	}

	return string(result), nil
}

// hasMinimumRequirements checks if any minimum requirements are set
func hasMinimumRequirements(opts PasswordGeneratorOptions) bool {
	return opts.MinLower > 0 || opts.MinUpper > 0 || opts.MinNumbers > 0 || opts.MinSpecial > 0
}

// ensureMinimumRequirements ensures the password meets minimum character requirements
func ensureMinimumRequirements(password string, opts PasswordGeneratorOptions) (string, error) {
	result := []rune(password)

	// Count current characters
	counts := countCharacterTypes(password)

	// Track positions that need replacement
	replacePositions := []int{}
	for i := 0; i < len(result); i++ {
		replacePositions = append(replacePositions, i)
	}

	// Shuffle positions randomly
	if err := shuffleSlice(replacePositions); err != nil {
		return "", err
	}

	posIndex := 0

	// Add missing lowercase
	if opts.MinLower > 0 && counts.lowercase < opts.MinLower {
		needed := opts.MinLower - counts.lowercase
		for i := 0; i < needed && posIndex < len(replacePositions); i++ {
			char, err := getRandomChar(lowercaseChars, opts.ExcludeAmbiguous)
			if err != nil {
				return "", err
			}
			result[replacePositions[posIndex]] = rune(char)
			posIndex++
		}
	}

	// Add missing uppercase
	if opts.MinUpper > 0 && counts.uppercase < opts.MinUpper {
		needed := opts.MinUpper - counts.uppercase
		for i := 0; i < needed && posIndex < len(replacePositions); i++ {
			char, err := getRandomChar(uppercaseChars, opts.ExcludeAmbiguous)
			if err != nil {
				return "", err
			}
			result[replacePositions[posIndex]] = rune(char)
			posIndex++
		}
	}

	// Add missing numbers
	if opts.MinNumbers > 0 && counts.numbers < opts.MinNumbers {
		needed := opts.MinNumbers - counts.numbers
		for i := 0; i < needed && posIndex < len(replacePositions); i++ {
			char, err := getRandomChar(numberChars, opts.ExcludeAmbiguous)
			if err != nil {
				return "", err
			}
			result[replacePositions[posIndex]] = rune(char)
			posIndex++
		}
	}

	// Add missing special
	if opts.MinSpecial > 0 && counts.special < opts.MinSpecial {
		needed := opts.MinSpecial - counts.special
		for i := 0; i < needed && posIndex < len(replacePositions); i++ {
			char, err := getRandomChar(specialChars, false)
			if err != nil {
				return "", err
			}
			result[replacePositions[posIndex]] = rune(char)
			posIndex++
		}
	}

	return string(result), nil
}

// characterCounts tracks counts of different character types
type characterCounts struct {
	lowercase int
	uppercase int
	numbers   int
	special   int
}

// countCharacterTypes counts how many of each character type are in a string
func countCharacterTypes(s string) characterCounts {
	counts := characterCounts{}

	for _, char := range s {
		if char >= 'a' && char <= 'z' {
			counts.lowercase++
		} else if char >= 'A' && char <= 'Z' {
			counts.uppercase++
		} else if char >= '0' && char <= '9' {
			counts.numbers++
		} else {
			counts.special++
		}
	}

	return counts
}

// getRandomChar gets a random character from a charset
func getRandomChar(charset string, excludeAmbiguous bool) (byte, error) {
	chars := charset
	if excludeAmbiguous {
		chars = removeAmbiguous(chars)
	}

	if len(chars) == 0 {
		return 0, errors.New("no characters available")
	}

	charsetLen := big.NewInt(int64(len(chars)))
	randomIndex, err := rand.Int(rand.Reader, charsetLen)
	if err != nil {
		return 0, err
	}

	return chars[randomIndex.Int64()], nil
}

// shuffleSlice randomly shuffles a slice using crypto/rand
func shuffleSlice(slice []int) error {
	n := len(slice)
	for i := n - 1; i > 0; i-- {
		jBig, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return err
		}
		j := jBig.Int64()
		slice[i], slice[j] = slice[j], slice[i]
	}
	return nil
}

// calculateEntropy calculates password entropy in bits
func calculateEntropy(charsetSize, length int) float64 {
	if charsetSize <= 0 || length <= 0 {
		return 0
	}

	// Entropy = log2(charsetSize^length) = length * log2(charsetSize)
	return float64(length) * math.Log2(float64(charsetSize))
}

// estimateStrength estimates password strength based on entropy
func estimateStrength(password string, entropy float64) PasswordStrength {
	// Check character variety
	counts := countCharacterTypes(password)
	variety := 0
	if counts.lowercase > 0 {
		variety++
	}
	if counts.uppercase > 0 {
		variety++
	}
	if counts.numbers > 0 {
		variety++
	}
	if counts.special > 0 {
		variety++
	}

	// length := len(password)

	// Entropy-based strength
	// < 28 bits: Very Weak
	// 28-35 bits: Weak
	// 36-59 bits: Medium
	// 60-127 bits: Strong
	// >= 128 bits: Very Strong

	switch {
	case entropy >= 128 && variety >= 4:
		return PasswordVeryStrong
	case entropy >= 80 && variety >= 3:
		return PasswordStrong
	case entropy >= 60 && variety >= 3:
		return PasswordStrong
	case entropy >= 36 && variety >= 2:
		return PasswordMedium
	case entropy >= 28:
		return PasswordWeak
	default:
		return PasswordVeryWeak
	}
}

// GenerateMultiplePasswords generates multiple passwords with the same options
func GenerateMultiplePasswords(
	opts PasswordGeneratorOptions,
	count int,
) ([]*GeneratedPassword, error) {
	if count < 1 {
		return nil, errors.New("count must be at least 1")
	}

	if count > 100 {
		return nil, errors.New("count cannot exceed 100")
	}

	passwords := make([]*GeneratedPassword, count)

	for i := 0; i < count; i++ {
		pwd, err := GeneratePassword(opts)
		if err != nil {
			return nil, fmt.Errorf("failed to generate password %d: %w", i+1, err)
		}
		passwords[i] = pwd
	}

	return passwords, nil
}

// GeneratePassphrase generates a memorable passphrase using random words
func GeneratePassphrase(
	wordCount int,
	separator string,
	capitalize bool,
	includeNumbers bool,
) (string, error) {
	if wordCount < 2 {
		return "", errors.New("word count must be at least 2")
	}

	if wordCount > 10 {
		return "", errors.New("word count cannot exceed 10")
	}

	// Common word list for passphrases
	words := []string{
		"correct", "horse", "battery", "staple", "dragon", "monkey", "tiger", "cloud",
		"ocean", "mountain", "river", "forest", "castle", "sword", "shield", "arrow",
		"thunder", "lightning", "storm", "wind", "fire", "water", "earth", "sky",
		"sun", "moon", "star", "planet", "galaxy", "universe", "atom", "quantum",
		"phoenix", "griffin", "unicorn", "kraken", "giant", "wizard", "knight", "ninja",
		"robot", "cyber", "digital", "virtual", "matrix", "code", "data", "network",
		"alpha", "beta", "gamma", "delta", "omega", "sigma", "theta", "lambda",
	}

	selectedWords := make([]string, wordCount)

	for i := 0; i < wordCount; i++ {
		// Generate random index
		indexBig, err := rand.Int(rand.Reader, big.NewInt(int64(len(words))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random index: %w", err)
		}

		word := words[indexBig.Int64()]

		if capitalize {
			// Capitalize first letter
			if len(word) > 0 {
				word = string(word[0]-32) + word[1:]
			}
		}

		selectedWords[i] = word
	}

	passphrase := ""
	for i, word := range selectedWords {
		if i > 0 {
			passphrase += separator
		}
		passphrase += word
	}

	// Add number if requested
	if includeNumbers {
		numBig, err := rand.Int(rand.Reader, big.NewInt(1000))
		if err != nil {
			return "", err
		}
		passphrase += separator + fmt.Sprintf("%d", numBig.Int64())
	}

	return passphrase, nil
}
