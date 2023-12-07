package openmic

import (
	"bufio"
	"fmt"
	"image"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/fogleman/gg"
	"github.com/jmacd/nerve/pru/program/data"
	"gonum.org/v1/gonum/stat/distuv"
)

const displayWidth = 60
const displayMargin = 2
const displayTotal = displayWidth + 2*displayMargin

var startText = `this is open mic nite; welcome. glad you came, we
have lots to do.  I think it would be nice if we could have a
gathering of makers too.  `

type OpenMic struct {
	*gg.Context
	lock    sync.Mutex
	display []string
	input   string
	combine bool
}

var (
	arrival = distuv.LogNormal{
		Mu:    0.15,
		Sigma: 0.6,
	}
	space = distuv.LogNormal{
		Mu:    0,
		Sigma: 0.4,
	}
	punct = distuv.LogNormal{
		Mu:    0.1,
		Sigma: 0.2,
	}
)

func New() *OpenMic {
	o := &OpenMic{
		Context: gg.NewContext(128, 128),
	}
	ft, err := data.LoadFontFace("resource/futura.ttf", 12)
	if err != nil {
		panic(err)
	}

	o.Context.SetFontFace(ft)
	o.input = startText

	go o.read()
	go o.write()
	return o
}

func (o *OpenMic) write() {
	for {
		o.clear()
		time.Sleep(time.Second)

		txt := o.Context.WordWrap(o.input, displayWidth)

		for len(txt) != 0 {
			line := txt[0]
			r, size := utf8.DecodeRuneInString(line)
			line = line[size:]

			if len(line) > 0 {
				o.stroke(r, false)
				txt[0] = line
			} else {
				o.stroke(r, true)
				txt = txt[1:]
			}

			rnd := 0.0
			if unicode.IsSpace(r) {
				rnd = arrival.Rand()
			} else if unicode.IsPunct(r) {
				rnd = punct.Rand()
			} else {
				rnd = space.Rand()
			}

			ii := time.Duration(rnd * float64(time.Millisecond) * 50)

			time.Sleep(ii)
		}
		time.Sleep(time.Second)
	}
}

func (o *OpenMic) read() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		t := scanner.Text()
		o.set(t)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}

func (o *OpenMic) clear() {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.display = []string{""}
	o.combine = false
}

func (o *OpenMic) stroke(ch rune, end bool) {
	o.lock.Lock()
	defer o.lock.Unlock()

	var b strings.Builder
	b.WriteString(o.display[len(o.display)-1])
	b.WriteRune(ch)
	o.display[len(o.display)-1] = b.String()

	if end {
		o.display = append(o.display, "")
	}

	o.combine = false
	if len(o.display) > 1 {
		w, _ := o.Context.MeasureString(o.display[len(o.display)-2] + " " + o.display[len(o.display)-1])
		if w <= displayWidth {
			o.combine = true
		}
	}

	lc := float64(len(o.display))
	if o.combine {
		lc--
	}
	if lc*o.Context.FontHeight() > displayTotal {
		copy(o.display[0:len(o.display)-1], o.display[1:])
		o.display = o.display[0 : len(o.display)-1]
	}
}

func (o *OpenMic) get() ([]string, bool) {
	o.lock.Lock()
	defer o.lock.Unlock()
	return o.display, o.combine
}

func (o *OpenMic) set(t string) {
	fmt.Println("SET:", t)
	o.lock.Lock()
	defer o.lock.Unlock()
	o.input = t
	o.combine = false
}

func (o *OpenMic) Draw(data *data.Data, img *image.RGBA) {
	o.Context.DrawRectangle(0, 0, 128, 128)
	o.Context.SetRGB(data.Sliders[0].Float(), data.Sliders[1].Float(), data.Sliders[2].Float())
	o.Context.Fill()

	o.Context.SetRGB(data.Sliders[3].Float(), data.Sliders[4].Float(), data.Sliders[5].Float())

	const lineSpacing = 1.15
	lines, combine := o.get()
	lc := len(lines)
	var l1, l2 string
	if combine {
		l1 = lines[lc-1]
		l2 = lines[lc-2]
		lines = lines[:lc-2]
	}
	for idx, line := range lines {
		o.Context.DrawStringAnchored(line, displayMargin, float64(idx+1)*o.Context.FontHeight()*lineSpacing, 0, 0)
	}
	if combine {
		o.Context.DrawStringAnchored(l2+" "+l1, displayMargin, float64(lc-1)*o.Context.FontHeight()*lineSpacing, 0, 0)
	}

	it := o.Context.Image().(*image.RGBA)
	it.Pix, img.Pix = img.Pix, it.Pix
}
