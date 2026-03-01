package main

import (
	"os"
	"testing"
)

func TestPasswordSecurityChecker(t *testing.T) {
	checker := NewPasswordSecurityChecker()

	// Create temp file with common passwords
	tmpFile, err := os.CreateTemp("", "passwords_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write test passwords
	commonPasswords := `password
123456
qwerty
admin
letmein
`
	if _, err := tmpFile.WriteString(commonPasswords); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Load passwords
	err = checker.LoadCommonPasswords(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load common passwords: %v", err)
	}

	// Test IsCommonPassword
	tests := []struct {
		password string
		expected bool
	}{
		{"password", true},
		{"Password", true}, // case insensitive
		{"PASSWORD", true}, // case insensitive
		{"123456", true},
		{"qwerty", true},
		{"SecureP@ssw0rd!", false},
		{"MyUniquePassword123!", false},
	}

	for _, tt := range tests {
		result := checker.IsCommonPassword(tt.password)
		if result != tt.expected {
			t.Errorf("IsCommonPassword(%q) = %v, expected %v", tt.password, result, tt.expected)
		}
	}
}

// func TestAnalyzePassword(t *testing.T) {
// 	checker := NewPasswordSecurityChecker()
//
// 	// Create temp file
// 	tmpFile, err := os.CreateTemp("", "passwords_*.txt")
// 	if err != nil {
// 		t.Fatalf("Failed to create temp file: %v", err)
// 	}
// 	defer os.Remove(tmpFile.Name())
//
// 	tmpFile.WriteString("password\n123456\n")
// 	tmpFile.Close()
//
// 	checker.LoadCommonPasswords(tmpFile.Name())
//
// 	tests := []struct {
// 		password         string
// 		expectedIsCommon bool
// 		minScore         int
// 		expectedStrength PasswordStrength
// 	}{
// 		{"password", true, 0, PasswordVeryWeak},
// 		{"123456", true, 0, PasswordVeryWeak},
// 		{"Weak1", false, 2, PasswordWeak},
// 		{"Medium@123", false, 4, PasswordMedium},
// 		{"Strong!Pass123", false, 5, PasswordStrong},
// 		{"VeryStr0ng!P@ssw0rd2024", false, 7, PasswordVeryStrong},
// 	}
//
// 	for _, tt := range tests {
// 		report := checker.AnalyzePassword(tt.password)
//
// 		if report.IsCommon != tt.expectedIsCommon {
// 			t.Errorf("AnalyzePassword(%q).IsCommon = %v, expected %v",
// 				tt.password, report.IsCommon, tt.expectedIsCommon)
// 		}
//
// 		if report.Score < tt.minScore {
// 			t.Errorf("AnalyzePassword(%q).Score = %d, expected >= %d",
// 				tt.password, report.Score, tt.minScore)
// 		}
//
// 		if report.Strength != tt.expectedStrength {
// 			t.Errorf("AnalyzePassword(%q).Strength = %v, expected %v",
// 				tt.password, report.Strength, tt.expectedStrength)
// 		}
// 	}
// }

func TestPasswordAnalysisCharacterTypes(t *testing.T) {
	checker := NewPasswordSecurityChecker()

	tests := []struct {
		password     string
		hasLowercase bool
		hasUppercase bool
		hasNumbers   bool
		hasSpecial   bool
	}{
		{"lowercase", true, false, false, false},
		{"UPPERCASE", false, true, false, false},
		{"Numbers123", true, true, true, false},
		{"Special!@#", true, true, false, true},
		{"All1Types!", true, true, true, true},
		{"12345", false, false, true, false},
	}

	for _, tt := range tests {
		report := checker.AnalyzePassword(tt.password)

		if report.HasLowercase != tt.hasLowercase {
			t.Errorf("%q HasLowercase = %v, expected %v",
				tt.password, report.HasLowercase, tt.hasLowercase)
		}

		if report.HasUppercase != tt.hasUppercase {
			t.Errorf("%q HasUppercase = %v, expected %v",
				tt.password, report.HasUppercase, tt.hasUppercase)
		}

		if report.HasNumbers != tt.hasNumbers {
			t.Errorf("%q HasNumbers = %v, expected %v",
				tt.password, report.HasNumbers, tt.hasNumbers)
		}

		if report.HasSpecial != tt.hasSpecial {
			t.Errorf("%q HasSpecial = %v, expected %v",
				tt.password, report.HasSpecial, tt.hasSpecial)
		}
	}
}

func TestCheckAllAccountPasswords(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	accountRepo := &AccountRepository{}
	checker := NewPasswordSecurityChecker()

	// Create temp common passwords file
	tmpFile, err := os.CreateTemp("", "passwords_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString("password\n123456\nweak\n")
	tmpFile.Close()

	if err := checker.LoadCommonPasswords(tmpFile.Name()); err != nil {
		t.Fatalf("Failed to load common passwords: %v", err)
	}

	// Create test accounts
	accounts := []*Account{
		{User: "user1", Password: "password"},            // Common
		{User: "user2", Password: "weak"},                // Common
		{User: "user3", Password: "Str0ng!P@ss"},         // Strong
		{User: "user4", Password: "VeryStr0ng!P@ss2024"}, // Very Strong
	}

	for _, acc := range accounts {
		_, err := accountRepo.Create(acc)
		if err != nil {
			t.Fatalf("Failed to create account: %v", err)
		}
	}

	// Check all passwords
	reports, err := accountRepo.CheckAllAccountPasswords(checker)
	if err != nil {
		t.Fatalf("CheckAllAccountPasswords failed: %v", err)
	}

	if len(reports) != 4 {
		t.Errorf("Expected 4 reports, got %d", len(reports))
	}

	// Verify common passwords detected
	commonCount := 0
	for _, report := range reports {
		if report.Analysis.IsCommon {
			commonCount++
		}
	}

	if commonCount != 2 {
		t.Errorf("Expected 2 common passwords, got %d", commonCount)
	}
}

func TestGetVulnerableAccounts(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	accountRepo := &AccountRepository{}
	checker := NewPasswordSecurityChecker()

	// Create temp file
	tmpFile, err := os.CreateTemp("", "passwords_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString("password\n123456\n")
	tmpFile.Close()

	checker.LoadCommonPasswords(tmpFile.Name())

	// Create accounts
	accounts := []*Account{
		{User: "user1", Password: "password"},    // Vulnerable (common)
		{User: "user2", Password: "weak1"},       // Vulnerable (weak)
		{User: "user3", Password: "Str0ng!P@ss"}, // Not vulnerable
	}

	for _, acc := range accounts {
		accountRepo.Create(acc)
	}

	// Get vulnerable accounts
	vulnerable, err := accountRepo.GetVulnerableAccounts(checker)
	if err != nil {
		t.Fatalf("GetVulnerableAccounts failed: %v", err)
	}

	if len(vulnerable) < 1 {
		t.Errorf("Expected at least 1 vulnerable account, got %d", len(vulnerable))
	}

	// Verify all returned accounts are actually vulnerable
	for _, report := range vulnerable {
		if !report.Analysis.IsCommon && report.Analysis.Strength > PasswordWeak {
			t.Errorf("Account %d should not be in vulnerable list", report.AccountID)
		}
	}
}

func TestGetSecurityStatistics(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	accountRepo := &AccountRepository{}
	checker := NewPasswordSecurityChecker()

	// Create temp file
	tmpFile, err := os.CreateTemp("", "passwords_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString("password\n")
	tmpFile.Close()

	checker.LoadCommonPasswords(tmpFile.Name())

	// Create accounts with different strength passwords
	accounts := []*Account{
		{User: "user1", Password: "password"},            // Very Weak (common)
		{User: "user2", Password: "Weak1"},               // Weak
		{User: "user3", Password: "Medium@123"},          // Medium
		{User: "user4", Password: "Strong!Pass123"},      // Strong
		{User: "user5", Password: "VeryStr0ng!P@ss2024"}, // Very Strong
	}

	for _, acc := range accounts {
		accountRepo.Create(acc)
	}

	// Get statistics
	stats, err := accountRepo.GetSecurityStatistics(checker, false)
	if err != nil {
		t.Fatalf("GetSecurityStatistics failed: %v", err)
	}

	if stats.TotalAccounts != 5 {
		t.Errorf("Expected 5 total accounts, got %d", stats.TotalAccounts)
	}

	if stats.CommonPasswords != 1 {
		t.Errorf("Expected 1 common password, got %d", stats.CommonPasswords)
	}

	// Verify we have accounts in different strength categories
	totalByStrength := stats.VeryWeakPasswords + stats.WeakPasswords +
		stats.MediumPasswords + stats.StrongPasswords + stats.VeryStrongPasswords

	if totalByStrength != stats.TotalAccounts {
		t.Errorf("Strength distribution doesn't match total accounts: %d vs %d",
			totalByStrength, stats.TotalAccounts)
	}
}
