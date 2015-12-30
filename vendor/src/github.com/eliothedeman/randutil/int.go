package randutil

import "math/rand"

// IntRange returns a random integer in the range from min to max.
func IntRange(min, size int) int {
	return (rand.Int() % size) + min
}
