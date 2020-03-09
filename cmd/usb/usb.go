package main

import (
	"fmt"
	"io"
	"log"

	"github.com/jmacd/nerve/lctlxl"
	"gitlab.com/gomidi/midi/mid"
)

func main() {
	lc, err := lctlxl.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer lc.Stop()

	lc.Start()
	fmt.Println("Started...")

	rd := mid.NewReader(mid.SetLogger(lc))

	// rd.Msg.Each = func(_ *mid.Position, msg midi.Message) {
	// 	fmt.Println(msg)
	// }

	// rd.Msg.Unknown = func(_ *mid.Position, msg midi.Message) {
	// 	fmt.Println(msg)
	// }

	// wr := mid.NewWriter(lc.OutEndpoint)
	// wr.Start()

	// go func() {
	// 	wr := mid.NewWriter(pipewr)
	// 	wr.SetChannel(11) // sets the channel for the next messages
	// 	wr.NoteOn(120, 50)
	// 	time.Sleep(time.Second)
	// 	wr.NoteOff(120) // let the note ring for 1 sec
	// 	pipewr.Close()  // finishes the writing
	// }()

	// data, err := libusb.Bulk_Transfer(hdl, ep_in[0].ep.BEndpointAddress, data, 10000)

	go func() {
		for {
			if rd.ReadAllFrom(lc.Reader()) == io.EOF {
				fmt.Println("EOF!!")
				break
			}
		}
	}()

	select {}
}
