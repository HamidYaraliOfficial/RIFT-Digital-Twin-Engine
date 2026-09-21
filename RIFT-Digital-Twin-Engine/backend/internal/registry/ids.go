package registry

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// NewID returns a short, unique, dependency-free identifier: "<prefix>_<12 hex chars>".
// It uses crypto/rand so collisions are effectively impossible without pulling
// in an external UUID library.
func NewID(prefix string) string {
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		// crypto/rand failing means the OS entropy source is broken; fall back
		// to a timestamp-based suffix rather than crashing the process.
		return fmt.Sprintf("%s_fallback", prefix)
	}
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(b))
}
