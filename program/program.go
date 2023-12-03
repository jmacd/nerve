package program

import (
	"github.com/jmacd/launchmidi/launchctl/xl"
	"github.com/jmacd/nerve/pru/artnet"
	"github.com/lucasb-eyer/go-colorful"
)

type (
	Color = colorful.Color

	Buffer struct {
		Width  int
		Height int
		Pixels []Color
	}

	Controls struct {
		params []Parameter
	}

	Parameter struct {
		control xl.Control
		color   xl.Color
	}

	Player interface {
		Sender() *artnet.Sender
		Controller() *xl.LaunchControl
	}

	Runner interface {
		Apply(Player)
		Draw(Player) error
	}
)

func (p *Controls) AddParameter(control xl.Control, color xl.Color) {
	p.params = append(p.params, Parameter{
		control: control,
		color:   color,
	})
}

func (p *Controls) Apply(player Player) {
	lc := player.Controller()
	lc.Clear(0)
	for _, p := range p.params {
		lc.SetColor(0, p.control, xl.FlashUnknown(p.color))
	}
	lc.SwapBuffers(0)
}
