package util

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRandomInt(t *testing.T) {
	min, max := int64(-100), int64(100)
	result := RandomInt(min, max)
	require.GreaterOrEqual(t, result, min)
	require.LessOrEqual(t, result, max)
}

func TestRandomString(t *testing.T) {
	n := 10
	result := RandomString(n)
	require.Len(t, result, n)

	for _, char := range result {
		require.Contains(t, alphabet, string(char))
	}
}

func TestRandomOwner(t *testing.T) {
	result := RandomOwner()
	require.Len(t, result, 6)
}

func TestRandomMoney(t *testing.T) {
	min, max := int64(-5000), int64(5000)
	result := RandomMoney(min, max)

	val, err := strconv.ParseInt(result, 10, 64)
	require.NoError(t, err)
	require.GreaterOrEqual(t, val, min)
	require.LessOrEqual(t, val, max)
}

func TestRandomAmount(t *testing.T) {
	result, err := strconv.ParseInt(RandomAmount(500), 10, 64)
	require.NoError(t, err)
	require.Greater(t, result, int64(0))
}

func TestRandomCurrency(t *testing.T) {
	currencies := []string{"EUR", "USD", "BHD"}
	result := RandomCurrency()
	require.NotEmpty(t, result)
	require.Contains(t, currencies, result)
}
