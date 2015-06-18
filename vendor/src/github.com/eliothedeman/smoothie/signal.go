package smoothie

import (
	"math"
	"math/cmplx"
	"math/rand"

	"github.com/mjibson/go-dsp/fft"
)

// create a new data frame filled with a signal given a frequencey and a phase
func NewSignal(length int, freq float64) *DataFrame {
	df := NewDataFrame(length)

	for i := 0; i < length; i++ {
		df.Insert(i, math.Sin(float64(i)/float64(df.Len())*math.Pi*2*freq))
	}

	return df
}

func Noise(length int) *DataFrame {
	df := NewDataFrame(length)

	for i := 0; i < length; i++ {
		df.Insert(i, rand.Float64())
	}

	return df
}

// Return the top frequencies found in the dataframe
func (d *DataFrame) FFTTopFreqs() *DataFrame {
	f := d.FFT()

	// cut out imposible freqs
	f = f[0 : len(f)/2]

	// create a dataframe out of the absolute values from the fft
	r := NewDataFrame(len(f))
	for _, i := range f {
		r.Push(cmplx.Abs(i))
	}

	stdDev := r.StdDev()
	top := NewDataFrame(0)
	for i := 1; i < r.Len(); i++ {
		if r.Index(i) > stdDev*5 {
			top.Grow(1)
			top.Push(float64(i))
		}
	}

	return top

}

// Return the real and imaginary parts of the fft
func (d *DataFrame) FFT() []complex128 {
	data := d.Data()

	return fft.FFTReal(data)
}
