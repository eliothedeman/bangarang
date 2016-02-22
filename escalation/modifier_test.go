package escalation

import (
	"testing"

	"github.com/eliothedeman/smoothie"
)

func BenchmarkDerivative100(b *testing.B) {
	df := smoothie.NewSignal(100, 4)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		derivative(df)
	}
}

func BenchmarkDerivative1000(b *testing.B) {
	df := smoothie.NewSignal(1000, 4)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		derivative(df)
	}
}

func BenchmarkNonNegativeDerivative100(b *testing.B) {
	df := smoothie.NewSignal(100, 4)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		nonNegativeDerivative(df)
	}
}

func BenchmarkNonNegativeDerivative1000(b *testing.B) {
	df := smoothie.NewSignal(1000, 4)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		nonNegativeDerivative(df)
	}
}
func BenchmarkMovingAverage100(b *testing.B) {
	df := smoothie.NewSignal(100, 4)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		movingAverage(df)
	}
}

func BenchmarkMovingAverage1000(b *testing.B) {
	df := smoothie.NewSignal(1000, 4)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		movingAverage(df)
	}
}
func BenchmarkSingleExponential100(b *testing.B) {
	df := smoothie.NewSignal(100, 4)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		singleExponentialSmooting(df)
	}
}
func BenchmarkSingleExponential1000(b *testing.B) {
	df := smoothie.NewSignal(1000, 4)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		singleExponentialSmooting(df)
	}
}
func BenchmarkHoltWinters100(b *testing.B) {
	df := smoothie.NewSignal(100, 4)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		holtWinters(df)
	}
}

func BenchmarkHoltWinters1000(b *testing.B) {
	df := smoothie.NewSignal(1000, 4)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		holtWinters(df)
	}
}
