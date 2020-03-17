module github.com/jmacd/nerve

replace github.com/jmacd/go-artnet => ../go-artnet

replace gitlab.com/gomidi/midi => ../../../gitlab.com/gomidi/midi

go 1.13

require (
	github.com/deadsy/libusb v0.0.0-20180330230923-04d6a756f531
	github.com/hsluv/hsluv-go v2.0.0+incompatible
	github.com/jsimonetti/go-artnet v0.0.0-20200229173917-43ec7447138c
	github.com/lucasb-eyer/go-colorful v1.0.3
	gitlab.com/gomidi/midi v1.14.1
	go.opentelemetry.io/otel v0.2.3
)
