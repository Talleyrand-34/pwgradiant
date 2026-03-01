package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

func generateTOTP(secret string) (string, error) {
	// 1. Decodificar la semilla Base32
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		return "", err
	}

	// 2. Obtener el intervalo de tiempo (ventanas de 30 segundos)
	epoch := time.Now().Unix()
	interval := uint64(math.Floor(float64(epoch) / 30.0))

	// Convertir el intervalo a un arreglo de bytes (Big Endian)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, interval)

	// 3. Calcular el HMAC-SHA1
	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	sum := mac.Sum(nil)

	// 4. "Dynamic Truncation" para obtener el código de 6 dígitos
	offset := sum[len(sum)-1] & 0xf
	binaryCode := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff

	otp := binaryCode % 1000000
	return fmt.Sprintf("%06d", otp), nil
}

func main() {
	// Ejemplo de semilla (Base32)
	seed := "JBSWY3DPEHPK3PXP"
	code, _ := generateTOTP(seed)
	fmt.Printf("Tu código OTP actual es: %s\n", code)
}
