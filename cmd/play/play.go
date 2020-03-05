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
)

var keypoints = GradientTable{
	{MustParseHex("#9e0142"), 0.0},
	{MustParseHex("#d53e4f"), 0.1},
	{MustParseHex("#f46d43"), 0.2},
	{MustParseHex("#fdae61"), 0.3},
	{MustParseHex("#fee090"), 0.4},
	{MustParseHex("#ffffbf"), 0.5},
	{MustParseHex("#e6f598"), 0.6},
	{MustParseHex("#abdda4"), 0.7},
	{MustParseHex("#66c2a5"), 0.8},
	{MustParseHex("#3288bd"), 0.9},
	{MustParseHex("#5e4fa2"), 1.0},
}

type Buffer [pixels]colorful.Color

type Sender struct {
	dest *net.UDPAddr
	conn *net.UDPConn

	Buffer
	packet.ArtDMXPacket
}

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

	for {
		for i := 0; i < 100; i++ {
			c := keypoints.GetInterpolatedColorFor(float64(i) / float64(100))

			for i := 0; i < pixels; i++ {
				buffer[i] = c
			}

			sender.send()
			time.Sleep(time.Millisecond * 160)
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
			data[i*3+0] = byte(c.B * 255)
			data[i*3+1] = byte(c.G * 255)
			data[i*3+2] = byte(c.R * 255)
		}

		s.ArtDMXPacket.SubUni++

		b, _ := s.ArtDMXPacket.MarshalBinary()

		_, err := s.conn.WriteTo(b, s.dest)
		if err != nil {
			panic(fmt.Sprint("error writing packet: ", err))
		}

		//fmt.Println("Packet at P", p)
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

func MustParseHex(s string) colorful.Color {
	c, err := colorful.Hex(s)
	if err != nil {
		panic("MustParseHex: " + err.Error())
	}
	return c
}
