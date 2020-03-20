package putil

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandIntn(min int, max int) int {
	return min + rand.Intn(max-min+1)
}

func RandInt63n(min int64, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}
