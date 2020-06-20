package main

import (
	"context"
	"log"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	flag "github.com/spf13/pflag"

	"github.com/jmacd/launchmidi/launchctl/xl"
	"github.com/jmacd/nerve/artnet"
	"github.com/jmacd/nerve/program"
	"github.com/jmacd/nerve/program/colors"
	"github.com/jmacd/nerve/program/strobe"
	"github.com/jmacd/nerve/program/tilesnake"
	"github.com/jmacd/nerve/video"
	"github.com/lucasb-eyer/go-colorful"

	"github.com/faiface/pixel/pixelgl"
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
	width  = 20
	height = 15
	pixels = width * height
)

type (
	Color = colorful.Color

	Buffer struct {
		Pixels []Color
	}
)

var (
	senderMode *string = flag.String("mode", "artnet", "e.g., argnet,video")
	ipAddr     *string = flag.String("artnetip", "192.168.0.25", "artnet IP address")
)

func main() {
	flag.Parse()

	pixelgl.Run(play)
}

func play() {
	lc, err := xl.Open()
	if err != nil {
		log.Fatalf("error while opening connection to launchctl: %v", err)
	}
	defer lc.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	grp, ctx := errgroup.WithContext(ctx)

	var sender Sender

	switch {
	case strings.EqualFold("artnet", *senderMode):
		sender = artnet.NewSender(*ipAddr)
	case strings.EqualFold("video", *senderMode):
		sender = video.New()
	default:
		log.Fatalln("could not configure a sender:", *senderMode)
	}

	grp.Go(func() error {
		err := sender.Run(ctx)
		if err != nil {
			log.Println("sender:", err)
		}
		return err
	})

	grp.Go(func() error {
		err := lc.Run(ctx)
		if err != nil {
			log.Println("launch control XL:", err)
		}
		return err
	})

	grp.Go(func() error {
		bp := newPlayProgram(sender, lc)
		err := bp.Run(ctx)
		if err != nil {
			log.Println("play program", err)
		}
		return err
	})

	log.Println("play group:", grp.Wait())
}

type Sender interface {
	Run(context.Context) error
	Input() chan []Color
}

type PlayProgram struct {
	lock    sync.Mutex
	sender  Sender // e.g., *artnet.Sender
	lc      *xl.LaunchControl
	current int // Top button last pressed, [0-7].

	programs [8]program.Runner
}

func newPlayProgram(sender Sender, lc *xl.LaunchControl) *PlayProgram {
	// The set of patterns
	snake := tilesnake.New(width, height)
	strobe := strobe.New(width, height)
	colors := colors.New(width, height)

	sender.Input() <- make([]Color, pixels)

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

func (bp *PlayProgram) Run(ctx context.Context) error {
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
			return ctx.Err()
		default:
		}

		bp.programs[bp.current].Draw(bp)
		bp.programs[bp.current].CopyTo(buffer)
		bp.sender.Input() <- buffer.Pixels[:]
	}
	return nil
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

func (bp *PlayProgram) Sender() Sender {
	return bp.sender
}
