package circle

import (
	"image"

	"github.com/fogleman/gg"
	"github.com/jmacd/launchmidi/midi/controller"
	"github.com/jmacd/nerve/pru/program/data"
)

type Circle struct {
}

func New() *Circle {
	return &Circle{}
}

func (c *Circle) Inputs() []controller.Control {
	return data.StandardControls
}

func (c *Circle) Draw(data *data.Data, img *image.RGBA) {
	ggctx := gg.NewContextForRGBA(img)
	ggctx.DrawRectangle(0, 0, 128, 128)
	ggctx.SetRGB(
		data.Sliders[3].Float(),
		data.Sliders[4].Float(),
		data.Sliders[5].Float(),
	)
	ggctx.Fill()

	ggctx.DrawCircle(
		float64(data.KnobsRow1[0]),
		float64(data.KnobsRow1[1]),
		float64(data.KnobsRow1[2]),
	)
	ggctx.SetRGB(
		data.Sliders[0].Float(),
		data.Sliders[1].Float(),
		data.Sliders[2].Float(),
	)
	ggctx.Fill()
}
