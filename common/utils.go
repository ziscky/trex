package common

import (
	"crypto/sha1"
	"encoding/hex"
)

//NameHash gets the sha1 hex of a name
func NameHash(name string) string {
	h := sha1.New()
	h.Write([]byte(name))
	return hex.EncodeToString(h.Sum(nil))

}
