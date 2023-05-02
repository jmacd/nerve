// fractal copied from https://github.com/joweich/fractal
package fractal

import (
	"image"
	"image/color"
	"math"
	"math/cmplx"
	"runtime"
)

// Configuration
const (
	// Quality
	imgWidth     = 128
	imgHeight    = 128
	maxIter      = 200
	samples      = 50
	hueOffset    = 0.0 // hsl color model; float in range [0,1)
	linearMixing = true
)

const (
	ratio = float64(imgWidth) / float64(imgHeight)
)

func Fractal(img *image.RGBA, loc Location) {
	jobs := make(chan int)

	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			for y := range jobs {
				for x := 0; x < imgWidth; x++ {
					var r, g, b int
					for i := 0; i < samples; i++ {
						nx := 3*(1/loc.Zoom)*ratio*((float64(x)+RandFloat64())/float64(imgWidth)-0.5) + loc.XCenter
						ny := 3*(1/loc.Zoom)*((float64(y)+RandFloat64())/float64(imgHeight)-0.5) - loc.YCenter
						c := paint(mandelbrotIterComplex(nx, ny, maxIter))
						if linearMixing {
							r += int(RGBToLinear(c.R))
							g += int(RGBToLinear(c.G))
							b += int(RGBToLinear(c.B))
						} else {
							r += int(c.R)
							g += int(c.G)
							b += int(c.B)
						}
					}
					var cr, cg, cb uint8
					if linearMixing {
						cr = LinearToRGB(uint16(float64(r) / float64(samples)))
						cg = LinearToRGB(uint16(float64(g) / float64(samples)))
						cb = LinearToRGB(uint16(float64(b) / float64(samples)))
					} else {
						cr = uint8(float64(r) / float64(samples))
						cg = uint8(float64(g) / float64(samples))
						cb = uint8(float64(b) / float64(samples))
					}
					img.SetRGBA(x, y, color.RGBA{R: cr, G: cg, B: cb, A: 255})
				}
			}
		}()
	}

	for y := 0; y < imgHeight; y++ {
		jobs <- y
	}
}

func paint(magnitude float64, n int) color.RGBA {
	if magnitude > 2 {
		// adapted http://linas.org/art-gallery/escape/escape.html
		nu := math.Log(math.Log(magnitude)) / math.Log(2)
		hue := (float64(n)+1-nu)/float64(maxIter) + hueOffset
		return hslToRGB(hue, 1, 0.5)
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
