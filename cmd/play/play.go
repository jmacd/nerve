package main

import (
	"fmt"
	"net"

	"github.com/jmacd/go-artnet/packet"
	"github.com/lucasb-eyer/go-colorful"
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

func main() {

	dst := fmt.Sprintf("%s:%d", "192.168.0.14", packet.ArtNetPort)
	node, _ := net.ResolveUDPAddr("udp", dst)
	src := fmt.Sprintf("%s:%d", "", packet.ArtNetPort)
	localAddr, _ := net.ResolveUDPAddr("udp", src)

	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		fmt.Printf("error opening udp: %s\n", err)
		return
	}

	// set channels 1 and 4 to FL, 2, 3 and 5 to FD
	// on my colorBeam this sets output 1 to fullbright red with zero strobing

	p := &packet.ArtDMXPacket{
		Sequence: 0,
		SubUni:   0,
		Net:      0,
	}

	for i := 0; i < 170; i++ {
		c := keypoints.GetInterpolatedColorFor(float64(i) / float64(170))

		p.Data[i*3+0] = byte(c.B * 255)
		p.Data[i*3+1] = byte(c.G * 255)
		p.Data[i*3+2] = byte(c.R * 255)
	}

	// fmt.Println("LOOK", p.Data)
	b, err := p.MarshalBinary()

	n, err := conn.WriteTo(b, node)
	if err != nil {
		fmt.Printf("error writing packet: %s\n", err)
		return
	}
	fmt.Printf("packet sent, wrote %d bytes\n", n)
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
