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
	"fmt"
	"log"
	"os"
	"time"

	// Note: from when I borrowed Tracy's APC Mini controller
	// xl "github.com/jmacd/nerve/pru/apc/mini"

	"github.com/jmacd/launchmidi/launchctl/xl"
	"github.com/jmacd/nerve/pru/artnet"
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

	var err error

	buf := gpixio.NewBuffer()
	state, err := newAppState(buf)
	if err != nil {
		return err
	}

	recvFrom := os.Getenv("ARTNET_RECVFROM")

	if recvFrom != "" {
		recv, err := artnet.NewReceiver(recvFrom, buf.RGBA)
		if err != nil {
			return err
		}
		ctx := context.Background()
		if err = recv.Start(ctx); err != nil {
			return err
		}

		go func() {
			for {
				recv.Draw()

				bank := state.waitReady()

				const gamma = 2.2 // @@@ not sure lol

				buf.Copy0(gamma, &state.frames[bank])

				state.finish(bank)
				// yawn
				time.Sleep(time.Second / 30)
			}
		}()

	} else {
		input, err = xl.Open()
		if err != nil || input == nil {
			return fmt.Errorf("error while opening connection to launchctl: %w", err)
		}
		defer input.Close()

		go func() {
			err := input.Run(context.Background())
			if err != nil {
				log.Println("LX control run:", err)
			}
			log.Println("LX control exit")
		}()

		player := player.New(input)

		go func() {
			for {
				player.Draw(buf.RGBA)

				bank := state.waitReady()

				buf.Copy0(1+2*player.Data.KnobsRow3[7].Float(), &state.frames[bank])

				state.finish(bank)
			}
		}()
	}

	return state.run()
}

func main() {
	if err := Main(); err != nil {
		log.Println("error:", err)
	}
}
