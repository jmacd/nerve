package artnet

import (
	"context"
	"fmt"
	"image"
	"net"
	"sync"

	"github.com/jmacd/go-artnet/packet"
	"github.com/jmacd/go-artnet/packet/code"
)

type Receiver struct {
	conn *net.UDPConn
	out  *image.RGBA
	in   *image.RGBA
	cpy  *image.RGBA
	wg   sync.WaitGroup

	lsu uint8
	off uint64
}

func NewReceiver(hostIP string, out *image.RGBA) (*Receiver, error) {
	src := fmt.Sprintf("%s:%d", hostIP, packet.ArtNetPort)
	localAddr, _ := net.ResolveUDPAddr("udp", src)

	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		fmt.Printf("error opening udp: %s\n", err)
		return nil, err
	}
	return &Receiver{
		conn: conn,
		out:  out,
		cpy:  image.NewRGBA(out.Bounds()),
		in:   image.NewRGBA(out.Bounds()),
	}, nil
}

func (r *Receiver) Start(ctx context.Context) error {
	recvCh := make(chan []byte, 1000)
	r.wg.Add(2)

	go func() {
		defer r.wg.Done()
		buf := make([]byte, 1024)
		for {
			n, _, err := r.conn.ReadFromUDP(buf) // first packet you read will be your own
			if err != nil {
				fmt.Printf("error reading packet: %s\n", err)
				continue

			}
			// fmt.Printf("packet received from %v, read %d bytes\n", addr.IP, n)

			// if addr.IP.Equal(localAddr.IP) {
			// 	// skip messages from myself
			// 	continue
			// }
			select {
			case <-ctx.Done():
				return
			case recvCh <- buf[:n]:
				buf = make([]byte, 1024)
			}
		}
	}()

	go func() {
		defer r.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case b := <-recvCh:
				p, err := packet.Unmarshal(b)
				if err != nil {
					fmt.Printf("error unmarshalling packet: %s\n", err)
					continue
				}
				switch p.GetOpCode() {
				case code.OpDMX:
					dmx := p.(*packet.ArtDMXPacket)
					off := uint64(0)
					switch dmx.SubUni {
					case 0:
						off = 0
						r.lsu = 0
						r.off = 0
					case r.lsu + 1:
						off = r.off
						r.lsu++
					default:
						fmt.Println("artnet: dmx reset", dmx.SubUni, dmx.Length, r.off, r.lsu)
						off = 0
						r.off = 0
						r.lsu = dmx.SubUni

					}

					for i := uint64(0); i*3+2 < uint64(dmx.Length); i++ {
						o := (off + i) * 4
						if o+2 < uint64(len(r.in.Pix)) {
							r.in.Pix[o+0] = dmx.Data[i*3+0]
							r.in.Pix[o+1] = dmx.Data[i*3+1]
							r.in.Pix[o+2] = dmx.Data[i*3+2]
							r.off += 1
						} else {
							break
						}
					}

					if int(r.off*4) == len(r.in.Pix) {
						copy(r.cpy.Pix, r.in.Pix)
					}
				default:
					fmt.Printf("artnet: %v %#v\n", p.GetOpCode(), p)
				}
			}
		}
	}()

	return nil
}

func (r *Receiver) Draw() error {
	// @@ TODO no synchronization here
	copy(r.out.Pix, r.cpy.Pix)
	return nil
}
