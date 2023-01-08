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
				case which[0] == 'b':
					cname = "blues"
				case which[0] == 'g':
					cname = "greens"
				}
				order := which[1]
				fmt.Printf("  %s.choose(J%d_%c, f, %d),\n", cname, i+1, order, gpio.bit)
			}
		}

		fmt.Printf(")\n")
	}

	return nil
}

func main() {
	if err := Main(); err != nil {
		log.Print("mkmap:", err)
		os.Exit(1)
	}
}
