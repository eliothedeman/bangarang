package smoothie

import "testing"

func TestDerivative(t *testing.T) {
	df := NewDataFrame(10)

	for i := 0; i < 10; i++ {
		df.Push(float64((i * 10)))
	}

	der := df.Derivative()

	der.ForEach(func(v float64, i int) {
		if v != 10 {
			t.Fail()
		}
	})
}
