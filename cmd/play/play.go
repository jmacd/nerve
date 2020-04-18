package main

import (
	"context"
	"fmt"
	"image"
	"log"
	"math"
	"net"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/fogleman/gg"
	"github.com/hsluv/hsluv-go"
	"github.com/jkl1337/go-chromath"
	"github.com/jmacd/launchmidi/launchctl/xl"

	"github.com/jsimonetti/go-artnet/packet"
	"github.com/lucasb-eyer/go-colorful"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/metric"
	"go.opentelemetry.io/otel/exporters/metric/stdout"
)

const (
	//ipAddr = "192.168.1.167" // Bldg
	ipAddr = "192.168.0.23" // Home

	width  = 20
	height = 15
	pixels = width * height

	maxPerPacket = 170

	zeroOffset = 0
	epsilon    = 0.00001
)

var (
	meter  = global.Meter("main")
	frames = metric.Must(meter).NewInt64Counter("frames").Bind()
)

type (
	Color = colorful.Color

	Sender struct {
		dest *net.UDPAddr
		conn *net.UDPConn

		Buffer [pixels]colorful.Color
		packet.ArtDMXPacket
	}
)

func newSender() *Sender {
	dst := fmt.Sprintf("%s:%d", ipAddr, packet.ArtNetPort)
	node, _ := net.ResolveUDPAddr("udp", dst)

	src := fmt.Sprintf("%s:%d", "", packet.ArtNetPort)
	localAddr, _ := net.ResolveUDPAddr("udp", src)

	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		panic(fmt.Sprint("error opening udp: ", err))
	}

	return &Sender{
		dest: node,
		conn: conn,
	}
}

func telemetry() func() {
	controller, err := stdout.NewExportPipeline(stdout.Config{}, time.Second*1)
	if err != nil {
		log.Fatal(err)
	}
	global.SetMeterProvider(controller)

	if err != nil {
		log.Fatal(err)
	}
	return controller.Stop
}

func main() {
	var sender *Sender

	// stop := telemetry()
	// defer stop()

	sender = newSender()

	l, err := xl.Open()
	if err != nil {
		log.Fatalf("error while openning connection to launchctl: %v", err)
	}
	defer l.Close()

	go l.Run(context.Background())

	scroller(sender, l)
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

func tilesnake(sender *Sender, lc *xl.LaunchControl) {
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

				r, g, b := hsluv.HsluvToRGB(360*cangle, 100*lc.Get(xl.ControlSlider[1]), 100*lc.Get(xl.ControlSlider[2]))
				c0 := Color{R: r, G: g, B: b}

				cxyz := rgb2xyz.Convert(chromath.RGB{c0.R, c0.G, c0.B})

				crgb := xyz2rgb.Invert(cxyz)

				c1 := Color{R: crgb[0], G: crgb[1], B: crgb[2]}

				for x := 0; x < tw; x++ {
					for y := 0; y < th; y++ {
						idx := (j*th+y)*width + (i*tw + x)
						sender.Buffer[idx] = c1
					}
				}
			}
		}

		// Haha
		// rand.Shuffle(pixels, func(i, j int) {
		// 	sender.Buffer[i], sender.Buffer[j] = sender.Buffer[j], sender.Buffer[i]
		// })

		sender.send()
		time.Sleep(time.Duration(float64(5*time.Millisecond) * lc.Get(xl.ControlKnobSendA[1])))
	}
}

func luv(sender *Sender, lc *xl.LaunchControl) {
	for {
		level := lc.Get(xl.ControlKnobSendA[0])

		wX, wY, wZ := colorful.XyyToXyz(0.5*lc.Get(xl.ControlKnobSendA[1]), 0.5*lc.Get(xl.ControlKnobSendA[2]), 1)
		wref := [3]float64{wX, wY, wZ}

		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				var c Color

				c = colorful.LuvWhiteRef(
					level,
					(float64(y)+zeroOffset)/(height-1+zeroOffset),
					(float64(x)+zeroOffset)/(width-1+zeroOffset),
					wref,
				)

				sender.Buffer[y*width+x] = c.Clamped()

			}
		}

		sender.send()
		time.Sleep(time.Millisecond * 10)
	}
}

func hcl(sender *Sender, lc *xl.LaunchControl) {
	for {
		level := lc.Get(xl.ControlKnobSendA[0])

		wref := [3]float64{0.5 + lc.Get(xl.ControlKnobSendA[1]), 1.00000, 0.5 + lc.Get(xl.ControlKnobSendA[2])}

		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				var c Color

				c = colorful.HclWhiteRef(
					360*(float64(x)+zeroOffset)/(width-1+zeroOffset),
					(float64(y)+zeroOffset)/(height-1+zeroOffset),
					level,
					wref,
				)

				sender.Buffer[y*width+x] = c.Clamped()

			}
		}

		sender.send()
		time.Sleep(time.Millisecond * 10)
	}
}

func lab(sender *Sender, lc *xl.LaunchControl) {
	for {
		level := lc.Get(xl.ControlKnobSendA[0])

		wX, wY, wZ := colorful.XyyToXyz(0.5*lc.Get(xl.ControlKnobSendA[1]), 0.5*lc.Get(xl.ControlKnobSendA[2]), 1)
		wref := [3]float64{wX, wY, wZ}

		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				var c Color

				c = colorful.LabWhiteRef(
					level,
					(float64(y)+zeroOffset)/(height-1+zeroOffset),
					(float64(x)+zeroOffset)/(width-1+zeroOffset),
					wref,
				)

				sender.Buffer[y*width+x] = c.Clamped()

			}
		}

		sender.send()
		time.Sleep(time.Millisecond * 10)
	}
}

func showHsluv(sender *Sender, lc *xl.LaunchControl) {
	for frame := 0; ; frame++ {
		wX, wY, wZ := colorful.XyyToXyz(0.5*lc.Get(xl.ControlKnobSendA[0]), 0.5*lc.Get(xl.ControlKnobSendA[1]), 1)
		wref := [3]float64{wX, wY, wZ}

		r, g, b := hsluv.HsluvToRGB(360*lc.Get(xl.ControlSlider[0]), 100*lc.Get(xl.ControlSlider[1]), 100*lc.Get(xl.ControlSlider[2]))
		c := Color{R: r, G: g, B: b}
		//fmt.Println("COLOR IN", c, wref)
		x1, y1, z1 := c.Xyz()
		//fmt.Println("XYZ", x1, y1, z1)
		x2, y2, Y2 := colorful.XyzToXyyWhiteRef(x1, y1, z1, wref)
		//fmt.Println("XYY", x2, y2, Y2)
		c = colorful.Xyy(x2, y2, Y2)
		//fmt.Println("RGB", c.R, c.G, c.B)

		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				sender.Buffer[y*width+x] = c
			}
		}

		sender.send()
		time.Sleep(10 * time.Millisecond)
	}
}

func gamma(sender *Sender, lc *xl.LaunchControl) {
	for frame := 0; ; frame++ {
		gamma := math.Max(lc.Get(xl.ControlKnobSendA[0])*3, 0.001)
		//gamma = 1.0

		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				sender.Buffer[y*width+x] = Color{
					R: math.Pow(float64(x)/float64(width), gamma),
					G: math.Pow(float64(x)/float64(width), gamma),
					B: math.Pow(float64(x)/float64(width), gamma),
				}
			}
		}

		sender.send()
		time.Sleep(10 * time.Millisecond)
	}
}

func walk(sender *Sender, lc *xl.LaunchControl) {
	last := time.Now()
	elapsed := 0.0

	for frame := 0; ; frame++ {
		now := time.Now()
		delta := now.Sub(last)

		last = now

		rate := 100 * lc.Get(xl.ControlKnobSendA[0])

		elapsed += rate * float64(delta.Seconds())

		spot := int64(elapsed) % pixels
		sender.Buffer[spot] = Color{
			R: lc.Get(xl.ControlSlider[0]) * lc.Get(xl.ControlKnobSendA[1]),
			G: lc.Get(xl.ControlSlider[1]) * lc.Get(xl.ControlKnobSendA[1]),
			B: lc.Get(xl.ControlSlider[2]) * lc.Get(xl.ControlKnobSendA[1]),
		}

		sender.send()
		time.Sleep(time.Duration(250*lc.Get(xl.ControlKnobSendA[0])) * time.Millisecond)
	}
}

func strobe2(sender *Sender, lc *xl.LaunchControl) {
	for frame := 0; ; frame++ {

		if frame%2 == 0 {
			for i := 0; i < pixels; i++ {
				sender.Buffer[i] = Color{
					R: lc.Get(xl.ControlSlider[3]) * lc.Get(xl.ControlKnobSendA[1]),
					G: lc.Get(xl.ControlSlider[4]) * lc.Get(xl.ControlKnobSendA[1]),
					B: lc.Get(xl.ControlSlider[5]) * lc.Get(xl.ControlKnobSendA[1]),
				}
			}
		} else {
			for i := 0; i < pixels; i++ {
				sender.Buffer[i] = Color{
					R: lc.Get(xl.ControlSlider[0]) * lc.Get(xl.ControlKnobSendA[1]),
					G: lc.Get(xl.ControlSlider[1]) * lc.Get(xl.ControlKnobSendA[1]),
					B: lc.Get(xl.ControlSlider[2]) * lc.Get(xl.ControlKnobSendA[1]),
				}
			}
		}

		sender.send()
		time.Sleep(time.Duration(250*lc.Get(xl.ControlKnobSendA[0])) * time.Millisecond)
	}
}

func colors(sender *Sender, lc *xl.LaunchControl) {
	start := time.Now()

	for {
		seconds := time.Now().Sub(start) / time.Second

		pattern := (seconds / 15) % 4

		for twoCount := 0.0; twoCount < 2; twoCount += 0.005 {
			level := lc.Get(xl.ControlKnobSendA[0])

			for y := 0; y < height; y++ {
				for x := 0; x < width; x++ {
					var c Color

					switch pattern {
					case 0:
						c = colorful.Lab(
							level,
							2*(float64(x)/(width-1)-0.5),
							2*(float64(y)/(height-1)-0.5),
						)
					case 1:
						c = colorful.Xyy(
							(float64(y)+zeroOffset)/(height-1+zeroOffset),
							(float64(x)+zeroOffset)/(width-1+zeroOffset),
							level,
						)
					case 2:
						c = colorful.Xyz(
							float64(x)/(width-1),
							float64(y)/(height-1),
							level,
						)
					case 3:
						c = colorful.Luv(
							level,
							2*(float64(x)/(width-1)-0.5),
							2*(float64(y)/(height-1)-0.5),
						)
					}

					sender.Buffer[y*width+x] = c.Clamped()

				}
			}

			sender.send()
			time.Sleep(time.Millisecond * 10)
		}
	}
}

func (s *Sender) send() {
	data := s.ArtDMXPacket.Data[:]
	s.ArtDMXPacket.SubUni = 0

	frames.Add(context.Background(), 1)

	for p := 0; p < pixels; {

		num := maxPerPacket
		if pixels-p < num {
			num = pixels - p
		}

		for i := 0; i < num; i++ {
			c := s.Buffer[p+i]
			data[i*3+0] = byte(c.R * 255)
			data[i*3+1] = byte(c.G * 255)
			data[i*3+2] = byte(c.B * 255)
		}
		s.ArtDMXPacket.Length = uint16(num)

		b, _ := s.ArtDMXPacket.MarshalBinary()

		_, err := s.conn.WriteTo(b, s.dest)
		if err != nil {
			panic(fmt.Sprint("error writing packet: ", err))
		}

		s.ArtDMXPacket.SubUni++
		p += num
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

func scroller(sender *Sender, lc *xl.LaunchControl) {
	img := image.NewRGBA(image.Rectangle{
		Min: image.Point{0, 0},
		Max: image.Point{width, height},
	})

	const S = height
	const N = height * 2

	const T = `
Hello, hello, 
and welcome to my show.
Yes, I'm belting while not melting,
even though I'm made of snow.
Perhaps you might recall,
I once needed my own flurry.
But now you need not worry...

Who can handle this enormous candle?
Unmeltable me.
Who's not sweating this sweltering setting?
Unmeltable me.
Yes, Elsa's powers grew,
and it's a whole new situation.
Because of our relation,
I now have self-refrigeration.

Who is present, but not liquescent?
(That's right, I just learned to read, and I like the big words.)
Unmeltable me.
Who's super cuddly and not all puddly?
Unmeltable me.
Who's inexhaustible, indefrostable, humble and fun?
It's unmeltable --
Oh, thank goodness you're here. Grab a seat, just about to bring it on home --
Meeeee!
`

	dc := gg.NewContextForRGBA(img)

	render, _ := prepareString(dc, T)

	// Position of the start of the string
	offset := float64(width / 2)
	lastTime := time.Now()

	currentFontSize := 0.0

	for {
		fontSize := 6 + lc.Get(xl.ControlKnobSendA[6])*20

		if currentFontSize != fontSize {
			if err := dc.LoadFontFace("/System/Library/Fonts/Avenir.ttc", fontSize); err != nil {
				panic(err)
			}
			currentFontSize = fontSize
		}

		rate := (lc.Get(xl.ControlKnobSendA[7]) - 0.5) * 200
		// fmt.Println("RATE", rate, lc.Get(xl.ControlKnobSendA[7]), lc.Get(xl.ControlKnobSendA[7])-0.5)

		now := time.Now()
		delta := now.Sub(lastTime).Seconds()
		lastTime = now
		offset += rate * delta

		// fmt.Println("OFFSET", offset)

		dc.SetRGB(lc.Get(xl.ControlSlider[5]), lc.Get(xl.ControlSlider[6]), lc.Get(xl.ControlSlider[7]))
		dc.Clear()
		dc.SetRGB(lc.Get(xl.ControlSlider[0]), lc.Get(xl.ControlSlider[1]), lc.Get(xl.ControlSlider[2]))

		// for i := 0; i <= N; i++ {
		// 	t := float64(i) / N
		// 	d := t*S*10*lc.Get(xl.ControlKnobSendA[0]) + lc.Get(xl.ControlKnobSendA[1])
		// 	a := t * math.Pi * 10 * lc.Get(xl.ControlKnobSendA[2])
		// 	x := width/2 + math.Cos(a)*d
		// 	y := height/2 + math.Sin(a)*d
		// 	r := t * lc.Get(xl.ControlKnobSendA[3]) * 8
		// 	dc.DrawCircle(x, y, r)
		// }
		// dc.Fill()

		dc.DrawStringAnchored(render, offset, S/2, 0, 0.5)

		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				idx := y*width + x
				value := img.RGBAAt(x, y)
				sender.Buffer[idx] = colorful.Color{
					R: float64(value.R) / 255,
					G: float64(value.G) / 255,
					B: float64(value.B) / 255,
				}
			}
		}

		sender.send()
		time.Sleep(time.Duration(250*lc.Get(xl.ControlKnobSendA[0])) * time.Millisecond)
	}
}
