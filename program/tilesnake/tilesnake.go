package tilesnake

import (
	"math"
	"time"

	"github.com/hsluv/hsluv-go"
	"github.com/jkl1337/go-chromath"
	"github.com/jmacd/launchmidi/launchctl/xl"
	"github.com/jmacd/nerve/program"
	"github.com/lucasb-eyer/go-colorful"
)

type Tilesnake struct {
	program.Pattern

	patternW int
	patternH int
	sections int
	rgb2xyz  *chromath.RGBTransformer
	setX     float64
	setY     float64
	tw       int
	th       int
	elapsed  time.Duration
	last     time.Time
}

func New(width, height int) *Tilesnake {
	snake := &Tilesnake{
		Pattern: program.New(width, height),
	}

	snake.AddParameter(xl.ControlKnobSendA[0], xl.ColorBrightGreen)
	snake.AddParameter(xl.ControlKnobSendA[1], xl.ColorBrightGreen)
	snake.AddParameter(xl.ControlKnobSendA[2], xl.ColorBrightOrange)
	snake.AddParameter(xl.ControlKnobSendA[3], xl.ColorBrightOrange)

	snake.AddParameter(xl.ControlSlider[0], 0)
	snake.AddParameter(xl.ControlSlider[1], 0)

	wf := factors(width)
	hf := factors(height)

	// TODO
	// tw := wf[len(wf)-1]
	// th := hf[len(hf)-1]
	snake.tw = wf[0]
	snake.th = hf[0]

	snake.patternW = width / snake.tw
	snake.patternH = height / snake.th
	snake.sections = (width * height) / (snake.tw * snake.th)

	snake.rgb2xyz = chromath.NewRGBTransformer(
		&chromath.SpaceSRGB,
		&chromath.AdaptationBradford,
		&chromath.IlluminantRefD65,
		nil,
		1.0,
		chromath.SRGBCompander.Init(&chromath.SpaceSRGB))

	D65 := colorful.D65
	snake.setX, snake.setY, _ = colorful.XyzToXyy(D65[0], D65[1], D65[2])

	snake.last = time.Now()

	return snake
}

func factors(n int) []int {
	sq := math.Sqrt(float64(n))
	var fs []int

	for i := 2; i <= int(sq); i++ {
		if n%i != 0 {
			continue
		}
		fs = append(fs, i)
	}
	return fs
}

func (snake *Tilesnake) Draw(player program.Player) {
	lc := player.Controller()

	now := time.Now()
	delta := now.Sub(snake.last)
	snake.last = now
	snake.elapsed += time.Duration(50 * lc.Get(xl.ControlKnobSendA[0]) * float64(delta))

	wX, wY, wZ := colorful.XyyToXyz(snake.setX+(lc.Get(xl.ControlKnobSendA[2])-0.5)/10, snake.setY+(lc.Get(xl.ControlKnobSendA[3])-0.5)/10, 1)

	targetIlluminant := &chromath.IlluminantRef{
		XYZ:      chromath.XYZ{wX, wY, wZ},
		Observer: chromath.CIE2,
		Standard: nil,
	}

	xyz2rgb := chromath.NewRGBTransformer(
		&chromath.SpaceSRGB,
		&chromath.AdaptationBradford,
		targetIlluminant,
		nil,
		1.0,
		chromath.SRGBCompander.Init(&chromath.SpaceSRGB))

	for i := 0; i < snake.patternW; i++ {
		for j := 0; j < snake.patternH; j++ {
			var cidx int

			if j%2 == 0 {
				cidx = j*snake.patternW + i
			} else {
				cidx = (j+1)*snake.patternW - 1 - i
			}

			cangle := ((float64(cidx) + snake.elapsed.Seconds()) / float64(snake.sections))
			cangle -= float64(int64(cangle))

			r, g, b := hsluv.HsluvToRGB(360*cangle, 100*lc.Get(xl.ControlSlider[0]), 100*lc.Get(xl.ControlSlider[1]))
			c0 := program.Color{R: r, G: g, B: b}

			cxyz := snake.rgb2xyz.Convert(chromath.RGB{c0.R, c0.G, c0.B})

			crgb := xyz2rgb.Invert(cxyz)

			c1 := program.Color{R: crgb[0], G: crgb[1], B: crgb[2]}

			for x := 0; x < snake.tw; x++ {
				for y := 0; y < snake.th; y++ {
					idx := (j*snake.th+y)*snake.Buffer.Width + (i*snake.tw + x)
					snake.Buffer.Pixels[idx] = c1
				}
			}
		}
	}
}
