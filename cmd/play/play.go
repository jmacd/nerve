package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/jmacd/nerve/lctlxl"
	"github.com/jsimonetti/go-artnet/packet"
	"github.com/lucasb-eyer/go-colorful"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/exporters/metric/stdout"
)

const (
	// ipAddr = "192.168.1.167"  // Minna
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

	walk(sender, lc)
}

func walk(sender *Sender, lc *lctlxl.LaunchControl) {
	last := time.Now()
	elapsed := 0.0

	for {
		now := time.Now()
		delta := now.Sub(last)

		last = now

		rate := 100 * lc.SendA[0]

		elapsed += rate * float64(delta.Seconds())

		spot := int64(elapsed) % pixels

		sender.Buffer = Buffer{}

		sender.Buffer[spot] = Color{
			R: lc.Slide[0],
			G: lc.Slide[1],
			B: lc.Slide[2],
		}

		sender.send()
		time.Sleep(time.Duration(10*lc.SendA[1]) * time.Millisecond)
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
