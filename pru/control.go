// MIT License
//
// Copyright (C) Joshua MacDonald
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"github.com/jmacd/launchmidi/launchctl/xl"
	"github.com/jmacd/nerve/pru/gpixio"
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

	Frameset    = gpixio.Frameset
	Schedule    = gpixio.Schedule
	Frame       = gpixio.Frame
	DoubleRow   = gpixio.DoubleRow
	DoublePixel = gpixio.DoublePixel
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
		log.Fatalf("error while opening connection to launchctl: %v", err)
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
			time.Sleep(10 * time.Second)
			after := atomic.LoadUint32(&ctrl.frameCount)
			log.Println("frames/sec", after-before)
			before = after
		}
	}()

	r := 0.8
	g := 0.05
	b := 0.15

	focus := 0.0

	//frames.setBernoulli(r, g, b)

	// Sliders 0-2 control R, G, B
	l.AddCallback(xl.AllChannels, xl.ControlSlider[0], func(ch int, control xl.Control, value xl.Value) {
		r = value.Float()
	})
	l.AddCallback(xl.AllChannels, xl.ControlSlider[1], func(ch int, control xl.Control, value xl.Value) {
		g = value.Float()
	})
	l.AddCallback(xl.AllChannels, xl.ControlSlider[2], func(ch int, control xl.Control, value xl.Value) {
		b = value.Float()
	})
	// Track buttons 0-2 set the dithering mode.
	l.AddCallback(xl.AllChannels, xl.ControlButtonTrackFocus[0], func(ch int, control xl.Control, value xl.Value) {
		focus = 0
	})
	l.AddCallback(xl.AllChannels, xl.ControlButtonTrackFocus[1], func(ch int, control xl.Control, value xl.Value) {
		focus = 1
	})
	l.AddCallback(xl.AllChannels, xl.ControlButtonTrackFocus[2], func(ch int, control xl.Control, value xl.Value) {
		focus = 2
	})
	go func() {
		err := l.Run(context.Background())
		if err != nil {
			log.Println("LX control run:", err)
		}
		log.Println("LX control exit")
	}()

	go func() {
		buf := gpixio.NewBuffer()

		_ = focus

		for s := 0; ; s = (s + 1) % 8 {

			buf.SetRGB(0.2, 0.2, 0.2)
			buf.DrawCircle(64, 64, 50)
			buf.SetRGB(r, g, b)
			buf.Fill()

			buf.Copy(&frames[s])
		}
	}()

	select {}
}

func main() {
	if err := Main(); err != nil {
		log.Println("error:", err)
	}
}
