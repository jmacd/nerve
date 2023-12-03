// fractal copied from https://github.com/joweich/fractal
package fractal

import (
	"image"
	"image/color"
	"math"

	"github.com/jmacd/launchmidi/midi/controller"
	"github.com/jmacd/nerve/pru/program/data"
	colorful "github.com/lucasb-eyer/go-colorful"
)

const (
	imgWidth  = 128
	imgHeight = 128
	maxIter   = 5000
)

type locSet struct {
	num int
	za  controller.Value
	zb  controller.Value
	zc  controller.Value
	zd  controller.Value
}

type Fractal struct {
	loc   locSet
	iters [128][128]float64
	hues  [maxIter + 2]float64
}

func New() *Fractal {
	return &Fractal{
		loc: locSet{
			num: -1,
		},
	}
}

func (f *Fractal) Draw(data *data.Data, img *image.RGBA) {
	loc := locSet{
		num: int(data.KnobsRow1[0]),
		za:  data.KnobsRow2[0],
		zb:  data.KnobsRow2[1],
		zc:  data.KnobsRow2[2],
		zd:  data.KnobsRow2[3],
	}
	if f.loc != loc {
		f.computeHues(loc)
	}
	f.render(data, img)
}

func vary(knob controller.Value) float64 {
	return 1 // knob.Float() + 0.5
}

func (f *Fractal) computeHues(loc locSet) {
	f.loc = loc
	seed := Seeds[loc.num%len(Seeds)]
	var num [maxIter + 2]uint16
	var escaped int
	for y := 0; y < imgHeight; y++ {
		for x := 0; x < imgWidth; x++ {
			nx := (3/(seed.Zoom*vary(loc.za)))*((float64(x)+0.5)/float64(imgWidth)-0.5) + (seed.XCenter * vary(loc.zc))
			ny := (3/(seed.Zoom*vary(loc.zb)))*((float64(y)+0.5)/float64(imgHeight)-0.5) - (seed.YCenter * vary(loc.zd))

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

	hue := 0.0
	for i, cnt := range num {
		hue += float64(cnt) / float64(escaped)
		f.hues[i] = hue
	}
}

func (f *Fractal) render(data *data.Data, img *image.RGBA) {
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

			col := colorful.HSLuv(
				(hue+data.Sliders[0].Float())*360,
				data.Sliders[1].Float(),
				data.Sliders[2].Float(),
			)
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
