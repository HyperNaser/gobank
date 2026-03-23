// Package util provides utils for database operations/testing
package util

import (
	"math/rand"
	"strconv"
	"strings"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for range n {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

func RandomOwner() string {
	return RandomString(6)
}

func RandomBalance() string {
	return strconv.FormatInt(RandomInt(0, 5000), 10)
}

func RandomCurrency() string {
	currencies := []string{"EUR", "USD", "BHD"}
	return currencies[rand.Intn(len(currencies))]
}
