package artnet

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/jmacd/go-artnet/packet"
	"github.com/lucasb-eyer/go-colorful"
)

const (
	maxPerPacket = 170
)

type (
	Sender struct {
		destStr string
		srcStr  string

		dest *net.UDPAddr
		conn *net.UDPConn

		inputCh chan []Color

		lastLog time.Time

		packet.ArtDMXPacket
	}

	Color = colorful.Color
)

func NewSender(ipAddr string) *Sender {
	return &Sender{
		destStr: fmt.Sprintf("%s:%d", ipAddr, packet.ArtNetPort),
		srcStr:  fmt.Sprintf("%s:%d", "", packet.ArtNetPort),
		inputCh: make(chan []Color),
	}
}

func (s *Sender) Send(buffer []Color) {
	s.inputCh <- buffer
}

func (s *Sender) Run(ctx context.Context) error {
	grp, ctx := errgroup.WithContext(ctx)

	if err := s.reconnect(); err != nil {
		log.Println("artnet sender:", err)
	}

	grp.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case buffer := <-s.inputCh:
				var err error
				if s.conn == nil {
					err = s.reconnect()
				}
				if err == nil {
					err = s.send(buffer)
				}
				if err != nil {
					now := time.Now()
					if now.Sub(s.lastLog) >= time.Second {
						log.Printf("send: %v\n", err)
						s.lastLog = now
					}
				}
			}
		}
	})
	return grp.Wait()
}

func (s *Sender) reconnect() error {
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
	return nil
}

func (s *Sender) send(buffer []Color) error {
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
