package data

const WelcomeText = `Open-Mic Game Night Welcome!
......
THIS is a game, you will see,
It runs on creativity.
Press the buttons, you will get,
a view into the Mandelbrot set.
Picking colors, many ways,
even Caspar Mayonnaise.
Want to learn how this thing emits?
(Color depth at least six bits?)
Won't you agree, it's a fine spot?
Does anyone want to build a robot?
......
If you feel this way as well,
come on now, Show and Tell.
`

const Manifesto = `
Hackers Unite!
...
A "Hack Night" invitation.
...
Are you making something unique, or interesting?
Interested in technology, computers, electronics?
Tell us why and show us!  Come and see?
...
Does anyone want to build a modular synthesizer?
Radio control car, boat, plane, other thing?
Radio station?
Reverse oscilliscope?
...
I turn MIDI signals into flashy lights.
...
2D game development.
Pacman vs Blinky.
...
4-20mA current loop vs Modbus.
...
Do "Zephyr and Micropython" mean anything to you?
Or, two BeagleConnect Freedom devices and no idea what I'm doing.
...
Cross compiling with Bazel :emoji:
...
The Caspar Water open source software stack
...
Do you have a logic analyzer?
The other six panels want to talk.
...
Autonomous Scarecrow, and other projects.
...
`

const Technical = `THIS device is a Beaglebone Black (BBB)
single board computer, an open hardware design based on the Texas
Instruments am3358 "system on a chip" (SoC), it's an $80 computer with
2x46 pins for connecting "capes" , which are 3rd party expansion
boards.

The Beaglebone Black is in the same category as the Rasperry Pi, but
with a more industrial focus.  The thing about this am335x chip and
its relatives, what sets it apart for this this application, is a
secondary micro controller called the Programmable Realtime Unit, or
PRU.

There are two PRU micro controllers on the Beaglebone Black.  Unlike the
main CPU on the device, the PRU executes a single instruction at a
time, with a predictable five nanosecond cycle time.  That's 200
megahertz.  PRUs can be programmed in C or assembly code, but firmware
"text" is limited to 8 kilobytes, so PRU programs must be extremely
small.

The two PRUs share access to 20 kilobytes of local memory, which can
be read in several deterministic cycles, and the PRUs have access to
other parts of the system on a chip, such as the general purpose I/O
controller, the pulse width modulation controller, the analog to
digital controller, and so on.

Here, the cape is a variant of the "Octoscroller", which supports up
to 8 LED panels.  These panels use an interface called "HUB75".
Ordinarily, they would be used to make "video walls", with panels
controlled by custom hardware.  These panels are 32x64 pixels each,
and there are two in the current installation, for a total of 4096
pixels.  There ought to be four times as many, 16384 pixels, but after
two panels something electrical goes wrong.

The panels feature simple red, green, and blue LEDs, one of each color
per pixel.  By modulating pins of the HUB75 connector, an application
can set each pixel off or on at full brightness.  The panel uses what
is called a "1/16 scan" design, meaning at any moment in time, only
one out of 16 pixels is actually on.  The appearance of brightness and
color control is because our eyes are slow and the PRU is fast.

The am3358 organizes its general purpose output pins in four 32-bit
"banks", meaning it can modulate all pins by writing only four 32-bit
"words" to four special addresses.

Each 32x64 panel is organized as a 16 pairs of 64-long scan lines
a.k.a. "shift registers".  There are four address lines to select the
current scan line, CLOCK, LATCH, ENABLE pins, and six pins
corresponding with the pair of red, green, and blue pixels.

The hardware or software is meant to rapidly cycle through each of the
16 scan line pairs, writing 64 pixels pairs per scan line per pass.
Imagine there are eight panels, not two.  It takes four 32-bit words
to encode a pair of pixel values for all panels, or 16 bytes.  It
takes 64 of those to complete a single scan line, or 1024 bytes.  It
takes 16 of those to complete a whole frame, which is 16 kilobytes.
But with only 6144 of information per frame, the frame has 37.5%
information density.

To generate colors other than red, green, and blue, the program has to
repeatedly draw frames that average to the target color at each pixel.
To accomplish this, a program running on the main processor (a 1 GhZ
ARM chip) computes a sequence of 256 frames, occupying four megabytes.
In order to facilitate smooth updates, it "double buffers" into two of
these. In total, there are 8 megabytes of dedicated video memory.

The PRU is able to write approximately 1600 frames per second, but
this wasn't easy.  The PRU has only a small amount of memory that with
fast, predictable access.  For the panel to project uniform
brightness, we need latency to be predictable.

The PRU program sets up two small buffers in local memory and
configures special hardware called the Direct Memory Access (DMA)
controller to copy one section of a time into alternating buffers.
While the PRU writes the current buffer in local memory, the DMA fills
the next local buffer from main memory.  You'll notice there is some
glitchiness -- it gets worse the busier the main CPU is -- this
results from contention on the memory bus, and it can probably be
fixed with better DMA programming.

The animation is generated using an open-source software renderer with
a simple program written in Golang.  Software rendering isn't fast
enough for most purposes, but at least we're not talking about OpenGL
here.

The attached MIDI device simply provides knobs and sliders to the
application, though it brings in some gnarly dependencies.  TL;DR
"libalsa".  Ugh.  This had to be cross compiled on a bigger machine,
for silly reasons involving Cmake.

The ARM chip is not very powerful, compared with my Apple MacBook,
currently running the animation.  A protocol called ArtNet is used to
send frames from the MacBook to the BeagleBone, just something I had
used in the past.

There you have it -- its name is "nerve" -- find me at github.com
jmacd/nerve.
`
