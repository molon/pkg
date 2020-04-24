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
