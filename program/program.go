package program

import "github.com/jmacd/launchmidi/launchctl/xl"

type (
	Program struct {
		name     string
		lc       *xl.LaunchControl
		controls []Control
	}

	Control struct {
		device xl.Control
	}
)

func NewProgram(name string, lc *xl.LaunchControl) *Program {
	return &Program{
		name: name,
		lc:   lc,
	}
}

func (p *Program) AddControl(control xl.Control, color xl.Color) {
	p.controls = append(p.controls, Control{control})
	p.lc.SetColor(0, control, xl.FlashUnknown(color))
}
