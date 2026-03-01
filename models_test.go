package main

import (
	"encoding/json"
	"testing"
	"time"
)

func TestAccountJSONMarshaling(t *testing.T) {
	email := "test@example.com"
	url := "https://example.com"
	notes := "Test notes"
	
	account := Account{
		ID:       1,
		Email:    &email,
		User:     "testuser",
		Password: "testpass",
		URL:      &url,
		Notes:    &notes,
		Expire:   "2025-12-31T23:59:59Z",
		Tags:     []Tag{{ID: 1, Name: "work"}},
	}
	
	// Marshal to JSON
	data, err := json.Marshal(account)
	if err != nil {
		t.Fatalf("Failed to marshal account: %v", err)
	}
	
	// Unmarshal back
	var decoded Account
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal account: %v", err)
	}
	
	// Verify fields
	if decoded.ID != account.ID {
		t.Errorf("ID mismatch: got %d, want %d", decoded.ID, account.ID)
	}
	if *decoded.Email != *account.Email {
		t.Errorf("Email mismatch: got %s, want %s", *decoded.Email, *account.Email)
	}
	if decoded.User != account.User {
		t.Errorf("User mismatch: got %s, want %s", decoded.User, account.User)
	}
}

func TestAccountNullableFields(t *testing.T) {
	// Account with no nullable fields set
	account := Account{
		ID:       1,
		User:     "testuser",
		Password: "testpass",
		Expire:   "2025-12-31T23:59:59Z",
	}
	
	data, err := json.Marshal(account)
	if err != nil {
		t.Fatalf("Failed to marshal account: %v", err)
	}
	
	var decoded Account
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal account: %v", err)
	}
	
	// Verify nullable fields are nil
	if decoded.Email != nil {
		t.Error("Email should be nil")
	}
	if decoded.URL != nil {
		t.Error("URL should be nil")
	}
	if decoded.Notes != nil {
		t.Error("Notes should be nil")
	}
}

func TestTagJSONMarshaling(t *testing.T) {
	tag := Tag{
		ID:   1,
		Name: "work",
	}
	
	data, err := json.Marshal(tag)
	if err != nil {
		t.Fatalf("Failed to marshal tag: %v", err)
	}
	
	var decoded Tag
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal tag: %v", err)
	}
	
	if decoded.ID != tag.ID {
		t.Errorf("ID mismatch: got %d, want %d", decoded.ID, tag.ID)
	}
	if decoded.Name != tag.Name {
		t.Errorf("Name mismatch: got %s, want %s", decoded.Name, tag.Name)
	}
}

func TestTOTPJSONMarshaling(t *testing.T) {
	totp := TOTP{
		ID:        1,
		AccountID: 10,
		TOTPSeed:  "JBSWY3DPEHPK3PXP",
	}
	
	data, err := json.Marshal(totp)
	if err != nil {
		t.Fatalf("Failed to marshal totp: %v", err)
	}
	
	var decoded TOTP
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal totp: %v", err)
	}
	
	if decoded.ID != totp.ID {
		t.Errorf("ID mismatch: got %d, want %d", decoded.ID, totp.ID)
	}
	if decoded.AccountID != totp.AccountID {
		t.Errorf("AccountID mismatch: got %d, want %d", decoded.AccountID, totp.AccountID)
	}
	if decoded.TOTPSeed != totp.TOTPSeed {
		t.Errorf("TOTPSeed mismatch: got %s, want %s", decoded.TOTPSeed, totp.TOTPSeed)
	}
}

func TestAccountHistoryJSONMarshaling(t *testing.T) {
	accountID := int64(1)
	email := "old@example.com"
	user := "olduser"
	password := "oldpass"
	url := "https://old.com"
	notes := "old notes"
	expire := "2024-12-31T23:59:59Z"
	validTo := "2025-01-01T00:00:00Z"
	changeReason := "Password expired"
	
	history := AccountHistory{
		HistoryID:    1,
		AccountID:    &accountID,
		Email:        &email,
		User:         &user,
		Password:     &password,
		URL:          &url,
		Notes:        &notes,
		Expire:       &expire,
		ValidFrom:    "2024-01-01T00:00:00Z",
		ValidTo:      validTo,
		ChangeReason: &changeReason,
	}
	
	data, err := json.Marshal(history)
	if err != nil {
		t.Fatalf("Failed to marshal account history: %v", err)
	}
	
	var decoded AccountHistory
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal account history: %v", err)
	}
	
	if *decoded.AccountID != *history.AccountID {
		t.Errorf("AccountID mismatch")
	}
	if *decoded.ChangeReason != *history.ChangeReason {
		t.Errorf("ChangeReason mismatch")
	}
}

func TestCurrentTimestamp(t *testing.T) {
	ts := currentTimestamp()
	
	// Parse the timestamp
	_, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		t.Fatalf("currentTimestamp() returned invalid RFC3339 format: %v", err)
	}
	
	// Verify it's recent (within last 5 seconds)
	parsedTime, _ := time.Parse(time.RFC3339, ts)
	diff := time.Since(parsedTime)
	if diff > 5*time.Second || diff < -5*time.Second {
		t.Errorf("currentTimestamp() seems to be off by %v", diff)
	}
}

func TestAccountWithTOTP(t *testing.T) {
	totp := &TOTP{
		ID:        1,
		AccountID: 1,
		TOTPSeed:  "JBSWY3DPEHPK3PXP",
	}
	
	account := Account{
		ID:       1,
		User:     "testuser",
		Password: "testpass",
		Expire:   "2025-12-31T23:59:59Z",
		TOTP:     totp,
	}
	
	data, err := json.Marshal(account)
	if err != nil {
		t.Fatalf("Failed to marshal account with TOTP: %v", err)
	}
	
	var decoded Account
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal account with TOTP: %v", err)
	}
	
	if decoded.TOTP == nil {
		t.Fatal("TOTP should not be nil")
	}
	if decoded.TOTP.TOTPSeed != totp.TOTPSeed {
		t.Errorf("TOTP seed mismatch")
	}
}

func TestTagAssociationJSONMarshaling(t *testing.T) {
	assoc := TagAssociation{
		AccountID: 1,
		TagID:     2,
	}
	
	data, err := json.Marshal(assoc)
	if err != nil {
		t.Fatalf("Failed to marshal tag association: %v", err)
	}
	
	var decoded TagAssociation
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal tag association: %v", err)
	}
	
	if decoded.AccountID != assoc.AccountID {
		t.Errorf("AccountID mismatch")
	}
	if decoded.TagID != assoc.TagID {
		t.Errorf("TagID mismatch")
	}
}

func TestTOTPHistoryJSONMarshaling(t *testing.T) {
	history := TOTPHistory{
		HistoryID: 1,
		TOTPID:    2,
		TOTPSeed:  "NEWSEED123",
		ValidFrom: "2024-01-01T00:00:00Z",
		ValidTo:   "2025-01-01T00:00:00Z",
	}
	
	data, err := json.Marshal(history)
	if err != nil {
		t.Fatalf("Failed to marshal TOTP history: %v", err)
	}
	
	var decoded TOTPHistory
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal TOTP history: %v", err)
	}
	
	if decoded.TOTPID != history.TOTPID {
		t.Errorf("TOTPID mismatch")
	}
}
