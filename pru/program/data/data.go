package data

import "github.com/jmacd/launchmidi/launchctl/xl"

var rnd = rand.New(rand.NewSource(1333))

func randValue() xl.Value {
	return xl.Value(rnd.Intn(256))
}

type Data struct {
	Sliders   [8]xl.Value
	KnobsRow1 [8]xl.Value
	KnobsRow2 [8]xl.Value
	KnobsRow3 [8]xl.Value
}

func (d *Data) Init(rnd *rand.Rand) {
	for i := 0; i < 8; i++ {
		d.Sliders[i] = randValue()
		d.KnobsRow1[i] = randValue()
		d.KnobsRow2[i] = randValue()
		d.KnobsRow3[i] = randValue()
	}
}
