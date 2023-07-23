package data

import (
	"math/rand"

	"github.com/jmacd/nerve/pru/program/player/input"
)

var rnd = rand.New(rand.NewSource(1333))

func randValue() input.Value {
	return input.Value(rnd.Intn(256))
}

type Data struct {
	Sliders       [8]input.Value
	KnobsRow1     [8]input.Value
	KnobsRow2     [8]input.Value
	KnobsRow3     [8]input.Value
	Slider9       input.Value
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
