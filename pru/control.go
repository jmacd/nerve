package main

import (
	"fmt"
	"log"
	"os"
)

const deviceName = "/dev/rpmsg_pru30"

type RPMsgDevice struct {
	file *os.File
}

func openRPMsgDevice() (*RPMsgDevice, error) {
	file, err := os.OpenFile(deviceName, os.O_RDWR, 0666)
	return &RPMsgDevice{
		file: file,
	}, err
}

func (r *RPMsgDevice) write(data []byte) error {
	n, err := r.file.Write(data)
	if err != nil {
		return nil
	}
	if n != len(data) {
		return fmt.Errorf("short write: %d != %d", n, len(data))
	}
	return nil
}

func (r *RPMsgDevice) read() ([]byte, error) {
	var data [32]byte // TODO: sizeof control struct
	n, err := r.file.Read(data[:])
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func Main() error {
	rpm, err := openRPMsgDevice()
	if err != nil {
		return err
	}

	// pru.c does not parse the message, this delivers the
	// interrupt which causes pru.c to respond with its two
	// carveout addresses.
	if err := rpm.write([]byte("wakeup")); err != nil {
		return err
	}

	for {
		data, err := rpm.read()
		if err != nil {
			return err
		}
		// TODO: actually a control struct
		fmt.Println("Read data: ", string(data))
	}

	return nil
}

func main() {
	if err := Main(); err != nil {
		log.Println("error:", err)
	}
}
