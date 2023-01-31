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
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/fogleman/gg"
	"github.com/jmacd/launchmidi/launchctl/xl"
	"github.com/jmacd/nerve/pru/gpixio"
)

var (
	noInput = flag.Bool("no_input", false, "do not use MIDI input device")
)

type (
	Frameset    = gpixio.Frameset
	Schedule    = gpixio.Schedule
	Frame       = gpixio.Frame
	DoubleRow   = gpixio.DoubleRow
	DoublePixel = gpixio.DoublePixel
)

func Main() error {
	var input *xl.LaunchControl

	flag.Parse()

	if !*noInput {
		var err error
		input, err = xl.Open()
		if err != nil {
			return fmt.Errorf("error while opening connection to launchctl: %w", err)
		}
		defer input.Close()
	}

	buf := gpixio.NewBuffer()
	state, err := newAppState(buf)
	if err != nil {
		return err
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

		for {
			ggctx.DrawRectangle(0, 0, 128, 128)
			ggctx.SetRGB(0.8, 0.8, 0)
			ggctx.Fill()

			ggctx.DrawCircle(64, 64, 60)
			ggctx.SetRGB(r, g, b)
			ggctx.Fill()

			bank := state.waitReady()

			t := time.Now()
			for s := 0; s < 4; s++ {
				buf.Copy(&state.frames[bank][s])

				state.test(&state.frames[bank][s])
			}
			a := time.Now()
			fmt.Println("render in", a.Sub(t))
			state.finish(bank)
		}
	}()

	return state.run()
}

func add4(bp *byte) {
	(*bp) += 4 // 4 = 2 bits shifted (64 frames => 256 levels)

	// overflow indicates max brightness, b/c we counted 4 per up to 64 frames
	if *bp == 0 {
		*bp = 255
	}
}
func main() {
	if err := Main(); err != nil {
		log.Println("error:", err)
	}
}
