package main

/*
 * Portions of this code are inspired by or derived from 'didiercrunch/paillier'
 * https://github.com/didiercrunch/paillier
 * * Copyright (c) 2014 Didier Amyot
 * Licensed under the MIT License
 */

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"time"
)

// --- PAILLIER CRYPTOSYSTEM ---

type (
	// PaillierPublicKey represents the public key for Paillier encryption
	PaillierPublicKey struct {
		N *big.Int
	}

	// PaillierPrivateKey represents the private key for Paillier encryption
	PaillierPrivateKey struct {
		PublicKey PaillierPublicKey
		Lambda    *big.Int
	}

	// PaillierCypher represents an encrypted value
	PaillierCypher struct {
		C *big.Int
	}
)

var (
	ZERO = big.NewInt(0)
	ONE  = big.NewInt(1)
)

// CreatePaillierKeyPair generates a new Paillier key pair with the given bit size
func CreatePaillierKeyPair(bits int) (*PaillierPrivateKey, error) {
	// Generate two prime numbers
	p, err := rand.Prime(rand.Reader, bits/2)
	if err != nil {
		return nil, fmt.Errorf("failed to generate prime p: %w", err)
	}

	q, err := rand.Prime(rand.Reader, bits/2)
	if err != nil {
		return nil, fmt.Errorf("failed to generate prime q: %w", err)
	}

	return CreatePaillierPrivateKey(p, q), nil
}

// CreatePaillierPrivateKey creates a private key from two primes
func CreatePaillierPrivateKey(p, q *big.Int) *PaillierPrivateKey {
	n := new(big.Int).Mul(p, q)
	lambda := new(big.Int).Mul(new(big.Int).Sub(p, ONE), new(big.Int).Sub(q, ONE))
	return &PaillierPrivateKey{
		PublicKey: PaillierPublicKey{N: n},
		Lambda:    lambda,
	}
}

// GetNSquare returns N^2
func (pk *PaillierPublicKey) GetNSquare() *big.Int {
	return new(big.Int).Mul(pk.N, pk.N)
}

// Encrypt encrypts a message using the public key
func (pk *PaillierPublicKey) Encrypt(m *big.Int, random io.Reader) (*PaillierCypher, error) {
	r, err := GetRandomNumberInMultiplicativeGroup(pk.N, random)
	if err != nil {
		return nil, err
	}

	nSquare := pk.GetNSquare()
	g := new(big.Int).Add(pk.N, ONE)
	gm := new(big.Int).Exp(g, m, nSquare)
	rn := new(big.Int).Exp(r, pk.N, nSquare)

	c := new(big.Int).Mod(new(big.Int).Mul(rn, gm), nSquare)
	return &PaillierCypher{C: c}, nil
}

// Add performs homomorphic addition on encrypted values
func (pk *PaillierPublicKey) Add(cyphers ...*PaillierCypher) *PaillierCypher {
	accumulator := big.NewInt(1)
	nSq := pk.GetNSquare()

	for _, c := range cyphers {
		accumulator.Mod(accumulator.Mul(accumulator, c.C), nSq)
	}

	return &PaillierCypher{C: accumulator}
}

// Mul performs homomorphic scalar multiplication
func (pk *PaillierPublicKey) Mul(cypher *PaillierCypher, scalar *big.Int) *PaillierCypher {
	return &PaillierCypher{C: new(big.Int).Exp(cypher.C, scalar, pk.GetNSquare())}
}

// Decrypt decrypts a cypher using the private key
func (priv *PaillierPrivateKey) Decrypt(cypher *PaillierCypher) *big.Int {
	mu := new(big.Int).ModInverse(priv.Lambda, priv.PublicKey.N)
	tmp := new(big.Int).Exp(cypher.C, priv.Lambda, priv.PublicKey.GetNSquare())
	return new(big.Int).Mod(new(big.Int).Mul(L(tmp, priv.PublicKey.N), mu), priv.PublicKey.N)
}

// L function for Paillier decryption
func L(u, n *big.Int) *big.Int {
	return new(big.Int).Div(new(big.Int).Sub(u, ONE), n)
}

// GetRandomNumberInMultiplicativeGroup generates a random number in the multiplicative group
func GetRandomNumberInMultiplicativeGroup(n *big.Int, random io.Reader) (*big.Int, error) {
	for {
		r, err := rand.Int(random, n)
		if err != nil {
			return nil, err
		}
		if r.Cmp(ZERO) > 0 && new(big.Int).GCD(nil, nil, r, n).Cmp(ONE) == 0 {
			return r, nil
		}
	}
}

// --- HOMOMORPHIC TOTP FUNCTIONS ---

// HomomorphicTOTPData stores encrypted TOTP data
type HomomorphicTOTPData struct {
	EncryptedSeed string // Hex-encoded encrypted seed
	PaillierN     string // Hex-encoded Paillier N
	PlaintextSeed string // Original Base32 seed for backwards compatibility
}

// GenerateTOTPFromParts generates a TOTP code from seed bytes and time interval
func GenerateTOTPFromParts(seedBytes []byte, interval uint64) string {
	// Convert interval to bytes (Big Endian) as per RFC 6238
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, interval)

	// HMAC-SHA1
	mac := hmac.New(sha1.New, seedBytes)
	mac.Write(buf)
	sum := mac.Sum(nil)

	// Dynamic truncation
	offset := sum[len(sum)-1] & 0xf
	binaryCode := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff

	otp := binaryCode % 1000000
	return fmt.Sprintf("%06d", otp)
}

// GenerateStandardTOTPGeneric generates either a standard or homomorphic TOTP code.
// Note: For homomorphic, it assumes the private key is accessible via a helper (e.g., global or retrieved).
func GenerateStandardTOTPGeneric(
	seedBase32 string,
	cseed string,
	pailliern string,
	homomorphic bool,
) (string, error) {
	if homomorphic {
		// In a real scenario, you must retrieve the private key from your secure vault/storage
		// For this implementation, we assume a helper exists to fetch it.
		// priv, err := GetStoredPaillierPrivateKey()
		// if err != nil {
		// 	return "", fmt.Errorf("failed to retrieve private key for homomorphic generation: %w", err)
		// }
		return GenerateHomomorphicTOTP(cseed, pailliern, priv)
	}

	return GenerateStandardTOTP(seedBase32)
}

// GenerateHomomorphicTOTP generates a TOTP code using homomorphic encryption
func GenerateHomomorphicTOTP(
	encryptedSeedHex, paillierNHex string,
	privateKey *PaillierPrivateKey,
) (string, error) {
	// 1. Parse encrypted seed (C) from Hex
	encSeedBytes, err := hex.DecodeString(encryptedSeedHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted seed: %w", err)
	}
	cSeed := &PaillierCypher{C: new(big.Int).SetBytes(encSeedBytes)}

	// 2. Parse Paillier N from Hex
	nBytes, err := hex.DecodeString(paillierNHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode Paillier N: %w", err)
	}
	n := new(big.Int).SetBytes(nBytes)

	// 3. Security: Verify N matches the provided private key
	if n.Cmp(privateKey.PublicKey.N) != 0 {
		return "", fmt.Errorf("Paillier N mismatch: encrypted data belongs to a different key")
	}

	// 4. Get current time interval
	interval := time.Now().Unix() / 30
	intervalInt := big.NewInt(interval)

	// 5. Encrypt current time interval (T)
	cTime, err := privateKey.PublicKey.Encrypt(intervalInt, rand.Reader)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt time interval: %w", err)
	}

	// 6. Define a Shift (Base) to separate the Seed and the Time in the homomorphic result.
	// We use a power of 2 larger than the maximum possible time interval/seed
	// to ensure bits don't overlap. 2^256 is safe for standard TOTP seeds.
	shift := new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil)

	// 7. Homomorphic operation: C_final = (C_seed ^ shift) * C_time (mod N^2)
	// This represents E(Seed * shift + Time)
	cShifted := privateKey.PublicKey.Mul(cSeed, shift)
	cFinal := privateKey.PublicKey.Add(cShifted, cTime)

	// 8. Decrypt the result to get the combined Plaintext: (Seed << 256) | Time
	combinedResult := privateKey.Decrypt(cFinal)

	// 9. Recover the parts via modular arithmetic
	recoveredInterval := new(big.Int).Mod(combinedResult, shift).Uint64()
	recoveredSeedInt := new(big.Int).Div(combinedResult, shift)
	recoveredSeedBytes := recoveredSeedInt.Bytes()

	// 10. Generate the final 6-digit TOTP code using standard HMAC-SHA1
	return GenerateTOTPFromParts(recoveredSeedBytes, recoveredInterval), nil
}

// GetStoredPaillierPrivateKey is a placeholder for your logic to retrieve the
// private key (e.g., from a database or environment variable).
// func GetStoredPaillierPrivateKey() (*PaillierPrivateKey, error) {
// 	// Example: return DeserializePrivateKey(os.Getenv("PAILLIER_N"), os.Getenv("PAILLIER_LAMBDA"))
// 	return nil, fmt.Errorf("private key retrieval not implemented")
// }
// // GenerateStandardTOTP generates a standard TOTP code from a Base32 seed
// func GenerateStandardTOTPGeneric(
// 	seedBase32 string,
// 	cseed string,
// 	pailliern string,
// 	homomorphic bool,
// ) (string, error) {
// 	if homomorphic {
// 		return "", nil // GenerateHomomorphicTOTP()
// 	}
//
// 	return GenerateStandardTOTP(seedBase32)
// }

// GenerateStandardTOTP generates a standard TOTP code from a Base32 seed
func GenerateStandardTOTP(seedBase32 string) (string, error) {
	seedBytes, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(seedBase32)
	if err != nil {
		return "", fmt.Errorf("failed to decode Base32 seed: %w", err)
	}

	interval := uint64(time.Now().Unix() / 30)
	return GenerateTOTPFromParts(seedBytes, interval), nil
}

// EncryptTOTPSeed encrypts a TOTP seed using Paillier encryption
func EncryptTOTPSeed(
	seedBase32 string,
	privateKey *PaillierPrivateKey,
) (*HomomorphicTOTPData, error) {
	// Decode Base32 seed to bytes
	seedBytes, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(seedBase32)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Base32 seed: %w", err)
	}

	// Convert seed bytes to big.Int
	seedInt := new(big.Int).SetBytes(seedBytes)

	// Encrypt the seed
	encryptedSeed, err := privateKey.PublicKey.Encrypt(seedInt, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt seed: %w", err)
	}

	return &HomomorphicTOTPData{
		EncryptedSeed: hex.EncodeToString(encryptedSeed.C.Bytes()),
		PaillierN:     hex.EncodeToString(privateKey.PublicKey.N.Bytes()),
		PlaintextSeed: seedBase32,
	}, nil
}

// GenerateHomomorphicTOTP generates a TOTP code using homomorphic encryption
// func GenerateHomomorphicTOTP(
// 	encryptedSeedHex, paillierNHex string,
// 	privateKey *PaillierPrivateKey,
// ) (string, error) {
// 	// Parse encrypted seed
// 	encSeedBytes, err := hex.DecodeString(encryptedSeedHex)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to decode encrypted seed: %w", err)
// 	}
// 	cSeed := &PaillierCypher{C: new(big.Int).SetBytes(encSeedBytes)}
//
// 	// Parse Paillier N
// 	nBytes, err := hex.DecodeString(paillierNHex)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to decode Paillier N: %w", err)
// 	}
// 	n := new(big.Int).SetBytes(nBytes)
//
// 	// Verify N matches the private key
// 	if n.Cmp(privateKey.PublicKey.N) != 0 {
// 		return "", fmt.Errorf("Paillier N mismatch")
// 	}
//
// 	// Get current time interval
// 	interval := time.Now().Unix() / 30
// 	intervalInt := big.NewInt(interval)
//
// 	// Encrypt current time interval
// 	cTime, err := privateKey.PublicKey.Encrypt(intervalInt, rand.Reader)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to encrypt time interval: %w", err)
// 	}
//
// 	// Use a large shift to separate seed and time (10^20)
// 	shift := new(big.Int).Exp(big.NewInt(10), big.NewInt(20), nil)
//
// 	// Homomorphic operation: (cSeed * shift) + cTime
// 	cShifted := privateKey.PublicKey.Mul(cSeed, shift)
// 	cFinal := privateKey.PublicKey.Add(cShifted, cTime)
//
// 	// Decrypt the result
// 	result := privateKey.Decrypt(cFinal)
//
// 	// Recover the parts using arithmetic
// 	recoveredInterval := new(big.Int).Mod(result, shift).Uint64()
// 	recoveredSeedInt := new(big.Int).Div(result, shift)
// 	recoveredSeedBytes := recoveredSeedInt.Bytes()
//
// 	// Generate TOTP code
// 	otpCode := GenerateTOTPFromParts(recoveredSeedBytes, recoveredInterval)
//
// 	return otpCode, nil
// }

// GenerateRandomTOTPSeed generates a random Base32-encoded TOTP seed
func GenerateRandomTOTPSeed() (string, error) {
	// Generate 20 random bytes (160 bits)
	seedBytes := make([]byte, 20)
	_, err := rand.Read(seedBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random seed: %w", err)
	}

	// Encode to Base32 without padding
	seed := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(seedBytes)
	return seed, nil
}

// VerifyTOTP verifies a TOTP code against a seed
func VerifyTOTP(code, seedBase32 string, windowSize int) (bool, error) {
	currentInterval := time.Now().Unix() / 30

	// Check current time and window
	for i := -windowSize; i <= windowSize; i++ {
		interval := uint64(currentInterval + int64(i))

		seedBytes, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(seedBase32)
		if err != nil {
			return false, fmt.Errorf("failed to decode seed: %w", err)
		}

		expectedCode := GenerateTOTPFromParts(seedBytes, interval)
		if code == expectedCode {
			return true, nil
		}
	}

	return false, nil
}

// SerializePrivateKey converts a private key to hex strings for storage
func SerializePrivateKey(key *PaillierPrivateKey) (n, lambda string) {
	return hex.EncodeToString(key.PublicKey.N.Bytes()), hex.EncodeToString(key.Lambda.Bytes())
}

// DeserializePrivateKey reconstructs a private key from hex strings
func DeserializePrivateKey(nHex, lambdaHex string) (*PaillierPrivateKey, error) {
	nBytes, err := hex.DecodeString(nHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode N: %w", err)
	}

	lambdaBytes, err := hex.DecodeString(lambdaHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Lambda: %w", err)
	}

	n := new(big.Int).SetBytes(nBytes)
	lambda := new(big.Int).SetBytes(lambdaBytes)

	return &PaillierPrivateKey{
		PublicKey: PaillierPublicKey{N: n},
		Lambda:    lambda,
	}, nil
}
