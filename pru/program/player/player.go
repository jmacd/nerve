package player

import (
	"image"
	"sync"
	"time"

	"github.com/jmacd/launchmidi/launchctl/xl"
	"github.com/jmacd/launchmidi/midi/controller"
	"github.com/jmacd/nerve/pru/program/circle"
	"github.com/jmacd/nerve/pru/program/data"
	"github.com/jmacd/nerve/pru/program/fractal"
	"github.com/jmacd/nerve/pru/program/gradient"
	"github.com/jmacd/nerve/pru/program/openmic"
	"github.com/jmacd/nerve/pru/program/panes"
)

const fastPressLimit = 500 * time.Millisecond

type Program interface {
	Draw(*data.Data, *image.RGBA)
	Inputs() []controller.Control
}

type Player struct {
	inp  controller.Input
	lock sync.Mutex

	pnum      int
	pvar      int
	programs  [8]Program
	current   data.Data
	shadows   [8]data.Data
	lastpress time.Time
	active    map[controller.Control]bool
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

func (e *emptyProgram) Inputs() []controller.Control {
	return nil
}

func New(inp controller.Input) *Player {
	p := &Player{
		inp:    inp,
		active: map[controller.Control]bool{},
	}

	for i := range p.programs {
		p.programs[i] = newEmptyProgram()
		p.shadows[i].Init()
	}

	p.programs[0] = openmic.New(data.WelcomeText, false)
	p.programs[1] = fractal.New()
	p.programs[2] = panes.New()
	p.programs[3] = openmic.New(data.Manifesto, false)
	p.programs[4] = openmic.New(data.Technical, false)
	p.programs[5] = gradient.New()
	p.programs[6] = circle.New()
	p.programs[7] = openmic.New("", true)

	p.pnum = 1
	p.setProgram(0)

	for i := 0; i < 8; i++ {
		i := i
		p.withLock(controller.Control(xl.ControlKnobSendA[i]), func(control controller.Control, value controller.Value) {
			p.press()
			p.current.KnobsRow1[i] = value
			p.shadows[p.pnum].KnobsRow1[i] = value
		})
		p.withLock(controller.Control(xl.ControlKnobSendB[i]), func(control controller.Control, value controller.Value) {
			p.press()
			p.current.KnobsRow2[i] = value
			p.shadows[p.pnum].KnobsRow2[i] = value
		})
		p.withLock(controller.Control(xl.ControlKnobPanDevice[i]), func(control controller.Control, value controller.Value) {
			p.press()
			p.current.KnobsRow3[i] = value
			p.shadows[p.pnum].KnobsRow3[i] = value
		})
		p.withLock(controller.Control(xl.ControlSlider[i]), func(control controller.Control, value controller.Value) {
			p.press()
			p.current.Sliders[i] = value
			p.shadows[p.pnum].Sliders[i] = value
		})
		p.withLock(controller.Control(xl.ControlButtonTrackFocus[i]), func(control controller.Control, value controller.Value) {
			if value == 0 {
				// Ignore button-up
				return
			}
			p.press()
			num := int(control - controller.Control(xl.ControlButtonTrackFocus[0]))
			p.setProgram(num)
		})
		p.withLock(controller.Control(xl.ControlButtonTrackControl[i]), func(control controller.Control, value controller.Value) {

			if value == 0 {
				// Ignore button-up
				return
			}
			if !p.press() {
				p.current.ButtonsToggle[i] = !p.current.ButtonsToggle[i]
			}
			p.current.ButtonsToggleMod4[i]++
			p.current.ButtonsToggleMod4[i] %= 4

			p.shadows[p.pnum].ButtonsToggle[i] = p.current.ButtonsToggle[i]
			p.shadows[p.pnum].ButtonsToggleMod4[i] = p.current.ButtonsToggleMod4[i]

			inp.SetColor(0, controller.Control(xl.ControlButtonTrackControl[i]), p.buttonColorsFor(i))
		})
	}
	return p
}

// Data returns current data
func (p *Player) Data() data.Data {
	p.lock.Lock()
	defer p.lock.Unlock()
	return p.shadows[p.pnum]
}

func (p *Player) Draw(img *image.RGBA) {
	p.lock.Lock()
	dat := p.shadows[p.pnum]
	num := p.pnum
	p.lock.Unlock()

	p.programs[num].Draw(&dat, img)
}

func (p *Player) setProgram(num int) {
	p.inp.SetColor(0, xl.ControlButtonTrackFocus[num], controller.Color(p.chooseProgramColor()))
	if p.pnum == num {
		return
	}
	for _, act := range p.programs[p.pnum].Inputs() {
		if p.active[act] {
			p.inp.SetColor(0, act, 0)
		}
	}

	p.current = p.shadows[num]
	p.inp.SetColor(0, controller.Control(xl.ControlButtonTrackFocus[p.pnum]), 0)
	p.pnum = num

	for i := 0; i < 8; i++ {
		p.inp.SetColor(0, controller.Control(xl.ControlButtonTrackControl[i]), p.buttonColorsFor(i))
	}

	p.active = map[controller.Control]bool{}
	for _, act := range p.programs[p.pnum].Inputs() {
		p.active[act] = true
		p.inp.SetColor(0, act, xl.FourBrightColors[act%4])
	}
}

func (p *Player) chooseProgramColor() (col xl.Color) {
	col = xl.FourBrightColors[p.pvar%4]
	p.pvar++
	return
}

// press returns true if the press was fast
func (p *Player) press() bool {
	now := time.Now()
	last := p.lastpress
	p.lastpress = now
	return now.Sub(last) < fastPressLimit
}

func (p *Player) buttonColorsFor(n int) xl.Color {
	if !p.current.ButtonsToggle[n] {
		return 0
	}
	return xl.FourBrightColors[p.current.ButtonsToggleMod4[n]]
}
