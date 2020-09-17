package token

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

var (
	ErrRawTooShort = errors.New(`token: raw too short`)
)

func NewGCM(key string) (cipher.AEAD, error) {
	c, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}

	return cipher.NewGCM(c)
}

func encrypt(raw []byte, gcm cipher.AEAD) ([]byte, error) {
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, raw, nil), nil
}

func decrypt(raw []byte, gcm cipher.AEAD) ([]byte, error) {
	ns := gcm.NonceSize()
	if len(raw) < ns {
		return nil, ErrRawTooShort
	}

	nonce, raw := raw[:ns], raw[ns:]
	return gcm.Open(nil, nonce, raw, nil)
}
