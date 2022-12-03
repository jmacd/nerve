module github.com/jmacd/nerve

// replace github.com/jmacd/go-artnet => ../go-artnet
// replace github.com/rakyll/portmidi => ../portmidi
// replace github.com/jmacd/launchmidi => ../launchmidi
// replace github.com/fogleman/gg => ../../fogleman/gg

go 1.18

require (
	github.com/hsluv/hsluv-go v2.0.0+incompatible
	github.com/jkl1337/go-chromath v0.0.0-20140428033135-240283655afd
	github.com/jmacd/go-artnet v0.0.0-20220707060336-6bfd9f54a67f
	github.com/jmacd/launchmidi v0.0.0-20200418073604-5904f9815af2
	github.com/lucasb-eyer/go-colorful v1.0.3
)

require github.com/rakyll/portmidi v0.0.0-20191102002215-74e95e8bc9b1 // indirect
