package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	pinFile = "bbb-pins.csv"
	mapFile = "fpp-octo.json"
)

type gpioNum struct {
	bank int
	bit  int
}

type capePinType struct {
	Pin  string `json:"pin"`
	Type string `json:"type"`
}

type capeOutput struct {
	Pins map[string]string `json:"pins"`
}

type cape struct {
	Name     string `json:"name"`
	LongName string `json:"longName"`
	PRU      int    `json:"pru"`
	Timing   int    `json:"timing"`
	Controls struct {
		GPIO  int         `json:"gpio"`
		Latch capePinType `json:"latch"`
		OE    capePinType `json:"oe"`
		Clock capePinType `json:"clock"`
		Sel0  capePinType `json:"sel0"`
		Sel1  capePinType `json:"sel1"`
		Sel2  capePinType `json:"sel2"`
		Sel3  capePinType `json:"sel3"`
		Sel4  capePinType `json:"sel4"`
	} `json:"controls"`
	Outputs []capeOutput `json:"outputs"`
}

func ParsePins() (map[string]gpioNum, error) {
	pin2gpio := map[string]gpioNum{}

	f, err := os.Open(pinFile)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", pinFile, err)
	}
	pins, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read csv %s: %w", pinFile, err)
	}
	for _, row := range pins {
		ps := row[0]
		gs := row[9]
		if !strings.HasPrefix(gs, "gpio") {
			continue
		}
		pinAdd := 0
		if strings.HasSuffix(ps, ".1") {
			ps = ps[:len(ps)-2]
			// This +50 convention is used in the fpp config file.
			pinAdd = 50
		}
		hdr := ps[:2]
		pinNum, err := strconv.Atoi(ps[3:])
		if err != nil {
			return nil, fmt.Errorf("invalid pin: %s", ps)
		}
		pin := fmt.Sprintf("%s-%02d", hdr, pinNum+pinAdd)

		bank := int(gs[4] - '0')
		if bank < 0 || bank > 3 {
			return nil, fmt.Errorf("invalid gpio: %s", gs)
		}
		bit, err := strconv.Atoi(gs[6:])
		if err != nil {
			return nil, fmt.Errorf("invalid gpio: %s", gs)
		}

		pin2gpio[pin] = gpioNum{
			bank: bank,
			bit:  bit,
		}
	}
	return pin2gpio, nil
}

func ParseCape() (cape, error) {
	var c cape
	ccfg, err := os.ReadFile(mapFile)
	if err != nil {
		return c, err
	}

	if err := json.Unmarshal(ccfg, &c); err != nil {
		return c, err
	}
	return c, nil
}

func Main() error {
	p2g, err := ParsePins()
	if err != nil {
		return err
	}
	c, err := ParseCape()
	if err != nil {
		return err
	}

	for i, output := range c.Outputs {
		for which, pin := range output.Pins {
			gpio, ok := p2g[pin]
			if !ok {
				return fmt.Errorf("unknown pin: J%d-%s: %s", i+1, which, pin)
			}

			fmt.Printf("J%d %s gpio%d/%d\n", i+1, which, gpio.bank, gpio.bit)
		}
	}

	for bank := 0; bank < 4; bank++ {
		fmt.Printf("dp.Gpio%d = combine(\n", bank)

		for i, output := range c.Outputs {
			for which, pin := range output.Pins {
				gpio := p2g[pin]
				if gpio.bank != bank {
					continue
				}
				var cname string
				switch {
				case which[0] == 'r':
					cname = "reds"
				case which[0] == 'g':
					cname = "greens"
				case which[0] == 'b':
					cname = "blues"
				}
				order := which[1]
				fmt.Printf("  %s.choose(J%d_%c, f, %d),\n", cname, i+1, order, gpio.bit)
			}
		}
		fmt.Printf(")\n")
	}

	for i, output := range c.Outputs {
		for which, pin := range output.Pins {
			gpio := p2g[pin]

			var add int
			switch {
			case which[0] == 'r':
				add = 0
			case which[0] == 'g':
				add = 1
			case which[0] == 'b':
				add = 2
			}

			fmt.Printf("  if dp.Gpio%d & (1<<%d) != 0 {\n", gpio.bank, gpio.bit)
			fmt.Printf("     add4(&img.Pix[j%d%cOff+%d])\n", i+1, which[1], add)
			fmt.Printf("  }\n")
		}
	}

	for i, output := range c.Outputs {
		fmt.Println("")
		for _, which := range []string{"r1", "g1", "b1", "r2", "g2", "b2"} {
			pin := output.Pins[which]
			gpio := p2g[pin]

			value := 0
			if which[0] == 'b' {
				value = 1
			}
			fmt.Printf("      pixptr->gpv%d.bits.j%d_%s = %d;\n", gpio.bank, i+1, which, value)
		}
	}

	for i, output := range c.Outputs {
		fmt.Println("")

		for _, which := range []string{"r1", "g1", "b1", "r2", "g2", "b2"} {
			pin := output.Pins[which]
			gpio := p2g[pin]

			value := map[string]string{
				"r1": "1 ^ quad",
				"g1": "0",
				"b1": "0 ^ quad",
				"r2": "0",
				"g2": "0 ^ quad",
				"b2": "1 ^ quad",
			}[which]
			fmt.Printf("      pixptr->gpv%d.bits.j%d_%s = %s;\n", gpio.bank, i+1, which, value)
		}
	}

	for bank := 0; bank < 4; bank++ {
		fmt.Println(`
typedef union {
  volatile uint32_t word;

  volatile struct {`)
		for bit := 0; bit < 32; bit++ {
			found := false
			for i, output := range c.Outputs {
				for which, pin := range output.Pins {
					gpio := p2g[pin]
					if gpio.bank != bank {
						continue
					}
					if gpio.bit != bit {
						continue
					}
					fmt.Printf("    unsigned j%d_%s : 1; // %d\n", i+1, which, bit)
					found = true
				}
			}
			if !found {
				fmt.Printf("    unsigned _bit%d : 1; // %d\n", bit, bit)
			}
		}

		fmt.Printf("  } bits;\n")
		fmt.Printf("} gpio%d_t;\n", bank)
	}

	return nil
}

func main() {
	if err := Main(); err != nil {
		log.Print("mkmap:", err)
		os.Exit(1)
	}
}
