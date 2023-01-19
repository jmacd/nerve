//go:build darwin

package main

import (
	"image"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"github.com/jmacd/nerve/pru/gpixio"
)

type appState struct {
	frames *Frameset

	inputWindow  fyne.Window
	outputWindow fyne.Window

	inputImage  *canvas.Image
	outputImage *canvas.Image

	outputPixels *image.RGBA
}

func newAppState(buf *gpixio.Buffer) (*appState, error) {
	app := app.New()

	outputPixels := image.NewRGBA(image.Rect(0, 0, 128, 128))

	inputWindow := app.NewWindow("Image")
	inputImage := canvas.NewImageFromImage(buf.RGBA)
	inputImage.FillMode = canvas.ImageFillOriginal
	inputWindow.SetContent(inputImage)

	outputWindow := app.NewWindow("Visage")
	outputImage := canvas.NewImageFromImage(outputPixels)
	outputImage.FillMode = canvas.ImageFillOriginal
	outputWindow.SetContent(outputImage)

	return &appState{
		frames:       &Frameset{},
		inputWindow:  inputWindow,
		outputWindow: outputWindow,
		inputImage:   inputImage,
		outputImage:  outputImage,
		outputPixels: outputPixels,
	}, nil
}

func (state *appState) test(schedule int) {
	testRender(&state.frames[schedule], state.outputPixels)

	canvas.Refresh(state.inputImage)
	canvas.Refresh(state.outputImage)
}

func (state *appState) run() error {
	state.outputWindow.Show()
	state.inputWindow.ShowAndRun()
	return nil
}
