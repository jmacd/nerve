package player

import (
	"image"
	"sync"

	"github.com/jmacd/launchmidi/launchctl/xl"
	"github.com/jmacd/nerve/pru/program/fractal"
)

type Player struct {
	input *xl.LaunchControl
	lock  sync.Mutex
	pat   int
	frac  *fractal.Fractal

	r, g, b float64
}

func (p *Player) withLock(trigger xl.Control, callback func(control xl.Control, value xl.Value)) {
	p.input.AddCallback(xl.AllChannels, trigger, func(_ int, actual xl.Control, value xl.Value) {
		p.lock.Lock()
		defer p.lock.Unlock()
		callback(actual, value)
	})
}

func New(input *xl.LaunchControl) *Player {
	p := &Player{
		input: input,
		pat:   0,
		frac:  nil,
	}

	p.withLock(xl.ControlKnobSendA[0], func(control xl.Control, value xl.Value) {
		p.pat = int(value)
		p.frac = nil
	})
	p.withLock(xl.ControlSlider[0], func(control xl.Control, value xl.Value) {
		p.r = value.Float()
	})
	p.withLock(xl.ControlSlider[1], func(control xl.Control, value xl.Value) {
		p.g = value.Float()
	})
	p.withLock(xl.ControlSlider[1], func(control xl.Control, value xl.Value) {
		p.b = value.Float()
	})

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
