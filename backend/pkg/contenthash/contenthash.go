package contenthash

import (
	"crypto/sha256"
	"encoding/hex"
)

func SHA256(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}
