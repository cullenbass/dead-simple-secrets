package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"log/slog"
)

type crypto struct {
	gcmCipher cipher.AEAD
}

func GenNewKey() []byte {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		panic(err)
	}
	return key
}

func NewCrypto(key []byte) *crypto {
	c, err := aes.NewCipher(key)
	if err != nil {
		slog.Error("Cannot create cipher")
		panic(err)
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		slog.Error("Cannot create GCM")
		panic(err)
	}
	return &crypto{gcm}
}

func (c *crypto) Encrypt(plaintext string) (encrypted []byte, nonce []byte) {
	nonce = make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err)
	}
	encrypted = c.gcmCipher.Seal(nil, nonce, []byte(plaintext), nil)
	return
}

func (c *crypto) Decrypt(ciphertext []byte, nonce []byte) (plaintext string) {
	plainBytes, err := c.gcmCipher.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err)
	}
	return string(plainBytes)
}
