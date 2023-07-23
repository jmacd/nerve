package player

import (
	"image"
	"sync"

	//  "github.com/jmacd/launchmidi/launchctl/xl"
	xl "github.com/jmacd/nerve/pru/apc/mini"
	"github.com/jmacd/nerve/pru/program/circle"
	"github.com/jmacd/nerve/pru/program/data"
	"github.com/jmacd/nerve/pru/program/fractal"
	"github.com/jmacd/nerve/pru/program/panelnum"
	"github.com/jmacd/nerve/pru/program/panes"
	"github.com/jmacd/nerve/pru/program/player/input"
)

type Program interface {
	Draw(*data.Data, *image.RGBA)
}

type Player struct {
	inp  input.Input
	lock sync.Mutex

	playing  int
	programs [16]Program

	data.Data
}

func (p *Player) withLock(trigger input.Control, callback func(control input.Control, value input.Value)) {
	p.inp.AddCallback(0, trigger, func(_ int, actual input.Control, value input.Value) {
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

func New(inp input.Input) *Player {
	p := &Player{
		inp: inp,
	}

	for i := range p.programs {
		p.programs[i] = newEmptyProgram()
	}

	p.programs[0] = fractal.New()
	p.programs[1] = panes.New()
	p.programs[6] = circle.New()
	p.programs[7] = panelnum.New()

	p.Data.Init()

	inp.SetColor(0, input.Control(xl.ControlButtonTrackFocus[0]), input.Color(xl.ColorBrightRed))

	p.withLock(input.Control(xl.ControlSlider[8]), func(control input.Control, value input.Value) {
		p.Data.Slider9 = value
	})

	for i := 0; i < 8; i++ {
		i := i
		// p.withLock(input.Control(xl.ControlKnobSendA[i]), func(control input.Control, value input.Value) {
		// 	p.Data.KnobsRow1[i] = value
		// })
		// p.withLock(input.Control(xl.ControlKnobSendB[i]), func(control input.Control, value input.Value) {
		// 	p.Data.KnobsRow2[i] = value
		// })
		// p.withLock(input.Control(xl.ControlKnobPanDevice[i]), func(control input.Control, value input.Value) {
		// 	p.Data.KnobsRow3[i] = value
		// })
		p.withLock(input.Control(xl.ControlSlider[i]), func(control input.Control, value input.Value) {
			p.Data.Sliders[i] = value
			p.Data.KnobsRow1[i] = value
			p.Data.KnobsRow2[i] = value
			p.Data.KnobsRow3[i] = value

		})
		p.withLock(input.Control(xl.ControlButtonTrackFocus[i]), func(control input.Control, value input.Value) {
			if value == 0 {
				return
			}
			if p.Data.ButtonsRadio == i {
				return
			}
			inp.SetColor(0, input.Control(xl.ControlButtonTrackFocus[p.Data.ButtonsRadio]), 0)
			inp.SetColor(0, control, input.Color(xl.ColorBrightRed))
			p.Data.ButtonsRadio = int(control - input.Control(xl.ControlButtonTrackFocus[0]))
		})
		p.withLock(input.Control(xl.ControlButtonTrackControl[i]), func(control input.Control, value input.Value) {
			if value == 0 {
				return
			}
			if p.Data.ButtonsToggle[i] {
				p.Data.ButtonsToggle[i] = false
				inp.SetColor(0, input.Control(xl.ControlButtonTrackControl[i]), 0)
			} else {
				p.Data.ButtonsToggle[i] = true
				inp.SetColor(0, input.Control(xl.ControlButtonTrackControl[i]), input.Color(xl.ColorBrightYellow))
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
