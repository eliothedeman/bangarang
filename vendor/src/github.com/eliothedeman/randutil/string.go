package randutil

import "math/rand"

const (
	// Set of characters to use for generating random strings
	Alphabet     = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	Lower        = "abcdefghijklmnopqrstuvwxyz"
	Upper        = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Numerals     = "1234567890"
	Alphanumeric = Alphabet + Numerals
	Ascii        = Alphanumeric + "~!@#$%^&*()-_+={}[]\\|<,>.?/\"';:`"
)

// String returns a random string n characters long, composed of entities
// from charset.
func String(length int, charset string) string {
	randstr := make([]byte, length) // Random string to return
	size := len(charset)
	for i := 0; i < length; i++ {
		randstr[i] = charset[rand.Int()%size]
	}
	return string(randstr)
}

// AlphaString returns a random alphanumeric string n characters long.
func AlphaString(length int) string {
	return String(length, Alphanumeric)
}
