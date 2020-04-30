package colors

import (
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
	colors.AddParameter(xl.ControlSlider[0], xl.ColorBrightGreen)
	colors.AddParameter(xl.ControlSlider[1], xl.ColorBrightYellow)
	colors.AddParameter(xl.ControlSlider[2], xl.ColorBrightOrange)
	colors.AddParameter(xl.ControlSlider[6], xl.ColorBrightRed)
	colors.AddParameter(xl.ControlSlider[7], xl.ColorBrightOrange)
	return colors
}

func (s *Colors) Draw(player program.Player) {
	switch s.Feature {
	case 0:
		s.luv(player)
	case 5:
		s.hclStripe(player)
	case 2, 6:
	case 3, 7:
	}
}

func (s *Colors) luv(player program.Player) {
	lc := player.Controller()
	level1 := lc.Get(xl.ControlSlider[7])

	wX, wY, wZ := colorful.XyyToXyz(
		lc.Get(xl.ControlSlider[0]),
		lc.Get(xl.ControlSlider[1]),
		lc.Get(xl.ControlSlider[2]),
	)
	wref := [3]float64{wX, wY, wZ}

	for y := 0; y < s.Height; y++ {
		for x := 0; x < s.Width; x++ {
			var c program.Color

			c = colorful.LuvWhiteRef(
				level1,
				(float64(y))/float64(s.Height-1),
				(float64(x))/float64(s.Width-1),
				wref,
			)

			s.Buffer.Pixels[y*s.Width+x] = c.Clamped()
		}
	}
}

func (s *Colors) hclStripe(player program.Player) {
	lc := player.Controller()

	wX, wY, wZ := colorful.XyyToXyz(
		lc.Get(xl.ControlSlider[0]),
		lc.Get(xl.ControlSlider[1]),
		lc.Get(xl.ControlSlider[2]),
	)
	wref := [3]float64{wX, wY, wZ}

	for y := 0; y < s.Height; y++ {
		for x := 0; x < s.Width; x++ {
			var c program.Color

			c = colorful.HclWhiteRef(
				360*(float64(x))/float64(s.Width-1),
				lc.Get(xl.ControlSlider[6]),
				lc.Get(xl.ControlSlider[7]),
				wref,
			)

			s.Buffer.Pixels[y*s.Width+x] = c.Clamped()

		}
	}
}

// func lab(sender *Sender, lc *xl.LaunchControl) {
// 	for {
// 		level := lc.Get(xl.ControlKnobSendA[0])

// 		wX, wY, wZ := colorful.XyyToXyz(0.5*lc.Get(xl.ControlKnobSendA[1]), 0.5*lc.Get(xl.ControlKnobSendA[2]), 1)
// 		wref := [3]float64{wX, wY, wZ}

// 		for y := 0; y < height; y++ {
// 			for x := 0; x < width; x++ {
// 				var c Color

// 				c = colorful.LabWhiteRef(
// 					level,
// 					(float64(y))/(height-1),
// 					(float64(x))/(width-1),
// 					wref,
// 				)

// 				sender.Buffer[y*width+x] = c.Clamped()

// 			}
// 		}

// 		sender.send()
// 		time.Sleep(time.Millisecond * 10)
// 	}
// }

// func hsluvPalette(sender *Sender, lc *xl.LaunchControl) {
// 	for frame := 0; ; frame++ {
// 		wX, wY, wZ := colorful.XyyToXyz(0.5*lc.Get(xl.ControlKnobSendA[0]), 0.5*lc.Get(xl.ControlKnobSendA[1]), 1)
// 		wref := [3]float64{wX, wY, wZ}

// 		r, g, b := hsluv.HsluvToRGB(360*lc.Get(xl.ControlSlider[0]), 100*lc.Get(xl.ControlSlider[1]), 100*lc.Get(xl.ControlSlider[2]))
// 		c := Color{R: r, G: g, B: b}
// 		//fmt.Println("COLOR IN", c, wref)
// 		x1, y1, z1 := c.Xyz()
// 		//fmt.Println("XYZ", x1, y1, z1)
// 		x2, y2, Y2 := colorful.XyzToXyyWhiteRef(x1, y1, z1, wref)
// 		//fmt.Println("XYY", x2, y2, Y2)
// 		c = colorful.Xyy(x2, y2, Y2)
// 		//fmt.Println("RGB", c.R, c.G, c.B)

// 		for x := 0; x < width; x++ {
// 			for y := 0; y < height; y++ {
// 				sender.Buffer[y*width+x] = c
// 			}
// 		}

// 		sender.send()
// 		time.Sleep(10 * time.Millisecond)
// 	}
// }
