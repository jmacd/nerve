package player

import (
	"image"
	"sync"

	"github.com/jmacd/launchmidi/launchctl/xl"
	"github.com/jmacd/nerve/pru/program/circle"
	"github.com/jmacd/nerve/pru/program/data"
	"github.com/jmacd/nerve/pru/program/fractal"
	"github.com/jmacd/nerve/pru/program/panelnum"
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

func (e *emptyProgram) Draw(*data.Data, *image.RGBA) {
}

func New(input *xl.LaunchControl) *Player {
	p := &Player{
		input: input,
	}

	for i := range p.programs {
		p.programs[i] = newEmptyProgram()
	}

	p.programs[0] = fractal.New()
	p.programs[6] = circle.New()
	p.programs[7] = panelnum.New()

	p.Data.Init()

	input.SetColor(0, xl.ControlButtonTrackFocus[0], xl.ColorBrightRed)

	for i := 0; i < 8; i++ {
		i := i
		p.withLock(xl.ControlKnobSendA[i], func(control xl.Control, value xl.Value) {
			p.Data.KnobsRow1[i] = value
		})
		p.withLock(xl.ControlKnobSendB[i], func(control xl.Control, value xl.Value) {
			p.Data.KnobsRow2[i] = value
		})
		p.withLock(xl.ControlKnobPanDevice[i], func(control xl.Control, value xl.Value) {
			p.Data.KnobsRow3[i] = value
		})
		p.withLock(xl.ControlSlider[i], func(control xl.Control, value xl.Value) {
			p.Data.Sliders[i] = value
		})
		p.withLock(xl.ControlButtonTrackFocus[i], func(control xl.Control, value xl.Value) {
			if value == 0 {
				return
			}
			if p.Data.ButtonsRadio == i {
				return
			}
			input.SetColor(0, xl.ControlButtonTrackFocus[p.Data.ButtonsRadio], 0)
			input.SetColor(0, control, xl.ColorBrightRed)
			p.Data.ButtonsRadio = int(control - xl.ControlButtonTrackFocus[0])
		})
		p.withLock(xl.ControlButtonTrackControl[i], func(control xl.Control, value xl.Value) {
			if value == 0 {
				return
			}
			if p.Data.ButtonsToggle[i] {
				p.Data.ButtonsToggle[i] = false
				input.SetColor(0, xl.ControlButtonTrackControl[i], 0)
			} else {
				p.Data.ButtonsToggle[i] = true
				input.SetColor(0, xl.ControlButtonTrackControl[i], xl.ColorBrightYellow)
			}
		})
	}

	return p
}

func (p *Player) Draw(img *image.RGBA) {
	p.lock.Lock()
	data := p.Data
	p.lock.Unlock()

	p.programs[data.ButtonsRadio].Draw(&data, img)
}
