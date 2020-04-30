package program

import (
	"github.com/jmacd/launchmidi/launchctl/xl"
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

	Pattern struct {
		Controls
		Buffer

		Feature int
	}

	Parameter struct {
		control xl.Control
		color   xl.Color
	}

	Player interface {
		Controller() *xl.LaunchControl
	}

	Runner interface {
		Apply(Player)
		Draw(Player)
		SetFeature(int) // 0-7
		CopyTo(*Buffer)
	}
)

func New(width, height int) Pattern {
	return Pattern{
		Buffer: Buffer{
			Width:  width,
			Height: height,
			Pixels: make([]Color, width*height),
		},
	}
}

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

func (p *Pattern) SetAllColor(c Color) {
	for i := range p.Buffer.Pixels {
		p.Buffer.Pixels[i] = c
	}
}

func (p *Pattern) CopyTo(buffer *Buffer) {
	*buffer = p.Buffer
	buffer.Pixels = make([]Color, len(p.Buffer.Pixels))
	copy(buffer.Pixels, p.Buffer.Pixels)
}

func (p *Pattern) SetFeature(feature int) {
	p.Feature = feature
}
