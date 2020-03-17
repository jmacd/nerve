package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"net"
	"time"

	"github.com/hsluv/hsluv-go"
	"github.com/jmacd/nerve/lctlxl"
	"github.com/jsimonetti/go-artnet/packet"
	"github.com/lucasb-eyer/go-colorful"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/exporters/metric/stdout"
)

const (
	// ipAddr = "192.168.1.167"  // Bldg
	ipAddr = "192.168.0.21" // Home

	pixels = 300
	width  = 20
	height = 15

	epsilon = 0.00001
)

var (
	meter  = global.MeterProvider().Meter("main")
	frames = meter.NewInt64Counter("frames").Bind(meter.Labels())
)

type (
	Buffer [pixels]colorful.Color

	Color = colorful.Color

	Sender struct {
		dest *net.UDPAddr
		conn *net.UDPConn

		Buffer
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
	stop := telemetry()
	defer stop()

	sender := newSender()

	lc, err := lctlxl.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer lc.Stop()

	lc.Start()

	white(sender, lc)
}

func white(sender *Sender, lc *lctlxl.LaunchControl) {
	for {
		level := lc.SendA[0]

		wref := [3]float64{0.5 + lc.SendA[1], 1.00000, 0.5 + lc.SendA[2]}

		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				var c Color
				// Hmm

				c = colorful.Color{level, level, level}
				c = colorful.Xyy(c.XyyWhiteRef(wref))

				sender.Buffer[y*width+x] = c.Clamped()

			}
		}

		sender.send()
		time.Sleep(time.Millisecond * 10)
	}
}

func hcl(sender *Sender, lc *lctlxl.LaunchControl) {
	for {
		level := lc.SendA[0]

		wref := [3]float64{0.5 + lc.SendA[1], 1.00000, 0.5 + lc.SendA[2]}

		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				var c Color

				c = colorful.HclWhiteRef(
					360*(float64(x)+epsilon)/(width-1+epsilon),
					(float64(y)+epsilon)/(height-1+epsilon),
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

func luv(sender *Sender, lc *lctlxl.LaunchControl) {
	for {
		level := lc.SendA[0]

		wref := [3]float64{0.5 + lc.SendA[1], 1.00000, 0.5 + lc.SendA[2]}

		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				var c Color

				c = colorful.LuvWhiteRef(
					level,
					(float64(y)+epsilon)/(height-1+epsilon),
					(float64(x)+epsilon)/(width-1+epsilon),
					wref,
				)

				sender.Buffer[y*width+x] = c.Clamped()

			}
		}

		sender.send()
		time.Sleep(time.Millisecond * 10)
	}
}

func lab(sender *Sender, lc *lctlxl.LaunchControl) {
	for {
		level := lc.SendA[0]

		wref := [3]float64{0.5 + lc.SendA[1], 1.00000, 0.5 + lc.SendA[2]}

		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				var c Color

				c = colorful.LabWhiteRef(
					level,
					(float64(y)+epsilon)/(height-1+epsilon),
					(float64(x)+epsilon)/(width-1+epsilon),
					wref,
				)

				sender.Buffer[y*width+x] = c.Clamped()

			}
		}

		sender.send()
		time.Sleep(time.Millisecond * 10)
	}
}

func showHsluv(sender *Sender, lc *lctlxl.LaunchControl) {
	for frame := 0; ; frame++ {
		r, g, b := hsluv.HsluvToRGB(360*lc.Slide[0], 100*lc.Slide[1], 100*lc.Slide[2])

		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				sender.Buffer[y*width+x] = Color{
					R: r,
					G: g,
					B: b,
				}
			}
		}

		sender.send()
		time.Sleep(10 * time.Millisecond)
	}
}

func gamma(sender *Sender, lc *lctlxl.LaunchControl) {
	for frame := 0; ; frame++ {
		gamma := math.Max(lc.SendA[0]*3, 0.001)
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

func walk(sender *Sender, lc *lctlxl.LaunchControl) {
	last := time.Now()
	elapsed := 0.0

	for frame := 0; ; frame++ {
		now := time.Now()
		delta := now.Sub(last)

		last = now

		rate := 100 * lc.SendA[0]

		elapsed += rate * float64(delta.Seconds())

		spot := int64(elapsed) % pixels
		sender.Buffer[spot] = Color{
			R: lc.Slide[0] * lc.SendA[1],
			G: lc.Slide[1] * lc.SendA[1],
			B: lc.Slide[2] * lc.SendA[1],
		}

		sender.send()
		time.Sleep(time.Duration(250*lc.SendA[0]) * time.Millisecond)
	}
}

func strobe2(sender *Sender, lc *lctlxl.LaunchControl) {
	for frame := 0; ; frame++ {

		if frame%2 == 0 {
			for i := 0; i < pixels; i++ {
				sender.Buffer[i] = Color{
					R: lc.Slide[3] * lc.SendA[1],
					G: lc.Slide[4] * lc.SendA[1],
					B: lc.Slide[5] * lc.SendA[1],
				}
			}
		} else {
			for i := 0; i < pixels; i++ {
				sender.Buffer[i] = Color{
					R: lc.Slide[0] * lc.SendA[1],
					G: lc.Slide[1] * lc.SendA[1],
					B: lc.Slide[2] * lc.SendA[1],
				}
			}
		}

		sender.send()
		time.Sleep(time.Duration(250*lc.SendA[0]) * time.Millisecond)
	}
}

func colors(sender *Sender, lc *lctlxl.LaunchControl) {
	start := time.Now()

	for {
		seconds := time.Now().Sub(start) / time.Second

		pattern := (seconds / 15) % 4

		for twoCount := 0.0; twoCount < 2; twoCount += 0.005 {
			level := lc.SendA[0]

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
							(float64(y)+epsilon)/(height-1+epsilon),
							(float64(x)+epsilon)/(width-1+epsilon),
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

		num := 170
		if pixels-p < num {
			num = pixels - p
		}

		for i := 0; i < num; i++ {
			c := s.Buffer[p+i]
			data[i*3+0] = byte(c.R * 255)
			data[i*3+1] = byte(c.G * 255)
			data[i*3+2] = byte(c.B * 255)
		}

		b, _ := s.ArtDMXPacket.MarshalBinary()

		_, err := s.conn.WriteTo(b, s.dest)
		if err != nil {
			panic(fmt.Sprint("error writing packet: ", err))
		}

		s.ArtDMXPacket.SubUni++
		p += num
	}
}
