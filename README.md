# pico2ice - a tinygo package to support pico2ice features

## Overview

The [pico2-ice](https://pico2-ice.tinyvision.ai/) board features a
[RP2350B](https://datasheets.raspberrypi.com/rp2350/rp2350-datasheet.pdf)
microcontroller and a
[iCE40UP5K](https://www.latticesemi.com/en/Products/FPGAandCPLD/iCE40UltraPlus)
FPGA. This package provides some APIs for setting up these devices.

## Example programs

Light up the pico2-ice onboard LEDs using the `hello.bin` FPGA
bitstream data:

```
$ git clone https://github.com/tinkerator/pico2ice.git
$ cd pico2ice/examples
$ ~/go/bin/tinygo flash -target=pico2-ice cram.go && ~/go/bin/tinygo monitor
Connected to /dev/ttyACM0. Press Ctrl-C to exit.
Hello, World!
FPGA program running (flashing FPGA Blue LED)
Blinking the RP Tricolor LEDs randomly
```

Once flashed, the RP2350B starts running and briefly flashes the Red
RP LED while the cram-load of the `hello.bin` FPGA bitstream is
injected into the FPGA. It then lets the FPGA run causing it to flash
its Blue LED. After this, the cram.go code randomly flashes one color
at a time of the RP2350B connected Tricolor LED.

## The pio.go file

Part of the `pico2ice` package is generated code. The input for this
generation is the collection of `.pio` files in the [pio/](pio)
directory:

- [`pio/clock.pio`](pio/clock.pio) a 6-cycle clock output (pico2: 25 MHz)
- [`pio/spi.pio`](pio/spi.pio) a simple SPI read and write transfer loop (pico2 6.25 MHz)

To rebuild the `pio.go` file (assuming that you are in the
pico2ice/examples directory):

```
$ cd ../../
$ git clone https://github.com/tinkerator/pious.git
$ cd pious
$ go install examples/piocli.go
```

This downloads the [`pious`](https://zappem.net/pub/io/pious) package
sources and installs the built tool in `~/go/bin/piocli`. Go back to
the `pico2ice` sources and run this tool:

```
$ cd ../pico2ice
$ ~/go/bin/piocli --src=pio/spi.pio,pio/clock.pio --name pico2ice --tinygo > pio.go
```

## TODO

- Add some notes on installing tinygo. This should have been easy, but
  it took some extra packages, and you have to use the version that
  supports the pico2-ice board.

- Add some notes on compiling, simulating verilog and generating
  bitfiles for the FPGA.

- Implement some verilog for using the FPGA's SPI pins so after
  cramming the FPGA code, we can use the same pins to communicate with
  the configured FPGA logic.

## Support

This is a personal project aiming at exploring the capabilities of the
[pico2-ice](http://pico2-ice.tinyvision.ai/) board. As such, **no
support can be expected**. That being said, if you find a bug, or want
to request a feature, use the [github pico2ice bug
tracker](https://github.com/tinkerator/pico2ice/issues).

## License information

See the [LICENSE](LICENSE) file: the same BSD 3-clause license as that
used by [golang](https://golang.org/LICENSE) itself.
