package main

import (
	"time"
)

// Account represents a password account entry
type Account struct {
	ID       int64   `json:"id"              db:"id"`
	Email    *string `json:"email,omitempty" db:"email"`
	User     string  `json:"user"            db:"user"`
	Password string  `json:"password"        db:"password"`
	URL      *string `json:"url,omitempty"   db:"url"`
	Notes    *string `json:"notes"           db:"notes"`
	Expire   string  `json:"expire"          db:"expire"`
	Tags     []Tag   `json:"tags,omitempty"`
	TOTP     *TOTP   `json:"totp,omitempty"`
}

// Tag represents a tag for categorizing accounts
type Tag struct {
	ID   int64  `json:"id"   db:"id"`
	Name string `json:"name" db:"name"`
}

// TagAssociation represents the relationship between accounts and tags
type TagAssociation struct {
	AccountID int64 `json:"account_id" db:"account_id"`
	TagID     int64 `json:"tag_id"     db:"tag_id"`
}

// TOTP represents a TOTP seed for 2FA linked to an account
type TOTP struct {
	ID             int64   `json:"id"                    db:"id"`
	AccountID      int64   `json:"account_id"            db:"account_id"`
	TOTPSeed       string  `json:"totp_seed"             db:"totp_seed"`       // Plaintext Base32 seed
	CTOTPSeed      *string `json:"c_totp_seed,omitempty" db:"c_totp_seed"`     // Encrypted seed (hex)
	PaillierN      *string `json:"paillier_n,omitempty"  db:"paillier_n"`      // Paillier N parameter (hex)
	UseHomomorphic bool    `json:"use_homomorphic"       db:"use_homomorphic"` // Flag to indicate which method to use
}

// AccountHistory represents historical versions of an account
type AccountHistory struct {
	HistoryID    int64   `json:"history_id"              db:"history_id"`
	AccountID    *int64  `json:"account_id,omitempty"    db:"account_id"`
	Email        *string `json:"email,omitempty"         db:"email"`
	User         *string `json:"user,omitempty"          db:"user"`
	Password     *string `json:"password,omitempty"      db:"password"`
	URL          *string `json:"url,omitempty"           db:"url"`
	Notes        *string `json:"notes"                   db:"notes"`
	Expire       *string `json:"expire,omitempty"        db:"expire"`
	ValidFrom    string  `json:"valid_from"              db:"valid_from"`
	ValidTo      string  `json:"valid_to"                db:"valid_to"`
	ChangeReason *string `json:"change_reason,omitempty" db:"change_reason"`
}

// TagAssociationHistory represents historical tag associations
type TagAssociationHistory struct {
	AccountID int64  `json:"account_id" db:"account_id"`
	TagID     int64  `json:"tag_id"     db:"tag_id"`
	ValidFrom string `json:"valid_from" db:"valid_from"`
	ValidTo   string `json:"valid_to"   db:"valid_to"`
}

// TOTPHistory represents historical TOTP seeds
type TOTPHistory struct {
	HistoryID      int64   `json:"history_id"            db:"history_id"`
	TOTPID         int64   `json:"totp_id"               db:"totp_id"`
	TOTPSeed       string  `json:"totp_seed"             db:"totp_seed"`
	CTOTPSeed      *string `json:"c_totp_seed,omitempty" db:"c_totp_seed"`
	PaillierN      *string `json:"paillier_n,omitempty"  db:"paillier_n"`
	UseHomomorphic bool    `json:"use_homomorphic"       db:"use_homomorphic"`
	ValidFrom      string  `json:"valid_from"            db:"valid_from"`
	ValidTo        string  `json:"valid_to"              db:"valid_to"`
}

// Helper function to get current timestamp
func currentTimestamp() string {
	return time.Now().Format(time.RFC3339)
}

// HidePassword creates a copy of the account with password hidden
func (a *Account) HidePassword() Account {
	hidden := *a
	hidden.Password = ""
	return hidden
}

// HidePasswords hides passwords in a slice of accounts
func HidePasswords(accounts []Account) []Account {
	hidden := make([]Account, len(accounts))
	for i, acc := range accounts {
		hidden[i] = acc.HidePassword()
	}
	return hidden
}
