module github.com/jmacd/nerve

go 1.18

require (
	github.com/hsluv/hsluv-go v2.0.0+incompatible
	github.com/jkl1337/go-chromath v0.0.0-20140428033135-240283655afd
	github.com/jmacd/go-artnet v0.0.0-20220707060336-6bfd9f54a67f
	github.com/jmacd/launchmidi v0.0.0-20221203062954-9a79ac9cf609
	github.com/lucasb-eyer/go-colorful v1.0.3
)

require gitlab.com/gomidi/midi/v2 v2.0.25 // indirect

replace github.com/jmacd/nerve/pru => ./pru