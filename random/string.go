package random

import (
	"math/rand"
	"time"
	"unsafe"
)

const Letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

// Generates a random string. Characters used are regular ASCII 1 byte runes, defined
// in the Letters constant
//
// Algorithm comes from https://stackoverflow.com/a/31832326/351295
func String(size int) string {
	buf := make([]byte, size)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := size-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(Letters) {
			buf[i] = Letters[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return *(*string)(unsafe.Pointer(&buf))
}
