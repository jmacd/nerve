package main

import (
	"image"
	"time"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
)

func main() {
	myApp := app.New()
	w := myApp.NewWindow("Image")

	src := image.NewRGBA(image.Rect(0, 0, 1000, 1000))

	// image := canvas.NewImageFromResource(theme.FyneLogo())
	// image := canvas.NewImageFromURI(uri)
	image := canvas.NewImageFromImage(src)
	// image := canvas.NewImageFromReader(reader, name)
	// image := canvas.NewImageFromFile(fileName)
	image.FillMode = canvas.ImageFillOriginal
	w.SetContent(image)

	go func() {
		for x := 0; ; x = (x + 1) % 256 {
			for i := range src.Pix {
				src.Pix[i] = byte(x)
			}
			canvas.Refresh(image)
			time.Sleep(10 * time.Millisecond)
		}
	}()

	w.ShowAndRun()
}
