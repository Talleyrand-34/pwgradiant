package main

import (
	"os"
	"testing"
)

// Setup helper function for tests
func setupTestDB(t *testing.T) func() {
	dbPath := "test_repo.db"
	encryptionKey := "test-key-repo"

	os.Remove(dbPath)

	err := InitDB(dbPath, encryptionKey)
	if err != nil {
		t.Fatalf("Failed to init DB: %v", err)
	}

	err = CreateSchema()
	if err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	// Return cleanup function
	return func() {
		CloseDB()
		os.Remove(dbPath)
	}
}

func TestAccountRepositoryCreate(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	repo := &AccountRepository{}

	email := "test@example.com"
	url := "https://example.com"
	notes := "Test notes"

	account := &Account{
		Email:    &email,
		User:     "testuser",
		Password: "testpass",
		URL:      &url,
		Notes:    &notes,
	}

	id, err := repo.Create(account)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if id <= 0 {
		t.Error("Expected positive ID")
	}
}

func TestAccountRepositoryGetByID(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	repo := &AccountRepository{}

	// Create account
	account := &Account{
		User:     "testuser",
		Password: "testpass",
	}

	id, err := repo.Create(account)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Get by ID
	retrieved, err := repo.GetByID(id)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved.ID != id {
		t.Errorf("ID mismatch: got %d, want %d", retrieved.ID, id)
	}
	if retrieved.User != account.User {
		t.Errorf("User mismatch: got %s, want %s", retrieved.User, account.User)
	}
}

func TestAccountRepositoryGetByIDNotFound(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	repo := &AccountRepository{}

	_, err := repo.GetByID(999)
	if err == nil {
		t.Error("Expected error for non-existent account")
	}
}

func TestAccountRepositoryGetAll(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	repo := &AccountRepository{}

	// Create multiple accounts
	for i := 0; i < 3; i++ {
		account := &Account{
			User:     "user" + string(rune('0'+i)),
			Password: "pass" + string(rune('0'+i)),
		}
		_, err := repo.Create(account)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	// Get all
	accounts, err := repo.GetAll()
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	if len(accounts) != 3 {
		t.Errorf("Expected 3 accounts, got %d", len(accounts))
	}
}

func TestAccountRepositoryUpdate(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	repo := &AccountRepository{}

	// Create account
	account := &Account{
		User:     "testuser",
		Password: "oldpass",
	}

	id, err := repo.Create(account)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Update account
	account.ID = id
	account.Password = "newpass"
	changeReason := "Testing update"

	err = repo.Update(account, &changeReason)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	retrieved, err := repo.GetByID(id)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved.Password != "newpass" {
		t.Errorf("Password not updated: got %s", retrieved.Password)
	}
}

func TestAccountRepositoryDelete(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	repo := &AccountRepository{}

	// Create account
	account := &Account{
		User:     "testuser",
		Password: "testpass",
	}

	id, err := repo.Create(account)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Delete
	err = repo.Delete(id)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deleted
	_, err = repo.GetByID(id)
	if err == nil {
		t.Error("Expected error for deleted account")
	}
}

func TestAccountRepositorySearch(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	repo := &AccountRepository{}

	// Create accounts with searchable data
	email1 := "john@example.com"
	email2 := "jane@test.com"

	accounts := []*Account{
		{User: "john_doe", Password: "pass1", Email: &email1},
		{User: "jane_smith", Password: "pass2", Email: &email2},
	}

	for _, acc := range accounts {
		_, err := repo.Create(acc)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	// Search
	results, err := repo.Search("john")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if len(results) > 0 && results[0].User != "john_doe" {
		t.Errorf("Wrong search result: got %s", results[0].User)
	}
}

func TestAccountRepositoryAddTag(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	accountRepo := &AccountRepository{}
	tagRepo := &TagRepository{}

	// Create account
	account := &Account{User: "testuser", Password: "testpass"}
	accountID, err := accountRepo.Create(account)
	if err != nil {
		t.Fatalf("Create account failed: %v", err)
	}

	// Create tag
	tag := &Tag{Name: "work"}
	tagID, err := tagRepo.Create(tag)
	if err != nil {
		t.Fatalf("Create tag failed: %v", err)
	}

	// Add tag to account
	err = accountRepo.AddTag(accountID, tagID)
	if err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}

	// Verify tag was added
	retrieved, err := accountRepo.GetByID(accountID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if len(retrieved.Tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(retrieved.Tags))
	}

	if len(retrieved.Tags) > 0 && retrieved.Tags[0].Name != "work" {
		t.Errorf("Wrong tag name: got %s", retrieved.Tags[0].Name)
	}
}

func TestAccountRepositoryRemoveTag(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	accountRepo := &AccountRepository{}
	tagRepo := &TagRepository{}

	// Create account and tag
	account := &Account{User: "testuser", Password: "testpass"}
	accountID, _ := accountRepo.Create(account)

	tag := &Tag{Name: "work"}
	tagID, _ := tagRepo.Create(tag)

	// Add tag
	accountRepo.AddTag(accountID, tagID)

	// Remove tag
	err := accountRepo.RemoveTag(accountID, tagID)
	if err != nil {
		t.Fatalf("RemoveTag failed: %v", err)
	}

	// Verify tag was removed
	retrieved, err := accountRepo.GetByID(accountID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if len(retrieved.Tags) != 0 {
		t.Errorf("Expected 0 tags, got %d", len(retrieved.Tags))
	}
}

func TestAccountRepositoryGetHistory(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	repo := &AccountRepository{}

	// Create account
	account := &Account{User: "testuser", Password: "pass1"}
	id, _ := repo.Create(account)

	// Update account (creates history)
	account.ID = id
	account.Password = "pass2"
	changeReason := "First update"
	repo.Update(account, &changeReason)

	account.Password = "pass3"
	changeReason2 := "Second update"
	repo.Update(account, &changeReason2)

	// Get history
	history, err := repo.GetHistory(id)
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}

	if len(history) < 3 {
		t.Errorf("Expected at least 3 history entries, got %d", len(history))
	}
}

func TestTagRepositoryCreate(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	repo := &TagRepository{}

	tag := &Tag{Name: "work"}
	id, err := repo.Create(tag)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if id <= 0 {
		t.Error("Expected positive ID")
	}
}

func TestTagRepositoryGetByID(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	repo := &TagRepository{}

	tag := &Tag{Name: "personal"}
	id, _ := repo.Create(tag)

	retrieved, err := repo.GetByID(id)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved.Name != "personal" {
		t.Errorf("Name mismatch: got %s", retrieved.Name)
	}
}

func TestTagRepositoryGetByName(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	repo := &TagRepository{}

	tag := &Tag{Name: "urgent"}
	repo.Create(tag)

	retrieved, err := repo.GetByName("urgent")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}

	if retrieved.Name != "urgent" {
		t.Errorf("Name mismatch: got %s", retrieved.Name)
	}
}

func TestTagRepositoryGetAll(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	repo := &TagRepository{}

	tags := []string{"work", "personal", "urgent"}
	for _, name := range tags {
		repo.Create(&Tag{Name: name})
	}

	allTags, err := repo.GetAll()
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	if len(allTags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(allTags))
	}
}

func TestTagRepositoryUpdate(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	repo := &TagRepository{}

	tag := &Tag{Name: "oldname"}
	id, _ := repo.Create(tag)

	tag.ID = id
	tag.Name = "newname"

	err := repo.Update(tag)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	retrieved, _ := repo.GetByID(id)
	if retrieved.Name != "newname" {
		t.Errorf("Name not updated: got %s", retrieved.Name)
	}
}

func TestTagRepositoryDelete(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	repo := &TagRepository{}

	tag := &Tag{Name: "todelete"}
	id, _ := repo.Create(tag)

	err := repo.Delete(id)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = repo.GetByID(id)
	if err == nil {
		t.Error("Expected error for deleted tag")
	}
}

func TestTOTPRepositoryCreate(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	accountRepo := &AccountRepository{}
	totpRepo := &TOTPRepository{}

	// Create account first
	account := &Account{User: "testuser", Password: "testpass"}
	accountID, _ := accountRepo.Create(account)

	// Create TOTP
	totp := &TOTP{
		AccountID: accountID,
		TOTPSeed:  "JBSWY3DPEHPK3PXP",
	}

	id, err := totpRepo.Create(totp)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if id <= 0 {
		t.Error("Expected positive ID")
	}
}

func TestTOTPRepositoryGetByAccountID(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	accountRepo := &AccountRepository{}
	totpRepo := &TOTPRepository{}

	account := &Account{User: "testuser", Password: "testpass"}
	accountID, _ := accountRepo.Create(account)

	totp := &TOTP{AccountID: accountID, TOTPSeed: "SEED123"}
	totpRepo.Create(totp)

	retrieved, err := totpRepo.GetByAccountID(accountID)
	if err != nil {
		t.Fatalf("GetByAccountID failed: %v", err)
	}

	if retrieved.TOTPSeed != "SEED123" {
		t.Errorf("Seed mismatch: got %s", retrieved.TOTPSeed)
	}
}

func TestTOTPRepositoryUpdate(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	accountRepo := &AccountRepository{}
	totpRepo := &TOTPRepository{}

	account := &Account{User: "testuser", Password: "testpass"}
	accountID, _ := accountRepo.Create(account)

	totp := &TOTP{AccountID: accountID, TOTPSeed: "OLDSEED"}
	id, _ := totpRepo.Create(totp)

	totp.ID = id
	totp.TOTPSeed = "NEWSEED"

	err := totpRepo.Update(totp)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	retrieved, _ := totpRepo.GetByAccountID(accountID)
	if retrieved.TOTPSeed != "NEWSEED" {
		t.Errorf("Seed not updated: got %s", retrieved.TOTPSeed)
	}
}

func TestTOTPRepositoryDelete(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	accountRepo := &AccountRepository{}
	totpRepo := &TOTPRepository{}

	account := &Account{User: "testuser", Password: "testpass"}
	accountID, _ := accountRepo.Create(account)

	totp := &TOTP{AccountID: accountID, TOTPSeed: "SEED"}
	id, _ := totpRepo.Create(totp)

	err := totpRepo.Delete(id)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = totpRepo.GetByAccountID(accountID)
	if err == nil {
		t.Error("Expected error for deleted TOTP")
	}
}

func TestTOTPRepositoryGetHistory(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	accountRepo := &AccountRepository{}
	totpRepo := &TOTPRepository{}

	account := &Account{User: "testuser", Password: "testpass"}
	accountID, _ := accountRepo.Create(account)

	totp := &TOTP{AccountID: accountID, TOTPSeed: "SEED1"}
	id, _ := totpRepo.Create(totp)

	// Update to create history
	totp.ID = id
	totp.TOTPSeed = "SEED2"
	totpRepo.Update(totp)

	history, err := totpRepo.GetHistory(id)
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}

	if len(history) < 2 {
		t.Errorf("Expected at least 2 history entries, got %d", len(history))
	}
}

func TestForeignKeyConstraints(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	accountRepo := &AccountRepository{}

	// Try to add tag to non-existent account
	err := accountRepo.AddTag(999, 1)
	if err == nil {
		t.Error("Expected foreign key constraint error")
	}
}

func TestTOTPHomomorphicFields(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	accountRepo := &AccountRepository{}
	totpRepo := &TOTPRepository{}

	// Create account
	account := &Account{
		User:     "testuser",
		Password: "testpass",
	}
	accountID, _ := accountRepo.Create(account)

	// Create TOTP with proper homomorphic encryption
	seed := "JBSWY3DPEHPK3PXP"
	privateKey, err := CreatePaillierKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to create key pair: %v", err)
	}

	homomorphicData, err := EncryptTOTPSeed(seed, privateKey)
	if err != nil {
		t.Fatalf("Failed to encrypt seed: %v", err)
	}

	totp := &TOTP{
		AccountID:      accountID,
		TOTPSeed:       homomorphicData.PlaintextSeed,
		CTOTPSeed:      &homomorphicData.EncryptedSeed,
		PaillierN:      &homomorphicData.PaillierN,
		UseHomomorphic: true,
	}

	totpID, err := totpRepo.Create(totp)
	if err != nil {
		t.Fatalf("Create TOTP failed: %v", err)
	}

	// Retrieve and verify
	retrieved, err := totpRepo.GetByAccountID(accountID)
	if err != nil {
		t.Fatalf("GetByAccountID failed: %v", err)
	}

	if !retrieved.UseHomomorphic {
		t.Error("UseHomomorphic flag not set")
	}

	if retrieved.CTOTPSeed == nil {
		t.Error("Encrypted seed is nil")
	}

	if retrieved.PaillierN == nil {
		t.Error("Paillier N is nil")
	}

	// Verify we can generate TOTP from the encrypted data
	code, err := GenerateHomomorphicTOTP(*retrieved.CTOTPSeed, *retrieved.PaillierN, privateKey)
	if err != nil {
		t.Fatalf("Failed to generate homomorphic TOTP: %v", err)
	}

	if len(code) != 6 {
		t.Errorf("Expected 6-digit code, got %d digits", len(code))
	}

	// Check history
	history, err := totpRepo.GetHistory(totpID)
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}

	if len(history) != 1 {
		t.Errorf("Expected 1 history entry, got %d", len(history))
	}

	if !history[0].UseHomomorphic {
		t.Error("History UseHomomorphic flag not set")
	}
}

// func TestTOTPUpdateHomomorphicToStandard(t *testing.T) {
// 	cleanup := setupTestDB(t)
// 	defer cleanup()
//
// 	accountRepo := &AccountRepository{}
// 	totpRepo := &TOTPRepository{}
//
// 	// Create account
// 	account := &Account{
// 		User:     "testuser",
// 		Password: "testpass",
// 	}
// 	accountID, _ := accountRepo.Create(account)
//
// 	// Create homomorphic TOTP properly using the encryption functions
// 	seed := "JBSWY3DPEHPK3PXP"
// 	privateKey, err := CreatePaillierKeyPair(2048)
// 	if err != nil {
// 		t.Fatalf("Failed to create key pair: %v", err)
// 	}
//
// 	// Encrypt the seed properly
// 	homomorphicData, err := EncryptTOTPSeed(seed, privateKey)
// 	if err != nil {
// 		t.Fatalf("Failed to encrypt seed: %v", err)
// 	}
//
// 	// Create TOTP with proper homomorphic encryption
// 	totp := &TOTP{
// 		AccountID:      accountID,
// 		TOTPSeed:       homomorphicData.PlaintextSeed,
// 		CTOTPSeed:      &homomorphicData.EncryptedSeed,
// 		PaillierN:      &homomorphicData.PaillierN,
// 		UseHomomorphic: true,
// 	}
//
// 	totpID, err := totpRepo.Create(totp)
// 	if err != nil {
// 		t.Fatalf("Create failed: %v", err)
// 	}
//
// 	// Update to standard TOTP (remove homomorphic encryption)
// 	totp.ID = totpID
// 	totp.CTOTPSeed = nil
// 	totp.PaillierN = nil
// 	totp.UseHomomorphic = false
//
// 	err = totpRepo.Update(totp)
// 	if err != nil {
// 		t.Fatalf("Update failed: %v", err)
// 	}
//
// 	// Verify update
// 	retrieved, err := totpRepo.GetByAccountID(accountID)
// 	if err != nil {
// 		t.Fatalf("GetByAccountID failed: %v", err)
// 	}
//
// 	if retrieved.UseHomomorphic {
// 		t.Error("Expected UseHomomorphic to be false")
// 	}
//
// 	if retrieved.CTOTPSeed != nil {
// 		t.Error("Expected CTOTPSeed to be nil")
// 	}
//
// 	// Verify history
// 	history, err := totpRepo.GetHistory(totpID)
// 	if err != nil {
// 		t.Fatalf("GetHistory failed: %v", err)
// 	}
//
// 	if len(history) != 2 {
// 		t.Fatalf("Expected 2 history entries, got %d", len(history))
// 	}
//
// 	// Most recent should be standard (index 0, ordered by valid_from DESC)
// 	if history[0].UseHomomorphic {
// 		t.Errorf(
// 			"Latest history entry should be standard TOTP, got UseHomomorphic=%v",
// 			history[0].UseHomomorphic,
// 		)
// 	}
//
// 	// Oldest should be homomorphic (index 1)
// 	if !history[1].UseHomomorphic {
// 		t.Errorf(
// 			"Original history entry should be homomorphic, got UseHomomorphic=%v",
// 			history[1].UseHomomorphic,
// 		)
// 	}
// }

func TestAddTagByName(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	accountRepo := &AccountRepository{}

	// Create account
	account := &Account{
		User:     "testuser",
		Password: "testpass",
	}
	accountID, _ := accountRepo.Create(account)

	// Add tag by name (should create tag automatically)
	err := accountRepo.AddTagByName(accountID, "work")
	if err != nil {
		t.Fatalf("AddTagByName failed: %v", err)
	}

	// Verify tag was created and associated
	retrieved, _ := accountRepo.GetByID(accountID)

	if len(retrieved.Tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(retrieved.Tags))
	}

	if retrieved.Tags[0].Name != "work" {
		t.Errorf("Expected tag 'work', got '%s'", retrieved.Tags[0].Name)
	}
}

func TestAddTagByNameExisting(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	accountRepo := &AccountRepository{}
	tagRepo := &TagRepository{}

	// Create tag first
	tag := &Tag{Name: "existing"}
	tagID, _ := tagRepo.Create(tag)

	// Create account
	account := &Account{
		User:     "testuser",
		Password: "testpass",
	}
	accountID, _ := accountRepo.Create(account)

	// Add existing tag by name (should not create duplicate)
	err := accountRepo.AddTagByName(accountID, "existing")
	if err != nil {
		t.Fatalf("AddTagByName failed: %v", err)
	}

	// Verify tag was associated with correct ID
	retrieved, _ := accountRepo.GetByID(accountID)

	if len(retrieved.Tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(retrieved.Tags))
	}

	if retrieved.Tags[0].ID != tagID {
		t.Errorf("Expected tag ID %d, got %d", tagID, retrieved.Tags[0].ID)
	}
}

func TestRemoveTagByName(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	accountRepo := &AccountRepository{}

	// Create account
	account := &Account{
		User:     "testuser",
		Password: "testpass",
	}
	accountID, _ := accountRepo.Create(account)

	// Add tag by name
	accountRepo.AddTagByName(accountID, "work")

	// Remove tag by name
	err := accountRepo.RemoveTagByName(accountID, "work")
	if err != nil {
		t.Fatalf("RemoveTagByName failed: %v", err)
	}

	// Verify tag was removed
	retrieved, _ := accountRepo.GetByID(accountID)

	if len(retrieved.Tags) != 0 {
		t.Errorf("Expected 0 tags, got %d", len(retrieved.Tags))
	}
}

func TestRemoveTagByNameNotFound(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	accountRepo := &AccountRepository{}

	// Create account
	account := &Account{
		User:     "testuser",
		Password: "testpass",
	}
	accountID, _ := accountRepo.Create(account)

	// Try to remove non-existent tag
	err := accountRepo.RemoveTagByName(accountID, "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent tag")
	}
}

func TestGetAccountsByTagName(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	accountRepo := &AccountRepository{}

	// Create multiple accounts
	for i := 0; i < 3; i++ {
		account := &Account{
			User:     "user" + string(rune('0'+i)),
			Password: "pass",
		}
		accountID, _ := accountRepo.Create(account)

		// Add "work" tag to first 2 accounts
		if i < 2 {
			accountRepo.AddTagByName(accountID, "work")
		}
	}

	// Get accounts by tag name
	accounts, err := accountRepo.GetAccountsByTagName("work")
	if err != nil {
		t.Fatalf("GetAccountsByTagName failed: %v", err)
	}

	if len(accounts) != 2 {
		t.Errorf("Expected 2 accounts, got %d", len(accounts))
	}

	// Verify all returned accounts have the tag
	for _, account := range accounts {
		hasWorkTag := false
		for _, tag := range account.Tags {
			if tag.Name == "work" {
				hasWorkTag = true
				break
			}
		}
		if !hasWorkTag {
			t.Errorf("Account %d missing 'work' tag", account.ID)
		}
	}
}

func TestGetAccountsByTagNameNotFound(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	accountRepo := &AccountRepository{}

	// Try to get accounts by non-existent tag
	_, err := accountRepo.GetAccountsByTagName("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent tag")
	}
}

func TestHasTag(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	accountRepo := &AccountRepository{}

	// Create account
	account := &Account{
		User:     "testuser",
		Password: "testpass",
	}
	accountID, _ := accountRepo.Create(account)

	// Add tag
	accountRepo.AddTagByName(accountID, "work")

	// Check if account has tag
	hasTag, err := accountRepo.HasTag(accountID, "work")
	if err != nil {
		t.Fatalf("HasTag failed: %v", err)
	}

	if !hasTag {
		t.Error("Expected account to have 'work' tag")
	}

	// Check for tag account doesn't have
	hasTag, err = accountRepo.HasTag(accountID, "personal")
	if err != nil {
		t.Fatalf("HasTag failed: %v", err)
	}

	if hasTag {
		t.Error("Expected account to not have 'personal' tag")
	}
}

func TestMultipleTagsWorkflow(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	accountRepo := &AccountRepository{}

	// Create account
	account := &Account{
		User:     "testuser",
		Password: "testpass",
	}
	accountID, _ := accountRepo.Create(account)

	// Add multiple tags by name
	tags := []string{"work", "important", "finance"}
	for _, tagName := range tags {
		err := accountRepo.AddTagByName(accountID, tagName)
		if err != nil {
			t.Fatalf("Failed to add tag '%s': %v", tagName, err)
		}
	}

	// Verify all tags were added
	retrieved, _ := accountRepo.GetByID(accountID)

	if len(retrieved.Tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(retrieved.Tags))
	}

	// Remove one tag
	accountRepo.RemoveTagByName(accountID, "important")

	// Verify tag was removed
	retrieved, _ = accountRepo.GetByID(accountID)

	if len(retrieved.Tags) != 2 {
		t.Errorf("Expected 2 tags after removal, got %d", len(retrieved.Tags))
	}

	// Verify correct tags remain
	tagNames := make(map[string]bool)
	for _, tag := range retrieved.Tags {
		tagNames[tag.Name] = true
	}

	if !tagNames["work"] || !tagNames["finance"] {
		t.Error("Expected 'work' and 'finance' tags to remain")
	}

	if tagNames["important"] {
		t.Error("Expected 'important' tag to be removed")
	}
}

func TestBooleanStorageInSQLite(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	accountRepo := &AccountRepository{}
	totpRepo := &TOTPRepository{}

	// Create account
	account := &Account{
		User:     "testuser",
		Password: "testpass",
	}
	accountID, _ := accountRepo.Create(account)

	// Create TOTP with UseHomomorphic = true
	totp := &TOTP{
		AccountID:      accountID,
		TOTPSeed:       "TESTSEED",
		UseHomomorphic: false,
	}

	_, err := totpRepo.Create(totp)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Retrieve and check
	retrieved, err := totpRepo.GetByAccountID(accountID)
	if err != nil {
		t.Fatalf("GetByAccountID failed: %v", err)
	}

	if retrieved.UseHomomorphic {
		t.Errorf("Expected UseHomomorphic=false, got %v", retrieved.UseHomomorphic)
	}

	// Update to false
	retrieved.UseHomomorphic = true
	err = totpRepo.Update(retrieved)
	if err == nil {
		t.Fatalf("Update has been succesful when it should be wrong: %v", err)
	}

	// Retrieve again
	retrieved2, err := totpRepo.GetByAccountID(accountID)
	if err != nil {
		t.Fatalf("GetByAccountID failed: %v", err)
	}

	if retrieved2.UseHomomorphic {
		t.Errorf("Expected UseHomomorphic=false, got %v", retrieved2.UseHomomorphic)
	}

	// // Check history
	// history, err := totpRepo.GetHistory(totpID)
	// if err != nil {
	// 	t.Fatalf("GetHistory failed: %v", err)
	// }
	//
	// if len(history) != 1 {
	// 	t.Fatalf("Expected 1 history entries, got %d", len(history))
	// }
	//
	// t.Logf("History[0].UseHomomorphic (most recent): %v", history[0].UseHomomorphic)
	// t.Logf("History[1].UseHomomorphic (oldest): %v", history[1].UseHomomorphic)
	//
	// // Most recent (index 0) should be false
	// if history[0].UseHomomorphic != false {
	// 	t.Errorf("Expected history[0].UseHomomorphic=false, got %v", history[0].UseHomomorphic)
	// }
	//
	// // Oldest (index 1) should be true
	// if history[1].UseHomomorphic != true {
	// 	t.Errorf("Expected history[1].UseHomomorphic=true, got %v", history[1].UseHomomorphic)
	// }
}

func TestTOTPValidationRejectsInconsistentState(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	accountRepo := &AccountRepository{}
	totpRepo := &TOTPRepository{}

	// Create account
	account := &Account{
		User:     "testuser",
		Password: "testpass",
	}
	accountID, _ := accountRepo.Create(account)

	// Test 1: UseHomomorphic=true but missing encrypted fields
	encSeed := "encrypted"
	totp1 := &TOTP{
		AccountID:      accountID,
		TOTPSeed:       "JBSWY3DPEHPK3PXP",
		CTOTPSeed:      &encSeed,
		PaillierN:      nil, // Missing!
		UseHomomorphic: true,
	}

	_, err := totpRepo.Create(totp1)
	if err == nil {
		t.Error("Expected error when UseHomomorphic=true but PaillierN is nil")
	}

	// Test 2: UseHomomorphic=false but has encrypted fields
	paillierN := "paillier"
	totp2 := &TOTP{
		AccountID:      accountID,
		TOTPSeed:       "JBSWY3DPEHPK3PXP",
		CTOTPSeed:      &encSeed,
		PaillierN:      &paillierN,
		UseHomomorphic: false, // Inconsistent!
	}

	_, err = totpRepo.Create(totp2)
	if err == nil {
		t.Error("Expected error when UseHomomorphic=false but encrypted fields are set")
	}

	// Test 3: Valid standard TOTP (should succeed)
	totp3 := &TOTP{
		AccountID:      accountID,
		TOTPSeed:       "JBSWY3DPEHPK3PXP",
		CTOTPSeed:      nil,
		PaillierN:      nil,
		UseHomomorphic: false,
	}

	_, err = totpRepo.Create(totp3)
	if err != nil {
		t.Errorf("Valid standard TOTP should be created, got error: %v", err)
	}
}
