module github.com/jmacd/nerve

replace github.com/jmacd/go-artnet => ../go-artnet

replace gitlab.com/gomidi/midi => ../../../gitlab.com/gomidi/midi

go 1.13

require (
	github.com/deadsy/libusb v0.0.0-20180330230923-04d6a756f531
	github.com/go-gl/example v0.0.0-20191129173604-c307114f3462 // indirect
	github.com/go-gl/gl v0.0.0-20190320180904-bf2b1f2f34d7 // indirect
	github.com/go-gl/glfw/v3.3/glfw v0.0.0-20200222043503-6f7a984d4dc4 // indirect
	github.com/go-gl/mathgl v0.0.0-20190713194549-592312d8590a // indirect
	github.com/hsluv/hsluv-go v2.0.0+incompatible
	github.com/jkl1337/go-chromath v0.0.0-20140428033135-240283655afd
	github.com/jsimonetti/go-artnet v0.0.0-20200229173917-43ec7447138c
	github.com/lucasb-eyer/go-colorful v1.0.3
	gitlab.com/gomidi/midi v1.14.1
	go.opentelemetry.io/otel v0.2.3
	golang.org/x/image v0.0.0-20200119044424-58c23975cae1 // indirect
)
