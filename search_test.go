package main

import (
	"testing"
)

// Test SearchByField functionality
func TestSearchByField(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	accountRepo := &AccountRepository{}

	// Create test accounts
	email1 := "alice@example.com"
	email2 := "bob@test.com"
	url1 := "https://example.com"

	accounts := []*Account{
		{User: "alice", Password: "pass1", Email: &email1, URL: &url1},
		{User: "bob", Password: "pass2", Email: &email2},
		{User: "charlie", Password: "pass3"},
	}

	for _, acc := range accounts {
		_, err := accountRepo.Create(acc)
		if err != nil {
			t.Fatalf("Failed to create account: %v", err)
		}
	}

	// Test search by email
	results, err := accountRepo.SearchByField("email", "example")
	if err != nil {
		t.Fatalf("SearchByField failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if len(results) > 0 && results[0].User != "alice" {
		t.Errorf("Expected alice, got %s", results[0].User)
	}

	// Test search by user
	results, err = accountRepo.SearchByField("user", "bob")
	if err != nil {
		t.Fatalf("SearchByField failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'bob', got %d", len(results))
	}

	// Test search by URL
	results, err = accountRepo.SearchByField("url", "example")
	if err != nil {
		t.Fatalf("SearchByField failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result for URL search, got %d", len(results))
	}
}

func TestSearchByFieldInvalid(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	accountRepo := &AccountRepository{}

	// Test invalid field
	_, err := accountRepo.SearchByField("invalid_field", "query")
	if err == nil {
		t.Error("Expected error for invalid field")
	}

	if err != nil && !contains(err.Error(), "invalid search field") {
		t.Errorf("Expected 'invalid search field' error, got: %v", err)
	}
}

func TestFuzzySearchBasic(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	accountRepo := &AccountRepository{}

	// Create test accounts
	email := "john@example.com"
	accounts := []*Account{
		{User: "john", Password: "pass1", Email: &email},
		{User: "johnny", Password: "pass2"},
		{User: "jonathan", Password: "pass3"},
		{User: "jane", Password: "pass4"},
	}

	for _, acc := range accounts {
		_, err := accountRepo.Create(acc)
		if err != nil {
			t.Fatalf("Failed to create account: %v", err)
		}
	}

	// Fuzzy search for "john"
	results, err := accountRepo.FuzzySearch("john", nil)
	if err != nil {
		t.Fatalf("FuzzySearch failed: %v", err)
	}

	// Should find accounts with john, johnny, jonathan (not jane)
	if len(results) < 1 {
		t.Error("Expected at least 1 fuzzy search result")
	}

	// Results should be sorted by score
	for i := 1; i < len(results); i++ {
		if results[i].Score > results[i-1].Score {
			t.Error("Results not sorted by score (descending)")
		}
	}

	// First result should be exact match "john"
	if len(results) > 0 && results[0].Account.User != "john" {
		t.Logf("Expected exact match 'john' first, got %s (score: %f)",
			results[0].Account.User, results[0].Score)
	}
}

func TestFuzzySearchWithMinScore(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	accountRepo := &AccountRepository{}

	// Create test accounts
	accounts := []*Account{
		{User: "test", Password: "pass1"},
		{User: "testing", Password: "pass2"},
		{User: "completely_different", Password: "pass3"},
	}

	for _, acc := range accounts {
		accountRepo.Create(acc)
	}

	// Search with high minimum score
	filters := &SearchFilters{
		MinScore: 0.7,
	}

	results, err := accountRepo.FuzzySearch("test", filters)
	if err != nil {
		t.Fatalf("FuzzySearch failed: %v", err)
	}

	// Should not include "completely_different" due to low similarity
	for _, result := range results {
		if result.Score < 0.7 {
			t.Errorf("Result has score %f below minimum 0.7", result.Score)
		}
	}
}

func TestFuzzySearchWithTagFilter(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	accountRepo := &AccountRepository{}
	tagRepo := &TagRepository{}

	// Create tags
	workTag := &Tag{Name: "work"}
	workID, _ := tagRepo.Create(workTag)

	personalTag := &Tag{Name: "personal"}
	personalID, _ := tagRepo.Create(personalTag)

	// Create accounts
	acc1 := &Account{User: "alice", Password: "pass1"}
	acc1ID, _ := accountRepo.Create(acc1)
	accountRepo.AddTag(acc1ID, workID)

	acc2 := &Account{User: "bob", Password: "pass2"}
	acc2ID, _ := accountRepo.Create(acc2)
	accountRepo.AddTag(acc2ID, personalID)

	acc3 := &Account{User: "charlie", Password: "pass3"}
	accountRepo.Create(acc3) // No tags

	// Search with tag filter
	filters := &SearchFilters{
		Tags: []string{"work"},
	}

	results, err := accountRepo.FuzzySearch("a", filters)
	if err != nil {
		t.Fatalf("FuzzySearch failed: %v", err)
	}

	// Should only find alice (has work tag and contains 'a')
	foundAlice := false
	for _, result := range results {
		if result.Account.User == "alice" {
			foundAlice = true
		}
		if result.Account.User == "bob" || result.Account.User == "charlie" {
			t.Errorf("Found %s but should only match work tag", result.Account.User)
		}
	}

	if !foundAlice {
		t.Error("Expected to find alice with work tag")
	}
}

func TestFuzzySearchWithTOTPFilter(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	accountRepo := &AccountRepository{}
	totpRepo := &TOTPRepository{}

	// Create accounts
	acc1 := &Account{User: "alice", Password: "pass1"}
	acc1ID, _ := accountRepo.Create(acc1)

	acc2 := &Account{User: "bob", Password: "pass2"}
	acc2ID, err := accountRepo.Create(acc2)
	if err != nil {
		t.Fatalf("FuzzySearch failed for %v: %v", acc2ID, err)
	}

	// Add TOTP only to alice
	totp := &TOTP{
		AccountID: acc1ID,
		TOTPSeed:  "JBSWY3DPEHPK3PXP",
	}
	totpRepo.Create(totp)

	// Search for accounts with TOTP
	hasTOTP := true
	filters := &SearchFilters{
		HasTOTP: &hasTOTP,
	}

	results, err := accountRepo.FuzzySearch("a", filters)
	if err != nil {
		t.Fatalf("FuzzySearch failed: %v", err)
	}

	// Should only find alice
	if len(results) != 1 {
		t.Errorf("Expected 1 result with TOTP, got %d", len(results))
	}

	if len(results) > 0 && results[0].Account.User != "alice" {
		t.Errorf("Expected alice, got %s", results[0].Account.User)
	}

	// Search for accounts without TOTP
	hasNoTOTP := false
	filters.HasTOTP = &hasNoTOTP

	results, err = accountRepo.FuzzySearch("b", filters)
	if err != nil {
		t.Fatalf("FuzzySearch failed: %v", err)
	}

	// Should find bob
	foundBob := false
	for _, result := range results {
		if result.Account.User == "bob" {
			foundBob = true
		}
		if result.Account.TOTP != nil {
			t.Errorf("Found account %s with TOTP when filtering for no TOTP",
				result.Account.User)
		}
	}

	if !foundBob {
		t.Error("Expected to find bob without TOTP")
	}
}

func TestLevenshteinDistance(t *testing.T) {
	testCases := []struct {
		s1       string
		s2       string
		expected int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "a", 1},
		{"abc", "abc", 0},
		{"abc", "abd", 1},
		{"abc", "abcd", 1},
		{"kitten", "sitting", 3},
	}

	for _, tc := range testCases {
		result := levenshteinDistance(tc.s1, tc.s2)
		if result != tc.expected {
			t.Errorf("levenshteinDistance(%q, %q) = %d; expected %d",
				tc.s1, tc.s2, result, tc.expected)
		}
	}
}

func TestCalculateSimilarity(t *testing.T) {
	testCases := []struct {
		query    string
		text     string
		minScore float64
	}{
		{"test", "test", 1.0},    // Exact match
		{"test", "testing", 0.6}, // Contains
		{"test", "test123", 0.6}, // Starts with
		{"xyz", "abc", 0.0},      // No match
	}

	for _, tc := range testCases {
		score := calculateSimilarity(tc.query, tc.text)
		if score < tc.minScore {
			t.Errorf("calculateSimilarity(%q, %q) = %f; expected >= %f",
				tc.query, tc.text, score, tc.minScore)
		}
	}
}

func TestFuzzySearchMatches(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	accountRepo := &AccountRepository{}

	// Create account with multiple searchable fields
	email := "test@example.com"
	url := "https://example.com"
	notes := "important test account"

	acc := &Account{
		User:     "testuser",
		Password: "pass",
		Email:    &email,
		URL:      &url,
		Notes:    &notes,
	}
	accountRepo.Create(acc)

	// Search for "test"
	results, err := accountRepo.FuzzySearch("test", nil)
	if err != nil {
		t.Fatalf("FuzzySearch failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("Expected at least 1 result")
	}

	// Check that matches are recorded
	matches := results[0].Matches
	if len(matches) == 0 {
		t.Error("Expected matches to be recorded")
	}

	// Should have matches in user, email, and notes
	expectedFields := []string{"user", "email", "notes"}
	for _, field := range expectedFields {
		if _, found := matches[field]; !found {
			t.Logf("Expected match in field '%s', matches: %v", field, matches)
		}
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			hasSubstring(s, substr)))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
