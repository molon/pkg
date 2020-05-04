package putil

import (
	"math/rand"
	"strings"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandLowerString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func SplitLineBreak(str string) []string {
	return SplitTrimNonEmpty(str, "\n")
}

func SplitTrimNonEmpty(str string, sep string) []string {
	ret := []string{}
	lines := strings.Split(str, sep)
	for _, line := range lines {
		line := strings.TrimSpace(line)
		if line == "" {
			continue
		}
		ret = append(ret, line)
	}
	return ret
}

func MaskMiddle(str string) string {
	if len(str) <= 0 {
		return ""
	}

	half := len(str) / 2

	for len(str)-2*half <= 0 {
		half--
	}

	if half > 6 {
		half = 6
	}
	strLen := len(str) - 2*half
	if strLen > 20 {
		strLen = 20
	}

	ret := str[:half]
	for idx := 0; idx < strLen; idx++ {
		ret += "*"
	}
	ret += str[len(str)-half:]
	return ret
}
