package openmic

import (
	"bufio"
	"fmt"
	"image"
	"os"
	"sync"

	"github.com/fogleman/gg"
	"github.com/jmacd/nerve/pru/program/data"
)

var text = `this is open mic nite; welcome. glad you came, we have lots to do.
I think it would be nice if we could have a gathering of makers too.
`

type OpenMic struct {
	*gg.Context
	lock sync.Mutex
	text string
}

func New() *OpenMic {
	o := &OpenMic{
		Context: gg.NewContext(128, 128),
		text:    text,
	}
	ft, err := data.LoadFontFace("resource/futura.ttf", 11)
	if err != nil {
		panic(err)
	}

	o.Context.SetFontFace(ft)
	go o.read()
	return o
}

func (o *OpenMic) read() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		t := scanner.Text()
		o.set(t)
		fmt.Println("READ:", t)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}

func (o *OpenMic) get() string {
	o.lock.Lock()
	defer o.lock.Unlock()
	return o.text
}

func (o *OpenMic) set(t string) {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.text = t
}

func (o *OpenMic) Draw(data *data.Data, img *image.RGBA) {
	o.Context.DrawRectangle(0, 0, 128, 128)
	o.Context.SetRGB(data.Sliders[0].Float(), data.Sliders[1].Float(), data.Sliders[2].Float())
	o.Context.Fill()

	o.Context.SetRGB(data.Sliders[3].Float(), data.Sliders[4].Float(), data.Sliders[5].Float())

	o.Context.DrawStringWrapped(o.get(), 3, 3, 0, 0, 61, 1.1, gg.AlignLeft)

	// for i := 0; i < 8; i++ {
	// 	x := (i/4)*64 + 32
	// 	y := (i%4)*32 + 16
	// 	o.Context.DrawString(fmt.Sprint(i+1), float64(x), float64(y))
	// }

	it := o.Context.Image().(*image.RGBA)
	it.Pix, img.Pix = img.Pix, it.Pix

}
