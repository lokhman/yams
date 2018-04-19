package utils

import (
	"unicode"
)

func IsBinaryString(s string) bool {
	for _, r := range s {
		if !unicode.IsSpace(r) && !unicode.IsPrint(r) {
			return true
		}
	}
	return false
}

func IsTextString(s string) bool {
	return !IsBinaryString(s)
}
