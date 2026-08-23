package query

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

type Cursor struct {
	LastID   string
	Snapshot string
}

func EncodeCursor(c Cursor) string {
	raw := c.LastID + "|" + c.Snapshot
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}
func DecodeCursor(value string) (Cursor, error) {
	if value == "" {
		return Cursor{}, nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return Cursor{}, fmt.Errorf("cursor decode: %w", err)
	}
	parts := strings.Split(string(raw), "|")
	if len(parts) != 2 || parts[0] == "" {
		return Cursor{}, fmt.Errorf("invalid cursor")
	}
	return Cursor{LastID: parts[0], Snapshot: parts[1]}, nil
}
func SnapshotToken(value string) string {
	sum := sha256.Sum256([]byte(value))
	return fmt.Sprintf("%x", sum[:8])
}
func ParseLimit(value string, defaultLimit, max int) int {
	n, err := strconv.Atoi(value)
	if err != nil || n < 1 {
		return defaultLimit
	}
	if n > max {
		return max
	}
	return n
}
