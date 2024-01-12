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
	return controller.Value(14 + rnd.Intn(100))
}

type Data struct {
	Sliders           [8]controller.Value
	KnobsRow1         [8]controller.Value
	KnobsRow2         [8]controller.Value
	KnobsRow3         [8]controller.Value
	ButtonsToggle     [8]bool
	ButtonsToggleMod4 [8]int // 0-3

	// Program number is internal to player.go
	// ButtonsRadio      int // 0-7
}

func (d *Data) Init() {
	for i := 0; i < 8; i++ {
		d.Sliders[i] = randValue()
		d.KnobsRow1[i] = randValue()
		d.KnobsRow2[i] = randValue()
		d.KnobsRow3[i] = randValue()
		d.ButtonsToggle[i] = rnd.Intn(2) == 0
		d.ButtonsToggleMod4[i] = rnd.Intn(4)
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
