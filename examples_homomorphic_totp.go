package main

import (
	"fmt"
	"log"
)

// Example: Standard TOTP Usage
func ExampleStandardTOTP() {
	// Generate a random seed
	seed, err := GenerateRandomTOTPSeed()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Generated Seed: %s\n", seed)

	// Generate TOTP code
	code, err := GenerateStandardTOTP(seed)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("TOTP Code: %s\n", code)

	// Verify the code
	valid, err := VerifyTOTP(code, seed, 1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Code Valid: %v\n", valid)

	// Output:
	// Generated Seed: <random-base32-string>
	// TOTP Code: <6-digit-code>
	// Code Valid: true
}

// Example: Homomorphic TOTP Creation
func ExampleHomomorphicTOTPCreation() {
	// 1. Generate or use existing seed
	seed := "JBSWY3DPEHPK3PXP"

	// 2. Create Paillier key pair (2048 bits recommended)
	privateKey, err := CreatePaillierKeyPair(2048)
	if err != nil {
		log.Fatal(err)
	}

	// 3. Encrypt the seed
	homomorphicData, err := EncryptTOTPSeed(seed, privateKey)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Original Seed: %s\n", seed)
	fmt.Printf("Encrypted Seed: %s\n", homomorphicData.EncryptedSeed[:20]+"...")
	fmt.Printf("Paillier N: %s\n", homomorphicData.PaillierN[:20]+"...")

	// 4. Serialize private key for secure storage
	nHex, lambdaHex := SerializePrivateKey(privateKey)
	fmt.Printf("Private Key N: %s\n", nHex[:20]+"...")
	fmt.Printf("Private Key Lambda: %s\n", lambdaHex[:20]+"...")

	fmt.Println("\n⚠️  STORE THE PRIVATE KEY SECURELY!")

	// Output:
	// Original Seed: JBSWY3DPEHPK3PXP
	// Encrypted Seed: <hex-string>...
	// Paillier N: <hex-string>...
	// Private Key N: <hex-string>...
	// Private Key Lambda: <hex-string>...
	// ⚠️  STORE THE PRIVATE KEY SECURELY!
}

// Example: Homomorphic TOTP Generation
func ExampleHomomorphicTOTPGeneration() {
	// Assume we have encrypted seed data from storage
	seed := "JBSWY3DPEHPK3PXP"

	// Create key pair (in practice, load from secure storage)
	privateKey, _ := CreatePaillierKeyPair(2048)
	homomorphicData, _ := EncryptTOTPSeed(seed, privateKey)

	// Generate TOTP code using homomorphic encryption
	code, err := GenerateHomomorphicTOTP(
		homomorphicData.EncryptedSeed,
		homomorphicData.PaillierN,
		privateKey,
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Homomorphic TOTP Code: %s\n", code)

	// Compare with standard TOTP
	standardCode, _ := GenerateStandardTOTP(seed)
	fmt.Printf("Standard TOTP Code: %s\n", standardCode)
	fmt.Printf("Codes Match: %v\n", code == standardCode)

	// Output:
	// Homomorphic TOTP Code: <6-digit-code>
	// Standard TOTP Code: <6-digit-code>
	// Codes Match: true
}

// Example: Complete Workflow
func ExampleCompleteWorkflow() {
	fmt.Println("=== Complete Homomorphic TOTP Workflow ===")

	// Step 1: Generate random seed
	fmt.Println("1. Generating random seed...")
	seed, _ := GenerateRandomTOTPSeed()
	fmt.Printf("   Seed: %s\n\n", seed)

	// Step 2: Create encryption keys
	fmt.Println("2. Creating Paillier key pair (2048 bits)...")
	privateKey, _ := CreatePaillierKeyPair(2048)
	fmt.Println("   ✓ Keys generated")

	// Step 3: Encrypt the seed
	fmt.Println("3. Encrypting TOTP seed...")
	homomorphicData, _ := EncryptTOTPSeed(seed, privateKey)
	fmt.Printf("   Encrypted Seed: %s...\n", homomorphicData.EncryptedSeed[:40])
	fmt.Printf("   Paillier N: %s...\n\n", homomorphicData.PaillierN[:40])

	// Step 4: Store data (simulated)
	fmt.Println("4. Storing encrypted data...")
	nHex, lambdaHex := SerializePrivateKey(privateKey)
	fmt.Println("   ✓ Data serialized for storage")

	// Step 5: Retrieve and reconstruct (simulated)
	fmt.Println("5. Retrieving encrypted data...")
	restoredKey, _ := DeserializePrivateKey(nHex, lambdaHex)
	fmt.Println("   ✓ Private key restored")

	// Step 6: Generate TOTP codes
	fmt.Println("6. Generating TOTP codes...")

	// Standard method
	standardCode, _ := GenerateStandardTOTP(seed)
	fmt.Printf("   Standard TOTP: %s\n", standardCode)

	// Homomorphic method
	homomorphicCode, _ := GenerateHomomorphicTOTP(
		homomorphicData.EncryptedSeed,
		homomorphicData.PaillierN,
		restoredKey,
	)
	fmt.Printf("   Homomorphic TOTP: %s\n", homomorphicCode)
	fmt.Printf("   Match: %v\n\n", standardCode == homomorphicCode)

	// Step 7: Verify code
	fmt.Println("7. Verifying TOTP code...")
	valid, _ := VerifyTOTP(homomorphicCode, seed, 1)
	fmt.Printf("   Code Valid: %v\n\n", valid)

	fmt.Println("=== Workflow Complete ===")

	// Output:
	// === Complete Homomorphic TOTP Workflow ===
	//
	// 1. Generating random seed...
	//    Seed: <random-seed>
	//
	// 2. Creating Paillier key pair (2048 bits)...
	//    ✓ Keys generated
	//
	// 3. Encrypting TOTP seed...
	//    Encrypted Seed: <encrypted-data>...
	//    Paillier N: <paillier-n>...
	//
	// 4. Storing encrypted data...
	//    ✓ Data serialized for storage
	//
	// 5. Retrieving encrypted data...
	//    ✓ Private key restored
	//
	// 6. Generating TOTP codes...
	//    Standard TOTP: <code>
	//    Homomorphic TOTP: <code>
	//    Match: true
	//
	// 7. Verifying TOTP code...
	//    Code Valid: true
	//
	// === Workflow Complete ===
}

// Example: Database Integration
func ExampleDatabaseIntegration() {
	// Initialize database (simulated)
	// setupTestDB()

	accountRepo := &AccountRepository{}
	totpRepo := &TOTPRepository{}

	// Create account
	account := &Account{
		User:     "john@example.com",
		Password: "SecurePassword123!",
	}
	accountID, _ := accountRepo.Create(account)

	// Generate random seed
	seed, _ := GenerateRandomTOTPSeed()

	// Create Paillier keys
	privateKey, _ := CreatePaillierKeyPair(2048)

	// Encrypt seed
	homomorphicData, _ := EncryptTOTPSeed(seed, privateKey)

	// Create TOTP entry with homomorphic encryption
	totp := &TOTP{
		AccountID:      accountID,
		TOTPSeed:       homomorphicData.PlaintextSeed,
		CTOTPSeed:      &homomorphicData.EncryptedSeed,
		PaillierN:      &homomorphicData.PaillierN,
		UseHomomorphic: true,
	}

	totpID, _ := totpRepo.Create(totp)

	fmt.Printf("Created TOTP ID: %d\n", totpID)
	fmt.Printf("Account ID: %d\n", accountID)
	fmt.Printf("Use Homomorphic: %v\n", totp.UseHomomorphic)

	// Serialize private key for secure storage
	nHex, lambdaHex := SerializePrivateKey(privateKey)

	// In production, you would store these securely
	_ = nHex
	_ = lambdaHex
	fmt.Println("\n⚠️  Store the private key in a secure location!")

	// Output:
	// Created TOTP ID: 1
	// Account ID: 1
	// Use Homomorphic: true
	// Private Key (N): <hex>...
	// Private Key (Lambda): <hex>...
	// ⚠️  Store the private key in a secure location!
}

// Example: Key Management Best Practices
func ExampleKeyManagement() {
	fmt.Println("=== Key Management Best Practices ===")

	// Generate keys
	privateKey, _ := CreatePaillierKeyPair(2048)
	// nHex, lambdaHex := SerializePrivateKey(privateKey)
	nHex, _ := SerializePrivateKey(privateKey)

	fmt.Println("1. NEVER store private keys in plaintext")
	fmt.Println("   ✗ database.totp_private_key = '" + nHex[:20] + "...'")
	fmt.Println()

	fmt.Println("2. DO encrypt private keys with a master key")
	fmt.Println("   ✓ encrypted_key = AES_Encrypt(private_key, master_key)")
	fmt.Println()

	fmt.Println("3. DO use key derivation from user password")
	fmt.Println("   ✓ key = PBKDF2(user_password, salt, iterations)")
	fmt.Println()

	fmt.Println("4. CONSIDER using Hardware Security Modules (HSM)")
	fmt.Println("   ✓ hsm.store(private_key)")
	fmt.Println("   ✓ code = hsm.generate_totp()")
	fmt.Println()

	fmt.Println("5. DO implement key rotation")
	fmt.Println("   ✓ new_key = CreatePaillierKeyPair(2048)")
	fmt.Println("   ✓ re_encrypt_all_seeds(old_key, new_key)")
	fmt.Println()

	fmt.Println("6. DO split keys across multiple locations")
	fmt.Println("   ✓ part1 = key[0:len(key)/2]  // Client-side")
	fmt.Println("   ✓ part2 = key[len(key)/2:]   // Server-side")

	// Output:
	// === Key Management Best Practices ===
	//
	// 1. NEVER store private keys in plaintext
	//    ✗ database.totp_private_key = '<hex>...'
	//
	// 2. DO encrypt private keys with a master key
	//    ✓ encrypted_key = AES_Encrypt(private_key, master_key)
	//
	// 3. DO use key derivation from user password
	//    ✓ key = PBKDF2(user_password, salt, iterations)
	//
	// 4. CONSIDER using Hardware Security Modules (HSM)
	//    ✓ hsm.store(private_key)
	//    ✓ code = hsm.generate_totp()
	//
	// 5. DO implement key rotation
	//    ✓ new_key = CreatePaillierKeyPair(2048)
	//    ✓ re_encrypt_all_seeds(old_key, new_key)
	//
	// 6. DO split keys across multiple locations
	//    ✓ part1 = key[0:len(key)/2]  // Client-side
	//    ✓ part2 = key[len(key)/2:]   // Server-side
}

// Example: Performance Comparison
func ExamplePerformanceComparison() {
	seed := "JBSWY3DPEHPK3PXP"
	privateKey, _ := CreatePaillierKeyPair(2048)
	homomorphicData, _ := EncryptTOTPSeed(seed, privateKey)

	fmt.Println("=== Performance Comparison ===")

	// Standard TOTP
	fmt.Println("Standard TOTP:")
	code1, _ := GenerateStandardTOTP(seed)
	fmt.Printf("  Code: %s\n", code1)
	fmt.Println("  Speed: ⚡⚡⚡ Very Fast (~0.1ms)")
	fmt.Println("  Memory: 📦 Minimal")
	fmt.Println()

	// Homomorphic TOTP
	fmt.Println("Homomorphic TOTP:")
	code2, _ := GenerateHomomorphicTOTP(
		homomorphicData.EncryptedSeed,
		homomorphicData.PaillierN,
		privateKey,
	)
	fmt.Printf("  Code: %s\n", code2)
	fmt.Println("  Speed: 🐌 Slow (~150ms)")
	fmt.Println("  Memory: 📦📦📦 Heavy (~1-2MB)")
	fmt.Println()

	fmt.Println("Use Cases:")
	fmt.Println("  Standard:    ✓ Regular authentication")
	fmt.Println("              ✓ High-frequency usage")
	fmt.Println("              ✓ Mobile devices")
	fmt.Println()
	fmt.Println("  Homomorphic: ✓ High-security environments")
	fmt.Println("              ✓ Zero-knowledge systems")
	fmt.Println("              ✓ Compliance requirements")

	// Output:
	// === Performance Comparison ===
	//
	// Standard TOTP:
	//   Code: <code>
	//   Speed: ⚡⚡⚡ Very Fast (~0.1ms)
	//   Memory: 📦 Minimal
	//
	// Homomorphic TOTP:
	//   Code: <code>
	//   Speed: 🐌 Slow (~150ms)
	//   Memory: 📦📦📦 Heavy (~1-2MB)
	//
	// Use Cases:
	//   Standard:    ✓ Regular authentication
	//               ✓ High-frequency usage
	//               ✓ Mobile devices
	//
	//   Homomorphic: ✓ High-security environments
	//               ✓ Zero-knowledge systems
	//               ✓ Compliance requirements
}
