package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"
)

func itoa(n int) string { return strconv.Itoa(n) }

func itoa64(n int64) string { return strconv.FormatInt(n, 10) }

// hashStrings builds a stable digest of an ordered list of strings. The
// fingerprint must not depend on map iteration order, so callers pass a
// slice rather than a map.
func hashStrings(parts []string) string {
	h := sha256.New()
	for _, p := range parts {
		h.Write([]byte(p))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))[:32]
}

// NowFunc is the clock used for every persisted timestamp. It is a variable
// rather than a direct call to time.Now so that tests can freeze it and
// assert on deterministic records.
var NowFunc = func() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// Now returns the current time in the format used across the persisted
// records. One format everywhere keeps the store diffable.
func Now() string { return NowFunc() }
