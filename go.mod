module github.com/jmacd/nerve

replace github.com/jmacd/go-artnet => ../go-artnet

replace github.com/rakyll/portmidi => ../portmidi
replace github.com/jmacd/launchmidi => ../launchmidi

replace github.com/fogleman/gg => ../../fogleman/gg

go 1.14

require (
	github.com/deadsy/libusb v0.0.0-20180330230923-04d6a756f531
	github.com/fogleman/gg v1.3.0
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/hsluv/hsluv-go v2.0.0+incompatible
	github.com/jkl1337/go-chromath v0.0.0-20140428033135-240283655afd
	github.com/jmacd/launchmidi v0.0.0-20200418073604-5904f9815af2
	github.com/jsimonetti/go-artnet v0.0.0-20200229173917-43ec7447138c
	github.com/lucasb-eyer/go-colorful v1.0.3
	gitlab.com/gomidi/midi v1.14.1
	go.opentelemetry.io/otel v0.4.2
	golang.org/x/image v0.0.0-20200119044424-58c23975cae1 // indirect
)
