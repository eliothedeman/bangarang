package smoothie

import "testing"

func TestFFT(t *testing.T) {
	df := NewSignal(1000, 4)

	plotSingle(df, "fft_test_raw")
	plotSingle(df.FFTTopFreqs(), "fft_test_freqs")
}

func TestFFTWithNoise(t *testing.T) {
	df := NewSignal(1000, 4).Add(Noise(1000))

	plotSingle(df, "fft_test_noise_raw")
	plotSingle(df.FFTTopFreqs(), "fft_test_noise_freqs")
}

func TestNewSignal(t *testing.T) {
	df := NewSignal(1000, 3)

	plotSingle(df, "new_signal")

}
