// Program cram writes a bit file to the FPGA on a pico2-ice board,
// and executes it.
//
// The verilog code for the default FPGA loaded logic is in
// ../v/hello.v. See the README.md file for how to simulate and build
// it.
//
// The hello.v code expects a clock and a reset input. The code below
// holds the uploaded FPGA logic reset using using a GPIO. After the
// FPGA code is Cram-loaded, reset is de-asserted and the logic can
// run, flashing the blue FPGA LED. On the pico2-ice RP2350B GPIO29 is
// connected to the FPGA ICE11 and we use this pin for the (asserted
// low) reset signal.
//
// An alternate set of logic is provided in ../v/comm.v, which
// supports SPI communication over the same pins that the cram loading
// uses to initialize the FPGA. This logic listens to SPI and will set
// the FPGA LEDs based on the byte of data written to the FPGA
// device. This is what the SPI write of the color value causes. With
// the hello.v logic, this write is ignored by the FPGA, but it
// becomes active if comm.v is the source for the FPGA logic. The
// comm.v logic returns the prior value of the FPGA LEDs over SPI, and
// the cram.go code recognizes and displays 9 pairs of these values.
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

	bits := map[machine.Pin]byte{
		machine.LED_BLUE:  1 << 0,
		machine.LED_GREEN: 1 << 1,
		machine.LED_RED:   1 << 2,
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
		println("CramFPGA failed with error: ", err)
	}

	// Enable the logic
	reset.High()

	// Clear the RED RP LED after programming ended.
	machine.LED_RED.High()

	// Blink the different LEDs.
	println("Blinking the RP Tricolor LEDs randomly")
	data := make([]byte, 2)
	once := false
	for i := 0; ; i++ {
		// A new random order for leds.
		for led := range leds {
			led.Low()
			data[1] = bits[led]

			// ignored if using the hello.v logic.
			pico2ice.SPIxF(data, data)

			if !once {
				if data[1] == 0xff {
					println("FPGA program running (flashing FPGA Blue LED)")
				} else {
					println("FPGA program mirrors RP LEDs")
				}
				once = true
			}
			if i < 3 && data[1] != 0xff {
				// The FPGA logic responded with
				// something other than 0xff.
				println("wrote:", bits[led], " read:", data[1])
			}

			time.Sleep(time.Millisecond * 400)
			led.High()
			time.Sleep(time.Millisecond * 500)
		}
	}
}
