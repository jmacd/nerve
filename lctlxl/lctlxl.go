package lctlxl

import (
	"fmt"

	"github.com/google/gousb"
	"github.com/google/gousb/usbid"
)

const (
	InNum  = 0x81
	OutNum = 0x2

	VendorID  gousb.ID = 0x1235
	ProductID gousb.ID = 0x61

	ProductName = "Launch Control XL"
)

type (
	LaunchControl struct {
		*gousb.Context
		*gousb.Device
		*gousb.Config
		*gousb.InEndpoint
		*gousb.OutEndpoint
	}
)

func Open() (*LaunchControl, error) {
	// Initialize a new Context.
	uctx := gousb.NewContext()

	// Open any device with a given VID/PID using a convenience function.
	dev, err := uctx.OpenDeviceWithVIDPID(VendorID, ProductID)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	if dev == nil {
		return nil, fmt.Errorf("can't find %v", ProductName)
	}

	if len(dev.Desc.Configs) != 1 {
		return nil, fmt.Errorf("%v: unexpected configs: %d", dev, len(dev.Desc.Configs))
	}

	var cfgNum int
	var cfgDesc gousb.ConfigDesc
	for cfgNum, cfgDesc = range dev.Desc.Configs {
		fmt.Printf("  %s:\n", cfgDesc)
		for _, intf := range cfgDesc.Interfaces {
			fmt.Printf("    --------------\n")
			for _, ifSetting := range intf.AltSettings {
				fmt.Printf("    %s\n", ifSetting)
				fmt.Printf("      %s\n", usbid.Classify(ifSetting))
				for _, end := range ifSetting.Endpoints {
					fmt.Printf("      %s\n", end)
				}
			}
		}
	}

	cfg, err := dev.Config(cfgNum)
	if err != nil {
		return nil, fmt.Errorf("%v config: %w", dev, err)
	}

	intf, err := cfg.Interface(1, 0)
	if err != nil {
		return nil, fmt.Errorf("%v default interface: %w", dev, err)
	}

	in, err := intf.InEndpoint(InNum)
	if err != nil {
		return nil, fmt.Errorf("%v in endpoint : %w", dev, err)
	}

	out, err := intf.OutEndpoint(OutNum)
	if err != nil {
		return nil, fmt.Errorf("%v out endpoint : %w", dev, err)
	}

	return &LaunchControl{
		Context:     uctx,
		Device:      dev,
		Config:      cfg,
		InEndpoint:  in,
		OutEndpoint: out,
	}, nil
}
