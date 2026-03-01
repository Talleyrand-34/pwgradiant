package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

// PasswordSecurityChecker manages common password checking
type PasswordSecurityChecker struct {
	commonPasswords map[string]bool
	mu              sync.RWMutex
	loaded          bool
}

// NewPasswordSecurityChecker creates a new password security checker
func NewPasswordSecurityChecker() *PasswordSecurityChecker {
	return &PasswordSecurityChecker{
		commonPasswords: make(map[string]bool),
		loaded:          false,
	}
}

// LoadCommonPasswords loads the common passwords list from a file
func (psc *PasswordSecurityChecker) LoadCommonPasswords(filepath string) error {
	psc.mu.Lock()
	defer psc.mu.Unlock()

	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open common passwords file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	count := 0

	for scanner.Scan() {
		password := strings.TrimSpace(scanner.Text())
		if password != "" {
			// Store both original and lowercase versions
			psc.commonPasswords[password] = true
			psc.commonPasswords[strings.ToLower(password)] = true
			count++
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading common passwords file: %w", err)
	}

	psc.loaded = true
	fmt.Printf("Loaded %d common passwords\n", count)
	return nil
}

// IsCommonPassword checks if a password is in the common passwords list
func (psc *PasswordSecurityChecker) IsCommonPassword(password string) bool {
	psc.mu.RLock()
	defer psc.mu.RUnlock()

	if !psc.loaded {
		return false
	}

	// Check exact match
	if psc.commonPasswords[password] {
		return true
	}

	// Check lowercase match
	if psc.commonPasswords[strings.ToLower(password)] {
		return true
	}

	return false
}

// PasswordStrength represents the strength level of a password
type PasswordStrength int

const (
	PasswordVeryWeak PasswordStrength = iota
	PasswordWeak
	PasswordMedium
	PasswordStrong
	PasswordVeryStrong
)

func (ps PasswordStrength) String() string {
	switch ps {
	case PasswordVeryWeak:
		return "Very Weak"
	case PasswordWeak:
		return "Weak"
	case PasswordMedium:
		return "Medium"
	case PasswordStrong:
		return "Strong"
	case PasswordVeryStrong:
		return "Very Strong"
	default:
		return "Unknown"
	}
}

// PasswordSecurityReport contains detailed analysis of password security
type PasswordSecurityReport struct {
	IsCommon       bool             `json:"is_common"`
	Strength       PasswordStrength `json:"strength"`
	StrengthText   string           `json:"strength_text"`
	Length         int              `json:"length"`
	HasLowercase   bool             `json:"has_lowercase"`
	HasUppercase   bool             `json:"has_uppercase"`
	HasNumbers     bool             `json:"has_numbers"`
	HasSpecial     bool             `json:"has_special"`
	Score          int              `json:"score"`
	Warnings       []string         `json:"warnings"`
	Recommendations []string        `json:"recommendations"`
}

// AnalyzePassword performs comprehensive password security analysis
func (psc *PasswordSecurityChecker) AnalyzePassword(password string) PasswordSecurityReport {
	report := PasswordSecurityReport{
		Warnings:        []string{},
		Recommendations: []string{},
		Length:          len(password),
	}

	// Check if it's a common password
	report.IsCommon = psc.IsCommonPassword(password)
	if report.IsCommon {
		report.Warnings = append(report.Warnings, "This password appears in common password lists")
		report.Recommendations = append(report.Recommendations, "Use a unique, randomly generated password")
	}

	// Check character types
	for _, char := range password {
		if char >= 'a' && char <= 'z' {
			report.HasLowercase = true
		} else if char >= 'A' && char <= 'Z' {
			report.HasUppercase = true
		} else if char >= '0' && char <= '9' {
			report.HasNumbers = true
		} else {
			report.HasSpecial = true
		}
	}

	// Calculate score
	report.Score = 0

	// Length scoring
	if report.Length >= 8 {
		report.Score += 1
	}
	if report.Length >= 12 {
		report.Score += 1
	}
	if report.Length >= 16 {
		report.Score += 1
	}

	// Character type scoring
	if report.HasLowercase {
		report.Score += 1
	} else {
		report.Recommendations = append(report.Recommendations, "Add lowercase letters (a-z)")
	}

	if report.HasUppercase {
		report.Score += 1
	} else {
		report.Recommendations = append(report.Recommendations, "Add uppercase letters (A-Z)")
	}

	if report.HasNumbers {
		report.Score += 1
	} else {
		report.Recommendations = append(report.Recommendations, "Add numbers (0-9)")
	}

	if report.HasSpecial {
		report.Score += 1
	} else {
		report.Recommendations = append(report.Recommendations, "Add special characters (!@#$%^&*)")
	}

	// Penalize if common
	if report.IsCommon {
		report.Score = 0
		report.Strength = PasswordVeryWeak
		report.StrengthText = "Very Weak"
		return report
	}

	// Determine strength based on score
	switch {
	case report.Score >= 6:
		report.Strength = PasswordVeryStrong
		report.StrengthText = "Very Strong"
	case report.Score >= 5:
		report.Strength = PasswordStrong
		report.StrengthText = "Strong"
	case report.Score >= 3:
		report.Strength = PasswordMedium
		report.StrengthText = "Medium"
		report.Warnings = append(report.Warnings, "Password could be stronger")
	case report.Score >= 2:
		report.Strength = PasswordWeak
		report.StrengthText = "Weak"
		report.Warnings = append(report.Warnings, "Weak password - consider using a stronger one")
	default:
		report.Strength = PasswordVeryWeak
		report.StrengthText = "Very Weak"
		report.Warnings = append(report.Warnings, "Very weak password - should be changed immediately")
	}

	// Additional warnings
	if report.Length < 8 {
		report.Warnings = append(report.Warnings, "Password is too short (minimum 8 characters recommended)")
		report.Recommendations = append(report.Recommendations, "Use at least 12 characters for better security")
	}

	return report
}

// CheckAllAccountPasswords checks security of all account passwords
func (r *AccountRepository) CheckAllAccountPasswords(checker *PasswordSecurityChecker) ([]AccountSecurityReport, error) {
	accounts, err := r.GetAll()
	if err != nil {
		return nil, err
	}

	reports := make([]AccountSecurityReport, 0, len(accounts))

	for _, account := range accounts {
		analysis := checker.AnalyzePassword(account.Password)
		
		report := AccountSecurityReport{
			AccountID:   account.ID,
			User:        account.User,
			Email:       account.Email,
			URL:         account.URL,
			Analysis:    analysis,
		}

		reports = append(reports, report)
	}

	return reports, nil
}

// AccountSecurityReport combines account info with password analysis
type AccountSecurityReport struct {
	AccountID int64                  `json:"account_id"`
	User      string                 `json:"user"`
	Email     *string                `json:"email,omitempty"`
	URL       *string                `json:"url,omitempty"`
	Analysis  PasswordSecurityReport `json:"analysis"`
}

// GetVulnerableAccounts returns accounts with weak or common passwords
func (r *AccountRepository) GetVulnerableAccounts(checker *PasswordSecurityChecker) ([]AccountSecurityReport, error) {
	allReports, err := r.CheckAllAccountPasswords(checker)
	if err != nil {
		return nil, err
	}

	vulnerable := make([]AccountSecurityReport, 0)

	for _, report := range allReports {
		// Include if common password or weak strength
		if report.Analysis.IsCommon || report.Analysis.Strength <= PasswordWeak {
			vulnerable = append(vulnerable, report)
		}
	}

	return vulnerable, nil
}

// GetSecurityStatistics returns overall security statistics
type SecurityStatistics struct {
	TotalAccounts      int                        `json:"total_accounts"`
	CommonPasswords    int                        `json:"common_passwords"`
	VeryWeakPasswords  int                        `json:"very_weak_passwords"`
	WeakPasswords      int                        `json:"weak_passwords"`
	MediumPasswords    int                        `json:"medium_passwords"`
	StrongPasswords    int                        `json:"strong_passwords"`
	VeryStrongPasswords int                       `json:"very_strong_passwords"`
	VulnerableAccounts []AccountSecurityReport    `json:"vulnerable_accounts,omitempty"`
}

func (r *AccountRepository) GetSecurityStatistics(checker *PasswordSecurityChecker, includeVulnerable bool) (*SecurityStatistics, error) {
	reports, err := r.CheckAllAccountPasswords(checker)
	if err != nil {
		return nil, err
	}

	stats := &SecurityStatistics{
		TotalAccounts: len(reports),
	}

	vulnerableAccounts := make([]AccountSecurityReport, 0)

	for _, report := range reports {
		if report.Analysis.IsCommon {
			stats.CommonPasswords++
		}

		switch report.Analysis.Strength {
		case PasswordVeryWeak:
			stats.VeryWeakPasswords++
		case PasswordWeak:
			stats.WeakPasswords++
		case PasswordMedium:
			stats.MediumPasswords++
		case PasswordStrong:
			stats.StrongPasswords++
		case PasswordVeryStrong:
			stats.VeryStrongPasswords++
		}

		// Add to vulnerable if common or weak
		if report.Analysis.IsCommon || report.Analysis.Strength <= PasswordWeak {
			vulnerableAccounts = append(vulnerableAccounts, report)
		}
	}

	if includeVulnerable {
		stats.VulnerableAccounts = vulnerableAccounts
	}

	return stats, nil
}
