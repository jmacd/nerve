package gpixio

import (
	"testing"

	"github.com/fogleman/gg"
)

func BenchmarkPlay(b *testing.B) {
	buf := NewBuffer()
	dc := gg.NewContextForRGBA(buf.RGBA)

	dc.SetRGB(0, 0, 1)
	dc.DrawCircle(64, 64, 50)
	dc.SetRGB(1, 1, 0)
	dc.Fill()

	var fb FrameBank

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Copy(&fb)
	}
}
