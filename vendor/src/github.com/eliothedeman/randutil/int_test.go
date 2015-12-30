package randutil

import (
	"math/rand"
	"strings"
	"testing"
)

func TestIntRange(t *testing.T) {
	x := rand.Int()
	size := 10000

	for i := 0; i < 10000; i++ {
		r := IntRange(x, size)
		if r > x+size {
			t.Fail()
		}

		if r < x {
			t.Fail()
		}
	}
}

func TestStringLength(t *testing.T) {

	for i := 0; i < 10000; i++ {
		size := rand.Int() % 100

		s := String(size, Ascii)
		if len(s) != size {
			t.Fail()
		}
	}
}
func TestStringChars(t *testing.T) {

	chars := []string{
		Alphabet,
		Alphanumeric,
		Ascii,
		Upper,
		Lower,
		";laksdjfl;kajdfl",
	}

	for i := 0; i < 10000; i++ {
		size := rand.Int() % 100
		charSet := chars[rand.Int()%len(chars)]

		s := String(size, charSet)
		for i := 0; i < len(s); i++ {
			if !strings.Contains(charSet, string([]byte{s[i]})) {
				t.Fatalf("%s is not in %s", s[i], charSet)
			}
		}
	}
}
