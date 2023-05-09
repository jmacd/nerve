package player

import (
	"image"
	"sync"

	"github.com/jmacd/launchmidi/launchctl/xl"
	"github.com/jmacd/nerve/pru/program/data"
	"github.com/jmacd/nerve/pru/program/fractal"
)

type Program interface {
	Draw(*data.Data, *image.RGBA)
}

type Player struct {
	input *xl.LaunchControl
	lock  sync.Mutex

	playing  int
	programs [16]Program

	data.Data
}

func (p *Player) withLock(trigger xl.Control, callback func(control xl.Control, value xl.Value)) {
	p.input.AddCallback(xl.AllChannels, trigger, func(_ int, actual xl.Control, value xl.Value) {
		p.lock.Lock()
		defer p.lock.Unlock()
		callback(actual, value)
	})
}

type emptyProgram struct{}

var _ Program = &emptyProgram{}

func newEmptyProgram() Program {
	return &emptyProgram{}
}

func New(input *xl.LaunchControl) *Player {
	p := &Player{
		input: input,
	}

	for i := range p.programs {
		p.programs[i] = newEmptyProgram()
	}

	p.programs[p.playing] = fractal.New()

	for i := 0; i < 0; i++ {
		p.Data.Init(rnd)

		p.withLock(xl.ControlKnobSendA[i], func(control xl.Control, value xl.Value) {
			p.knobsRow1[i] = value
		})
		p.withLock(xl.ControlKnobSendB[i], func(control xl.Control, value xl.Value) {
			p.knobsRow2[i] = value
		})
		p.withLock(xl.ControlKnobPanDevice[i], func(control xl.Control, value xl.Value) {
			p.knobsRow3[i] = value
		})
		p.withLock(xl.ControlSlider[i], func(control xl.Control, value xl.Value) {
			p.sliders[i] = value
		})
	}

	// p.pat = int(value)
	// p.frac = nil

	return p
}

func (p *Player) Draw(pix *image.RGBA) {
	p.lock.Lock()
	pat := p.pat % len(fractal.Seeds)
	frac := p.frac
	r := p.r
	g := p.g
	b := p.b
	p.lock.Unlock()

	if frac == nil {
		frac = fractal.New(fractal.Seeds[pat])

		p.lock.Lock()
		p.frac = frac
		p.lock.Unlock()
	}

	p.frac.Draw(pix, r, g, b)
}
