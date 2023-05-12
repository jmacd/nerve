package gpixio

type residue [3]uint8 // R, G, B
type residues [16]residue
type intermediate [16][64]residues

// func pixelOffsetFor1(x, y, pos int) int {
// 	panelX := pos / 8
// 	panelY := pos % 8
// 	pixY := (panelY * 16) + y
// 	pixX := (panelX * 64) + x
// 	return 4 * (128*pixY + pixX)
// }

func cp(r *residue, p []byte, o int) {
	// r[0] = p[o]
	// r[1] = p[o+1]
	// r[2] = p[o+2]
	copy(r[:], p[o:o+3])
}

func (b *Buffer) Copy1(fb *FrameBank) {
	var inter intermediate
	var offset int

	// 4 bytes/pix * 128 pix/row * 16 row/panel
	ys := 4 * 128 * 16
	xs := 4 * 64

	pix := b.RGBA.Pix

	// Gather into GPIO-pixel-order
	for y := 0; y < 16; y++ {
		row := &inter[y]

		for x := 0; x < 64; x++ {
			cp(&row[x][0], pix, offset+0*ys)
			cp(&row[x][1], pix, offset+1*ys)
			cp(&row[x][2], pix, offset+2*ys)
			cp(&row[x][3], pix, offset+3*ys)
			cp(&row[x][4], pix, offset+4*ys)
			cp(&row[x][5], pix, offset+5*ys)
			cp(&row[x][6], pix, offset+6*ys)
			cp(&row[x][7], pix, offset+7*ys)

			cp(&row[x][8], pix, offset+0*ys+xs)
			cp(&row[x][9], pix, offset+1*ys+xs)
			cp(&row[x][10], pix, offset+2*ys+xs)
			cp(&row[x][11], pix, offset+3*ys+xs)
			cp(&row[x][12], pix, offset+4*ys+xs)
			cp(&row[x][13], pix, offset+5*ys+xs)
			cp(&row[x][14], pix, offset+6*ys+xs)
			cp(&row[x][15], pix, offset+7*ys+xs)
		}
	}

	// Calculate 2K 16-position 3x1-bit values x 256 frames
	for {
	}
}
