package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"github.com/jmacd/launchmidi/launchctl/xl"
)

const deviceName = "/dev/rpmsg_pru30"

type (
	RPMsgDevice struct {
		file *os.File
	}

	controlStruct struct {
		framebufsAddr uint32
		framebufsSize uint32
		frameCount    uint32
		dmaWait       uint32
	}

	Frameset  [2]Framebank
	Framebank [256]Framebuf
	Framebuf  [16]DoubleRow
	DoubleRow [64]DoublePixel

	DoublePixel struct {
		Gpio0 uint32
		Gpio1 uint32
		Gpio2 uint32
		Gpio3 uint32
	}
)

func openRPMsgDevice() (*RPMsgDevice, error) {
	file, err := os.OpenFile(deviceName, os.O_RDWR, 0666)
	return &RPMsgDevice{
		file: file,
	}, err
}

func (r *RPMsgDevice) write(data []byte) error {
	n, err := r.file.Write(data)
	if err != nil {
		return nil
	}
	if n != len(data) {
		return fmt.Errorf("short write: %d != %d", n, len(data))
	}
	return nil
}

func mmap(file *os.File, addr uint32, size int) ([]byte, error) {
	return syscall.Mmap(
		int(file.Fd()),
		int64(addr),
		size,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED,
	)
}

func (r *RPMsgDevice) readControl() (*controlStruct, *Frameset, error) {
	var data [32]byte
	n, err := r.file.Read(data[:])
	if err != nil {
		return nil, nil, err
	}
	if n != 4 {
		return nil, nil, fmt.Errorf("expected 4 bytes control address")
	}
	addr := binary.LittleEndian.Uint32(data[0:4])

	mem, err := os.OpenFile("/dev/mem", os.O_RDWR, 0)
	if err != nil {
		return nil, nil, err
	}

	cdata, err := mmap(mem, addr, int(unsafe.Sizeof(controlStruct{})))
	if err != nil {
		return nil, nil, err
	}

	ctrl := (*controlStruct)(unsafe.Pointer(&cdata[0]))

	fdata, err := mmap(mem, ctrl.framebufsAddr, int(ctrl.framebufsSize))
	if err != nil {
		return nil, nil, err
	}
	if ctrl.framebufsSize != uint32(unsafe.Sizeof(Frameset{})) {
		return nil, nil, fmt.Errorf("frameset size mismatch: C=%d, Go=%d", ctrl.framebufsSize, unsafe.Sizeof(Frameset{}))
	}

	framebuf := (*Frameset)(unsafe.Pointer(&fdata[0]))

	// Note: should drop privileges now.
	return ctrl, framebuf, nil
}

func Main() error {
	l, err := xl.Open()
	if err != nil {
		log.Fatalf("error while openning connection to launchctl: %v", err)
	}
	defer l.Close()

	rpm, err := openRPMsgDevice()
	if err != nil {
		return err
	}

	// pru.c does not parse the message, this delivers the
	// interrupt which causes pru.c to respond with its two
	// carveout addresses.
	if err := rpm.write([]byte("wakeup")); err != nil {
		return err
	}

	var ctrl *controlStruct
	var frames *Frameset

	ctrl, frames, err = rpm.readControl()
	if err != nil {
		return err
	}

	log.Println("Starting fill...")

	go func() {
		before := atomic.LoadUint32(&ctrl.frameCount)
		for {
			time.Sleep(time.Second)
			after := atomic.LoadUint32(&ctrl.frameCount)
			log.Println("frames/sec", after-before)
			before = after
		}
	}()

	set := func(r, g, b float64) {

		for bankNo := 0; bankNo < 2; bankNo++ {
			for frameNo := 0; frameNo < 256; frameNo++ {
				for dblrowNo := 0; dblrowNo < 16; dblrowNo++ {
					for pix := 0; pix < 64; pix++ {
						pixel := &(*frames)[bankNo][frameNo][dblrowNo][pix]
						pixel.j1r1(rand.Float64() < r)
						pixel.j1g1(rand.Float64() < g)
						pixel.j1b1(rand.Float64() < b)
						pixel.j1r2(rand.Float64() < r)
						pixel.j1g2(rand.Float64() < g)
						pixel.j1b2(rand.Float64() < b)
						pixel.j3r1(rand.Float64() < r)
						pixel.j3g1(rand.Float64() < g)
						pixel.j3b1(rand.Float64() < b)
						pixel.j3r2(rand.Float64() < r)
						pixel.j3g2(rand.Float64() < g)
						pixel.j3b2(rand.Float64() < b)
					}
				}
			}
		}
	}

	r := 0.8
	g := 0.05
	b := 0.15

	set(r, g, b)

	l.AddCallback(xl.AllChannels, xl.ControlSlider[0], func(ch int, control xl.Control, value xl.Value) {
		r = value.Float()
	})
	l.AddCallback(xl.AllChannels, xl.ControlSlider[1], func(ch int, control xl.Control, value xl.Value) {
		g = value.Float()
	})
	l.AddCallback(xl.AllChannels, xl.ControlSlider[2], func(ch int, control xl.Control, value xl.Value) {
		b = value.Float()
	})
	go func() {
		err := l.Run(context.Background())
		if err != nil {
			log.Println("LX control run:", err)
		}
		log.Println("LX control exit")
	}()

	go func() {
		for {
			log.Println("Render", r, g, b)
			set(r, g, b)
		}
	}()

	select {}
}

func main() {
	if err := Main(); err != nil {
		log.Println("error:", err)
	}
}

func set(ptr *uint32, on bool, pos uint32) {
	// var val uint32
	// if on {
	// 	val++
	// }
	// *ptr = (*ptr &^ (uint32(1) << pos)) | (val << pos)

	if on {
		*ptr |= 1 << pos
	} else {
		*ptr &= ^(1 << pos)
	}
}

func (p *DoublePixel) j1r1(on bool) { set(&p.Gpio2, on, 2) }
func (p *DoublePixel) j1g1(on bool) { set(&p.Gpio2, on, 3) }
func (p *DoublePixel) j1g2(on bool) { set(&p.Gpio2, on, 4) }
func (p *DoublePixel) j1b1(on bool) { set(&p.Gpio2, on, 5) }
func (p *DoublePixel) j1r2(on bool) { set(&p.Gpio0, on, 23) }
func (p *DoublePixel) j1b2(on bool) { set(&p.Gpio0, on, 26) }

func (p *DoublePixel) j3r2(on bool) { set(&p.Gpio1, on, 16) }
func (p *DoublePixel) j3g1(on bool) { set(&p.Gpio1, on, 18) }
func (p *DoublePixel) j3g2(on bool) { set(&p.Gpio0, on, 3) }
func (p *DoublePixel) j3b2(on bool) { set(&p.Gpio0, on, 5) }
func (p *DoublePixel) j3r1(on bool) { set(&p.Gpio0, on, 30) }
func (p *DoublePixel) j3b1(on bool) { set(&p.Gpio0, on, 31) }
