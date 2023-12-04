package data

import (
	"embed"
	"math/rand"

	"github.com/golang/freetype/truetype"
	"github.com/jmacd/launchmidi/midi/controller"
	"golang.org/x/image/font"
)

var rnd = rand.New(rand.NewSource(1333))

func randValue() controller.Value {
	return controller.Value(rnd.Intn(256))
}

type Data struct {
	Sliders   [8]controller.Value
	KnobsRow1 [8]controller.Value
	KnobsRow2 [8]controller.Value
	KnobsRow3 [8]controller.Value
	//Slider9       controller.Value
	ButtonsRadio  int // 0-7
	ButtonsToggle [8]bool
}

func (d *Data) Init() {
	for i := 0; i < 8; i++ {
		d.Sliders[i] = randValue()
		d.KnobsRow1[i] = randValue()
		d.KnobsRow2[i] = randValue()
		d.KnobsRow3[i] = randValue()
	}
}

//go:embed resource
var ResourceFS embed.FS

func LoadFontFace(path string, points float64) (font.Face, error) {
	fontBytes, err := ResourceFS.ReadFile(path)
	if err != nil {
		return nil, err
	}
	f, err := truetype.Parse(fontBytes)
	if err != nil {
		return nil, err
	}
	face := truetype.NewFace(f, &truetype.Options{
		Size: points,
	})
	return face, nil
}
