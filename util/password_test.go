package util

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	password := RandomString(12)

	hashedPassword1, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword1)

	err = CheckPassword(password, hashedPassword1)
	require.NoError(t, err)

	hashedPassword2, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEqual(t, hashedPassword1, hashedPassword2)

	wrongPassword := RandomString(8)

	err = CheckPassword(wrongPassword, hashedPassword1)
	require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())

	tooLongPassword := RandomString(100)

	hashedTooLongPassword, err := HashPassword(tooLongPassword)
	require.Empty(t, hashedTooLongPassword)
	require.EqualError(t, err, bcrypt.ErrPasswordTooLong.Error())
}
