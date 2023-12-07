package gradient

import (
	"image"

	"github.com/jmacd/nerve/pru/program/data"
	"github.com/lucasb-eyer/go-colorful"
)

type Gradient struct {
}

func New() *Gradient {
	return &Gradient{}
}

func (c *Gradient) Draw(data *data.Data, img *image.RGBA) {
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {

			col := colorful.HSLuv(data.Sliders[0].Float()*360, (float64(x))/63, (float64(y))/63)
			img.Set(x, y, col)
		}
	}
}
