package main

import (
	"strings"
	"testing"
)

func TestGeneratePassword(t *testing.T) {
	opts := DefaultPasswordOptions()

	generated, err := GeneratePassword(opts)
	if err != nil {
		t.Fatalf("GeneratePassword failed: %v", err)
	}

	if generated.Password == "" {
		t.Error("Generated password is empty")
	}

	if generated.Length != opts.Length {
		t.Errorf("Password length = %d, expected %d", generated.Length, opts.Length)
	}

	if generated.Entropy <= 0 {
		t.Error("Entropy should be > 0")
	}
}

func TestGeneratePasswordCharacterTypes(t *testing.T) {
	tests := []struct {
		name    string
		opts    PasswordGeneratorOptions
		checkFn func(string) bool
	}{
		{
			name: "Only lowercase",
			opts: PasswordGeneratorOptions{
				Length:       20,
				IncludeLower: true,
			},
			checkFn: func(pwd string) bool {
				for _, c := range pwd {
					if c < 'a' || c > 'z' {
						return false
					}
				}
				return true
			},
		},
		{
			name: "Only uppercase",
			opts: PasswordGeneratorOptions{
				Length:       20,
				IncludeUpper: true,
			},
			checkFn: func(pwd string) bool {
				for _, c := range pwd {
					if c < 'A' || c > 'Z' {
						return false
					}
				}
				return true
			},
		},
		{
			name: "Only numbers",
			opts: PasswordGeneratorOptions{
				Length:         20,
				IncludeNumbers: true,
			},
			checkFn: func(pwd string) bool {
				for _, c := range pwd {
					if c < '0' || c > '9' {
						return false
					}
				}
				return true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generated, err := GeneratePassword(tt.opts)
			if err != nil {
				t.Fatalf("GeneratePassword failed: %v", err)
			}

			if !tt.checkFn(generated.Password) {
				t.Errorf("Password '%s' doesn't match expected character types", generated.Password)
			}
		})
	}
}

func TestGeneratePasswordExcludeAmbiguous(t *testing.T) {
	opts := PasswordGeneratorOptions{
		Length:           100,
		IncludeLower:     true,
		IncludeUpper:     true,
		IncludeNumbers:   true,
		ExcludeAmbiguous: true,
	}

	generated, err := GeneratePassword(opts)
	if err != nil {
		t.Fatalf("GeneratePassword failed: %v", err)
	}

	// Check for ambiguous characters
	ambiguous := "il1Lo0O"
	for _, char := range generated.Password {
		if strings.ContainsRune(ambiguous, char) {
			t.Errorf("Password contains ambiguous character '%c'", char)
		}
	}
}

func TestGeneratePasswordValidation(t *testing.T) {
	tests := []struct {
		name        string
		opts        PasswordGeneratorOptions
		expectError bool
	}{
		{
			name: "Valid options",
			opts: PasswordGeneratorOptions{
				Length:       16,
				IncludeLower: true,
			},
			expectError: false,
		},
		{
			name: "Zero length",
			opts: PasswordGeneratorOptions{
				Length:       0,
				IncludeLower: true,
			},
			expectError: true,
		},
		{
			name: "Too long",
			opts: PasswordGeneratorOptions{
				Length:       300,
				IncludeLower: true,
			},
			expectError: true,
		},
		{
			name: "No character types",
			opts: PasswordGeneratorOptions{
				Length: 16,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GeneratePassword(tt.opts)
			if (err != nil) != tt.expectError {
				t.Errorf("Expected error: %v, got: %v", tt.expectError, err)
			}
		})
	}
}

func TestGenerateMultiplePasswords(t *testing.T) {
	opts := DefaultPasswordOptions()
	count := 5

	passwords, err := GenerateMultiplePasswords(opts, count)
	if err != nil {
		t.Fatalf("GenerateMultiplePasswords failed: %v", err)
	}

	if len(passwords) != count {
		t.Errorf("Expected %d passwords, got %d", count, len(passwords))
	}

	// Check uniqueness
	seen := make(map[string]bool)
	for _, pwd := range passwords {
		if seen[pwd.Password] {
			t.Errorf("Duplicate password generated: %s", pwd.Password)
		}
		seen[pwd.Password] = true
	}
}

func TestGeneratePassphrase(t *testing.T) {
	wordCount := 4
	separator := "-"
	capitalize := true
	includeNumbers := true

	passphrase, err := GeneratePassphrase(wordCount, separator, capitalize, includeNumbers)
	if err != nil {
		t.Fatalf("GeneratePassphrase failed: %v", err)
	}

	if passphrase == "" {
		t.Error("Generated passphrase is empty")
	}

	// Check for separator
	if !strings.Contains(passphrase, separator) {
		t.Error("Passphrase doesn't contain separator")
	}

	// Count separators (should be wordCount for words + possibly 1 for number)
	sepCount := strings.Count(passphrase, separator)
	if sepCount < wordCount-1 {
		t.Errorf("Expected at least %d separators, got %d", wordCount-1, sepCount)
	}

	// If capitalize is true, check first char is uppercase
	if capitalize && len(passphrase) > 0 {
		firstChar := rune(passphrase[0])
		if firstChar < 'A' || firstChar > 'Z' {
			t.Error("First character should be uppercase")
		}
	}
}

func TestGeneratePassphraseValidation(t *testing.T) {
	tests := []struct {
		name        string
		wordCount   int
		expectError bool
	}{
		{"Valid 4 words", 4, false},
		{"Valid 2 words", 2, false},
		{"Invalid 1 word", 1, true},
		{"Invalid 11 words", 11, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GeneratePassphrase(tt.wordCount, "-", true, false)
			if (err != nil) != tt.expectError {
				t.Errorf("Expected error: %v, got: %v", tt.expectError, err)
			}
		})
	}
}

func TestCalculateEntropy(t *testing.T) {
	tests := []struct {
		charsetSize int
		length      int
		minEntropy  float64
	}{
		{26, 8, 37.0},   // lowercase only, 8 chars
		{52, 8, 45.0},   // upper+lower, 8 chars
		{62, 12, 71.0},  // upper+lower+numbers, 12 chars
		{94, 16, 104.0}, // all characters, 16 chars
	}

	for _, tt := range tests {
		entropy := calculateEntropy(tt.charsetSize, tt.length)
		if entropy < tt.minEntropy {
			t.Errorf("Entropy for charset=%d, length=%d: got %.1f, expected >= %.1f",
				tt.charsetSize, tt.length, entropy, tt.minEntropy)
		}
	}
}

func TestPasswordUniqueness(t *testing.T) {
	// Generate many passwords and ensure they're all unique
	opts := DefaultPasswordOptions()
	count := 100

	passwords := make(map[string]bool)

	for i := 0; i < count; i++ {
		generated, err := GeneratePassword(opts)
		if err != nil {
			t.Fatalf("Failed to generate password %d: %v", i, err)
		}

		if passwords[generated.Password] {
			t.Fatalf("Duplicate password generated: %s", generated.Password)
		}
		passwords[generated.Password] = true
	}

	if len(passwords) != count {
		t.Errorf("Expected %d unique passwords, got %d", count, len(passwords))
	}
}

func TestPasswordStrengthEstimation(t *testing.T) {
	tests := []struct {
		opts           PasswordGeneratorOptions
		minStrength    PasswordStrength
	}{
		{
			opts: PasswordGeneratorOptions{
				Length:       8,
				IncludeLower: true,
			},
			minStrength: PasswordWeak,
		},
		{
			opts: PasswordGeneratorOptions{
				Length:         16,
				IncludeLower:   true,
				IncludeUpper:   true,
				IncludeNumbers: true,
				IncludeSpecial: true,
			},
			minStrength: PasswordStrong,
		},
		{
			opts: PasswordGeneratorOptions{
				Length:         24,
				IncludeLower:   true,
				IncludeUpper:   true,
				IncludeNumbers: true,
				IncludeSpecial: true,
			},
			minStrength: PasswordVeryStrong,
		},
	}

	for _, tt := range tests {
		generated, err := GeneratePassword(tt.opts)
		if err != nil {
			t.Fatalf("GeneratePassword failed: %v", err)
		}

		if generated.Strength < tt.minStrength {
			t.Errorf("Password strength %v is less than expected minimum %v",
				generated.Strength, tt.minStrength)
		}
	}
}

func BenchmarkGeneratePassword(b *testing.B) {
	opts := DefaultPasswordOptions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GeneratePassword(opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGeneratePassphrase(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GeneratePassphrase(4, "-", true, true)
		if err != nil {
			b.Fatal(err)
		}
	}
}
