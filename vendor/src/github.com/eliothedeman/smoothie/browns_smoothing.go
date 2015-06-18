package smoothie

import "math"

func (d *DataFrame) SingleExponentialSmooth(sf float64) *DataFrame {
	smoothed := EmptyDataFrame(d.Len())

	for i := 0; i < d.Len(); i++ {
		smoothed.Insert(i, d.singleSmoothPoint(i, sf, smoothed))
	}

	return smoothed
}

func (d *DataFrame) singleSmoothPoint(i int, sf float64, s *DataFrame) float64 {
	if i <= 1 {
		return (sf * d.Index(i)) + (1 - sf)
	}

	// check if the values has already been calculated
	if f := s.Index(i); !math.IsNaN(f) {
		return f
	}

	return ((sf * d.Index(i)) + ((1 - sf) * d.singleSmoothPoint(i-1, sf, s)))
}
