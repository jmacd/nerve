package video

import (
	// _ "image/png"

	"context"
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"time"

	"github.com/lucasb-eyer/go-colorful"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

type (
	Video struct {
		inputCh chan []Color
	}

	Color = colorful.Color
)

func New() *Video {
	return &Video{}
}

func (v *Video) Input() chan []Color {
	return v.inputCh
}

func (v *Video) Run(ctx context.Context) error {
	cfg := pixelgl.WindowConfig{
		Title:  "Pixel Rocks!",
		Bounds: pixel.R(0, 0, 1024, 768),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	spritesheet, err := loadPictureXXX("trees.png")
	if err != nil {
		panic(err)
	}

	batch := pixel.NewBatch(&pixel.TrianglesData{}, spritesheet)

	var treesFrames []pixel.Rect
	for x := spritesheet.Bounds().Min.X; x < spritesheet.Bounds().Max.X; x += 32 {
		for y := spritesheet.Bounds().Min.Y; y < spritesheet.Bounds().Max.Y; y += 32 {
			treesFrames = append(treesFrames, pixel.R(x, y, x+32, y+32))
		}
	}

	var (
		camPos       = pixel.ZV
		camSpeed     = 500.0
		camZoom      = 1.0
		camZoomSpeed = 1.2
	)

	var (
		frames = 0
		second = time.Tick(time.Second)
	)

	last := time.Now()
	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()

		cam := pixel.IM.Scaled(camPos, camZoom).Moved(win.Bounds().Center().Sub(camPos))
		win.SetMatrix(cam)

		if win.Pressed(pixelgl.MouseButtonLeft) {
			tree := pixel.NewSprite(spritesheet, treesFrames[rand.Intn(len(treesFrames))])
			mouse := cam.Unproject(win.MousePosition())
			tree.Draw(batch, pixel.IM.Scaled(pixel.ZV, 4).Moved(mouse))
		}
		if win.Pressed(pixelgl.KeyLeft) {
			camPos.X -= camSpeed * dt
		}
		if win.Pressed(pixelgl.KeyRight) {
			camPos.X += camSpeed * dt
		}
		if win.Pressed(pixelgl.KeyDown) {
			camPos.Y -= camSpeed * dt
		}
		if win.Pressed(pixelgl.KeyUp) {
			camPos.Y += camSpeed * dt
		}
		camZoom *= math.Pow(camZoomSpeed, win.MouseScroll().Y)

		win.Clear(colornames.Forestgreen)
		batch.Draw(win)
		win.Update()

		frames++
		select {
		case <-second:
			win.SetTitle(fmt.Sprintf("%s | FPS: %d", cfg.Title, frames))
			frames = 0
		default:
		}
	}
	// win.Clear(colornames.Skyblue)

	// for !win.Closed() {
	// 	win.Update()
	// }
	return nil
}

func loadPictureXXX(path string) (pixel.Picture, error) {
	// file, err := os.Open(path)
	// if err != nil {
	// 	return nil, err
	// }
	// defer file.Close()
	// img, _, err := image.Decode(file)
	// if err != nil {
	// 	return nil, err
	// }
	img := image.NewRGBA(image.Rectangle{
		Min: image.Pt(0, 0),
		Max: image.Pt(32, 32),
	})
	sigma := 1.0
	exd := 2 * sigma * sigma
	max := 1 / (math.Pi * exd)

	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			x0 := float64(x - 16)
			y0 := float64(y - 16)
			exn := x0*x0 + y0*y0
			exp := math.Exp(-exn / exd)
			g := exp / (math.Pi * exd)
			n := g / max

			a := uint8(n * 255)
			img.Set(x, y, color.RGBA{a, a, a, a})
		}
	}

	return pixel.PictureDataFromImage(img), nil
}
