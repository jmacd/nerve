package strobe

import (
	"github.com/jmacd/launchmidi/launchctl/xl"
	"github.com/jmacd/nerve/program"
)

type (
	Strobe struct {
		program.Pattern

		frame int64
	}
)

func New(width, height int) *Strobe {
	return &Strobe{
		Pattern: program.New(width, height),
	}
}

func (s *Strobe) Draw(player program.Player) {
	lc := player.Controller()
	adj := lc.Get(xl.ControlKnobSendA[0])
	s.frame++
	var c program.Color
	if s.frame%2 == 0 {
		c = program.Color{
			R: lc.Get(xl.ControlSlider[3]) * adj,
			G: lc.Get(xl.ControlSlider[4]) * adj,
			B: lc.Get(xl.ControlSlider[5]) * adj,
		}
	} else {
		c = program.Color{
			R: lc.Get(xl.ControlSlider[0]) * adj,
			G: lc.Get(xl.ControlSlider[1]) * adj,
			B: lc.Get(xl.ControlSlider[2]) * adj,
		}
	}

	s.Pattern.SetAllColor(c)
}
