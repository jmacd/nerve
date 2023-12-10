package player

import (
	"image"
	"sync"

	"github.com/jmacd/launchmidi/launchctl/xl"
	"github.com/jmacd/launchmidi/midi/controller"
	"github.com/jmacd/nerve/pru/program/circle"
	"github.com/jmacd/nerve/pru/program/data"
	"github.com/jmacd/nerve/pru/program/fractal"
	"github.com/jmacd/nerve/pru/program/gradient"
	"github.com/jmacd/nerve/pru/program/openmic"
	"github.com/jmacd/nerve/pru/program/panes"
)

type Program interface {
	Draw(*data.Data, *image.RGBA)
}

type Player struct {
	inp  controller.Input
	lock sync.Mutex

	programs [16]Program

	data.Data
}

func (p *Player) withLock(trigger controller.Control, callback func(control controller.Control, value controller.Value)) {
	p.inp.AddCallback(0, trigger, func(_ int, actual controller.Control, value controller.Value) {
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

func New(inp controller.Input) *Player {
	p := &Player{
		inp: inp,
	}

	for i := range p.programs {
		p.programs[i] = newEmptyProgram()
	}

	p.programs[0] = openmic.New(data.WelcomeText, false)
	p.programs[1] = fractal.New()
	p.programs[2] = panes.New()
	p.programs[3] = openmic.New(data.Manifesto, false)
	p.programs[4] = openmic.New(data.Technical, false)
	p.programs[5] = gradient.New()
	p.programs[6] = circle.New()
	p.programs[7] = openmic.New("", true)

	p.Data.Init()

	inp.SetColor(0, controller.Control(xl.ControlButtonTrackFocus[0]), controller.Color(xl.ColorBrightRed))

	for i := 0; i < 8; i++ {
		i := i
		p.withLock(controller.Control(xl.ControlKnobSendA[i]), func(control controller.Control, value controller.Value) {
			p.Data.KnobsRow1[i] = value
		})
		p.withLock(controller.Control(xl.ControlKnobSendB[i]), func(control controller.Control, value controller.Value) {
			p.Data.KnobsRow2[i] = value
		})
		p.withLock(controller.Control(xl.ControlKnobPanDevice[i]), func(control controller.Control, value controller.Value) {
			p.Data.KnobsRow3[i] = value
		})
		p.withLock(controller.Control(xl.ControlSlider[i]), func(control controller.Control, value controller.Value) {
			p.Data.Sliders[i] = value
		})
		p.withLock(controller.Control(xl.ControlButtonTrackFocus[i]), func(control controller.Control, value controller.Value) {
			if value == 0 {
				return
			}
			if p.Data.ButtonsRadio == i {
				return
			}
			inp.SetColor(0, controller.Control(xl.ControlButtonTrackFocus[p.Data.ButtonsRadio]), 0)
			inp.SetColor(0, control, controller.Color(xl.ColorBrightRed))
			p.Data.ButtonsRadio = int(control - controller.Control(xl.ControlButtonTrackFocus[0]))
		})
		p.withLock(controller.Control(xl.ControlButtonTrackControl[i]), func(control controller.Control, value controller.Value) {
			if value == 0 {
				return
			}
			if p.Data.ButtonsToggle[i] {
				p.Data.ButtonsToggle[i] = false
				inp.SetColor(0, controller.Control(xl.ControlButtonTrackControl[i]), 0)
			} else {
				p.Data.ButtonsToggle[i] = true
				inp.SetColor(0, controller.Control(xl.ControlButtonTrackControl[i]), controller.Color(xl.ColorBrightYellow))
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
