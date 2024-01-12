package main

import (
	"context"
	"fmt"
	"image"
	"time"

	"github.com/jmacd/nerve/pru/artnet"
)

func main() {
	img := image.NewRGBA(image.Rect(0, 0, 128, 128))
	r, err := artnet.NewReceiver("127.0.0.1", img)
	if err != nil {
		panic(err)
	}
	go r.Start(context.Background())

	go func() {
		for {
			if err := r.Draw(); err != nil {
				panic(err)
			}
			fmt.Println("Draw!")
			time.Sleep(time.Second)
		}
	}()

	select {}
}
