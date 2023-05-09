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

	"github.com/fogleman/gg"
	"github.com/jmacd/launchmidi/launchctl/xl"
	"github.com/jmacd/nerve/pru/gpixio"
	"github.com/jmacd/nerve/pru/program/player"
)

type (
	Frameset    = gpixio.Frameset
	Frame       = gpixio.Frame
	DoubleRow   = gpixio.DoubleRow
	DoublePixel = gpixio.DoublePixel
)

func Main() error {
	var input *xl.LaunchControl

	flag.Parse()

	var err error
	input, err = xl.Open()
	if err != nil || input == nil {
		return fmt.Errorf("error while opening connection to launchctl: %w", err)
	}
	defer input.Close()

	buf := gpixio.NewBuffer()
	state, err := newAppState(buf)
	if err != nil {
		return err
	}

	player := player.New(input)

	go func() {
		err := input.Run(context.Background())
		if err != nil {
			log.Println("LX control run:", err)
		}
		log.Println("LX control exit")
	}()

	go func() {
		for {
			player.Draw(buf.RGBA)

			bank := state.waitReady()

			//t := time.Now()
			buf.Copy0(&state.frames[bank])

			//state.test(&state.frames[bank])
			// a := time.Now()
			// fmt.Println("render in", a.Sub(t))
			state.finish(bank)
		}
	}()

	return state.run()
}

func save() {
	// ggctx.DrawRectangle(0, 0, 128, 128)
	// ggctx.SetRGB(0.8, 0.8, 0)
	// ggctx.Fill()

	// ggctx.DrawCircle(64, 64, 60)
	// ggctx.SetRGB(r, g, b)
	// ggctx.Fill()
	// Save
	_ = gg.NewContextForRGBA(nil)

}

// func imgEntropy(img *image.RGBA) string {
// 	var bins [3][256]uint16
// 	for y := 0; y < 128; y++ {
// 		for x := 0; x < 128; x++ {
// 			px := img.RGBAAt(x, y)
// 			bins[0][px.R]++
// 			bins[1][px.G]++
// 			bins[2][px.B]++
// 		}
// 	}
// 	const total = 128 * 128

// 	var ents [3]float64
// 	for idx, hist := range bins {
// 		var ps []float64
// 		for _, cnt := range hist {
// 			ps = append(ps, float64(cnt)/total)
// 		}
// 		ents[idx] = stat.Entropy(ps)
// 	}
// 	return fmt.Sprint("entropy", ents[0], ents[1], ents[2])
// }

func main() {
	if err := Main(); err != nil {
		log.Println("error:", err)
	}
}
