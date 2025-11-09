// Program cram writes a bit file to the FPGA on a pico2-ice board,
// and executes it.
//
// The verilog code for the FPGA loaded logic is in ../v/hello.v. See
// the README.md file for how to simulate and build it.
//
// The hello.v code expects a clock and a reset input. The code below
// holds the uploaded FPGA logic reset using using a GPIO. After the
// FPGA code is Cram-loaded, reset is de-asserted and the logic can
// run, flashing the blue FPGA LED. On the pico2-ice RP2350B GPIO29 is
// connected to the FPGA ICE11 and we use this pin for the (asserted
// low) reset signal.
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

	// The logic reset pin.
	reset := machine.ICE11
	machine.ICE11.Configure(machine.PinConfig{Mode: machine.PinOutput})
	// Prevent the logic from running (when we get there).
	reset.Low()

	pico2ice.Init()

	// Light the RED RP LED to indicate programming.
	machine.LED_RED.Low()

	if err := pico2ice.CramFPGA(ctx, fpgaBitstream); err != nil {
		println("CramFPGA returned: ", err)
	} else {
		println("FPGA program running (flashing FPGA Blue LED)")
	}

	// Enable the logic
	reset.High()

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
