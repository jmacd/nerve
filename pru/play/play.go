package play

import (
	"image"

	"github.com/fogleman/gg"
)

func Play() {
	img := image.NewRGBA(image.Rect(0, 0, 128, 128))

	dc := gg.NewContextForRGBA(img)
	dc.SetRGB(0, 0, 1)
	dc.DrawCircle(64, 64, 50)
	dc.SetRGB(1, 1, 0)
	dc.Fill()

	for rowSel := 0; rowSel < 16; rowSel++ {
		for rowQuad := 0; rowQuad < 4; rowQuad++ {

			firstOffset := 4 * ((rowSel * 128) + (rowQuad * 16))

			var R [16][16]byte
			var G [16][16]byte
			var B [16][16]byte

			for pix := 0; pix < 16; pix++ {

				R[0][pix] = img.Pix[firstOffset]
				G[0][pix] = img.Pix[firstOffset+1]
				G[0][pix] = img.Pix[firstOffset+2]
			}
		}
	}
}
