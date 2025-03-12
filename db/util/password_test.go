package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPassword(t *testing.T) {
	password := RandomString(6)

	// Test case untuk hash password berhasil
	hashedPassword1, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword1)

	// Test case untuk verifikasi password yang benar
	err = CheckPassword(password, hashedPassword1)
	require.NoError(t, err)

	// Test case untuk verifikasi password yang salah
	wrongPassword := "password456"
	err = CheckPassword(wrongPassword, hashedPassword1)
	require.Error(t, err)

	// Test case untuk memastikan hash yang berbeda untuk password yang sama
	hashedPassword2, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword2)
	require.NotEqual(t, hashedPassword1, hashedPassword2)

	// Test case untuk password kosong
	emptyPassword := ""
	hashedEmpty, err := HashPassword(emptyPassword)
	require.NoError(t, err)
	require.NotEmpty(t, hashedEmpty)
}
