package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// EncryptWithAES256 encrypts data using a AES256 cryptography.
func EncryptWithAES256(secretKey, data []byte) ([]byte, error) {
	if len(secretKey) != 32 {
		return nil, fmt.Errorf("secret key is not for AES-256: total %d bits", 8*len(secretKey))
	}

	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	cipherText := aesGCM.Seal(nonce, nonce, data, nil)

	return cipherText, nil
}

// DecryptWithAES256 decrypts data using a AES256 cryptography.
func DecryptWithAES256(secretKey, ciphertext []byte) ([]byte, error) {
	if len(secretKey) != 32 {
		return nil, fmt.Errorf("secret key is not for AES-256: total %d bits", 8*len(secretKey))
	}

	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plainText, err := aesgcm.Open(nil, ciphertext[:aesgcm.NonceSize()], ciphertext[aesgcm.NonceSize():], nil)
	if err != nil {
		return nil, err
	}

	return plainText, nil
}
