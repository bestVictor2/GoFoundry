package lock

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

type TokenGenerator interface {
	NextToken() (string, error)
}

type RandomTokenGenerator struct {
	length int
}

func NewRandomTokenGenerator(length int) *RandomTokenGenerator {
	if length <= 0 {
		length = defaultTokenLength
	}
	return &RandomTokenGenerator{length: length}
}

func (g *RandomTokenGenerator) NextToken() (string, error) {
	b := make([]byte, g.length)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return hex.EncodeToString(b), nil
}
