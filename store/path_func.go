package store

import (
	"encoding/hex"
	"hash"
	"path/filepath"
)

// this seems hacky. using two params so that
// i can either write by content addr or name
// maybe i don't need both...
type PathFunc func(hash.Hash, string) string

func awsContentPath(h hash.Hash, unused string) string {
	b := h.Sum(nil)
	topDir := hex.EncodeToString(b[:1])
	subDir := hex.EncodeToString(b[1:2])
	fname := hex.EncodeToString(b)
	return filepath.Join(topDir, subDir, fname)
}

func fileNamePath(h hash.Hash, key string) string {
	return key
}
