package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
)

func Message(data, key []byte) ([]byte, error) {
	h := hmac.New(sha256.New, key)

	if _, err := h.Write(data); err != nil {
		return nil, fmt.Errorf("invalid hashing sha256: %w", err)
	}

	dst := h.Sum(nil)

	return dst, nil
}

func CheckMessage(data, key, expectedHash []byte) (bool, error) {
	hash, err := Message(data, key)
	if err != nil {
		return false, fmt.Errorf("hashing message: %w", err)
	}

	return hmac.Equal(hash, expectedHash), nil
}
