package main

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"testing"
	"time"
)

// --- PAILLIER CRYPTOSYSTEM TESTS ---

func TestCreatePaillierKeyPair(t *testing.T) {
	keyPair, err := CreatePaillierKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to create key pair: %v", err)
	}

	if keyPair == nil {
		t.Fatal("Key pair is nil")
	}

	if keyPair.PublicKey.N == nil {
		t.Fatal("N is nil")
	}

	if keyPair.Lambda == nil {
		t.Fatal("Lambda is nil")
	}

	// N should be roughly 2048 bits
	bitLen := keyPair.PublicKey.N.BitLen()
	if bitLen < 2000 || bitLen > 2100 {
		t.Errorf("N bit length out of range: got %d, expected ~2048", bitLen)
	}
}

func TestPaillierEncryptDecrypt(t *testing.T) {
	// Create small key for faster testing
	p, _ := rand.Prime(rand.Reader, 512)
	q, _ := rand.Prime(rand.Reader, 512)
	privateKey := CreatePaillierPrivateKey(p, q)

	// Test with various messages
	testCases := []int64{0, 1, 42, 1000, 999999}

	for _, tc := range testCases {
		message := big.NewInt(tc)

		// Encrypt
		encrypted, err := privateKey.PublicKey.Encrypt(message, rand.Reader)
		if err != nil {
			t.Fatalf("Encryption failed for message %d: %v", tc, err)
		}

		// Decrypt
		decrypted := privateKey.Decrypt(encrypted)

		// Verify
		if decrypted.Cmp(message) != 0 {
			t.Errorf("Decryption mismatch: expected %d, got %s", tc, decrypted.String())
		}
	}
}

func TestPaillierHomomorphicAddition(t *testing.T) {
	p, _ := rand.Prime(rand.Reader, 512)
	q, _ := rand.Prime(rand.Reader, 512)
	privateKey := CreatePaillierPrivateKey(p, q)

	// Encrypt two numbers
	m1 := big.NewInt(100)
	m2 := big.NewInt(200)

	c1, _ := privateKey.PublicKey.Encrypt(m1, rand.Reader)
	c2, _ := privateKey.PublicKey.Encrypt(m2, rand.Reader)

	// Homomorphic addition
	cSum := privateKey.PublicKey.Add(c1, c2)

	// Decrypt result
	result := privateKey.Decrypt(cSum)

	// Expected: 100 + 200 = 300
	expected := big.NewInt(300)

	if result.Cmp(expected) != 0 {
		t.Errorf(
			"Homomorphic addition failed: expected %s, got %s",
			expected.String(),
			result.String(),
		)
	}
}

func TestPaillierHomomorphicScalarMultiplication(t *testing.T) {
	p, _ := rand.Prime(rand.Reader, 512)
	q, _ := rand.Prime(rand.Reader, 512)
	privateKey := CreatePaillierPrivateKey(p, q)

	// Encrypt a number
	m := big.NewInt(50)
	scalar := big.NewInt(3)

	c, _ := privateKey.PublicKey.Encrypt(m, rand.Reader)

	// Homomorphic scalar multiplication
	cMul := privateKey.PublicKey.Mul(c, scalar)

	// Decrypt result
	result := privateKey.Decrypt(cMul)

	// Expected: 50 * 3 = 150
	expected := big.NewInt(150)

	if result.Cmp(expected) != 0 {
		t.Errorf(
			"Homomorphic multiplication failed: expected %s, got %s",
			expected.String(),
			result.String(),
		)
	}
}

func TestPaillierMultipleOperations(t *testing.T) {
	p, _ := rand.Prime(rand.Reader, 512)
	q, _ := rand.Prime(rand.Reader, 512)
	privateKey := CreatePaillierPrivateKey(p, q)

	// Test: (10 * 5) + (20 * 3) = 50 + 60 = 110
	m1 := big.NewInt(10)
	m2 := big.NewInt(20)
	s1 := big.NewInt(5)
	s2 := big.NewInt(3)

	c1, _ := privateKey.PublicKey.Encrypt(m1, rand.Reader)
	c2, _ := privateKey.PublicKey.Encrypt(m2, rand.Reader)

	// (c1 * 5) + (c2 * 3)
	c1Mul := privateKey.PublicKey.Mul(c1, s1)
	c2Mul := privateKey.PublicKey.Mul(c2, s2)
	cFinal := privateKey.PublicKey.Add(c1Mul, c2Mul)

	result := privateKey.Decrypt(cFinal)
	expected := big.NewInt(110)

	if result.Cmp(expected) != 0 {
		t.Errorf(
			"Multiple operations failed: expected %s, got %s",
			expected.String(),
			result.String(),
		)
	}
}

// --- TOTP GENERATION TESTS ---

func TestGenerateTOTPFromParts(t *testing.T) {
	// Standard test seed from RFC 6238
	seedBytes := []byte("12345678901234567890")
	interval := uint64(59 / 30) // Should be 1

	code := GenerateTOTPFromParts(seedBytes, interval)

	// Code should be 6 digits
	if len(code) != 6 {
		t.Errorf("Expected 6-digit code, got %d digits: %s", len(code), code)
	}

	// Should be numeric
	for _, c := range code {
		if c < '0' || c > '9' {
			t.Errorf("Code contains non-numeric character: %c", c)
		}
	}
}

func TestGenerateStandardTOTP(t *testing.T) {
	seedBase32 := "JBSWY3DPEHPK3PXP"

	code, err := GenerateStandardTOTP(seedBase32)
	if err != nil {
		t.Fatalf("Failed to generate standard TOTP: %v", err)
	}

	if len(code) != 6 {
		t.Errorf("Expected 6-digit code, got %d digits: %s", len(code), code)
	}
}

func TestGenerateStandardTOTPInvalidSeed(t *testing.T) {
	invalidSeed := "INVALID!@#$"

	_, err := GenerateStandardTOTP(invalidSeed)
	if err == nil {
		t.Error("Expected error for invalid Base32 seed")
	}
}

func TestVerifyTOTP(t *testing.T) {
	seed := "JBSWY3DPEHPK3PXP"

	// Generate a code
	code, err := GenerateStandardTOTP(seed)
	if err != nil {
		t.Fatalf("Failed to generate TOTP: %v", err)
	}

	// Verify it
	valid, err := VerifyTOTP(code, seed, 1)
	if err != nil {
		t.Fatalf("Verification failed: %v", err)
	}

	if !valid {
		t.Error("Expected code to be valid")
	}
}

func TestVerifyTOTPInvalidCode(t *testing.T) {
	seed := "JBSWY3DPEHPK3PXP"
	invalidCode := "000000"

	valid, err := VerifyTOTP(invalidCode, seed, 1)
	if err != nil {
		t.Fatalf("Verification failed: %v", err)
	}

	if valid {
		t.Error("Expected invalid code to fail verification")
	}
}

func TestGenerateRandomTOTPSeed(t *testing.T) {
	seed, err := GenerateRandomTOTPSeed()
	if err != nil {
		t.Fatalf("Failed to generate random seed: %v", err)
	}

	// Should be Base32
	if len(seed) == 0 {
		t.Error("Generated seed is empty")
	}

	// Should be decodable
	_, err = GenerateStandardTOTP(seed)
	if err != nil {
		t.Errorf("Generated seed is not valid Base32: %v", err)
	}
}

// --- HOMOMORPHIC TOTP TESTS ---

func TestEncryptTOTPSeed(t *testing.T) {
	seed := "JBSWY3DPEHPK3PXP"
	keyPair, _ := CreatePaillierKeyPair(2048)

	data, err := EncryptTOTPSeed(seed, keyPair)
	if err != nil {
		t.Fatalf("Failed to encrypt TOTP seed: %v", err)
	}

	if data.EncryptedSeed == "" {
		t.Error("Encrypted seed is empty")
	}

	if data.PaillierN == "" {
		t.Error("Paillier N is empty")
	}

	if data.PlaintextSeed != seed {
		t.Errorf("Plaintext seed mismatch: expected %s, got %s", seed, data.PlaintextSeed)
	}

	// Verify it's valid hex
	_, err = hex.DecodeString(data.EncryptedSeed)
	if err != nil {
		t.Errorf("Encrypted seed is not valid hex: %v", err)
	}
}

func TestGenerateHomomorphicTOTP(t *testing.T) {
	seed := "JBSWY3DPEHPK3PXP"
	keyPair, _ := CreatePaillierKeyPair(2048)

	// Encrypt the seed
	data, err := EncryptTOTPSeed(seed, keyPair)
	if err != nil {
		t.Fatalf("Failed to encrypt seed: %v", err)
	}

	// Generate TOTP using homomorphic encryption
	code, err := GenerateHomomorphicTOTP(data.EncryptedSeed, data.PaillierN, keyPair)
	if err != nil {
		t.Fatalf("Failed to generate homomorphic TOTP: %v", err)
	}

	if len(code) != 6 {
		t.Errorf("Expected 6-digit code, got %d digits: %s", len(code), code)
	}

	// Compare with standard TOTP (should match within time window)
	standardCode, _ := GenerateStandardTOTP(seed)

	// They should match if generated at the same time
	// Note: There's a small chance they differ if crossing a 30-second boundary
	if code != standardCode {
		t.Logf(
			"Warning: Homomorphic (%s) and standard (%s) codes differ (might be due to timing)",
			code,
			standardCode,
		)

		// Verify both codes are valid
		valid1, _ := VerifyTOTP(code, seed, 1)
		valid2, _ := VerifyTOTP(standardCode, seed, 1)

		if !valid1 || !valid2 {
			t.Error("One or both codes are invalid")
		}
	}
}

func TestGenerateHomomorphicTOTPWithMismatchedKey(t *testing.T) {
	seed := "JBSWY3DPEHPK3PXP"
	keyPair1, _ := CreatePaillierKeyPair(2048)
	keyPair2, _ := CreatePaillierKeyPair(2048)

	// Encrypt with keyPair1
	data, _ := EncryptTOTPSeed(seed, keyPair1)

	// Try to generate with keyPair2 (should fail)
	_, err := GenerateHomomorphicTOTP(data.EncryptedSeed, data.PaillierN, keyPair2)
	if err == nil {
		t.Error("Expected error when using mismatched key")
	}
}

func TestSerializeDeserializePrivateKey(t *testing.T) {
	keyPair, _ := CreatePaillierKeyPair(2048)

	// Serialize
	nHex, lambdaHex := SerializePrivateKey(keyPair)

	// Deserialize
	restored, err := DeserializePrivateKey(nHex, lambdaHex)
	if err != nil {
		t.Fatalf("Failed to deserialize key: %v", err)
	}

	// Compare
	if keyPair.PublicKey.N.Cmp(restored.PublicKey.N) != 0 {
		t.Error("N values don't match after serialization")
	}

	if keyPair.Lambda.Cmp(restored.Lambda) != 0 {
		t.Error("Lambda values don't match after serialization")
	}
}

func TestHomomorphicTOTPConsistency(t *testing.T) {
	// Test that homomorphic TOTP produces consistent results
	seed := "JBSWY3DPEHPK3PXP"
	keyPair, _ := CreatePaillierKeyPair(2048)
	data, _ := EncryptTOTPSeed(seed, keyPair)

	// Generate multiple codes in rapid succession
	codes := make([]string, 5)
	for i := 0; i < 5; i++ {
		code, err := GenerateHomomorphicTOTP(data.EncryptedSeed, data.PaillierN, keyPair)
		if err != nil {
			t.Fatalf("Failed to generate code: %v", err)
		}
		codes[i] = code
		time.Sleep(100 * time.Millisecond)
	}

	// All codes should be the same (within the same 30-second window)
	firstCode := codes[0]
	for i, code := range codes {
		if code != firstCode {
			t.Errorf("Code %d differs: %s vs %s", i, code, firstCode)
		}
	}
}

func TestHomomorphicTOTPTimeProgression(t *testing.T) {
	// Test that codes change over time intervals
	seed := "JBSWY3DPEHPK3PXP"
	keyPair, _ := CreatePaillierKeyPair(2048)
	data, _ := EncryptTOTPSeed(seed, keyPair)

	// Get current code
	code1, _ := GenerateHomomorphicTOTP(data.EncryptedSeed, data.PaillierN, keyPair)

	// Wait for next interval (this test may take 30+ seconds)
	t.Log("Waiting for next TOTP interval (may take up to 30 seconds)...")
	currentInterval := time.Now().Unix() / 30
	for time.Now().Unix()/30 == currentInterval {
		time.Sleep(1 * time.Second)
	}

	// Get new code
	code2, _ := GenerateHomomorphicTOTP(data.EncryptedSeed, data.PaillierN, keyPair)

	// Codes should be different
	if code1 == code2 {
		t.Error("TOTP codes should change between time intervals")
	}
}

// --- INTEGRATION TESTS ---

func TestHomomorphicTOTPEndToEnd(t *testing.T) {
	// Complete workflow

	// 1. Generate random seed
	seed, err := GenerateRandomTOTPSeed()
	if err != nil {
		t.Fatalf("Failed to generate seed: %v", err)
	}

	// 2. Create key pair
	keyPair, err := CreatePaillierKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to create key pair: %v", err)
	}

	// 3. Encrypt seed
	data, err := EncryptTOTPSeed(seed, keyPair)
	if err != nil {
		t.Fatalf("Failed to encrypt seed: %v", err)
	}

	// 4. Serialize private key
	nHex, lambdaHex := SerializePrivateKey(keyPair)

	// 5. Simulate storage/retrieval by deserializing
	restoredKey, err := DeserializePrivateKey(nHex, lambdaHex)
	if err != nil {
		t.Fatalf("Failed to deserialize key: %v", err)
	}

	// 6. Generate homomorphic TOTP code
	homomorphicCode, err := GenerateHomomorphicTOTP(data.EncryptedSeed, data.PaillierN, restoredKey)
	if err != nil {
		t.Fatalf("Failed to generate homomorphic code: %v", err)
	}

	// 7. Generate standard TOTP code
	standardCode, err := GenerateStandardTOTP(seed)
	if err != nil {
		t.Fatalf("Failed to generate standard code: %v", err)
	}

	// 8. Verify both codes work
	valid1, err := VerifyTOTP(homomorphicCode, seed, 1)
	if err != nil || !valid1 {
		t.Error("Homomorphic code failed verification")
	}

	valid2, err := VerifyTOTP(standardCode, seed, 1)
	if err != nil || !valid2 {
		t.Error("Standard code failed verification")
	}

	t.Logf(
		"Successfully generated and verified codes: homomorphic=%s, standard=%s",
		homomorphicCode,
		standardCode,
	)
}

// --- BENCHMARK TESTS ---

func BenchmarkStandardTOTP(b *testing.B) {
	seed := "JBSWY3DPEHPK3PXP"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateStandardTOTP(seed)
	}
}

func BenchmarkHomomorphicTOTP(b *testing.B) {
	seed := "JBSWY3DPEHPK3PXP"
	keyPair, _ := CreatePaillierKeyPair(2048)
	data, _ := EncryptTOTPSeed(seed, keyPair)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateHomomorphicTOTP(data.EncryptedSeed, data.PaillierN, keyPair)
	}
}

func BenchmarkPaillierEncryption(b *testing.B) {
	keyPair, _ := CreatePaillierKeyPair(2048)
	message := big.NewInt(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		keyPair.PublicKey.Encrypt(message, rand.Reader)
	}
}

func BenchmarkPaillierDecryption(b *testing.B) {
	keyPair, _ := CreatePaillierKeyPair(2048)
	message := big.NewInt(12345)
	encrypted, _ := keyPair.PublicKey.Encrypt(message, rand.Reader)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		keyPair.Decrypt(encrypted)
	}
}
