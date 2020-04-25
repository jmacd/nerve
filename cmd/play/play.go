package main

import (
	"context"
	"log"
	"math"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/fogleman/gg"
	"github.com/hsluv/hsluv-go"
	"github.com/jkl1337/go-chromath"
	"github.com/jmacd/launchmidi/launchctl/xl"
	"github.com/jmacd/nerve/artnet"
	"github.com/jmacd/nerve/program"

	"github.com/lucasb-eyer/go-colorful"
)

const (
	ipAddr = "192.168.0.26"

	width  = 20
	height = 15
	pixels = width * height

	epsilon = 0.00001
)

type (
	Color = colorful.Color

	Buffer [pixels]colorful.Color
)

func main() {
	sender := artnet.NewSender(ipAddr)

	l, err := xl.Open()
	if err != nil {
		log.Fatalf("error while openning connection to launchctl: %v", err)
	}
	defer l.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go l.Run(ctx)

	tilesnake(sender, l)
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

func tilesnake(sender *artnet.Sender, lc *xl.LaunchControl) {
	prog := program.NewProgram("tilesnake", lc)
	prog.AddControl(xl.ControlKnobSendA[0], xl.ColorBrightGreen)
	prog.AddControl(xl.ControlKnobSendA[1], xl.ColorBrightGreen)
	prog.AddControl(xl.ControlKnobSendA[2], xl.ColorBrightOrange)
	prog.AddControl(xl.ControlKnobSendA[3], xl.ColorBrightOrange)

	prog.AddControl(xl.ControlSlider[0], 0)
	prog.AddControl(xl.ControlSlider[1], 0)

	lc.SwapBuffers(0)

	buffer := Buffer{}
	wf := factors(width)
	hf := factors(height)

	// tw := wf[len(wf)-1]
	// th := hf[len(hf)-1]
	tw := wf[0]
	th := hf[0]

	patW := pixels / height / tw
	patH := pixels / width / th

	cnt := pixels / tw / th

	last := time.Now()
	elapsed := 0.0

	rgb2xyz := chromath.NewRGBTransformer(
		&chromath.SpaceSRGB,
		&chromath.AdaptationBradford,
		&chromath.IlluminantRefD65,
		nil,
		1.0,
		chromath.SRGBCompander.Init(&chromath.SpaceSRGB))

	D65 := colorful.D65
	setX, setY, _ := colorful.XyzToXyy(D65[0], D65[1], D65[2])

	for {
		now := time.Now()
		delta := now.Sub(last)
		last = now

		elapsed += 50 * lc.Get(xl.ControlKnobSendA[0]) * float64(delta) / float64(time.Second)

		wX, wY, wZ := colorful.XyyToXyz(setX+(lc.Get(xl.ControlKnobSendA[2])-0.5)/10, setY+(lc.Get(xl.ControlKnobSendA[3])-0.5)/10, 1)

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

		for i := 0; i < patW; i++ {
			for j := 0; j < patH; j++ {
				var cidx int

				if j%2 == 0 {
					cidx = j*patW + i
				} else {
					cidx = (j+1)*patW - 1 - i
				}

				cangle := ((float64(cidx) + elapsed) / float64(cnt))
				cangle -= float64(int64(cangle))

				r, g, b := hsluv.HsluvToRGB(360*cangle, 100*lc.Get(xl.ControlSlider[0]), 100*lc.Get(xl.ControlSlider[1]))
				c0 := Color{R: r, G: g, B: b}

				cxyz := rgb2xyz.Convert(chromath.RGB{c0.R, c0.G, c0.B})

				crgb := xyz2rgb.Invert(cxyz)

				c1 := Color{R: crgb[0], G: crgb[1], B: crgb[2]}

				for x := 0; x < tw; x++ {
					for y := 0; y < th; y++ {
						idx := (j*th+y)*width + (i*tw + x)
						buffer[idx] = c1
					}
				}
			}
		}

		// Haha
		// rand.Shuffle(pixels, func(i, j int) {
		// 	sender.Buffer[i], sender.Buffer[j] = sender.Buffer[j], sender.Buffer[i]
		// })

		sender.Send(buffer[:])
		time.Sleep(time.Duration(float64(5*time.Millisecond) * lc.Get(xl.ControlKnobSendA[1])))
	}
}

type scrollFrag struct {
	pixWidth  float64
	pixOffset float64
	chars     string
}

func prepareString(dc *gg.Context, orig string) (render string, frags []scrollFrag) {
	var sb strings.Builder

	for len(orig) != 0 {
		r, size := utf8.DecodeRuneInString(orig)
		if unicode.IsSpace(r) {
			sb.WriteRune(' ')
		} else {
			sb.WriteRune(r)
		}
		orig = orig[size:]
	}

	orig = sb.String()
	render = orig
	offset := 0.0

	for len(orig) > 0 {
		prefixSize := 0
		leadingWidth := 0.0

		for {
			_, size := utf8.DecodeRuneInString(orig)
			prefixSize += size

			allWidth, _ := dc.MeasureString(orig)
			leadingWidth, _ = dc.MeasureString(orig[0:prefixSize])
			trailingWidth, _ := dc.MeasureString(orig[prefixSize:])

			if math.Abs(leadingWidth+trailingWidth-allWidth) >= epsilon {
				continue
			}

			break
		}

		frags = append(frags, scrollFrag{
			pixOffset: offset,
			pixWidth:  leadingWidth,
			chars:     orig[0:prefixSize],
		})

		offset += leadingWidth

		orig = orig[prefixSize:]
	}

	return
}
