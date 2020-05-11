package colors

import (
	"math"

	"github.com/hsluv/hsluv-go"
	"github.com/jmacd/launchmidi/launchctl/xl"
	"github.com/jmacd/nerve/program"
	"github.com/lucasb-eyer/go-colorful"
)

type (
	Colors struct {
		program.Pattern
	}
)

func New(width, height int) *Colors {
	colors := &Colors{
		Pattern: program.New(width, height),
	}
	for i := 0; i < 8; i++ {
		colors.AddParameter(xl.ControlSlider[i], xl.ColorBrightGreen)
	}
	return colors
}

func (s *Colors) Draw(player program.Player) {
	lc := player.Controller()

	// TODO consolidate with the wref logic of tilesnake.  Factor
	// chromath calls into a new class and use in hsluv method(s)
	// below.
	wX, wY, wZ := colorful.XyyToXyz(
		lc.Get(xl.ControlSlider[0]),
		lc.Get(xl.ControlSlider[1]),
		lc.Get(xl.ControlSlider[2]),
	)
	wref := [3]float64{wX, wY, wZ}

	switch s.Feature {
	case 0:
		s.hsvStripe(lc, wref)
	case 1:
		s.hsvRadial(lc, wref)
	case 2:
		s.luv(lc, wref)
	case 3:
		s.lab(lc, wref)
	case 4:
		s.hclStripe(lc, wref)
	case 5:
		s.hclRadials(lc, wref)
	case 6:
		s.hsluvStripe(lc, wref)
	case 7:
		s.hsluvRadials(lc, wref)
	}
}

func (s *Colors) hsvStripe(lc *xl.LaunchControl, wref [3]float64) {
	tplus := 360 * lc.Get(xl.ControlSlider[5])
	for x := 0; x < s.Width; x++ {
		for y := 0; y < s.Height; y++ {

			deg := 360*float64(x)/float64(s.Width-1) + tplus
			deg -= float64(int(deg/360)) * 360

			c := colorful.Hsv(
				deg,
				100*lc.Get(xl.ControlSlider[6]),
				100*lc.Get(xl.ControlSlider[7]),
			)

			x1, y1, z1 := c.Xyz()
			x2, y2, Y2 := colorful.XyzToXyyWhiteRef(x1, y1, z1, wref)
			c = colorful.Xyy(x2, y2, Y2)

			s.Buffer.Pixels[y*s.Width+x] = c.Clamped()
		}
	}
}

func (s *Colors) hsvRadial(lc *xl.LaunchControl, wref [3]float64) {
	tplus := 2 * math.Pi * (1 + lc.Get(xl.ControlSlider[5]))
	for y := 0; y < s.Height; y++ {
		for x := 0; x < s.Width; x++ {
			yf := float64(y)/float64(s.Height-1) - 0.5
			xf := float64(x)/float64(s.Width-1) - 0.5

			theta := math.Atan2(float64(yf), float64(xf)) + tplus
			theta = theta * 180 / math.Pi
			theta -= float64(int(theta/360)) * 360

			c := colorful.Hsv(
				theta,
				100*lc.Get(xl.ControlSlider[6]),
				100*lc.Get(xl.ControlSlider[7]),
			)

			x1, y1, z1 := c.Xyz()
			x2, y2, Y2 := colorful.XyzToXyyWhiteRef(x1, y1, z1, wref)
			c = colorful.Xyy(x2, y2, Y2)

			s.Buffer.Pixels[y*s.Width+x] = c.Clamped()
		}
	}
}

func (s *Colors) lab(lc *xl.LaunchControl, wref [3]float64) {
	for y := 0; y < s.Height; y++ {
		for x := 0; x < s.Width; x++ {
			var c program.Color

			c = colorful.LabWhiteRef(
				lc.Get(xl.ControlSlider[7]),
				(float64(y))/float64(s.Height-1),
				(float64(x))/float64(s.Width-1),
				wref,
			)

			s.Buffer.Pixels[y*s.Width+x] = c.Clamped()
		}
	}
}

func (s *Colors) luv(lc *xl.LaunchControl, wref [3]float64) {
	for y := 0; y < s.Height; y++ {
		for x := 0; x < s.Width; x++ {
			var c program.Color

			c = colorful.LuvWhiteRef(
				lc.Get(xl.ControlSlider[7]),
				(float64(y))/float64(s.Height-1),
				(float64(x))/float64(s.Width-1),
				wref,
			)

			s.Buffer.Pixels[y*s.Width+x] = c.Clamped()
		}
	}
}

func (s *Colors) hclStripe(lc *xl.LaunchControl, wref [3]float64) {
	tplus := 360 * lc.Get(xl.ControlSlider[5])
	for y := 0; y < s.Height; y++ {
		for x := 0; x < s.Width; x++ {
			var c program.Color

			c = colorful.HclWhiteRef(
				360*(float64(x))/float64(s.Width-1)+tplus,
				lc.Get(xl.ControlSlider[6]),
				lc.Get(xl.ControlSlider[7]),
				wref,
			)

			s.Buffer.Pixels[y*s.Width+x] = c.Clamped()

		}
	}
}

func (s *Colors) hclRadials(lc *xl.LaunchControl, wref [3]float64) {
	tplus := 2 * math.Pi * lc.Get(xl.ControlSlider[5])
	for y := 0; y < s.Height; y++ {
		for x := 0; x < s.Width; x++ {
			yf := float64(y)/float64(s.Height-1) - 0.5
			xf := float64(x)/float64(s.Width-1) - 0.5

			theta := math.Atan2(float64(yf), float64(xf)) + tplus

			var c program.Color

			c = colorful.HclWhiteRef(
				360*(theta+math.Pi)/(2*math.Pi),
				lc.Get(xl.ControlSlider[6]),
				lc.Get(xl.ControlSlider[7]),
				wref,
			)

			s.Buffer.Pixels[y*s.Width+x] = c.Clamped()
		}
	}
}

func (s *Colors) hsluvStripe(lc *xl.LaunchControl, wref [3]float64) {
	tplus := 360 * lc.Get(xl.ControlSlider[5])
	for x := 0; x < s.Width; x++ {
		for y := 0; y < s.Height; y++ {

			deg := 360*float64(x)/float64(s.Width-1) + tplus

			r, g, b := hsluv.HsluvToRGB(
				deg,
				100*lc.Get(xl.ControlSlider[6]),
				100*lc.Get(xl.ControlSlider[7]),
			)

			c := program.Color{R: r, G: g, B: b}

			x1, y1, z1 := c.Xyz()
			x2, y2, Y2 := colorful.XyzToXyyWhiteRef(x1, y1, z1, wref)
			c = colorful.Xyy(x2, y2, Y2)

			s.Buffer.Pixels[y*s.Width+x] = c.Clamped()
		}
	}
}

func (s *Colors) hsluvRadials(lc *xl.LaunchControl, wref [3]float64) {
	tplus := 2 * math.Pi * lc.Get(xl.ControlSlider[5])
	for y := 0; y < s.Height; y++ {
		for x := 0; x < s.Width; x++ {
			yf := float64(y)/float64(s.Height-1) - 0.5
			xf := float64(x)/float64(s.Width-1) - 0.5

			theta := math.Atan2(float64(yf), float64(xf)) + tplus

			r, g, b := hsluv.HsluvToRGB(
				theta*180/math.Pi,
				100*lc.Get(xl.ControlSlider[6]),
				100*lc.Get(xl.ControlSlider[7]),
			)

			c := program.Color{R: r, G: g, B: b}

			x1, y1, z1 := c.Xyz()
			x2, y2, Y2 := colorful.XyzToXyyWhiteRef(x1, y1, z1, wref)
			c = colorful.Xyy(x2, y2, Y2)

			s.Buffer.Pixels[y*s.Width+x] = c.Clamped()
		}
	}
}
