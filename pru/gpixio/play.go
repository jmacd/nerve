package gpixio

import (
	"image"
	"math"
	"math/bits"
)

type (
	Frameset  [8]Schedule
	Schedule  [64]Frame
	Frame     [16]DoubleRow
	DoubleRow [64]DoublePixel

	DoublePixel struct {
		Gpio0 uint32
		Gpio1 uint32
		Gpio2 uint32
		Gpio3 uint32
	}
)

const (
	J1_1 = 0
	J1_2 = 1
	J2_1 = 2
	J2_2 = 3
	J3_1 = 4
	J3_2 = 5
	J4_1 = 6
	J4_2 = 7
	J5_1 = 8
	J5_2 = 9
	J6_1 = 10
	J6_2 = 11
	J7_1 = 12
	J7_2 = 13
	J8_1 = 14
	J8_2 = 15
)

const deviceGamma = 2.3 // TODO

var degammaSix [256]uint8 = func() [256]uint8 {
	var d6 [256]uint8
	for i := range d6 {
		d6[i] = uint8(255*math.Pow(float64(i)/255, deviceGamma)) >> 2
	}
	return d6
}()

var sixBitPatterns [64]uint64 = func() [64]uint64 {
	var patterns [64]uint64
	for i := 1; i < 64; i++ {
		var p uint64
		stride := 64 / float64(i)
		offset := 0.0
		for j := 0; j < i; j++ {
			p |= 1 << int64(offset)
			offset += stride
		}
		patterns[i] = p
		if bits.OnesCount64(p) != i {
			panic("bad logic")
		}
	}
	return patterns
}()

type Buffer struct {
	*image.RGBA
}

func NewBuffer() *Buffer {
	img := image.NewRGBA(image.Rect(0, 0, 128, 128))
	return &Buffer{
		RGBA: img,
	}
}

type frameBits [16]uint64

func (f *frameBits) choose(position, frame, bit int) uint32 {
	if (*f)[position]&(1<<frame) == 0 {
		return 0
	}
	return 1 << bit
}

func pixelOffsetFor(rowSel, rowQuad, pos int) int {
	panelX := pos / 8
	panelY := pos % 8

	pixY := (panelY * 16) + rowSel
	pixX := (panelX * 64) + (rowQuad * 16)

	return 4 * (128*pixY + pixX)
}

func (b *Buffer) Copy(schedule *Schedule) {

	for rowSel := 0; rowSel < 16; rowSel++ {

		for rowQuad := 0; rowQuad < 4; rowQuad++ {

			// This loop activates 64 times, each time
			// constructing 768 bytes in three arrays.
			//
			//   64*768 == 3*(2**14)

			var R [16][16]byte
			var G [16][16]byte
			var B [16][16]byte

			// Here, step through 2 positions per panel
			// for 8 panels to yield 16 runs of 16 RGB
			// pixels.
			for pos := 0; pos < 16; pos++ {
				pR := &R[pos]
				pG := &G[pos]
				pB := &B[pos]

				pixOffset := pixelOffsetFor(rowSel, rowQuad, pos)

				// Gruesome: the next loop should be
				// done by a NEON 4-way extract
				// instruction, discarding one output
				// register (i.e., the alpha channel).
				for pix := 0; pix < 16; pix++ {
					(*pR)[pix] = b.Pix[pixOffset+0]
					(*pG)[pix] = b.Pix[pixOffset+1]
					(*pB)[pix] = b.Pix[pixOffset+2]
					pixOffset += 4
				}
			}

			// Loop body covers 1/64th of the 2**14 pixel image
			// (i.e., 256 pixels, 1KiB).
			//
			// We've constructed 16x16 pixels byte arrays for each
			// color channel.  768 bytes used (RGB), 256 bytes
			// skipped (A).
			//

			// The first dimension 16 pixels have the same row
			// selector & pixel number.  Each group translates into
			// 64-frame time slices (using 6 of 8 bits per color
			// channel), 16 bytes per timeslice. So, 16 adjacent
			// pixels each produces 64 * 16 == 1024 bytes,
			// generating a total of 16KiB output from the input 768
			// bytes.
			//
			// This loop body executes 64 times, yielding (64 *
			// 16KiB == 1MiB) for 64 temporal frames across all
			// pixels.  This is 1/8th of the frame buffer and
			// approximately 1/32 seconds.

			// For each of 16 pixels:
			for p := 0; p < 16; p++ {
				var reds frameBits
				var greens frameBits
				var blues frameBits

				for x := 0; x < 16; x++ {
					reds[x] = sixBitPatterns[degammaSix[R[x][p]]]
					greens[x] = sixBitPatterns[degammaSix[G[x][p]]]
					blues[x] = sixBitPatterns[degammaSix[B[x][p]]]
				}

				for f := 0; f < 64; f++ {
					dp := &schedule[f][rowSel][p+rowQuad*16]

					// Note: the code below is auto-generated by
					// ../cmd/mkmap.  Except add
					// uint32(rowSel)<<12, // Rowselect is bits 12:15
					// to gpio1.
					//
					// TODO The following works, i.e., subtracting
					// one from the rowSel.  why?
					// (uint32((rowSel+15)%16) << 12)

					dp.Gpio0 = blues.choose(J1_2, f, 26) |
						reds.choose(J1_2, f, 23) |
						reds.choose(J2_1, f, 27) |
						blues.choose(J2_1, f, 22) |
						reds.choose(J3_1, f, 30) |
						blues.choose(J3_1, f, 31) |
						greens.choose(J3_2, f, 3) |
						blues.choose(J3_2, f, 5) |
						blues.choose(J4_2, f, 4) |
						reds.choose(J4_1, f, 2) |
						greens.choose(J4_1, f, 15) |
						greens.choose(J5_1, f, 11) |
						blues.choose(J5_1, f, 10) |
						reds.choose(J5_2, f, 9) |
						greens.choose(J5_2, f, 8) |
						greens.choose(J8_2, f, 14)
					dp.Gpio1 = greens.choose(J3_1, f, 18) |
						reds.choose(J3_2, f, 16) |
						blues.choose(J4_1, f, 17) |
						(uint32((rowSel+15)%16) << 12)
					dp.Gpio2 = greens.choose(J1_2, f, 4) |
						reds.choose(J1_1, f, 2) |
						greens.choose(J1_1, f, 3) |
						blues.choose(J1_1, f, 5) |
						greens.choose(J2_1, f, 1) |
						reds.choose(J2_2, f, 22) |
						greens.choose(J2_2, f, 23) |
						blues.choose(J2_2, f, 24) |
						blues.choose(J5_2, f, 17) |
						reds.choose(J5_1, f, 25) |
						reds.choose(J6_1, f, 16) |
						greens.choose(J6_1, f, 15) |
						blues.choose(J6_1, f, 14) |
						reds.choose(J6_2, f, 13) |
						greens.choose(J6_2, f, 10) |
						blues.choose(J6_2, f, 12) |
						reds.choose(J7_1, f, 11) |
						greens.choose(J7_1, f, 9) |
						blues.choose(J7_1, f, 8) |
						reds.choose(J7_2, f, 6) |
						blues.choose(J7_2, f, 7)
					dp.Gpio3 = reds.choose(J4_2, f, 21) |
						greens.choose(J4_2, f, 19) |
						greens.choose(J7_2, f, 18) |
						reds.choose(J8_2, f, 14) |
						blues.choose(J8_2, f, 20) |
						reds.choose(J8_1, f, 17) |
						greens.choose(J8_1, f, 16) |
						blues.choose(J8_1, f, 15)
				}
			}
		}
	}
}
