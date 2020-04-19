package artnet

import (
	"fmt"
	"net"

	"github.com/jmacd/go-artnet/packet"
	"github.com/lucasb-eyer/go-colorful"
)

const (
	maxPerPacket = 170
)

type (
	Sender struct {
		dest *net.UDPAddr
		conn *net.UDPConn

		packet.ArtDMXPacket
	}

	Color = colorful.Color
)

func NewSender(ipAddr string) *Sender {
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

func (s *Sender) Send(buffer []Color) {
	data := s.ArtDMXPacket.Data[:]
	s.ArtDMXPacket.SubUni = 0
	pixels := len(buffer)
	for p := 0; p < pixels; {

		num := maxPerPacket
		if pixels-p < num {
			num = pixels - p
		}

		for i := 0; i < num; i++ {
			c := buffer[p+i]
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
