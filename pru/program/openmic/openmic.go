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

type OpenMic struct {
	lock sync.Mutex
	text string
}

func New() *OpenMic {
	o := &OpenMic{}
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
	ggctx := gg.NewContextForRGBA(img)

	ggctx.DrawRectangle(0, 0, 64, 64)
	ggctx.SetRGB(data.Sliders[0].Float(), data.Sliders[1].Float(), data.Sliders[2].Float())
	ggctx.Fill()

	ggctx.SetRGB(data.Sliders[3].Float(), data.Sliders[4].Float(), data.Sliders[5].Float())

	// if err := ggctx.LoadFontFace("/Library/Fonts/Arial.ttf", 12); err != nil {
	// 	panic(err)
	// }
	ggctx.DrawStringWrapped(o.get(), 3, 3, 0, 0, 56, 1.2, gg.AlignLeft)

	// for i := 0; i < 8; i++ {
	// 	x := (i/4)*64 + 32
	// 	y := (i%4)*32 + 16
	// 	ggctx.DrawString(fmt.Sprint(i+1), float64(x), float64(y))
	// }
}
