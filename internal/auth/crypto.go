package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

const (
	version         = 1
	versionByteSize = 1
	nonceSize       = 12
)

var ErrInvalidCiphertext = errors.New("invalid ciphertext")

func Encrypt(plaintext []byte, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("key must be 32 bytes")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// Format: [version][nonce][ciphertext]
	result := make([]byte, versionByteSize+nonceSize+len(ciphertext))
	result[0] = version
	copy(result[versionByteSize:versionByteSize+nonceSize], nonce)
	copy(result[versionByteSize+nonceSize:], ciphertext)

	return result, nil
}

func Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	if len(ciphertext) < versionByteSize+nonceSize {
		return nil, ErrInvalidCiphertext
	}

	if ciphertext[0] != version {
		return nil, errors.New("unsupported version")
	}

	if len(key) != 32 {
		return nil, errors.New("key must be 32 bytes")
	}

	nonce := ciphertext[versionByteSize : versionByteSize+nonceSize]
	data := ciphertext[versionByteSize+nonceSize:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
