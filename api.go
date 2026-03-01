package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	accountRepo = &AccountRepository{}
	tagRepo     = &TagRepository{}
	totpRepo    = &TOTPRepository{}
	securityChecker *PasswordSecurityChecker
)

// SetupRouter sets up the Gin router with all routes
func SetupRouter() *gin.Engine {
	router := gin.Default()

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Account routes
		accounts := v1.Group("/accounts")
		{
			accounts.POST("", createAccount)
			accounts.GET("", listAccounts)
			accounts.GET("/:id", getAccount)
			accounts.PUT("/:id", updateAccount)
			accounts.DELETE("/:id", deleteAccount)
			accounts.GET("/search", searchAccounts)
			accounts.GET("/search/field", searchAccountsByField)
			accounts.GET("/search/fuzzy", fuzzySearchAccounts)
			accounts.GET("/:id/history", getAccountHistory)
			accounts.POST("/:id/tags/:tag_id", addTagToAccount)
			accounts.DELETE("/:id/tags/:tag_id", removeTagFromAccount)
			accounts.POST("/:id/tag/:tag_name", addTagToAccountByName)
			accounts.DELETE("/:id/tag/:tag_name", removeTagFromAccountByName)
			accounts.GET("/by-tag/:tag_name", getAccountsByTag)
		}

		// Tag routes
		tags := v1.Group("/tags")
		{
			tags.POST("", createTag)
			tags.GET("", listTags)
			tags.GET("/:id", getTag)
			tags.PUT("/:id", updateTag)
			tags.DELETE("/:id", deleteTag)
		}

		// TOTP routes
		totps := v1.Group("/totp")
		{
			totps.POST("", createTOTP)
			totps.GET("/account/:account_id", getTOTPByAccount)
			totps.PUT("/:id", updateTOTP)
			totps.DELETE("/:id", deleteTOTP)
			totps.GET("/:id/history", getTOTPHistory)
			totps.POST("/:id/generate", generateTOTPCode)
			totps.POST("/:id/verify", verifyTOTPCode)
			totps.POST("/create-homomorphic", createHomomorphicTOTP)
			totps.POST("/:id/convert-to-homomorphic", convertToHomomorphicTOTP)
		}

		// Security routes
		security := v1.Group("/security")
		{
			security.POST("/check-password", checkPasswordSecurity)
			security.GET("/check-all", checkAllAccountsSecurity)
			security.GET("/vulnerable", getVulnerableAccounts)
			security.GET("/statistics", getSecurityStatistics)
		}
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return router
}

// Account handlers

func createAccount(c *gin.Context) {
	var account Account
	if err := c.ShouldBindJSON(&account); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := accountRepo.Create(&account)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	account.ID = id
	c.JSON(http.StatusCreated, account)
}

func listAccounts(c *gin.Context) {
	accounts, err := accountRepo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Hide passwords by default unless show_password=true
	showPassword := c.Query("show_password") == "true"
	if !showPassword {
		accounts = HidePasswords(accounts)
	}

	c.JSON(http.StatusOK, accounts)
}

func getAccount(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}

	account, err := accountRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Hide password by default unless show_password=true
	showPassword := c.Query("show_password") == "true"
	if !showPassword {
		hiddenAccount := account.HidePassword()
		c.JSON(http.StatusOK, hiddenAccount)
		return
	}

	c.JSON(http.StatusOK, account)
}

func updateAccount(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}

	var account Account
	if err := c.ShouldBindJSON(&account); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account.ID = id
	
	// Get change_reason from query parameter if provided
	changeReason := c.Query("change_reason")
	var changeReasonPtr *string
	if changeReason != "" {
		changeReasonPtr = &changeReason
	}

	if err := accountRepo.Update(&account, changeReasonPtr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, account)
}

func deleteAccount(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}

	if err := accountRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "account deleted successfully"})
}

func searchAccounts(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}

	accounts, err := accountRepo.Search(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Hide passwords by default unless show_password=true
	showPassword := c.Query("show_password") == "true"
	if !showPassword {
		accounts = HidePasswords(accounts)
	}

	c.JSON(http.StatusOK, accounts)
}

func searchAccountsByField(c *gin.Context) {
	field := c.Query("field")
	query := c.Query("q")
	
	if field == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'field' is required"})
		return
	}
	
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}

	accounts, err := accountRepo.SearchByField(field, query)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hide passwords by default unless show_password=true
	showPassword := c.Query("show_password") == "true"
	if !showPassword {
		accounts = HidePasswords(accounts)
	}

	c.JSON(http.StatusOK, accounts)
}

func fuzzySearchAccounts(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}

	// Parse optional filters
	filters := &SearchFilters{}
	
	// Min score filter
	if minScoreStr := c.Query("min_score"); minScoreStr != "" {
		var minScore float64
		if _, err := fmt.Sscanf(minScoreStr, "%f", &minScore); err == nil {
			filters.MinScore = minScore
		}
	}
	
	// Tags filter
	if tagsStr := c.Query("tags"); tagsStr != "" {
		filters.Tags = strings.Split(tagsStr, ",")
	}
	
	// Has TOTP filter
	if hasTOTPStr := c.Query("has_totp"); hasTOTPStr != "" {
		hasTOTP := hasTOTPStr == "true"
		filters.HasTOTP = &hasTOTP
	}
	
	// Expire before filter
	if expireBefore := c.Query("expire_before"); expireBefore != "" {
		filters.ExpireBefore = &expireBefore
	}
	
	// Expire after filter
	if expireAfter := c.Query("expire_after"); expireAfter != "" {
		filters.ExpireAfter = &expireAfter
	}

	results, err := accountRepo.FuzzySearch(query, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Hide passwords by default unless show_password=true
	showPassword := c.Query("show_password") == "true"
	if !showPassword {
		for i := range results {
			results[i].Account = results[i].Account.HidePassword()
		}
	}

	c.JSON(http.StatusOK, results)
}

func getAccountHistory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}

	history, err := accountRepo.GetHistory(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}

func addTagToAccount(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}

	tagID, err := strconv.ParseInt(c.Param("tag_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag id"})
		return
	}

	if err := accountRepo.AddTag(accountID, tagID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "tag added to account successfully"})
}

func removeTagFromAccount(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}

	tagID, err := strconv.ParseInt(c.Param("tag_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag id"})
		return
	}

	if err := accountRepo.RemoveTag(accountID, tagID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "tag removed from account successfully"})
}

func addTagToAccountByName(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}

	tagName := c.Param("tag_name")
	if tagName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tag name is required"})
		return
	}

	if err := accountRepo.AddTagByName(accountID, tagName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "tag added to account successfully",
		"tag_name": tagName,
	})
}

func removeTagFromAccountByName(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}

	tagName := c.Param("tag_name")
	if tagName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tag name is required"})
		return
	}

	if err := accountRepo.RemoveTagByName(accountID, tagName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "tag removed from account successfully",
		"tag_name": tagName,
	})
}

func getAccountsByTag(c *gin.Context) {
	tagName := c.Param("tag_name")
	if tagName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tag name is required"})
		return
	}

	accounts, err := accountRepo.GetAccountsByTagName(tagName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Hide passwords by default unless show_password=true
	showPassword := c.Query("show_password") == "true"
	if !showPassword {
		accounts = HidePasswords(accounts)
	}

	c.JSON(http.StatusOK, accounts)
}

// Tag handlers

func createTag(c *gin.Context) {
	var tag Tag
	if err := c.ShouldBindJSON(&tag); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := tagRepo.Create(&tag)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tag.ID = id
	c.JSON(http.StatusCreated, tag)
}

func listTags(c *gin.Context) {
	tags, err := tagRepo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tags)
}

func getTag(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag id"})
		return
	}

	tag, err := tagRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tag)
}

func updateTag(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag id"})
		return
	}

	var tag Tag
	if err := c.ShouldBindJSON(&tag); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tag.ID = id
	if err := tagRepo.Update(&tag); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tag)
}

func deleteTag(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag id"})
		return
	}

	if err := tagRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "tag deleted successfully"})
}

// TOTP handlers

func createTOTP(c *gin.Context) {
	var totp TOTP
	if err := c.ShouldBindJSON(&totp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := totpRepo.Create(&totp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totp.ID = id
	c.JSON(http.StatusCreated, totp)
}

func getTOTPByAccount(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("account_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}

	totp, err := totpRepo.GetByAccountID(accountID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, totp)
}

func updateTOTP(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid totp id"})
		return
	}

	var totp TOTP
	if err := c.ShouldBindJSON(&totp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	totp.ID = id
	if err := totpRepo.Update(&totp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, totp)
}

func deleteTOTP(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid totp id"})
		return
	}

	if err := totpRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "totp deleted successfully"})
}

func getTOTPHistory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid totp id"})
		return
	}

	history, err := totpRepo.GetHistory(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}

// generateTOTPCode generates a TOTP code from a TOTP entry
func generateTOTPCode(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid totp id"})
		return
	}

	// Get TOTP by ID first
	var totp *TOTP
	var accountID int64
	err = DB.QueryRow("SELECT account_id FROM totp WHERE id = ?", id).Scan(&accountID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "totp not found"})
		return
	}

	totp, err = totpRepo.GetByAccountID(accountID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var code string

	if totp.UseHomomorphic && totp.CTOTPSeed != nil && totp.PaillierN != nil {
		// Need private key - this would normally be stored securely or derived
		// For this example, we'll return an error asking for the private key
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "homomorphic TOTP requires private key in request body",
			"hint":  "POST with {\"paillier_n_hex\": \"...\", \"paillier_lambda_hex\": \"...\"}",
		})
		return
	} else {
		// Generate standard TOTP
		code, err = GenerateStandardTOTP(totp.TOTPSeed)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":            code,
		"use_homomorphic": totp.UseHomomorphic,
	})
}

// verifyTOTPCode verifies a TOTP code
func verifyTOTPCode(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid totp id"})
		return
	}

	var req struct {
		Code string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get TOTP
	var accountID int64
	err = DB.QueryRow("SELECT account_id FROM totp WHERE id = ?", id).Scan(&accountID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "totp not found"})
		return
	}

	totp, err := totpRepo.GetByAccountID(accountID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Verify using standard TOTP (window of 1 = 30 seconds before/after)
	valid, err := VerifyTOTP(req.Code, totp.TOTPSeed, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid": valid,
	})
}

// createHomomorphicTOTP creates a new TOTP with homomorphic encryption
func createHomomorphicTOTP(c *gin.Context) {
	var req struct {
		AccountID int64  `json:"account_id" binding:"required"`
		Seed      string `json:"seed"` // Optional: if not provided, generates random
		KeyBits   int    `json:"key_bits"` // Optional: default 2048
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default to 2048 bits
	if req.KeyBits == 0 {
		req.KeyBits = 2048
	}

	// Generate random seed if not provided
	if req.Seed == "" {
		seed, err := GenerateRandomTOTPSeed()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate seed"})
			return
		}
		req.Seed = seed
	}

	// Create Paillier key pair
	privateKey, err := CreatePaillierKeyPair(req.KeyBits)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create key pair"})
		return
	}

	// Encrypt the seed
	homomorphicData, err := EncryptTOTPSeed(req.Seed, privateKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encrypt seed"})
		return
	}

	// Create TOTP entry
	totp := &TOTP{
		AccountID:      req.AccountID,
		TOTPSeed:       homomorphicData.PlaintextSeed,
		CTOTPSeed:      &homomorphicData.EncryptedSeed,
		PaillierN:      &homomorphicData.PaillierN,
		UseHomomorphic: true,
	}

	totpID, err := totpRepo.Create(totp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totp.ID = totpID

	// Serialize private key for client storage
	nHex, lambdaHex := SerializePrivateKey(privateKey)

	c.JSON(http.StatusCreated, gin.H{
		"totp":               totp,
		"paillier_n_hex":     nHex,
		"paillier_lambda_hex": lambdaHex,
		"warning":            "Store the private key (n and lambda) securely! It's needed for homomorphic operations.",
	})
}

func convertToHomomorphicTOTP(c *gin.Context) {
	totpID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid totp id"})
		return
	}

	var req struct {
		KeyBits int `json:"key_bits"` // Optional: default 2048
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// Empty body is ok, use defaults
		req.KeyBits = 2048
	}

	// Default to 2048 bits
	if req.KeyBits == 0 {
		req.KeyBits = 2048
	}

	// Get existing TOTP
	var accountID int64
	var totpSeed string
	var cTOTPSeed, paillierN *string
	var useHomomorphic bool
	
	err = DB.QueryRow(
		"SELECT account_id, totp_seed, c_totp_seed, paillier_n, use_homomorphic FROM totp WHERE id = ?",
		totpID,
	).Scan(&accountID, &totpSeed, &cTOTPSeed, &paillierN, &useHomomorphic)
	
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "totp not found"})
		return
	}

	// Check if already homomorphic
	if useHomomorphic {
		c.JSON(http.StatusBadRequest, gin.H{"error": "TOTP is already using homomorphic encryption"})
		return
	}

	// Create Paillier key pair
	privateKey, err := CreatePaillierKeyPair(req.KeyBits)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create key pair"})
		return
	}

	// Encrypt the existing seed
	homomorphicData, err := EncryptTOTPSeed(totpSeed, privateKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encrypt seed"})
		return
	}

	// Update TOTP to use homomorphic encryption
	updatedTOTP := &TOTP{
		ID:             totpID,
		AccountID:      accountID,
		TOTPSeed:       homomorphicData.PlaintextSeed,
		CTOTPSeed:      &homomorphicData.EncryptedSeed,
		PaillierN:      &homomorphicData.PaillierN,
		UseHomomorphic: true,
	}

	err = totpRepo.Update(updatedTOTP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Serialize private key for client storage
	nHex, lambdaHex := SerializePrivateKey(privateKey)

	c.JSON(http.StatusOK, gin.H{
		"totp":                updatedTOTP,
		"paillier_n_hex":      nHex,
		"paillier_lambda_hex": lambdaHex,
		"warning":             "Store the private key (n and lambda) securely! It's needed for homomorphic operations.",
		"message":             "TOTP successfully converted to homomorphic encryption",
	})
}

// Security Check Handlers

func checkPasswordSecurity(c *gin.Context) {
	var req struct {
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if securityChecker == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Security checker not initialized"})
		return
	}

	report := securityChecker.AnalyzePassword(req.Password)
	c.JSON(http.StatusOK, report)
}

func checkAllAccountsSecurity(c *gin.Context) {
	if securityChecker == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Security checker not initialized"})
		return
	}

	reports, err := accountRepo.CheckAllAccountPasswords(securityChecker)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, reports)
}

func getVulnerableAccounts(c *gin.Context) {
	if securityChecker == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Security checker not initialized"})
		return
	}

	vulnerable, err := accountRepo.GetVulnerableAccounts(securityChecker)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count": len(vulnerable),
		"accounts": vulnerable,
	})
}

func getSecurityStatistics(c *gin.Context) {
	if securityChecker == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Security checker not initialized"})
		return
	}

	includeVulnerable := c.Query("include_vulnerable") == "true"

	stats, err := accountRepo.GetSecurityStatistics(securityChecker, includeVulnerable)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
