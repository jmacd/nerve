package lctlxl

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/deadsy/libusb"
	"gitlab.com/gomidi/midi/mid"
)

const (
	InNum  = 0x81
	OutNum = 0x2

	VendorID  = 0x1235
	ProductID = 0x61

	TimeoutMS = 1000

	ProductName = "Launch Control XL"
)

type (
	LaunchControl struct {
		ctx  libusb.Context
		hdl  libusb.Device_Handle
		rep  Endpoint
		wep  Endpoint
		wait *sync.WaitGroup
		stop chan struct{}
		ferr chan error

		reader io.Reader
		writer io.Writer

		rdbuf []byte
		wrbuf []byte

		midiRead *mid.Reader
		// midiWrite *mid.Writer

		SendA [8]float64 // 13-20
		SendB [8]float64 // 29-36
		SendC [8]float64 // 49-56
		Slide [8]float64 // 77-84
		NoteA [8]float64 // F1-C3
		NoteB [8]float64 // C#4-G#5
	}

	Endpoint struct {
		iface int
		ep    *libusb.Endpoint_Descriptor
	}
)

func Open() (*LaunchControl, error) {
	lc := &LaunchControl{
		wait: &sync.WaitGroup{},
		stop: make(chan struct{}),
		ferr: make(chan error),
	}
	for i := 0; i < 8; i++ {
		lc.SendA[i] = 1
		lc.SendB[i] = 1
		lc.SendC[i] = 1
		lc.Slide[i] = 1
	}

	var ctx libusb.Context

	err := libusb.Init(&ctx)
	if err != nil {
		return nil, fmt.Errorf("init libusb: %w", err)
	}

	lc.ctx = ctx
	lc.hdl = libusb.Open_Device_With_VID_PID(lc.ctx, VendorID, ProductID)

	if lc.hdl == nil {
		return nil, fmt.Errorf("can't find %v", ProductName)
	}

	dev := libusb.Get_Device(lc.hdl)
	if dev == nil {
		return nil, fmt.Errorf("could not get device")
	}

	dd, err := libusb.Get_Device_Descriptor(dev)
	if err != nil {
		return nil, fmt.Errorf("device descriptor: %w", err)
	}

	var readers []Endpoint
	var writers []Endpoint
	found := false

	for i := 0; i < int(dd.BNumConfigurations); i++ {
		cd, err := libusb.Get_Config_Descriptor(dev, uint8(i))
		if err != nil {
			return nil, fmt.Errorf("get config desc %w", err)
		}
		for _, itf := range cd.Interface {
			for _, id := range itf.Altsetting {
				if id.BInterfaceClass == libusb.CLASS_AUDIO && id.BInterfaceSubClass == 3 {
					found = true
				}
				for _, ep := range id.Endpoint {
					if ep.BEndpointAddress&libusb.ENDPOINT_IN != 0 {
						readers = append(readers, Endpoint{
							iface: int(id.BInterfaceNumber),
							ep:    ep,
						})
					} else {
						writers = append(writers, Endpoint{
							iface: int(id.BInterfaceNumber),
							ep:    ep,
						})
					}
				}
			}
		}

		libusb.Free_Config_Descriptor(cd)
	}

	if !found || len(readers) != 1 || len(writers) != 1 {
		return nil, fmt.Errorf("Wrong number of readers/writers: %d/%d", len(readers), len(writers))
	}

	libusb.Set_Auto_Detach_Kernel_Driver(lc.hdl, true)

	// claim the interfaces
	if err := libusb.Claim_Interface(lc.hdl, readers[0].iface); err != nil {
		return nil, fmt.Errorf("Could not get reader %w", err)
	}
	lc.rep = readers[0]
	lc.rdbuf = make([]byte, lc.rep.ep.WMaxPacketSize)

	if err := libusb.Claim_Interface(lc.hdl, writers[0].iface); err != nil {
		return nil, fmt.Errorf("Could not get writer %w", err)
	}
	lc.wep = writers[0]
	lc.wrbuf = make([]byte, lc.wep.ep.WMaxPacketSize)

	return lc, nil
}

func (lc *LaunchControl) Start() error {
	inrd, inwr := io.Pipe()
	outrd, outwr := io.Pipe()

	lc.reader = inrd

	lc.start(func() error { return lc.read(inwr) })

	lc.writer = outwr
	lc.start(func() error { return lc.write(outrd) })

	lc.setupMidi()

	return nil
}

func (lc *LaunchControl) start(f func() error) {
	lc.wait.Add(1)

	go func() {
		defer lc.wait.Done()

		for {
			select {
			case <-lc.stop:
				return
			default:
			}
			err := f()

			if err != nil {
				select {
				case lc.ferr <- err:
					// Set first error
				default:
					// Pass
				}
			}
		}
	}()
}

func (lc *LaunchControl) read(write io.Writer) error {
	data, err := libusb.Bulk_Transfer(lc.hdl, lc.rep.ep.BEndpointAddress, lc.rdbuf, TimeoutMS)

	if err != nil {
		return err
	}

	_, err = write.Write(data)
	return err
}

func (lc *LaunchControl) Reader() io.Reader {
	return lc.reader
}

func (lc *LaunchControl) write(write io.Reader) error {
	time.Sleep(time.Second)
	return nil
}

func (lc *LaunchControl) Stop() error {
	close(lc.stop)

	lc.wait.Wait()

	err0 := libusb.Release_Interface(lc.hdl, lc.rep.iface)
	err1 := libusb.Release_Interface(lc.hdl, lc.wep.iface)

	libusb.Close(lc.hdl)
	libusb.Exit(lc.ctx)

	if err0 != nil || err1 != nil {
		return fmt.Errorf("stop %w %w", err0, err1)
	}
	return nil
}

func (lc *LaunchControl) setupMidi() {
	lc.midiRead = mid.NewReader(mid.SetLogger(nil))

	// TODO Writer
	// wr := mid.NewWriter(lc.OutEndpoint)
	// wr.Start()

	lc.midiRead.Msg.Channel.ControlChange.Each =
		func(_ *mid.Position, channel, controller, value uint8) {
			// if channel != 8 {  //  This seems to vary
			// 	return
			// }

			if controller == 11 {
				// Imaginary (apparently).
				return
			}

			switch {
			case controller >= 13 && controller <= 20:
				lc.SendA[controller-13] = float64(value) / 127
			case controller >= 29 && controller <= 36:
				lc.SendB[controller-29] = float64(value) / 127
			case controller >= 49 && controller <= 56:
				lc.SendC[controller-49] = float64(value) / 127
			case controller >= 77 && controller <= 84:
				lc.Slide[controller-77] = float64(value) / 127
			}
		}

	go func() {
		for {
			if lc.midiRead.ReadAllFrom(lc.Reader()) == io.EOF {
				fmt.Println("EOF")
				break
			}
		}
	}()
}
