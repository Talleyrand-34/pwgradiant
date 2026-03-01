package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func setupTestAPI(t *testing.T) (*httptest.Server, func()) {
	dbPath := "test_api.db"
	encryptionKey := "test-key-api"
	
	os.Remove(dbPath)
	
	err := InitDB(dbPath, encryptionKey)
	if err != nil {
		t.Fatalf("Failed to init DB: %v", err)
	}
	
	err = CreateSchema()
	if err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}
	
	router := SetupRouter()
	server := httptest.NewServer(router)
	
	return server, func() {
		server.Close()
		CloseDB()
		os.Remove(dbPath)
	}
}

func TestHealthCheck(t *testing.T) {
	server, cleanup := setupTestAPI(t)
	defer cleanup()
	
	resp, err := http.Get(server.URL + "/health")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	
	if result["status"] != "ok" {
		t.Errorf("Expected status ok, got %v", result["status"])
	}
}

func TestCreateAccount(t *testing.T) {
	server, cleanup := setupTestAPI(t)
	defer cleanup()
	
	email := "test@example.com"
	account := Account{
		Email:    &email,
		User:     "testuser",
		Password: "testpass",
	}
	
	body, _ := json.Marshal(account)
	resp, err := http.Post(server.URL+"/api/v1/accounts", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}
	
	var created Account
	json.NewDecoder(resp.Body).Decode(&created)
	
	if created.ID <= 0 {
		t.Error("Expected positive ID")
	}
	if created.User != "testuser" {
		t.Errorf("User mismatch: got %s", created.User)
	}
}

func TestCreateAccountInvalidJSON(t *testing.T) {
	server, cleanup := setupTestAPI(t)
	defer cleanup()
	
	invalidJSON := []byte(`{"user": "test"`)
	resp, err := http.Post(server.URL+"/api/v1/accounts", "application/json", bytes.NewBuffer(invalidJSON))
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestListAccounts(t *testing.T) {
	server, cleanup := setupTestAPI(t)
	defer cleanup()
	
	// Create some accounts
	for i := 0; i < 3; i++ {
		account := Account{User: "user", Password: "pass"}
		body, _ := json.Marshal(account)
		http.Post(server.URL+"/api/v1/accounts", "application/json", bytes.NewBuffer(body))
	}
	
	// List accounts
	resp, err := http.Get(server.URL + "/api/v1/accounts")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	
	var accounts []Account
	json.NewDecoder(resp.Body).Decode(&accounts)
	
	if len(accounts) != 3 {
		t.Errorf("Expected 3 accounts, got %d", len(accounts))
	}
}

func TestGetAccount(t *testing.T) {
	server, cleanup := setupTestAPI(t)
	defer cleanup()
	
	// Create account
	account := Account{User: "testuser", Password: "testpass"}
	body, _ := json.Marshal(account)
	createResp, _ := http.Post(server.URL+"/api/v1/accounts", "application/json", bytes.NewBuffer(body))
	
	var created Account
	json.NewDecoder(createResp.Body).Decode(&created)
	createResp.Body.Close()
	
	// Get account
	resp, err := http.Get(server.URL + "/api/v1/accounts/" + string(rune(created.ID+'0')))
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestGetAccountNotFound(t *testing.T) {
	server, cleanup := setupTestAPI(t)
	defer cleanup()
	
	resp, err := http.Get(server.URL + "/api/v1/accounts/999")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestUpdateAccount(t *testing.T) {
	server, cleanup := setupTestAPI(t)
	defer cleanup()
	
	// Create account
	account := Account{User: "testuser", Password: "oldpass"}
	body, _ := json.Marshal(account)
	createResp, _ := http.Post(server.URL+"/api/v1/accounts", "application/json", bytes.NewBuffer(body))
	
	var created Account
	json.NewDecoder(createResp.Body).Decode(&created)
	createResp.Body.Close()
	
	// Update account
	created.Password = "newpass"
	updateBody, _ := json.Marshal(created)
	
	client := &http.Client{}
	req, _ := http.NewRequest("PUT", server.URL+"/api/v1/accounts/1", bytes.NewBuffer(updateBody))
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestDeleteAccount(t *testing.T) {
	server, cleanup := setupTestAPI(t)
	defer cleanup()
	
	// Create account
	account := Account{User: "testuser", Password: "testpass"}
	body, _ := json.Marshal(account)
	createResp, _ := http.Post(server.URL+"/api/v1/accounts", "application/json", bytes.NewBuffer(body))
	createResp.Body.Close()
	
	// Delete account
	client := &http.Client{}
	req, _ := http.NewRequest("DELETE", server.URL+"/api/v1/accounts/1", nil)
	
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestSearchAccounts(t *testing.T) {
	server, cleanup := setupTestAPI(t)
	defer cleanup()
	
	// Create accounts
	email := "john@example.com"
	account1 := Account{Email: &email, User: "john", Password: "pass"}
	account2 := Account{User: "jane", Password: "pass"}
	
	body1, _ := json.Marshal(account1)
	http.Post(server.URL+"/api/v1/accounts", "application/json", bytes.NewBuffer(body1))
	
	body2, _ := json.Marshal(account2)
	http.Post(server.URL+"/api/v1/accounts", "application/json", bytes.NewBuffer(body2))
	
	// Search
	resp, err := http.Get(server.URL + "/api/v1/accounts/search?q=john")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	
	var results []Account
	json.NewDecoder(resp.Body).Decode(&results)
	
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

func TestGetAccountHistory(t *testing.T) {
	server, cleanup := setupTestAPI(t)
	defer cleanup()
	
	// Create and update account to generate history
	account := Account{User: "testuser", Password: "pass1"}
	body, _ := json.Marshal(account)
	createResp, _ := http.Post(server.URL+"/api/v1/accounts", "application/json", bytes.NewBuffer(body))
	
	var created Account
	json.NewDecoder(createResp.Body).Decode(&created)
	createResp.Body.Close()
	
	// Get history
	resp, err := http.Get(server.URL + "/api/v1/accounts/1/history")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestCreateTag(t *testing.T) {
	server, cleanup := setupTestAPI(t)
	defer cleanup()
	
	tag := Tag{Name: "work"}
	body, _ := json.Marshal(tag)
	
	resp, err := http.Post(server.URL+"/api/v1/tags", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}
	
	var created Tag
	json.NewDecoder(resp.Body).Decode(&created)
	
	if created.ID <= 0 {
		t.Error("Expected positive ID")
	}
}

func TestListTags(t *testing.T) {
	server, cleanup := setupTestAPI(t)
	defer cleanup()
	
	// Create tags
	tags := []string{"work", "personal", "urgent"}
	for _, name := range tags {
		tag := Tag{Name: name}
		body, _ := json.Marshal(tag)
		http.Post(server.URL+"/api/v1/tags", "application/json", bytes.NewBuffer(body))
	}
	
	// List tags
	resp, err := http.Get(server.URL + "/api/v1/tags")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	
	var allTags []Tag
	json.NewDecoder(resp.Body).Decode(&allTags)
	
	if len(allTags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(allTags))
	}
}

func TestGetTag(t *testing.T) {
	server, cleanup := setupTestAPI(t)
	defer cleanup()
	
	tag := Tag{Name: "test"}
	body, _ := json.Marshal(tag)
	createResp, _ := http.Post(server.URL+"/api/v1/tags", "application/json", bytes.NewBuffer(body))
	createResp.Body.Close()
	
	resp, err := http.Get(server.URL + "/api/v1/tags/1")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestUpdateTag(t *testing.T) {
	server, cleanup := setupTestAPI(t)
	defer cleanup()
	
	tag := Tag{Name: "oldname"}
	body, _ := json.Marshal(tag)
	createResp, _ := http.Post(server.URL+"/api/v1/tags", "application/json", bytes.NewBuffer(body))
	createResp.Body.Close()
	
	tag.Name = "newname"
	updateBody, _ := json.Marshal(tag)
	
	client := &http.Client{}
	req, _ := http.NewRequest("PUT", server.URL+"/api/v1/tags/1", bytes.NewBuffer(updateBody))
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestDeleteTag(t *testing.T) {
	server, cleanup := setupTestAPI(t)
	defer cleanup()
	
	tag := Tag{Name: "todelete"}
	body, _ := json.Marshal(tag)
	createResp, _ := http.Post(server.URL+"/api/v1/tags", "application/json", bytes.NewBuffer(body))
	createResp.Body.Close()
	
	client := &http.Client{}
	req, _ := http.NewRequest("DELETE", server.URL+"/api/v1/tags/1", nil)
	
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestAddTagToAccount(t *testing.T) {
	server, cleanup := setupTestAPI(t)
	defer cleanup()
	
	// Create account
	account := Account{User: "testuser", Password: "testpass"}
	accountBody, _ := json.Marshal(account)
	http.Post(server.URL+"/api/v1/accounts", "application/json", bytes.NewBuffer(accountBody))
	
	// Create tag
	tag := Tag{Name: "work"}
	tagBody, _ := json.Marshal(tag)
	http.Post(server.URL+"/api/v1/tags", "application/json", bytes.NewBuffer(tagBody))
	
	// Add tag to account
	resp, err := http.Post(server.URL+"/api/v1/accounts/1/tags/1", "application/json", nil)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestRemoveTagFromAccount(t *testing.T) {
	server, cleanup := setupTestAPI(t)
	defer cleanup()
	
	// Create account and tag
	account := Account{User: "testuser", Password: "testpass"}
	accountBody, _ := json.Marshal(account)
	http.Post(server.URL+"/api/v1/accounts", "application/json", bytes.NewBuffer(accountBody))
	
	tag := Tag{Name: "work"}
	tagBody, _ := json.Marshal(tag)
	http.Post(server.URL+"/api/v1/tags", "application/json", bytes.NewBuffer(tagBody))
	
	// Add tag
	http.Post(server.URL+"/api/v1/accounts/1/tags/1", "application/json", nil)
	
	// Remove tag
	client := &http.Client{}
	req, _ := http.NewRequest("DELETE", server.URL+"/api/v1/accounts/1/tags/1", nil)
	
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestCreateTOTP(t *testing.T) {
	server, cleanup := setupTestAPI(t)
	defer cleanup()
	
	// Create account first
	account := Account{User: "testuser", Password: "testpass"}
	accountBody, _ := json.Marshal(account)
	http.Post(server.URL+"/api/v1/accounts", "application/json", bytes.NewBuffer(accountBody))
	
	// Create TOTP
	totp := TOTP{AccountID: 1, TOTPSeed: "JBSWY3DPEHPK3PXP"}
	body, _ := json.Marshal(totp)
	
	resp, err := http.Post(server.URL+"/api/v1/totp", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}
}

func TestGetTOTPByAccount(t *testing.T) {
	server, cleanup := setupTestAPI(t)
	defer cleanup()
	
	// Create account and TOTP
	account := Account{User: "testuser", Password: "testpass"}
	accountBody, _ := json.Marshal(account)
	http.Post(server.URL+"/api/v1/accounts", "application/json", bytes.NewBuffer(accountBody))
	
	totp := TOTP{AccountID: 1, TOTPSeed: "SEED123"}
	totpBody, _ := json.Marshal(totp)
	http.Post(server.URL+"/api/v1/totp", "application/json", bytes.NewBuffer(totpBody))
	
	// Get TOTP
	resp, err := http.Get(server.URL + "/api/v1/totp/account/1")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestUpdateTOTP(t *testing.T) {
	server, cleanup := setupTestAPI(t)
	defer cleanup()
	
	// Create account and TOTP
	account := Account{User: "testuser", Password: "testpass"}
	accountBody, _ := json.Marshal(account)
	http.Post(server.URL+"/api/v1/accounts", "application/json", bytes.NewBuffer(accountBody))
	
	totp := TOTP{AccountID: 1, TOTPSeed: "OLDSEED"}
	totpBody, _ := json.Marshal(totp)
	http.Post(server.URL+"/api/v1/totp", "application/json", bytes.NewBuffer(totpBody))
	
	// Update TOTP
	totp.TOTPSeed = "NEWSEED"
	updateBody, _ := json.Marshal(totp)
	
	client := &http.Client{}
	req, _ := http.NewRequest("PUT", server.URL+"/api/v1/totp/1", bytes.NewBuffer(updateBody))
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestDeleteTOTP(t *testing.T) {
	server, cleanup := setupTestAPI(t)
	defer cleanup()
	
	// Create account and TOTP
	account := Account{User: "testuser", Password: "testpass"}
	accountBody, _ := json.Marshal(account)
	http.Post(server.URL+"/api/v1/accounts", "application/json", bytes.NewBuffer(accountBody))
	
	totp := TOTP{AccountID: 1, TOTPSeed: "SEED"}
	totpBody, _ := json.Marshal(totp)
	http.Post(server.URL+"/api/v1/totp", "application/json", bytes.NewBuffer(totpBody))
	
	// Delete TOTP
	client := &http.Client{}
	req, _ := http.NewRequest("DELETE", server.URL+"/api/v1/totp/1", nil)
	
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}
