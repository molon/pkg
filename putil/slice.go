package putil

import (
	"math/rand"
	"strings"
)

func UniqueIntSlice(intSlice []int) []int {
	keys := make(map[int]bool)
	list := []int{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func UniqueStringSlice(stringSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range stringSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func NonEmptyStringSlice(stringSlice []string) []string {
	list := []string{}
	for _, entry := range stringSlice {
		if strings.TrimSpace(entry) != "" {
			list = append(list, entry)
		}
	}
	return list
}

func InStringSlice(s string, ss []string) bool {
	for _, a := range ss {
		if a == s {
			return true
		}
	}
	return false
}

func ShuffleStringSlice(words []string) []string {
	nws := make([]string, len(words))
	for idx, w := range words {
		nws[idx] = w
	}
	rand.Shuffle(len(nws), func(i, j int) {
		nws[i], nws[j] = nws[j], nws[i]
	})
	return nws
}
