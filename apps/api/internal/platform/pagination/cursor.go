// Package pagination provides cursor encoding/decoding for stable pagination.
package pagination

import (
	"encoding/base64"
	"encoding/json"
	"errors"
)

// Cursor identifies a stable position in a sorted result set.
type Cursor struct {
	Key string `json:"k"`
	ID  string `json:"i"`
}

// ErrInvalidCursor is returned when a cursor string cannot be decoded.
var ErrInvalidCursor = errors.New("invalid cursor")

// Encode returns a base64url-encoded JSON representation of the cursor.
func Encode(c Cursor) string {
	if c.ID == "" {
		return ""
	}
	b, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// Decode parses a base64url-encoded JSON cursor.
func Decode(s string) (Cursor, error) {
	if s == "" {
		return Cursor{}, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return Cursor{}, ErrInvalidCursor
	}
	var c Cursor
	if err := json.Unmarshal(b, &c); err != nil {
		return Cursor{}, ErrInvalidCursor
	}
	if c.ID == "" {
		return Cursor{}, ErrInvalidCursor
	}
	return c, nil
}
