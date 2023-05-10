package panelnum

import (
	"fmt"
	"image"

	"github.com/fogleman/gg"
	"github.com/jmacd/nerve/pru/program/data"
)

type PanelNum struct{}

func New() *PanelNum {
	return &PanelNum{}
}

func (c *PanelNum) Draw(data *data.Data, img *image.RGBA) {
	ggctx := gg.NewContextForRGBA(img)

	ggctx.DrawRectangle(0, 0, 128, 128)
	ggctx.SetRGB(.1, .1, .1)
	ggctx.Fill()

	ggctx.SetRGB(1, 1, 1)

	for i := 0; i < 8; i++ {
		x := (i/4)*64 + 32
		y := (i%4)*32 + 16
		ggctx.DrawString(fmt.Sprint(i), float64(x), float64(y))
	}
}
