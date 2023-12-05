//go:build !darwin

package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"github.com/jmacd/nerve/pru/gpixio"
)

const deviceName = "/dev/rpmsg_pru30"

type RPMsgDevice struct {
	file *os.File
}

// controlStruct should match ../control.h
type controlStruct struct {
	framebufsAddr uint32
	framebufsSize uint32
	frameCount    uint32
	dmaWait       uint32

	// readyBank is the next bank that the PRU will start.
	readyBank uint32

	// startBank is the most recent bank started by the PRU.
	startBank uint32

	// The ARM can be in two states:
	// 1. Waiting for startBank to equal readyBank.
	// 2. Writing to (readyBank^1) before updating readyBank.

	// The PRU will update its "currentBank" to the readyBank when
	// it reaches the end of a frame.  The PRU sets startBank to
	// the currentBank when it starts a frame.

}

type appState struct {
	frames *Frameset
	ctrl   *controlStruct
	rpm    *RPMsgDevice
}

func newAppState(buf *gpixio.Buffer) (*appState, error) {
	rpm, err := openRPMsgDevice()
	if err != nil {
		return nil, err
	}

	// pru.c does not parse the message, this delivers the
	// interrupt which causes pru.c to respond with its two
	// carveout addresses.
	if err := rpm.write([]byte("wakeup")); err != nil {
		return nil, err
	}

	ctrl, frames, err := rpm.readControl()
	if err != nil {
		return nil, err
	}

	go func() {
		before := atomic.LoadUint32(&ctrl.frameCount)
		last := time.Now()
		for {
			time.Sleep(10 * time.Second)
			now := time.Now()
			after := atomic.LoadUint32(&ctrl.frameCount)
			log.Println("frames/sec", float64(after-before)/now.Sub(last).Seconds(), "bank", ctrl.readyBank)
			before = after
			last = now
		}
	}()

	return &appState{
		ctrl:   ctrl,
		rpm:    rpm,
		frames: frames,
	}, nil
}

func (state *appState) finish(bank uint32) {
	atomic.StoreUint32(&state.ctrl.readyBank, bank)
}

func (state *appState) run() error {
	select {}
}

func (state *appState) waitReady() uint32 {
	ready := atomic.LoadUint32(&state.ctrl.readyBank)
	for ready != atomic.LoadUint32(&state.ctrl.startBank) {
		runtime.Gosched()
	}
	return ready ^ 1
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

func mmap(file *os.File, addr uint32, size int) ([]byte, error) {
	return syscall.Mmap(
		int(file.Fd()),
		int64(addr),
		size,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED,
	)
}

func (r *RPMsgDevice) readControl() (*controlStruct, *Frameset, error) {
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

	cdata, err := mmap(mem, addr, int(unsafe.Sizeof(controlStruct{})))
	if err != nil {
		return nil, nil, err
	}

	ctrl := (*controlStruct)(unsafe.Pointer(&cdata[0]))

	fdata, err := mmap(mem, ctrl.framebufsAddr, int(ctrl.framebufsSize))
	if err != nil {
		return nil, nil, err
	}
	if ctrl.framebufsSize != uint32(unsafe.Sizeof(Frameset{})) {
		return nil, nil, fmt.Errorf("frameset size mismatch: C=%d, Go=%d", ctrl.framebufsSize, unsafe.Sizeof(Frameset{}))
	}

	framebuf := (*Frameset)(unsafe.Pointer(&fdata[0]))

	// Note: should drop privileges now.
	return ctrl, framebuf, nil
}
