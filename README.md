# pico2ice - a tinygo package to support pico2ice features

## Overview

The [pico2-ice](https://pico2-ice.tinyvision.ai/) board features a
[RP2350B](https://datasheets.raspberrypi.com/rp2350/rp2350-datasheet.pdf)
microcontroller and a
[iCE40UP5K](https://www.latticesemi.com/en/Products/FPGAandCPLD/iCE40UltraPlus)
FPGA. This package provides some APIs for setting up these devices.

## Example program

Light up the pico2-ice onboard LEDs including the use of the
`hello.bin` FPGA bitstream logic.

- **IMPORTANT** the pico2-ice board ships with a default boot image
    that is not Tinygo compatible. The default image boots straight to
    MicroPython and skips the steps needed by Tinygo to intercept the
    boot. So, as you connect the USB cable from your computer to the
    board, you need to already have pressed and be holding down the
    `SW1` button (on the pico2 board, this button is labeled
    `BOOTSEL`). This is a little awkward to do in practice, but it
    should only be needed the first time. The bootloader re-program
    ready state is a single LED show "white" and once you see that,
    you can let go of the `SW1` button. The board is now patiently
    waiting for the `tinygo flash` command below.

Perform the following:
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
injected into the FPGA. When this is successful, a Green LED
illuminates to indicate that the FPGA is programmed. The `cram.go`
program then lets the FPGA logic run causing it to flash its Blue
LED. After this, and indefinitely, the cram.go code randomly flashes
one color at a time of the RP2350B connected Tricolor LED.

## The pio.go file

Part of the `pico2ice` package is generated code. The input for this
generation is the collection of `.pio` files in the `pio/` directory:

- [`pio/clock.pio`](pio/clock.pio) a 6-cycle clock output (pico2: 25 MHz)
- [`pio/spi.pio`](pio/spi.pio) a simple SPI read and write transfer loop (pico2 6.25 MHz)

To rebuild the `pio.go` file (assuming that you are in the
pico2ice/examples directory), do the following:

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

This shouldn't cause the file to change, but you can review if it has with:

```
$ git diff pio.go
```

## The hello.bin file

The `examples/hello.bin` file is an FPGA bitfile for the ice40 chip on
the pico2-ice board. It is built from some verilog file present in the
`v/` directory.

### Generating the hello.bin file

To convert `v/hello.v` into a `hello.bin` file we use the yosys
toolchain. Specifically the tools `yosys`, `nextpnr-ice40` and
`icepack`.

**NOTE** on Fedora, these tools are to be found in the `yosys`,
  `icestorm` and `nextpnr` packages. On Debian, these tools are to be
  found in the `yosys`, `fpga-icestorm` and `nextpnr` packages.

In the `v/` working directory do the following to rebuild the
`hello.bin` file:

```
$ yosys -p "synth_ice40 -top top -json hello.json" hello.v
$ nextpnr-ice40 --up5k  --package sg48 --json hello.json --pcf hello.pcf --asc hello.asc
$ icepack hello.asc hello.bin
```

- Note: the `hello.pcf` file is a script to map only the pins this
  `cram.go` + `hello.v` example uses. The full FPGA pinout file is
  referenced in a comment in that `hello.pcf` file, but conveniently
  [here
  too](https://github.com/tinyvision-ai-inc/pico-ice-sdk/blob/main/rtl/pico2_ice.pcf).

If you want to replace the default build of this file with the one
made above, you can do this:

```
$ cp hello.bin ../examples
```

### Simulating the hello.v logic

Typically, to simulate logic, you have some test wrapper for that
logic.  In our case we have the `hello_test.v` file. To perform that
test, we use the [Icarus
Verilog](https://steveicarus.github.io/iverilog/) suite. Most
specifically, the `iverilog` and `vvp` tools. We can also use GTKWave
(the `gtkwave` tool).

**NOTE** on Fedora and Debian, these tools are to be found in the
  `iverilog` and `gtkwave` packages.

To build the test binary (generate `a.out`) and run it do the
following:

```
$ iverilog hello_test.v hello.v
$ ./a.out
VCD info: dumpfile dump.vcd opened for output.
end of test
```

You can view this `dump.vcd` file with
[GTKWave](https://steveicarus.github.io/iverilog/usage/gtkwave.html),
and also with the [`twave`](https://zappem.net/pub/project/twave/)
tool. To only look at the top level signals we do the following:

```
$ ~/go/bin/twave --file dump.vcd --syms hello_test.top.clk,hello_test.reset,hello_test.red,hello_test.green,hello_test.blue
[] : [$version Icarus Verilog $end]
                hello_test.green-+
                  hello_test.red-|-+
                 hello_test.blue-|-|-+
                hello_test.reset-|-|-|-+
              hello_test.top.clk-|-|-|-|-+
                                 | | | | |
2025-11-08 15:29:26.000000000000 1 1 x x x
2025-11-08 15:29:26.000000010000 1 1 x 1 0
2025-11-08 15:29:26.000000020000 1 1 x 1 0
2025-11-08 15:29:26.000000030000 1 1 x 1 1
2025-11-08 15:29:26.000000040000 1 1 x 0 1
2025-11-08 15:29:26.000000050000 1 1 x 0 0
2025-11-08 15:29:26.000000060000 1 1 x 0 0
2025-11-08 15:29:26.000000070000 1 1 0 0 1
2025-11-08 15:29:26.000000080000 1 1 0 0 1
2025-11-08 15:29:26.000000090000 1 1 0 1 0
2025-11-08 15:29:26.000000100000 1 1 0 1 0
2025-11-08 15:29:26.000000110000 1 1 0 1 1
2025-11-08 15:29:26.000000120000 1 1 0 1 1
2025-11-08 15:29:26.000000130000 1 1 0 1 0
2025-11-08 15:29:26.000000140000 1 1 0 1 0
2025-11-08 15:29:26.000000150000 1 1 0 1 1
2025-11-08 15:29:26.000000160000 1 1 0 1 1
2025-11-08 15:29:26.000000170000 1 1 0 1 0
2025-11-08 15:29:26.000000180000 1 1 0 1 0
2025-11-08 15:29:26.000000190000 1 1 0 1 1
2025-11-08 15:29:26.000000200000 1 1 0 1 1
2025-11-08 15:29:26.000000210000 1 1 0 1 0
2025-11-08 15:29:26.000000220000 1 1 0 1 0
2025-11-08 15:29:26.000000230000 1 1 1 1 1
2025-11-08 15:29:26.000000240000 1 1 1 1 1
2025-11-08 15:29:26.000000250000 1 1 1 1 0
2025-11-08 15:29:26.000000260000 1 1 1 1 0
2025-11-08 15:29:26.000000270000 1 1 1 1 1
2025-11-08 15:29:26.000000280000 1 1 1 1 1
2025-11-08 15:29:26.000000290000 1 1 1 1 0
2025-11-08 15:29:26.000000300000 1 1 1 1 0
2025-11-08 15:29:26.000000310000 1 1 1 1 1
2025-11-08 15:29:26.000000320000 1 1 1 1 1
2025-11-08 15:29:26.000000330000 1 1 1 1 0
2025-11-08 15:29:26.000000340000 1 1 1 1 0
2025-11-08 15:29:26.000000350000 1 1 1 1 1
2025-11-08 15:29:26.000000360000 1 1 1 1 1
2025-11-08 15:29:26.000000370000 1 1 1 1 0
2025-11-08 15:29:26.000000380000 1 1 1 1 0
2025-11-08 15:29:26.000000390000 1 1 0 1 1
2025-11-08 15:29:26.000000400000 1 1 0 1 1
```

This list of bit values over time tracks the clock starting up, the
logic reset and eventually the Blue LED starting to toggle on and off.

## TODO

- Add some notes on installing tinygo. This should have been easy, but
  it took some extra packages, and you have to use the version that
  supports the pico2-ice board.

- Implement some verilog for using the FPGA's CRAM SPI pins. Then,
  after cramming the FPGA code, we can use the same pins to
  communicate with the configured FPGA logic. This will externalize
  (and fix/validate) the use of the `fpga.go:spiWR()` code.

## Support

This is a personal project aiming at exploring the capabilities of the
[pico2-ice](http://pico2-ice.tinyvision.ai/) board. As such, **no
support can be expected**. That being said, if you find a bug, or want
to request a feature, use the [github pico2ice bug
tracker](https://github.com/tinkerator/pico2ice/issues).

## License information

See the [LICENSE](LICENSE) file: the same BSD 3-clause license as that
used by [golang](https://golang.org/LICENSE) itself.
