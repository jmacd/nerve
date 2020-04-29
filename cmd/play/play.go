package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/jmacd/launchmidi/launchctl/xl"
	"github.com/jmacd/nerve/artnet"
	"github.com/jmacd/nerve/program"
	"github.com/jmacd/nerve/program/strobe2"
	"github.com/jmacd/nerve/program/tilesnake"
	"github.com/lucasb-eyer/go-colorful"
)

const (
	ipAddr = "192.168.0.23"

	width  = 20
	height = 15
	pixels = width * height

	epsilon = 0.00001
)

type (
	Color = colorful.Color
)

func main() {
	sender := artnet.NewSender(ipAddr)

	l, err := xl.Open()
	if err != nil {
		log.Fatalf("error while openning connection to launchctl: %v", err)
	}
	defer l.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go l.Run(ctx)

	bp := newPlayProgram(sender, l)

	sender.Send(make([]Color, pixels))

	go bp.Run(ctx)
	select {}
}

type PlayProgram struct {
	lock    sync.Mutex
	sender  *artnet.Sender
	lc      *xl.LaunchControl
	current int // Top button last pressed, [0-7].

	programs [8]program.Runner
}

func newPlayProgram(sender *artnet.Sender, lc *xl.LaunchControl) *PlayProgram {
	// @@@
	snake := tilesnake.New(width, height)
	strobe := strobe2.New(width, height)

	return &PlayProgram{
		sender:  sender,
		lc:      lc,
		current: -1,
		programs: [...]program.Runner{
			snake,
			strobe,
			snake,
			strobe,
			snake,
			strobe,
			snake,
			strobe,
		},
	}
}

func (bp *PlayProgram) Run(ctx context.Context) {
	for i := 0; i < 8; i++ {
		bp.lc.AddCallback(
			0,
			xl.ControlButtonTrackFocus[i],
			bp.topButton,
		)
	}

	bp.setButtonColors()

	buffer := &program.Buffer{}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if bp.current >= 0 {
			bp.programs[bp.current].Draw(bp)
			bp.programs[bp.current].CopyTo(buffer)
			bp.sender.Send(buffer.Pixels[:])
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (bp *PlayProgram) setButtonColors() {
	for i := 0; i < 8; i++ {
		var c xl.Color
		if i == bp.current {
			c = xl.FourBrightColors[i%4]
		} else {
			c = xl.FourDimColors[i%4]
		}
		bp.lc.SetColor(0,
			xl.ControlButtonTrackFocus[i],
			c,
		)
	}
}

func (bp *PlayProgram) topButton(_ int, control xl.Control, value xl.Value) {
	bp.lock.Lock()
	defer bp.lock.Unlock()
	defer bp.lc.SwapBuffers(0)

	idx := int(control - xl.ControlButtonTrackFocus[0])

	// Turn the LED off while while held down.
	if value == 127 {
		bp.lc.SetColor(0,
			xl.ControlButtonTrackFocus[idx],
			0,
		)
		return
	}

	bp.current = idx
	bp.programs[idx].Apply(bp)
	bp.setButtonColors()
}

func (bp *PlayProgram) Controller() *xl.LaunchControl {
	return bp.lc
}

func (bp *PlayProgram) Sender() *artnet.Sender {
	return bp.sender
}
