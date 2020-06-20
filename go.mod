module github.com/jmacd/nerve

replace github.com/jmacd/go-artnet => ../go-artnet

replace github.com/rakyll/portmidi => ../portmidi

replace github.com/jmacd/launchmidi => ../launchmidi

replace github.com/fogleman/gg => ../../fogleman/gg

go 1.14

require (
	github.com/faiface/pixel v0.9.0
	github.com/hsluv/hsluv-go v2.0.0+incompatible
	github.com/jkl1337/go-chromath v0.0.0-20140428033135-240283655afd
	github.com/jmacd/go-artnet v0.0.0-00010101000000-000000000000
	github.com/jmacd/launchmidi v0.0.0-20200418073604-5904f9815af2
	github.com/jsimonetti/go-artnet v0.0.0-20200229173917-43ec7447138c // indirect
	github.com/lucasb-eyer/go-colorful v1.0.3
	github.com/spf13/pflag v1.0.5
	gitlab.com/gomidi/midi v1.14.1
	golang.org/x/image v0.0.0-20200618115811-c13761719519
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
)
