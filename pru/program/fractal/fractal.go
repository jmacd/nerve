// fractal copied from https://github.com/joweich/fractal
package fractal

import (
	"image"
	"image/color"
	"math"

	//"maze.io/x/math32/cmplx32"
	colorful "github.com/lucasb-eyer/go-colorful"
)

const (
	imgWidth  = 128
	imgHeight = 128
	maxIter   = 5000
)

type Fractal struct {
	iters [128][128]float64
	hues  [maxIter + 2]float64
}

func New(loc Location) *Fractal {
	var num [maxIter + 2]uint16
	var escaped int
	f := &Fractal{}
	for y := 0; y < imgHeight; y++ {
		for x := 0; x < imgWidth; x++ {
			nx := (3/loc.Zoom)*((float64(x)+0.5)/float64(imgWidth)-0.5) + loc.XCenter
			ny := (3/loc.Zoom)*((float64(y)+0.5)/float64(imgHeight)-0.5) - loc.YCenter

			smooth := mandelbrotSmooth(nx, ny)
			if smooth != maxIter {
				num[int(smooth)]++
				escaped++
			} else {
				smooth = -1
			}

			f.iters[y][x] = float64(smooth)
		}
	}

	// var hues [maxIter + 2]float64
	hue := 0.0
	for i, cnt := range num {
		hue += float64(cnt) / float64(escaped)
		f.hues[i] = hue
	}
	return f
}

func (f *Fractal) Draw(img *image.RGBA, valA, valB, valC float64) {
	for y := 0; y < imgHeight; y++ {
		for x := 0; x < imgWidth; x++ {
			smooth := f.iters[y][x]
			if smooth == -1 {
				img.SetRGBA(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 255})
				continue
			}
			hue := linear(
				f.hues[int(smooth)],
				f.hues[int(smooth)+1],
				float64(smooth)-float64(int(smooth)),
			)

			col := colorful.HSLuv((hue+valA)*360, valB, valC)
			r, g, b := col.RGB255()
			img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
}

func linear(low, high, frac float64) float64 {
	return low + (high-low)*frac
}

func mandelbrotSmooth(x0, y0 float64) float64 {
	x := 0.0
	y := 0.0
	iteration := 0

	for x*x+y*y <= (1<<16) && iteration < maxIter {
		xtemp := x*x - y*y + x0
		y = 2*x*y + y0
		x = xtemp
		iteration++
	}

	if iteration == maxIter {
		return float64(iteration)
	}
	logZn := math.Log(x*x+y*y) / 2
	nu := math.Log(logZn*math.Log2E) * math.Log2E
	return float64(iteration) + 1 - nu
}
