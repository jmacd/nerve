package main

import (
	"fmt"
	"net"
	"time"

	"github.com/jsimonetti/go-artnet/packet"
	"github.com/lucasb-eyer/go-colorful"
)

const (
	ipAddr = "192.168.0.22"

	pixels = 300 // Lies (it's 299)
	width  = 20
	height = 15

	epsilon = 0.00001
)

type (
	Buffer [pixels]colorful.Color

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

func main() {
	sender := newSender()
	buffer := sender.Buffer[:]

	start := time.Now()

	for {
		seconds := time.Now().Sub(start) / time.Second

		pattern := (seconds / 15) % 4

		for twoCount := 0.0; twoCount < 2; twoCount += 0.005 {
			level := twoCount
			if level >= 1 {
				level = 2 - level
			}

			for y := 0; y < height; y++ {
				for x := 0; x < width; x++ {
					var c colorful.Color

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

					buffer[y*width+x] = c.Clamped()

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

type GradientTable []struct {
	Col colorful.Color
	Pos float64
}

func (self GradientTable) GetInterpolatedColorFor(t float64) colorful.Color {
	for i := 0; i < len(self)-1; i++ {
		c1 := self[i]
		c2 := self[i+1]
		if c1.Pos <= t && t <= c2.Pos {
			// We are in between c1 and c2. Go blend them!
			t := (t - c1.Pos) / (c2.Pos - c1.Pos)
			return c1.Col.BlendHcl(c2.Col, t).Clamped()
		}
	}

	// Nothing found? Means we're at (or past) the last gradient keypoint.
	return self[len(self)-1].Col
}
