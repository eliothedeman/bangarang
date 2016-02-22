package event

import (
	"fmt"
	"testing"
)

func TestTagSetSet(t *testing.T) {
	ts := TagSet{}
	for i := 0; i < 100; i++ {
		ts.Set("yes", fmt.Sprintf("%d", i))

	}

	x := 0
	ts.ForEach(func(k, v string) {
		if k != "yes" {
			t.Fatal(k)
		}

		if v != fmt.Sprintf("%d", x) {
			t.Fatal(v)
		}

		x++

	})

}
