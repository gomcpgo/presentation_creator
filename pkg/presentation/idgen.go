package presentation

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/gosimple/slug"
)

const (
	MaxSlugLength = 30
	SuffixLength  = 4
)

// GenerateID creates a unique presentation ID from a name
func GenerateID(name string, existsFunc func(string) bool) string {
	slugified := slug.Make(name)

	if len(slugified) > MaxSlugLength {
		slugified = slugified[:MaxSlugLength]
	}

	if slugified == "" {
		slugified = "presentation"
	}

	for i := 0; i < 100; i++ {
		suffix := generateRandomSuffix()
		id := fmt.Sprintf("%s-%s", slugified, suffix)
		if !existsFunc(id) {
			return id
		}
	}

	longSuffix := generateRandomSuffix() + generateRandomSuffix()
	return fmt.Sprintf("%s-%s", slugified, longSuffix)
}

func generateRandomSuffix() string {
	bytes := make([]byte, SuffixLength/2+1)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("%04x", len(bytes))
	}
	return hex.EncodeToString(bytes)[:SuffixLength]
}

// ValidateID checks if a presentation ID is valid
func ValidateID(id string) bool {
	if id == "" {
		return false
	}
	if !strings.Contains(id, "-") {
		return false
	}
	if len(id) > MaxSlugLength+SuffixLength+1 {
		return false
	}
	return true
}
