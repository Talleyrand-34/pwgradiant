package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"io"
	"math/big"
	"time"
)

// --- LÓGICA DE TOTP ---

func generateTOTPFromParts(seedBytes []byte, interval uint64) string {
	// Convertir el intervalo a bytes (Big Endian) como pide el RFC 6238
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, interval)

	// HMAC-SHA1
	mac := hmac.New(sha1.New, seedBytes)
	mac.Write(buf)
	sum := mac.Sum(nil)

	// Truncamiento dinámico
	offset := sum[len(sum)-1] & 0xf
	binaryCode := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff

	otp := binaryCode % 1000000
	return fmt.Sprintf("%06d", otp)
}

// --- MAIN ---

func main() {
	// 1. Configuración de Paillier (Llaves de 2048 bits para evitar desbordamiento)
	p, _ := rand.Prime(rand.Reader, 1024)
	q, _ := rand.Prime(rand.Reader, 1024)
	privateKey := CreatePrivateKey(p, q)

	// 2. Preparar Datos
	seedStr := "JBSWY3DPEHPK3PXP"
	// Decodificamos la semilla Base32 a bytes originales
	seedBytes, _ := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(seedStr)
	// Convertimos esos bytes a un número para Paillier
	seedInt := new(big.Int).SetBytes(seedBytes)

	interval := time.Now().Unix() / 30
	intervalInt := big.NewInt(interval)

	// 3. Cifrado y Operación Homomórfica
	// Usamos un shift de 10^20 para asegurar que no haya colisión entre semilla y tiempo
	shift := new(big.Int).Exp(big.NewInt(10), big.NewInt(20), nil)

	cSeed, _ := privateKey.Encrypt(seedInt, rand.Reader)
	cTime, _ := privateKey.Encrypt(intervalInt, rand.Reader)

	// Unión homomórfica: (cSeed * shift) + cTime
	cShifted := privateKey.PublicKey.Mul(cSeed, shift)
	cFinal := privateKey.PublicKey.Add(cShifted, cTime)

	// ---------------------------------------------------------
	// 4. PROCESO DE DESENCRIPTACIÓN Y GENERACIÓN DE OTP
	// ---------------------------------------------------------
	result := privateKey.Decrypt(cFinal)

	// Recuperar las partes usando aritmética
	recoveredInterval := new(big.Int).Mod(result, shift).Uint64()
	recoveredSeedInt := new(big.Int).Div(result, shift)
	recoveredSeedBytes := recoveredSeedInt.Bytes()

	// Generar el código final
	otpCode := generateTOTPFromParts(recoveredSeedBytes, recoveredInterval)

	fmt.Printf("--- Resultado Final ---\n")
	fmt.Printf("Semilla: %s | Intervalo: %d\n", seedStr, interval)
	fmt.Printf("Código TOTP generado desde datos cifrados: %s\n", otpCode)
}

// --- IMPLEMENTACIÓN DE PAILLIER (Tu librería corregida) ---

type (
	PublicKey  struct{ N *big.Int }
	PrivateKey struct {
		PublicKey
		Lambda *big.Int
	}
)
type Cypher struct{ C *big.Int }

var (
	ZERO = big.NewInt(0)
	ONE  = big.NewInt(1)
)

func (pk *PublicKey) GetNSquare() *big.Int { return new(big.Int).Mul(pk.N, pk.N) }

func (pk *PublicKey) Encrypt(m *big.Int, random io.Reader) (*Cypher, error) {
	r, err := GetRandomNumberInMultiplicativeGroup(pk.N, random)
	if err != nil {
		return nil, err
	}
	nSquare := pk.GetNSquare()
	g := new(big.Int).Add(pk.N, ONE)
	gm := new(big.Int).Exp(g, m, nSquare)
	rn := new(big.Int).Exp(r, pk.N, nSquare)
	return &Cypher{new(big.Int).Mod(new(big.Int).Mul(rn, gm), nSquare)}, nil
}

func (pk *PublicKey) Add(cypher ...*Cypher) *Cypher {
	accumulator := big.NewInt(1)
	nSq := pk.GetNSquare()
	for _, c := range cypher {
		accumulator.Mod(accumulator.Mul(accumulator, c.C), nSq)
	}
	return &Cypher{C: accumulator}
}

func (pk *PublicKey) Mul(cypher *Cypher, scalar *big.Int) *Cypher {
	return &Cypher{C: new(big.Int).Exp(cypher.C, scalar, pk.GetNSquare())}
}

func (priv *PrivateKey) Decrypt(cypher *Cypher) *big.Int {
	mu := new(big.Int).ModInverse(priv.Lambda, priv.N)
	tmp := new(big.Int).Exp(cypher.C, priv.Lambda, priv.GetNSquare())
	return new(big.Int).Mod(new(big.Int).Mul(L(tmp, priv.N), mu), priv.N)
}

func L(u, n *big.Int) *big.Int {
	return new(big.Int).Div(new(big.Int).Sub(u, ONE), n)
}

func CreatePrivateKey(p, q *big.Int) *PrivateKey {
	n := new(big.Int).Mul(p, q)
	lambda := new(big.Int).Mul(new(big.Int).Sub(p, ONE), new(big.Int).Sub(q, ONE))
	return &PrivateKey{PublicKey: PublicKey{N: n}, Lambda: lambda}
}

func GetRandomNumberInMultiplicativeGroup(n *big.Int, random io.Reader) (*big.Int, error) {
	for {
		r, _ := rand.Int(random, n)
		if r.Cmp(ZERO) > 0 && new(big.Int).GCD(nil, nil, r, n).Cmp(ONE) == 0 {
			return r, nil
		}
	}
}
