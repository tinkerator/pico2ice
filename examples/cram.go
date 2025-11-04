// Program cram writes a bit file to the FPGA on a pico2-ice board,
// and executes it for 2 seconds waits 1 second and repeats.
package main

import (
	"context"
	_ "embed"
	"machine"
	"time"

	"zappem.net/pub/io/pico2ice"
)

//go:embed hello.bin
var fpgaBitstream []byte

func main() {
	ctx := context.Background()
	time.Sleep(2 * time.Second)

	println("Hello, World!")

	leds := map[machine.Pin]string{
		machine.LED_GREEN: "Green",
		machine.LED_RED:   "Red",
		machine.LED_BLUE:  "Blue",
	}

	// Clear all RP LEDs.
	for led := range leds {
		led.Configure(machine.PinConfig{Mode: machine.PinOutput})
		led.High()
	}

	pico2ice.Init()

	// Light the RED RP LED to indicate programming.
	machine.LED_RED.Low()
	if err := pico2ice.CramFPGA(ctx, fpgaBitstream); err != nil {
		println("CramFPGA returned: ", err)
	} else {
		println("FPGA program running (flashing FPGA Blue LED)")
	}

	// Clear the RED RP LED after programming ended.
	machine.LED_RED.High()

	// Blink the different LEDs.
	println("Blinking the RP Tricolor LEDs randomly")
	for {
		// A new random order for leds.
		for led := range leds {
			led.Low()
			time.Sleep(time.Millisecond * 400)
			led.High()
			time.Sleep(time.Millisecond * 500)
		}
	}
}
