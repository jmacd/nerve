package video

import (
	// _ "image/png"

	"context"
	"fmt"
	"image"
	"math"
	"math/rand"
	"os"
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

	win.Clear(colornames.Skyblue)

	for !win.Closed() {
		win.Update()
	}
	return nil
}

func loadPictureXXX(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}

func runXXX() {
	cfg := pixelgl.WindowConfig{
		Title:  "Nerve!",
		Bounds: pixel.R(0, 0, 1024, 768),
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	// spritesheet, err := loadPictureXXX("trees.png")
	// if err != nil {
	// 	panic(err)
	// }

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
}

func mainXXX() {
	pixelgl.Run(runXXX)
}
