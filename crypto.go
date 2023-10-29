package main

import (
	"crypto/cipher"
	"crypto/aes"
	"crypto/rand"
	"encoding/hex"
)

type crypto struct {
	key []byte
}

// pass in hex string
func NewCrypto(key string) *crypto{
	if key == "" {
		panic("Key is empty")
	}
	kb, _:= hex.DecodeString(key)
	return &crypto{kb }
}

func (c *crypto) Encrypt(s string, nonce []byte) {
	
}