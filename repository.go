package main

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
)

// AccountRepository handles account database operations
type AccountRepository struct{}

// Create creates a new account and returns its ID
func (r *AccountRepository) Create(account *Account) (int64, error) {
	// If expire is empty, let database default handle it
	var expireValue interface{}
	if account.Expire == "" {
		expireValue = nil // Will use database default
	} else {
		expireValue = account.Expire
	}

	result, err := DB.Exec(
		"INSERT INTO account (email, user, password, url, notes, expire) VALUES (?, ?, ?, ?, ?, COALESCE(?, datetime('now')))",
		account.Email,
		account.User,
		account.Password,
		account.URL,
		account.Notes,
		expireValue,
	)
	if err != nil {
		return 0, fmt.Errorf("error creating account: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error getting last insert id: %w", err)
	}

	// Create history entry
	if err := r.createHistory(id, account, nil); err != nil {
		return 0, err
	}

	return id, nil
}

// GetByID retrieves an account by ID with its tags and TOTP
func (r *AccountRepository) GetByID(id int64) (*Account, error) {
	account := &Account{}
	err := DB.QueryRow(
		"SELECT id, email, user, password, url, notes, expire FROM account WHERE id = ?",
		id,
	).Scan(&account.ID, &account.Email, &account.User, &account.Password, &account.URL, &account.Notes, &account.Expire)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("account not found")
		}
		return nil, fmt.Errorf("error getting account: %w", err)
	}

	// Load tags
	tags, err := r.getTagsForAccount(id)
	if err != nil {
		return nil, err
	}
	account.Tags = tags

	// Load TOTP if exists
	totp, err := r.getTOTPForAccount(id)
	if err == nil {
		account.TOTP = totp
	}

	return account, nil
}

// GetAll retrieves all accounts
func (r *AccountRepository) GetAll() ([]Account, error) {
	rows, err := DB.Query(
		"SELECT id, email, user, password, url, notes, expire FROM account ORDER BY id",
	)
	if err != nil {
		return nil, fmt.Errorf("error querying accounts: %w", err)
	}
	defer rows.Close()

	accounts := []Account{}
	for rows.Next() {
		account := Account{}
		err := rows.Scan(
			&account.ID,
			&account.Email,
			&account.User,
			&account.Password,
			&account.URL,
			&account.Notes,
			&account.Expire,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning account: %w", err)
		}

		// Load tags
		tags, err := r.getTagsForAccount(account.ID)
		// if err != nil {
		// 	return nil, err
		// }
		// account.Tags = tags
		if err == nil {
			account.Tags = tags
		}

		// Load TOTP if exists
		totp, err := r.getTOTPForAccount(account.ID)
		if err == nil {
			account.TOTP = totp
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

// Update updates an existing account
func (r *AccountRepository) Update(account *Account, changeReason *string) error {
	// First, close the current history entry
	_, err := DB.Exec(
		"UPDATE account_history SET valid_to = datetime('now') WHERE account_id = ? AND valid_to > datetime('now')",
		account.ID,
	)
	if err != nil {
		return fmt.Errorf("error closing history: %w", err)
	}

	// Update the account
	_, err = DB.Exec(
		"UPDATE account SET email = ?, user = ?, password = ?, url = ?, notes = ?, expire = ? WHERE id = ?",
		account.Email,
		account.User,
		account.Password,
		account.URL,
		account.Notes,
		account.Expire,
		account.ID,
	)
	if err != nil {
		return fmt.Errorf("error updating account: %w", err)
	}

	// Create new history entry
	return r.createHistory(account.ID, account, changeReason)
}

// Delete deletes an account by ID
func (r *AccountRepository) Delete(id int64) error {
	result, err := DB.Exec("DELETE FROM account WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("error deleting account: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("account not found")
	}

	return nil
}

// Search searches accounts by email, user, url, or notes
func (r *AccountRepository) Search(query string) ([]Account, error) {
	searchPattern := "%" + query + "%"
	rows, err := DB.Query(
		"SELECT id, email, user, password, url, notes, expire FROM account WHERE email LIKE ? OR user LIKE ? OR url LIKE ? OR notes LIKE ? ORDER BY id",
		searchPattern,
		searchPattern,
		searchPattern,
		searchPattern,
	)
	if err != nil {
		return nil, fmt.Errorf("error searching accounts: %w", err)
	}
	defer rows.Close()

	accounts := []Account{}
	for rows.Next() {
		account := Account{}
		err := rows.Scan(
			&account.ID,
			&account.Email,
			&account.User,
			&account.Password,
			&account.URL,
			&account.Notes,
			&account.Expire,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning account: %w", err)
		}

		// Load tags
		tags, err := r.getTagsForAccount(account.ID)
		if err != nil {
			return nil, err
		}
		account.Tags = tags

		// Load TOTP if exists
		totp, err := r.getTOTPForAccount(account.ID)
		if err == nil {
			account.TOTP = totp
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

// SearchByField searches accounts by a specific field
func (r *AccountRepository) SearchByField(field, query string) ([]Account, error) {
	// Validate field to prevent SQL injection
	validFields := map[string]bool{
		"email":    true,
		"user":     true,
		"url":      true,
		"notes":    true,
		"password": true,
		"expire":   true,
	}

	if !validFields[field] {
		return nil, fmt.Errorf("invalid search field: %s", field)
	}

	searchPattern := "%" + query + "%"
	sqlQuery := fmt.Sprintf(
		"SELECT id, email, user, password, url, notes, expire FROM account WHERE %s LIKE ? ORDER BY id",
		field,
	)

	rows, err := DB.Query(sqlQuery, searchPattern)
	if err != nil {
		return nil, fmt.Errorf("error searching accounts by %s: %w", field, err)
	}
	defer rows.Close()

	accounts := []Account{}
	for rows.Next() {
		account := Account{}
		err := rows.Scan(
			&account.ID,
			&account.Email,
			&account.User,
			&account.Password,
			&account.URL,
			&account.Notes,
			&account.Expire,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning account: %w", err)
		}

		// Load tags
		tags, err := r.getTagsForAccount(account.ID)
		if err != nil {
			return nil, err
		}
		account.Tags = tags

		// Load TOTP if exists
		totp, err := r.getTOTPForAccount(account.ID)
		if err == nil {
			account.TOTP = totp
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

// SearchFilters defines filters for fuzzy search
type SearchFilters struct {
	MinScore     float64  // Minimum similarity score (0.0 to 1.0)
	Tags         []string // Filter by tag names
	HasTOTP      *bool    // Filter by TOTP presence
	ExpireBefore *string  // Filter by expiration date
	ExpireAfter  *string  // Filter by expiration date
}

// SearchResult represents a search result with score
type SearchResult struct {
	Account Account
	Score   float64
	Matches map[string]string // Field name -> matched text
}

// FuzzySearch performs fuzzy search with scoring
func (r *AccountRepository) FuzzySearch(
	query string,
	filters *SearchFilters,
) ([]SearchResult, error) {
	if filters == nil {
		filters = &SearchFilters{MinScore: 0.0}
	}

	// Get all accounts
	accounts, err := r.GetAll()
	if err != nil {
		return nil, err
	}

	results := []SearchResult{}
	queryLower := strings.ToLower(query)

	for _, account := range accounts {
		score := 0.0
		matches := make(map[string]string)
		matchCount := 0

		// Search in each field
		fields := map[string]string{
			"user": account.User,
		}

		if account.Email != nil {
			fields["email"] = *account.Email
		}
		if account.URL != nil {
			fields["url"] = *account.URL
		}
		if account.Notes != nil {
			fields["notes"] = *account.Notes
		}

		// Calculate similarity for each field
		for fieldName, fieldValue := range fields {
			fieldLower := strings.ToLower(fieldValue)

			// Check for match
			if strings.Contains(fieldLower, queryLower) {
				fieldScore := calculateSimilarity(queryLower, fieldLower)
				if fieldScore > score {
					score = fieldScore
				}
				matches[fieldName] = fieldValue
				matchCount++
			} else {
				// Try fuzzy matching
				fieldScore := calculateLevenshteinSimilarity(queryLower, fieldLower)
				if fieldScore > 0.5 { // Threshold for fuzzy match
					if fieldScore > score {
						score = fieldScore
					}
					matches[fieldName] = fieldValue
					matchCount++
				}
			}
		}

		// Search in tags
		for _, tag := range account.Tags {
			tagLower := strings.ToLower(tag.Name)
			if strings.Contains(tagLower, queryLower) {
				tagScore := calculateSimilarity(queryLower, tagLower)
				if tagScore > score {
					score = tagScore
				}
				matches["tag"] = tag.Name
				matchCount++
			}
		}

		// Apply filters
		if filters.Tags != nil && len(filters.Tags) > 0 {
			hasRequiredTag := false
			for _, requiredTag := range filters.Tags {
				for _, accountTag := range account.Tags {
					if strings.EqualFold(accountTag.Name, requiredTag) {
						hasRequiredTag = true
						break
					}
				}
				if hasRequiredTag {
					break
				}
			}
			if !hasRequiredTag {
				continue
			}
		}

		if filters.HasTOTP != nil {
			hasTOTP := account.TOTP != nil
			if hasTOTP != *filters.HasTOTP {
				continue
			}
		}

		if filters.ExpireBefore != nil {
			if account.Expire > *filters.ExpireBefore {
				continue
			}
		}

		if filters.ExpireAfter != nil {
			if account.Expire < *filters.ExpireAfter {
				continue
			}
		}

		// Only include if score meets minimum
		if matchCount > 0 && score >= filters.MinScore {
			results = append(results, SearchResult{
				Account: account,
				Score:   score,
				Matches: matches,
			})
		}
	}

	// Sort by score (highest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}

// calculateSimilarity calculates similarity score between query and text
func calculateSimilarity(query, text string) float64 {
	// Simple scoring algorithm
	// 1.0 = exact match
	// 0.8 = starts with query
	// 0.6 = contains query

	if query == text {
		return 1.0
	}

	if strings.HasPrefix(text, query) {
		return 0.8
	}

	if strings.Contains(text, query) {
		// Calculate based on query length vs text length
		ratio := float64(len(query)) / float64(len(text))
		return 0.6 + (ratio * 0.2)
	}

	return 0.0
}

// calculateLevenshteinSimilarity calculates similarity using Levenshtein distance
func calculateLevenshteinSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	distance := levenshteinDistance(s1, s2)
	maxLen := len(s1)
	if len(s2) > maxLen {
		maxLen = len(s2)
	}

	if maxLen == 0 {
		return 1.0
	}

	similarity := 1.0 - (float64(distance) / float64(maxLen))
	if similarity < 0 {
		similarity = 0
	}

	return similarity
}

// levenshteinDistance calculates the Levenshtein distance between two strings
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	// Initialize first row and column
	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

func min(values ...int) int {
	minVal := values[0]
	for _, v := range values[1:] {
		if v < minVal {
			minVal = v
		}
	}
	return minVal
}

// AddTag adds a tag to an account
func (r *AccountRepository) AddTag(accountID, tagID int64) error {
	_, err := DB.Exec(
		"INSERT OR IGNORE INTO tag_association (account_id, tag_id) VALUES (?, ?)",
		accountID, tagID,
	)
	if err != nil {
		return fmt.Errorf("error adding tag to account: %w", err)
	}

	// Create history entry with valid_to set to far future (still valid)
	_, err = DB.Exec(
		"INSERT INTO tag_association_history (account_id, tag_id, valid_from, valid_to) VALUES (?, ?, datetime('now'), datetime('9999-12-31'))",
		accountID,
		tagID,
	)
	if err != nil {
		return fmt.Errorf("error creating tag association history: %w", err)
	}

	return nil
}

// AddTagByName adds a tag to an account by tag name (creates tag if doesn't exist)
func (r *AccountRepository) AddTagByName(accountID int64, tagName string) error {
	tagRepo := &TagRepository{}

	// Try to get existing tag by name
	tag, err := tagRepo.GetByName(tagName)
	if err != nil {
		// Tag doesn't exist, create it
		newTag := &Tag{Name: tagName}
		tagID, createErr := tagRepo.Create(newTag)
		if createErr != nil {
			return fmt.Errorf("error creating tag '%s': %w", tagName, createErr)
		}
		tag = &Tag{ID: tagID, Name: tagName}
	}

	// Add the tag to the account
	return r.AddTag(accountID, tag.ID)
}

// RemoveTag removes a tag from an account
func (r *AccountRepository) RemoveTag(accountID, tagID int64) error {
	// Close history entry by setting valid_to to now
	_, err := DB.Exec(
		"UPDATE tag_association_history SET valid_to = datetime('now') WHERE account_id = ? AND tag_id = ? AND valid_to > datetime('now')",
		accountID,
		tagID,
	)
	if err != nil {
		return fmt.Errorf("error closing tag association history: %w", err)
	}

	// Remove association
	_, err = DB.Exec(
		"DELETE FROM tag_association WHERE account_id = ? AND tag_id = ?",
		accountID, tagID,
	)
	if err != nil {
		return fmt.Errorf("error removing tag from account: %w", err)
	}

	return nil
}

// RemoveTagByName removes a tag from an account by tag name
func (r *AccountRepository) RemoveTagByName(accountID int64, tagName string) error {
	tagRepo := &TagRepository{}

	// Get tag by name
	tag, err := tagRepo.GetByName(tagName)
	if err != nil {
		return fmt.Errorf("tag '%s' not found: %w", tagName, err)
	}

	// Remove the tag from the account
	return r.RemoveTag(accountID, tag.ID)
}

// GetAccountsByTagName retrieves all accounts that have a specific tag
func (r *AccountRepository) GetAccountsByTagName(tagName string) ([]Account, error) {
	tagRepo := &TagRepository{}

	// Get tag by name
	tag, err := tagRepo.GetByName(tagName)
	if err != nil {
		return nil, fmt.Errorf("tag '%s' not found: %w", tagName, err)
	}

	// Query accounts with this tag
	rows, err := DB.Query(
		`SELECT a.id, a.email, a.user, a.password, a.url, a.notes, a.expire 
		 FROM account a
		 INNER JOIN tag_association ta ON a.id = ta.account_id
		 WHERE ta.tag_id = ?
		 ORDER BY a.id`,
		tag.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("error querying accounts by tag: %w", err)
	}
	defer rows.Close()

	accounts := []Account{}
	for rows.Next() {
		account := Account{}
		err := rows.Scan(
			&account.ID,
			&account.Email,
			&account.User,
			&account.Password,
			&account.URL,
			&account.Notes,
			&account.Expire,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning account: %w", err)
		}

		// Load tags
		tags, err := r.getTagsForAccount(account.ID)
		if err != nil {
			return nil, err
		}
		account.Tags = tags

		// Load TOTP if exists
		totp, err := r.getTOTPForAccount(account.ID)
		if err == nil {
			account.TOTP = totp
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

// HasTag checks if an account has a specific tag
func (r *AccountRepository) HasTag(accountID int64, tagName string) (bool, error) {
	tagRepo := &TagRepository{}

	// Get tag by name
	tag, err := tagRepo.GetByName(tagName)
	if err != nil {
		return false, nil // Tag doesn't exist, so account doesn't have it
	}

	// Check if association exists
	var count int
	err = DB.QueryRow(
		"SELECT COUNT(*) FROM tag_association WHERE account_id = ? AND tag_id = ?",
		accountID, tag.ID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("error checking tag association: %w", err)
	}

	return count > 0, nil
}

// GetHistory retrieves account history
func (r *AccountRepository) GetHistory(accountID int64) ([]AccountHistory, error) {
	rows, err := DB.Query(
		"SELECT history_id, account_id, email, user, password, url, notes, expire, valid_from, valid_to, change_reason FROM account_history WHERE account_id = ? ORDER BY valid_from DESC",
		accountID,
	)
	if err != nil {
		return nil, fmt.Errorf("error querying account history: %w", err)
	}
	defer rows.Close()

	history := []AccountHistory{}
	for rows.Next() {
		h := AccountHistory{}
		err := rows.Scan(
			&h.HistoryID,
			&h.AccountID,
			&h.Email,
			&h.User,
			&h.Password,
			&h.URL,
			&h.Notes,
			&h.Expire,
			&h.ValidFrom,
			&h.ValidTo,
			&h.ChangeReason,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning account history: %w", err)
		}
		history = append(history, h)
	}

	return history, nil
}

// Helper functions

func (r *AccountRepository) getTagsForAccount(accountID int64) ([]Tag, error) {
	rows, err := DB.Query(
		`SELECT t.id, t.name FROM tag t 
		 INNER JOIN tag_association ta ON t.id = ta.tag_id 
		 WHERE ta.account_id = ?`,
		accountID,
	)
	if err != nil {
		return nil, fmt.Errorf("error querying tags: %w", err)
	}
	defer rows.Close()

	tags := []Tag{}
	for rows.Next() {
		tag := Tag{}
		err := rows.Scan(&tag.ID, &tag.Name)
		if err != nil {
			return nil, fmt.Errorf("error scanning tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

func (r *AccountRepository) getTOTPForAccount(accountID int64) (*TOTP, error) {
	totp := &TOTP{}
	err := DB.QueryRow(
		"SELECT id, account_id, totp_seed, c_totp_seed, paillier_n, use_homomorphic FROM totp WHERE account_id = ?",
		accountID,
	).Scan(&totp.ID, &totp.AccountID, &totp.TOTPSeed, &totp.CTOTPSeed, &totp.PaillierN, &totp.UseHomomorphic)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("totp not found")
		}
		return nil, fmt.Errorf("error getting totp: %w", err)
	}
	return totp, nil
}

func (r *AccountRepository) createHistory(
	accountID int64,
	account *Account,
	changeReason *string,
) error {
	_, err := DB.Exec(
		"INSERT INTO account_history (account_id, email, user, password, url, notes, expire, valid_from, valid_to, change_reason) VALUES (?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('9999-12-31'), ?)",
		accountID,
		account.Email,
		account.User,
		account.Password,
		account.URL,
		account.Notes,
		account.Expire,
		changeReason,
	)
	if err != nil {
		return fmt.Errorf("error creating account history: %w", err)
	}
	return nil
}

// TagRepository handles tag database operations
type TagRepository struct{}

// Create creates a new tag
func (r *TagRepository) Create(tag *Tag) (int64, error) {
	result, err := DB.Exec("INSERT INTO tag (name) VALUES (?)", tag.Name)
	if err != nil {
		return 0, fmt.Errorf("error creating tag: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error getting last insert id: %w", err)
	}

	return id, nil
}

// GetByID retrieves a tag by ID
func (r *TagRepository) GetByID(id int64) (*Tag, error) {
	tag := &Tag{}
	err := DB.QueryRow("SELECT id, name FROM tag WHERE id = ?", id).Scan(&tag.ID, &tag.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tag not found")
		}
		return nil, fmt.Errorf("error getting tag: %w", err)
	}
	return tag, nil
}

// GetByName retrieves a tag by name
func (r *TagRepository) GetByName(name string) (*Tag, error) {
	tag := &Tag{}
	err := DB.QueryRow("SELECT id, name FROM tag WHERE name = ?", name).Scan(&tag.ID, &tag.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tag not found")
		}
		return nil, fmt.Errorf("error getting tag: %w", err)
	}
	return tag, nil
}

// GetAll retrieves all tags
func (r *TagRepository) GetAll() ([]Tag, error) {
	rows, err := DB.Query("SELECT id, name FROM tag ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("error querying tags: %w", err)
	}
	defer rows.Close()

	tags := []Tag{}
	for rows.Next() {
		tag := Tag{}
		err := rows.Scan(&tag.ID, &tag.Name)
		if err != nil {
			return nil, fmt.Errorf("error scanning tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// Update updates an existing tag
func (r *TagRepository) Update(tag *Tag) error {
	_, err := DB.Exec("UPDATE tag SET name = ? WHERE id = ?", tag.Name, tag.ID)
	if err != nil {
		return fmt.Errorf("error updating tag: %w", err)
	}
	return nil
}

// Delete deletes a tag by ID
func (r *TagRepository) Delete(id int64) error {
	result, err := DB.Exec("DELETE FROM tag WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("error deleting tag: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("tag not found")
	}

	return nil
}

// TOTPRepository handles TOTP database operations
type TOTPRepository struct{}

// Create creates a new TOTP for an account
func (r *TOTPRepository) Create(totp *TOTP) (int64, error) {
	// Validate homomorphic encryption consistency
	if totp.UseHomomorphic {
		if totp.CTOTPSeed == nil || totp.PaillierN == nil {
			return 0, fmt.Errorf("homomorphic TOTP requires both c_totp_seed and paillier_n")
		}
	} else {
		if totp.CTOTPSeed != nil || totp.PaillierN != nil {
			return 0, fmt.Errorf("standard TOTP should not have c_totp_seed or paillier_n set")
		}
	}

	result, err := DB.Exec(
		"INSERT INTO totp (account_id, totp_seed, c_totp_seed, paillier_n, use_homomorphic) VALUES (?, ?, ?, ?, ?)",
		totp.AccountID,
		totp.TOTPSeed,
		totp.CTOTPSeed,
		totp.PaillierN,
		totp.UseHomomorphic,
	)
	if err != nil {
		return 0, fmt.Errorf("error creating totp: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error getting last insert id: %w", err)
	}

	// Create history entry
	_, err = DB.Exec(
		"INSERT INTO totp_history (totp_id, totp_seed, c_totp_seed, paillier_n, use_homomorphic, valid_from, valid_to) VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('9999-12-31'))",
		id,
		totp.TOTPSeed,
		totp.CTOTPSeed,
		totp.PaillierN,
		totp.UseHomomorphic,
	)
	if err != nil {
		return 0, fmt.Errorf("error creating totp history: %w", err)
	}

	return id, nil
}

// GetByAccountID retrieves TOTP for an account
func (r *TOTPRepository) GetByAccountID(accountID int64) (*TOTP, error) {
	totp := &TOTP{}
	err := DB.QueryRow(
		"SELECT id, account_id, totp_seed, c_totp_seed, paillier_n, use_homomorphic FROM totp WHERE account_id = ?",
		accountID,
	).Scan(&totp.ID, &totp.AccountID, &totp.TOTPSeed, &totp.CTOTPSeed, &totp.PaillierN, &totp.UseHomomorphic)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("totp not found")
		}
		return nil, fmt.Errorf("error getting totp: %w", err)
	}
	return totp, nil
}

// Update updates a TOTP seed
func (r *TOTPRepository) Update(totp *TOTP) error {
	// Validate homomorphic encryption consistency
	if totp.UseHomomorphic {
		if totp.CTOTPSeed == nil || totp.PaillierN == nil {
			return fmt.Errorf("homomorphic TOTP requires both c_totp_seed and paillier_n")
		}
	} else {
		if totp.CTOTPSeed != nil || totp.PaillierN != nil {
			return fmt.Errorf("standard TOTP should not have c_totp_seed or paillier_n set")
		}
	}

	// Close current history entry
	_, err := DB.Exec(
		"UPDATE totp_history SET valid_to = datetime('now') WHERE totp_id = ? AND valid_to > datetime('now')",
		totp.ID,
	)
	if err != nil {
		return fmt.Errorf("error closing totp history: %w", err)
	}

	// Update the totp
	_, err = DB.Exec(
		"UPDATE totp SET totp_seed = ?, c_totp_seed = ?, paillier_n = ?, use_homomorphic = ? WHERE id = ?",
		totp.TOTPSeed,
		totp.CTOTPSeed,
		totp.PaillierN,
		totp.UseHomomorphic,
		totp.ID,
	)
	if err != nil {
		return fmt.Errorf("error updating totp: %w", err)
	}

	// Create new history entry
	_, err = DB.Exec(
		"INSERT INTO totp_history (totp_id, totp_seed, c_totp_seed, paillier_n, use_homomorphic, valid_from, valid_to) VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('9999-12-31'))",
		totp.ID,
		totp.TOTPSeed,
		totp.CTOTPSeed,
		totp.PaillierN,
		totp.UseHomomorphic,
	)
	if err != nil {
		return fmt.Errorf("error creating totp history: %w", err)
	}

	return nil
}

// Delete deletes a TOTP by ID
func (r *TOTPRepository) Delete(id int64) error {
	result, err := DB.Exec("DELETE FROM totp WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("error deleting totp: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("totp not found")
	}

	return nil
}

// GetHistory retrieves TOTP history
func (r *TOTPRepository) GetHistory(totpID int64) ([]TOTPHistory, error) {
	rows, err := DB.Query(
		"SELECT history_id, totp_id, totp_seed, c_totp_seed, paillier_n, use_homomorphic, valid_from, valid_to FROM totp_history WHERE totp_id = ? ORDER BY valid_from DESC",
		totpID,
	)
	if err != nil {
		return nil, fmt.Errorf("error querying totp history: %w", err)
	}
	defer rows.Close()

	history := []TOTPHistory{}
	for rows.Next() {
		h := TOTPHistory{}
		err := rows.Scan(
			&h.HistoryID,
			&h.TOTPID,
			&h.TOTPSeed,
			&h.CTOTPSeed,
			&h.PaillierN,
			&h.UseHomomorphic,
			&h.ValidFrom,
			&h.ValidTo,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning totp history: %w", err)
		}
		history = append(history, h)
	}

	return history, nil
}
