package smoothie

import "testing"

func BenchmarkMovingAverage100(b *testing.B) {
	df := NewSignal(100, 4)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		df.MovingAverage(4)
	}
}

func BenchmarkMovingAverage1000(b *testing.B) {
	df := NewSignal(1000, 4)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		df.MovingAverage(4)
	}
}
func BenchmarkAvg100(b *testing.B) {
	df := NewSignal(100, 4)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		df.Avg()
	}
}

func BenchmarkAvg1000(b *testing.B) {
	df := NewSignal(1000, 4)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		df.Avg()
	}
}

func BenchmarkHoltWinters100(b *testing.B) {
	df := NewSignal(100, 4)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		df.HoltWinters(0.3, 0.2)
	}
}

func BenchmarkHoltWinters1000(b *testing.B) {
	df := NewSignal(1000, 4)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		df.HoltWinters(0.3, 0.2)
	}
}
