package yams

import (
	"math/rand"
	"time"
)

const (
	randBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	randIdxBits = 6
	randIdxMask = 1<<randIdxBits - 1
	randIdxMax  = 63 / randIdxBits
)

var randSource = rand.NewSource(time.Now().UnixNano())

// https://stackoverflow.com/a/31832326/1249581
func RandString(n int) string {
	out := make([]byte, n)

	for i, cache, remain := n-1, randSource.Int63(), randIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = randSource.Int63(), randIdxMax
		}
		if idx := int(cache & randIdxMask); idx < len(randBytes) {
			out[i] = randBytes[idx]
			i--
		}
		cache >>= randIdxBits
		remain--
	}

	return string(out)
}
