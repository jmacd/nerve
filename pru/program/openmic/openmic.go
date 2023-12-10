package openmic

import (
	"bufio"
	"fmt"
	"image"
	"io/fs"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/fogleman/gg"
	"github.com/jmacd/nerve/pru/program/data"
	"golang.org/x/image/font"
	"gonum.org/v1/gonum/stat/distuv"
)

const lineSpacing = 1.15
const lineSpacingVar = 0.25
const displayHeight = 60
const displayWidth = 60
const displayMargin = 2
const displayTotal = displayWidth + 2*displayMargin
const fontMin = 6.0
const fontMax = 24.0
const fontSize = 12.0

type OpenMic struct {
	*gg.Context
	lock    sync.Mutex
	display []string
	input   string
	fonts   []font.Face
	fnames  []string
	rate    float64
	fnum    int
	fsize   float64
	lspace  float64
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

func New(startText string, reader bool) *OpenMic {
	o := &OpenMic{
		Context: gg.NewContext(128, 128),
		input:   startText,
		fsize:   fontSize,
		lspace:  lineSpacing,
		rate:    0.5,
	}
	files, err := fs.Glob(data.ResourceFS, "resource/*.ttf")
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		ft, err := data.LoadFontFace(file, o.fsize)
		if err != nil {
			fmt.Printf("skipping font %q: %v\n", file, err)
			continue
		}
		o.fonts = append(o.fonts, ft)
		o.fnames = append(o.fnames, file)
	}

	if len(o.fonts) == 0 {
		panic("no fonts were loaded!")
	}

	if reader {
		go o.read()
	}
	go o.write()
	return o
}

func (o *OpenMic) wrapInput() []string {
	o.lock.Lock()
	defer o.lock.Unlock()

	o.Context.SetFontFace(o.fonts[o.fnum])

	return o.Context.WordWrap(o.input, displayWidth)
}

func (o *OpenMic) write() {
	for {
		o.clear()
		time.Sleep(time.Second)

		txt := o.wrapInput()

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

			ii := time.Duration(rnd * float64(time.Millisecond) * 200 * (1.5 - o.speed()))

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

func (o *OpenMic) speed() float64 {
	return o.rate
}

func (o *OpenMic) stroke(ch rune, end bool) {
	o.lock.Lock()
	defer o.lock.Unlock()

	o.Context.SetFontFace(o.fonts[o.fnum])

	var b strings.Builder
	b.WriteString(o.display[len(o.display)-1])
	b.WriteRune(ch)
	o.display[len(o.display)-1] = b.String()

	if end {
		o.display = append(o.display, "")
	}

	o.combine = false
	if len(o.display) > 1 {

		w := func() (rv float64) {
			// See https://github.com/golang/freetype/issues/87
			if ret := recover(); ret != nil {
				fmt.Printf("Could not measure %q %q\n", o.display[len(o.display)-2], o.display[len(o.display)-1])
				rv = displayWidth
				return
			}
			w, _ := o.Context.MeasureString(o.display[len(o.display)-2] + " " + o.display[len(o.display)-1])
			return w
		}()
		if w <= displayWidth {
			o.combine = true
		}
	}

	lc := float64(len(o.display))
	if o.combine {
		lc--
	}
	if lc*o.lspace*o.Context.FontHeight() > displayHeight {
		copy(o.display[0:len(o.display)-1], o.display[1:])
		o.display = o.display[0 : len(o.display)-1]
	}
}

func (o *OpenMic) getLocked() ([]string, bool) {
	return o.display, o.combine
}

func (o *OpenMic) set(t string) {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.input = t
	o.combine = false
}

func (o *OpenMic) Draw(dat *data.Data, img *image.RGBA) {
	o.lock.Lock()
	defer o.lock.Unlock()

	o.Context.DrawRectangle(0, 0, 128, 128)
	o.Context.SetRGB(dat.Sliders[0].Float(), dat.Sliders[1].Float(), dat.Sliders[2].Float())
	o.Context.Fill()

	o.Context.SetRGB(dat.Sliders[3].Float(), dat.Sliders[4].Float(), dat.Sliders[5].Float())

	nfsize := 0.0
	nfval := dat.KnobsRow1[6].Float() - 0.5
	if nfval < 0 {
		nfsize = fontSize + (fontSize-fontMin)*nfval
	} else {
		nfsize = fontSize + (fontMax-fontSize)*nfval
	}

	if nfsize != o.fsize {
		o.fsize = nfsize
		o.fonts = nil
		for _, file := range o.fnames {
			ft, _ := data.LoadFontFace(file, o.fsize)
			if ft != nil {
				o.fonts = append(o.fonts, ft)
			}
		}
	}

	o.rate = dat.KnobsRow1[4].Float()
	o.fnum = int(dat.KnobsRow1[7]) % len(o.fonts)
	o.lspace = lineSpacing + lineSpacingVar*(dat.KnobsRow1[5].Float()-0.5)
	o.Context.SetFontFace(o.fonts[o.fnum])

	lines, combine := o.getLocked()
	lc := len(lines)
	var l1, l2 string
	if combine {
		l1 = lines[lc-1]
		l2 = lines[lc-2]
		lines = lines[:lc-2]
	}
	for idx, line := range lines {
		o.Context.DrawStringAnchored(line, displayMargin, float64(idx+1)*o.Context.FontHeight()*o.lspace, 0, 0)
	}
	if combine {
		o.Context.DrawStringAnchored(l2+" "+l1, displayMargin, float64(lc-1)*o.Context.FontHeight()*o.lspace, 0, 0)
	}

	it := o.Context.Image().(*image.RGBA)
	it.Pix, img.Pix = img.Pix, it.Pix
}
