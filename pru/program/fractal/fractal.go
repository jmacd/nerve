// fractal copied from https://github.com/joweich/fractal
package fractal

import (
	"image"
	"image/color"
	"math"
	"time"

	//"maze.io/x/math32/cmplx32"
	colorful "github.com/lucasb-eyer/go-colorful"
)

const (
	imgWidth  = 128
	imgHeight = 128
	maxIter   = 5000
)

func Fractal(img *image.RGBA, loc Location, a, b, c float64) {
	time.Sleep(2 * time.Second)
	var (
		iters   [128][128]float64
		num     [maxIter + 2]uint16
		escaped int
	)

	for y := 0; y < imgHeight; y++ {
		for x := 0; x < imgWidth; x++ {
			nx := (3/loc.Zoom/b)*((float64(x)+0.5)/float64(imgWidth)-0.5) + loc.XCenter
			ny := (3/loc.Zoom/b)*((float64(y)+0.5)/float64(imgHeight)+0.5) - loc.YCenter

			// magnitude, iter := mandelbrot(nx, ny)
			// var smooth float64
			// if iter != maxIter {
			// 	smooth = normalizeIterations(magnitude, iter)
			// 	num[int(smooth)]++
			// 	escaped++
			// } else {
			// 	smooth = -1
			// }

			smooth := mandelbrotSmooth(nx, ny)
			if smooth != maxIter {
				num[int(smooth)]++
				escaped++
			} else {
				smooth = -1
			}

			iters[y][x] = float64(smooth)
		}
	}

	var hues [maxIter + 2]float64
	hue := 0.0
	for i, cnt := range num {
		hue += float64(cnt) / float64(escaped)
		hues[i] = hue
	}

	for y := 0; y < imgHeight; y++ {
		for x := 0; x < imgWidth; x++ {
			smooth := iters[y][x]
			if smooth == -1 {
				img.SetRGBA(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 255})
				continue
			}
			hue := 360 * linear(
				hues[int(smooth)],
				hues[int(smooth)+1],
				float64(smooth)-float64(int(smooth)),
			)
			col := colorful.HSLuv(hue+360*a, 1, c)
			r, g, b := col.RGB255()
			img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
}

func linear(low, high, frac float64) float64 {
	return low + (high-low)*frac
}

// func normalizeIterations(magnitude float64, iter int) float64 {
// 	// from http://linas.org/art-gallery/escape/escape.html
// 	// and/or https://en.wikipedia.org/wiki/Plotting_algorithms_for_the_Mandelbrot_set
// 	nu := math.Log(math.Log(magnitude)) * math.Log2E
// 	return float64(iter) + 1 - nu
// }

// func mandelbrot(px, py float64) (float64, int) {
// 	var current complex128
// 	pxpy := complex(px, py)

// 	for i := 0; i < maxIter; i++ {
// 		magnitude := cmplx.Abs(current)
// 		if magnitude > 2 {
// 			return magnitude, i
// 		}
// 		current = current*current + pxpy
// 	}

// 	magnitude := cmplx.Abs(current)
// 	return magnitude, maxIter
// }

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
