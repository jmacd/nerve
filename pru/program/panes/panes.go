package panes

import (
	"image"

	"github.com/fogleman/gg"
	"github.com/jmacd/launchmidi/midi/controller"
	"github.com/jmacd/nerve/pru/program/data"
	"github.com/lucasb-eyer/go-colorful"
)

type Panes struct {
}

func New() *Panes {
	return &Panes{}
}

func (p *Panes) Inputs() []controller.Control {
	return data.StandardControls
}

func (c *Panes) setColor(ggctx *gg.Context, data *data.Data, x, y, z float64) {
	var color colorful.Color
	switch {
	case data.ButtonsToggle[0]:
		color = colorful.Hsv(360*x, y, z)
	case data.ButtonsToggle[1]:
		color = colorful.Lab(x, y, z)
	case data.ButtonsToggle[2]:
		color = colorful.Luv(x, y*2-1, z*2-1)
	case data.ButtonsToggle[3]:
		color = colorful.LuvLCh(x*360, y*2-1, z*2-1)
	case data.ButtonsToggle[4]:
		color = colorful.Xyz(x, y, z)
	case data.ButtonsToggle[5]:
		color = colorful.Xyy(x, y, z)
	case data.ButtonsToggle[6]:
		color = colorful.LabWhiteRef(x, y, z, [3]float64{data.KnobsRow1[0].Float(), data.KnobsRow1[1].Float(), data.KnobsRow1[2].Float()})
	case data.ButtonsToggle[7]:
		color = colorful.LuvWhiteRef(x, y*2-1, z*2-1, [3]float64{data.KnobsRow1[0].Float(), data.KnobsRow1[1].Float(), data.KnobsRow1[2].Float()})
	}
	if color.R != 0 || color.G != 0 || color.B != 0 {
		x, y, z = color.R, color.G, color.B
	}

	ggctx.SetRGB(x, y, z)
}

func (c *Panes) Draw(data *data.Data, img *image.RGBA) {
	ggctx := gg.NewContextForRGBA(img)

	ggctx.DrawRectangle(0, 0, 64, 32)
	c.setColor(
		ggctx,
		data,
		data.Sliders[0].Float(),
		data.Sliders[1].Float(),
		data.Sliders[2].Float(),
	)
	ggctx.Fill()

	ggctx.DrawRectangle(0, 32, 64, 32)
	c.setColor(
		ggctx,
		data,
		data.Sliders[3].Float(),
		data.Sliders[4].Float(),
		data.Sliders[5].Float(),
	)
	ggctx.Fill()
}
