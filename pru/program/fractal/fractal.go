// fractal copied from https://github.com/joweich/fractal
package fractal

import (
	"image"
	"image/color"
	"math"
	"math/cmplx"
	"math/rand"
	//"maze.io/x/math32/cmplx32"
)

const (
	imgWidth  = 128
	imgHeight = 128
	maxIter   = 500
	samples   = 1
	hueOffset = 0.5 // hsl color model; float in range [0,1)
)

func Fractal(img *image.RGBA, loc Location) {
	rnd := rand.New(rand.NewSource(123))
	_ = rnd
	for y := 0; y < imgHeight; y++ {
		for x := 0; x < imgWidth; x++ {
			var r, g, b int
			for i := 0; i < samples; i++ {
				///nx := (3/loc.Zoom)*((float64(x)+rnd.Float64())/float64(imgWidth)-0.5) + loc.XCenter
				///ny := (3/loc.Zoom)*((float64(y)+rnd.Float64())/float64(imgHeight)-0.5) - loc.YCenter
				nx := (3/loc.Zoom)*((float64(x)+0.5)/float64(imgWidth)-0.5) + loc.XCenter
				ny := (3/loc.Zoom)*((float64(y)+0.5)/float64(imgHeight)+0.5) - loc.YCenter

				c := paint(mandelbrotIterComplex(nx, ny, maxIter))
				r += int(c.R)
				g += int(c.G)
				b += int(c.B)
			}
			var cr, cg, cb uint8
			cr = uint8(float64(r) / float64(samples))
			cg = uint8(float64(g) / float64(samples))
			cb = uint8(float64(b) / float64(samples))
			img.SetRGBA(x, y, color.RGBA{R: cr, G: cg, B: cb, A: 255})
		}
	}
}

func paint(magnitude float64, n int) color.RGBA {
	if magnitude > 2 {
		// adapted http://linas.org/art-gallery/escape/escape.html
		nu := math.Log(math.Log(float64(magnitude))) / math.Log(2)
		hue := (float64(n)+1-float64(nu))/float64(maxIter) + hueOffset
		return hslToRGB(float64(hue), 1, 0.5)
	}

	return color.RGBA{R: 0, G: 0, B: 0, A: 255}
}

func mandelbrotIterComplex(px, py float64, maxIter int) (float64, int) {
	var current complex128
	pxpy := complex(px, py)

	for i := 0; i < maxIter; i++ {
		magnitude := cmplx.Abs(current)
		if magnitude > 2 {
			return magnitude, i
		}
		current = current*current + pxpy
	}

	magnitude := cmplx.Abs(current)
	return magnitude, maxIter
}
