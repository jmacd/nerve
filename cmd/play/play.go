package main

import (
	"context"
	_ "image"
	"log"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/jmacd/launchmidi/launchctl/xl"
	"github.com/jmacd/nerve/artnet"
	"github.com/jmacd/nerve/program"
	"github.com/jmacd/nerve/program/colors"
	"github.com/jmacd/nerve/program/strobe"
	"github.com/jmacd/nerve/program/tilesnake"
	"github.com/lucasb-eyer/go-colorful"
)

// TODOs
// Buffer mgmt
// pool := &sync.Pool{
// 	New: func() interface{} {
// 		return new(Buffer)
// 	},
// }
// zbuf := pool.Get()
// defer pool.Put(zbuf)
//

const (
	ipAddr = "192.168.0.25"

	width  = 20
	height = 15
	pixels = width * height

	epsilon = 0.00001
)

type (
	Color = colorful.Color

	Buffer struct {
		Pixels []Color
	}
)

func main() {
	sender := artnet.NewSender(ipAddr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	grp, ctx := errgroup.WithContext(ctx)

	grp.Go(func() error {
		err := sender.Run(ctx)
		if err != nil {
			log.Println("artnet sender:", err)
		}
		return err
	})

	l, err := xl.Open()
	if err != nil {
		log.Fatalf("error while openning connection to launchctl: %v", err)
	}
	defer l.Close()

	go func() {
		err := l.Run(ctx)
		if err != nil {
			log.Println("launch control XL:", err)
		}
	}()

	bp := newPlayProgram(sender, l)

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
	// The set of patterns
	snake := tilesnake.New(width, height)
	strobe := strobe.New(width, height)
	colors := colors.New(width, height)

	sender.Send(make([]Color, pixels))

	return &PlayProgram{
		sender:  sender,
		lc:      lc,
		current: 0,
		programs: [...]program.Runner{
			snake,
			strobe,
			colors,
			snake,
			strobe,
			colors,
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
			bp.selectFeature,
		)
		bp.lc.AddCallback(
			0,
			xl.ControlButtonTrackControl[i],
			bp.selectProgram,
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

		// @@@ now := time.Now()

		bp.programs[bp.current].Draw(bp)
		bp.programs[bp.current].CopyTo(buffer)
		bp.sender.Send(buffer.Pixels[:])

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
			xl.ControlButtonTrackControl[i],
			c,
		)
		bp.lc.SetColor(0,
			xl.ControlButtonTrackFocus[i],
			c,
		)
	}
}

func (bp *PlayProgram) selectFeature(_ int, control xl.Control, value xl.Value) {
	bp.lock.Lock()
	defer bp.lock.Unlock()
	defer bp.lc.SwapBuffers(0)

	feature := int(control - xl.ControlButtonTrackFocus[0])

	if value == 0 {
		bp.setButtonColors()
		return
	}

	// Turn the LED off while while held down.
	bp.lc.SetColor(0,
		xl.ControlButtonTrackFocus[feature],
		0,
	)
	if bp.current >= 0 {
		bp.programs[bp.current].SetFeature(feature)
	}
}

func (bp *PlayProgram) selectProgram(_ int, control xl.Control, value xl.Value) {
	bp.lock.Lock()
	defer bp.lock.Unlock()
	defer bp.lc.SwapBuffers(0)

	idx := int(control - xl.ControlButtonTrackControl[0])

	// Turn the LED off while while held down.
	if value == 127 {
		bp.lc.SetColor(0,
			xl.ControlButtonTrackControl[idx],
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
