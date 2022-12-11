package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"syscall"
	"unsafe"
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

type controlStruct struct {
	u1 uint32
	u2 uint32
	u3 uint32
}

type Framebuf [1 << 21]uint32

func (r *RPMsgDevice) readControl() (*controlStruct, *Framebuf, error) {
	var data [32]byte
	n, err := r.file.Read(data[:])
	if err != nil {
		return nil, nil, err
	}
	if n != 4 {
		return nil, nil, fmt.Errorf("expected 4 bytes control address")
	}
	addr := binary.LittleEndian.Uint32(data[0:4])

	mem, err := os.OpenFile("/dev/mem", os.O_RDWR, 0)
	if err != nil {
		return nil, nil, err
	}

	cdata, err := syscall.Mmap(
		int(mem.Fd()),
		int64(addr),
		int(unsafe.Sizeof(controlStruct{})),
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED,
	)
	if err != nil {
		return nil, nil, err
	}

	ctrl := (*controlStruct)(unsafe.Pointer(&cdata[0]))

	frame := ctrl.u1 // First control word is the frame pointer.
	fdata, err := syscall.Mmap(
		int(mem.Fd()),
		int64(frame),
		int(1<<23), // hard-coded TODO
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED,
	)
	if err != nil {
		return nil, nil, err
	}
	framebuf := (*Framebuf)(unsafe.Pointer(&fdata[0]))

	// TODO: should drop privileges now.
	return ctrl, framebuf, nil
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
	fmt.Println("wakeup sent")

	var ctrl *controlStruct
	var frame *Framebuf

	ctrl, frame, err = rpm.readControl()
	if err != nil {
		return err
	}

	fmt.Println("OK, Go!", *ctrl)

	(*frame)[0] = 0

	return nil
}

func main() {
	if err := Main(); err != nil {
		log.Println("error:", err)
	}
}
