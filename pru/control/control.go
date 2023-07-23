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

	//"github.com/jmacd/launchmidi/launchctl/xl"
	xl "github.com/jmacd/nerve/pru/apc/mini"
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

			buf.Copy0(1+2*player.Data.Slider9.Float(), &state.frames[bank])

			state.finish(bank)
		}
	}()

	return state.run()
}

func main() {
	if err := Main(); err != nil {
		log.Println("error:", err)
	}
}
