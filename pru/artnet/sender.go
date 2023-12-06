package artnet

import (
	"fmt"
	"image"
	"log"
	"net"
	"time"

	"github.com/jmacd/go-artnet/packet"
	"github.com/lucasb-eyer/go-colorful"
)

type (
	Sender struct {
		destStr string
		srcStr  string

		dest *net.UDPAddr
		conn *net.UDPConn

		lastLog time.Time

		packet.ArtDMXPacket
	}

	Color = colorful.Color
)

func NewSender(ipAddr string) *Sender {
	return &Sender{
		destStr: fmt.Sprintf("%s:%d", ipAddr, packet.ArtNetPort),
		srcStr:  fmt.Sprintf("%s:%d", "", packet.ArtNetPort),
	}
}

func (s *Sender) Send(buffer *image.RGBA) error {
	err := s.send(buffer)
	if err != nil {
		now := time.Now()
		if now.Sub(s.lastLog) >= time.Second {
			log.Printf("send: %v\n", err)
			s.lastLog = now
		}
	}
	return err
}

func (s *Sender) send(buffer *image.RGBA) error {
	if s.conn == nil {
		node, err := net.ResolveUDPAddr("udp", s.destStr)
		if err != nil {
			return fmt.Errorf("error resolving local udp: %v", err)
		}
		localAddr, _ := net.ResolveUDPAddr("udp", s.srcStr)
		conn, err := net.ListenUDP("udp", localAddr)
		if err != nil {
			return fmt.Errorf("error resolving artnet udp: %v", err)
		}
		s.dest = node
		s.conn = conn
	}

	data := s.ArtDMXPacket.Data[:]
	s.ArtDMXPacket.SubUni = 0
	pixels := buffer.Rect.Dx() * buffer.Rect.Dy()
	for p := 0; p < pixels; {

		num := maxPerPacket
		if pixels-p < num {
			num = pixels - p
		}
		for i := 0; i < num; i++ {
			pi := 4 * (p + i)
			data[i*3+0] = buffer.Pix[pi+0]
			data[i*3+1] = buffer.Pix[pi+1]
			data[i*3+2] = buffer.Pix[pi+2]
		}
		s.ArtDMXPacket.Length = uint16(num * 3)

		b, _ := s.ArtDMXPacket.MarshalBinary()

		//fmt.Println("Pkt", len(b), "w", num)
		if len(b) > maxPacketSize {
			panic(fmt.Sprint("wrong size", len(b)))
		}

		_, err := s.conn.WriteTo(b, s.dest)
		if err != nil {
			s.conn.Close()
			s.conn = nil
			s.dest = nil
			return fmt.Errorf("error writing packet: %v", err)
		}
		s.ArtDMXPacket.SubUni++
		p += num
	}
	return nil
}
