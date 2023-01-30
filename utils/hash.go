package utils

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

func HashSha256(in string) string {
	h := sha256.New()
	h.Write([]byte(in))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func HashCRSString(s string) string {
	h := sha256.New()
	for _, line := range strings.Split(s, "\n") {
		line = strings.Trim(line, "\\")
		line = strings.TrimSpace(line)
		h.Write([]byte(line))
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
