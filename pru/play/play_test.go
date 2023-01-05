package play

import "testing"

func BenchmarkPlay(b *testing.B) {
	buf := NewBuffer()
	dc := buf.NewContext()

	dc.SetRGB(0, 0, 1)
	dc.DrawCircle(64, 64, 50)
	dc.SetRGB(1, 1, 0)
	dc.Fill()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Copy()
	}
}
