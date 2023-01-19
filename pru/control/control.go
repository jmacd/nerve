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
	"flag"
	"fmt"
	"image"
	"log"
	"math/rand"
	"os"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"github.com/fogleman/gg"
	"github.com/jmacd/launchmidi/launchctl/xl"
	"github.com/jmacd/nerve/pru/gpixio"
)

const deviceName = "/dev/rpmsg_pru30"

var (
	noInput = flag.Bool("no_input", false, "do not use MIDI input device")
	noBBB   = flag.Bool("no_bbb", false, "do not use BeagleBone device")
)

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
	var input *xl.LaunchControl

	flag.Parse()

	if !*noInput {
		input, err := xl.Open()
		if err != nil {
			log.Fatalf("error while opening connection to launchctl: %v", err)
		}
		defer input.Close()
	}

	var outputWindow fyne.Window
	var inputWindow fyne.Window
	var frames *Frameset
	var inputImage *canvas.Image
	var outputImage *canvas.Image
	var buf = gpixio.NewBuffer()
	outputPixels := image.NewRGBA(image.Rect(0, 0, 128, 128))

	if *noBBB {
		// frames is not in shared memory
		frames = &Frameset{}

		app := app.New()

		inputWindow = app.NewWindow("Image")
		inputImage = canvas.NewImageFromImage(buf.RGBA)
		inputImage.FillMode = canvas.ImageFillOriginal
		inputWindow.SetContent(inputImage)

		outputWindow = app.NewWindow("Visage")
		outputImage = canvas.NewImageFromImage(outputPixels)
		outputImage.FillMode = canvas.ImageFillOriginal
		outputWindow.SetContent(outputImage)

	} else {
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

		ctrl, frames, err = rpm.readControl()
		if err != nil {
			return err
		}

		go func() {
			before := atomic.LoadUint32(&ctrl.frameCount)
			for {
				time.Sleep(10 * time.Second)
				after := atomic.LoadUint32(&ctrl.frameCount)
				log.Println("frames/sec", after-before)
				before = after
			}
		}()
	}

	r := 0.8
	g := 0.05
	b := 0.15

	focus := 0.0

	if input != nil {
		// Sliders 0-2 control R, G, B
		input.AddCallback(xl.AllChannels, xl.ControlSlider[0], func(ch int, control xl.Control, value xl.Value) {
			r = value.Float()
		})
		input.AddCallback(xl.AllChannels, xl.ControlSlider[1], func(ch int, control xl.Control, value xl.Value) {
			g = value.Float()
		})
		input.AddCallback(xl.AllChannels, xl.ControlSlider[2], func(ch int, control xl.Control, value xl.Value) {
			b = value.Float()
		})
		// Track buttons 0-2 set the dithering mode.
		input.AddCallback(xl.AllChannels, xl.ControlButtonTrackFocus[0], func(ch int, control xl.Control, value xl.Value) {
			focus = 0
		})
		input.AddCallback(xl.AllChannels, xl.ControlButtonTrackFocus[1], func(ch int, control xl.Control, value xl.Value) {
			focus = 1
		})
		input.AddCallback(xl.AllChannels, xl.ControlButtonTrackFocus[2], func(ch int, control xl.Control, value xl.Value) {
			focus = 2
		})

		go func() {
			err := input.Run(context.Background())
			if err != nil {
				log.Println("LX control run:", err)
			}
			log.Println("LX control exit")
		}()
	} else {
		go func() {
			for {
				time.Sleep(time.Second)
				r, g, b = rand.Float64(), rand.Float64(), rand.Float64()
			}
		}()
	}

	go func() {
		ggctx := gg.NewContextForRGBA(buf.RGBA)

		_ = focus

		for s := 0; ; s = (s + 1) % 8 {
			ggctx.DrawRectangle(0, 0, 128, 128)
			ggctx.SetRGB(0.8, 0.8, 0)
			ggctx.Fill()

			ggctx.DrawCircle(64, 64, 60)
			ggctx.SetRGB(r, g, b)
			ggctx.Fill()

			canvas.Refresh(inputImage)

			time.Sleep(33 * time.Millisecond)

			buf.Copy(&frames[s])

			testRender(&frames[s], outputPixels)

			canvas.Refresh(outputImage)
		}
	}()

	if inputWindow != nil {
		outputWindow.Show()
		inputWindow.ShowAndRun()
		return nil
	} else {
		select {}
	}
}

func add4(bp *byte) {
	(*bp) += 4 // 4 = 2 bits shifted (64 frames => 256 levels)

	// overflow indicates max brightness, b/c we counted 4 per up to 64 frames
	if *bp == 0 {
		*bp = 255
	}
}

func testRender(sched *gpixio.Schedule, img *image.RGBA) {

	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i+0] = 0
		img.Pix[i+1] = 0
		img.Pix[i+2] = 0
		img.Pix[i+3] = 255
	}

	for fi := 0; fi < 64; fi++ {
		for dri := 0; dri < 16; dri++ {
			for dpi := 0; dpi < 64; dpi++ {
				dp := (*sched)[fi][dri][dpi]

				// 4 = bytes per pixel, 128 = image width

				// The following 16 positions are manual
				j11Off := 4 * (128*(dri+0) + dpi)
				j12Off := 4 * (128*(dri+16) + dpi)
				j21Off := 4 * (128*(dri+32) + dpi)
				j22Off := 4 * (128*(dri+48) + dpi)
				j31Off := 4 * (128*(dri+64) + dpi)
				j32Off := 4 * (128*(dri+80) + dpi)
				j41Off := 4 * (128*(dri+96) + dpi)
				j42Off := 4 * (128*(dri+112) + dpi)

				j51Off := 4 * (128*(dri+0) + 64 + dpi)
				j52Off := 4 * (128*(dri+16) + 64 + dpi)
				j61Off := 4 * (128*(dri+32) + 64 + dpi)
				j62Off := 4 * (128*(dri+48) + 64 + dpi)
				j71Off := 4 * (128*(dri+64) + 64 + dpi)
				j72Off := 4 * (128*(dri+80) + 64 + dpi)
				j81Off := 4 * (128*(dri+96) + 64 + dpi)
				j82Off := 4 * (128*(dri+112) + 64 + dpi)

				// The following code was generated by ../../cmd/mkmap
				if dp.Gpio2&(1<<4) != 0 {
					add4(&img.Pix[j12Off+1])
				}
				if dp.Gpio0&(1<<26) != 0 {
					add4(&img.Pix[j12Off+2])
				}
				if dp.Gpio2&(1<<2) != 0 {
					add4(&img.Pix[j11Off+0])
				}
				if dp.Gpio2&(1<<3) != 0 {
					add4(&img.Pix[j11Off+1])
				}
				if dp.Gpio2&(1<<5) != 0 {
					add4(&img.Pix[j11Off+2])
				}
				if dp.Gpio0&(1<<23) != 0 {
					add4(&img.Pix[j12Off+0])
				}
				if dp.Gpio2&(1<<23) != 0 {
					add4(&img.Pix[j22Off+1])
				}
				if dp.Gpio2&(1<<24) != 0 {
					add4(&img.Pix[j22Off+2])
				}
				if dp.Gpio0&(1<<27) != 0 {
					add4(&img.Pix[j21Off+0])
				}
				if dp.Gpio2&(1<<1) != 0 {
					add4(&img.Pix[j21Off+1])
				}
				if dp.Gpio0&(1<<22) != 0 {
					add4(&img.Pix[j21Off+2])
				}
				if dp.Gpio2&(1<<22) != 0 {
					add4(&img.Pix[j22Off+0])
				}
				if dp.Gpio1&(1<<18) != 0 {
					add4(&img.Pix[j31Off+1])
				}
				if dp.Gpio0&(1<<31) != 0 {
					add4(&img.Pix[j31Off+2])
				}
				if dp.Gpio1&(1<<16) != 0 {
					add4(&img.Pix[j32Off+0])
				}
				if dp.Gpio0&(1<<3) != 0 {
					add4(&img.Pix[j32Off+1])
				}
				if dp.Gpio0&(1<<5) != 0 {
					add4(&img.Pix[j32Off+2])
				}
				if dp.Gpio0&(1<<30) != 0 {
					add4(&img.Pix[j31Off+0])
				}
				if dp.Gpio0&(1<<2) != 0 {
					add4(&img.Pix[j41Off+0])
				}
				if dp.Gpio0&(1<<15) != 0 {
					add4(&img.Pix[j41Off+1])
				}
				if dp.Gpio1&(1<<17) != 0 {
					add4(&img.Pix[j41Off+2])
				}
				if dp.Gpio3&(1<<21) != 0 {
					add4(&img.Pix[j42Off+0])
				}
				if dp.Gpio3&(1<<19) != 0 {
					add4(&img.Pix[j42Off+1])
				}
				if dp.Gpio0&(1<<4) != 0 {
					add4(&img.Pix[j42Off+2])
				}
				if dp.Gpio0&(1<<11) != 0 {
					add4(&img.Pix[j51Off+1])
				}
				if dp.Gpio0&(1<<10) != 0 {
					add4(&img.Pix[j51Off+2])
				}
				if dp.Gpio0&(1<<9) != 0 {
					add4(&img.Pix[j52Off+0])
				}
				if dp.Gpio0&(1<<8) != 0 {
					add4(&img.Pix[j52Off+1])
				}
				if dp.Gpio2&(1<<17) != 0 {
					add4(&img.Pix[j52Off+2])
				}
				if dp.Gpio2&(1<<25) != 0 {
					add4(&img.Pix[j51Off+0])
				}
				if dp.Gpio2&(1<<16) != 0 {
					add4(&img.Pix[j61Off+0])
				}
				if dp.Gpio2&(1<<15) != 0 {
					add4(&img.Pix[j61Off+1])
				}
				if dp.Gpio2&(1<<14) != 0 {
					add4(&img.Pix[j61Off+2])
				}
				if dp.Gpio2&(1<<13) != 0 {
					add4(&img.Pix[j62Off+0])
				}
				if dp.Gpio2&(1<<10) != 0 {
					add4(&img.Pix[j62Off+1])
				}
				if dp.Gpio2&(1<<12) != 0 {
					add4(&img.Pix[j62Off+2])
				}
				if dp.Gpio2&(1<<6) != 0 {
					add4(&img.Pix[j72Off+0])
				}
				if dp.Gpio3&(1<<18) != 0 {
					add4(&img.Pix[j72Off+1])
				}
				if dp.Gpio2&(1<<7) != 0 {
					add4(&img.Pix[j72Off+2])
				}
				if dp.Gpio2&(1<<11) != 0 {
					add4(&img.Pix[j71Off+0])
				}
				if dp.Gpio2&(1<<9) != 0 {
					add4(&img.Pix[j71Off+1])
				}
				if dp.Gpio2&(1<<8) != 0 {
					add4(&img.Pix[j71Off+2])
				}
				if dp.Gpio0&(1<<14) != 0 {
					add4(&img.Pix[j82Off+1])
				}
				if dp.Gpio3&(1<<20) != 0 {
					add4(&img.Pix[j82Off+2])
				}
				if dp.Gpio3&(1<<17) != 0 {
					add4(&img.Pix[j81Off+0])
				}
				if dp.Gpio3&(1<<16) != 0 {
					add4(&img.Pix[j81Off+1])
				}
				if dp.Gpio3&(1<<15) != 0 {
					add4(&img.Pix[j81Off+2])
				}
				if dp.Gpio3&(1<<14) != 0 {
					add4(&img.Pix[j82Off+0])
				}
			}
		}
	}
}

func main() {
	if err := Main(); err != nil {
		log.Println("error:", err)
	}
}
