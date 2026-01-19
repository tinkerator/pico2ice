# pico2ice - a tinygo package to support pico2ice features

## Overview

The [pico2-ice](https://pico2-ice.tinyvision.ai/) board features a
[RP2350B](https://datasheets.raspberrypi.com/rp2350/rp2350-datasheet.pdf)
microcontroller and a
[iCE40UP5K](https://www.latticesemi.com/en/Products/FPGAandCPLD/iCE40UltraPlus)
FPGA. This [package](https://zappem.net/pub/io/pico2ice/) provides
some APIs for setting up these devices.

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
    ready state is a single LED shows "white" (all three RGB LEDs lit)
    and once you see that, you can let go of the `SW1` button. The
    board is now patiently waiting for the `tinygo flash` command
    below.

- **NOTE** this README assumes that `~/go/bin` is in your `$PATH`
    variable. Built tools like `piocli` and `twave` will be installed
    there. Also, in the case of Fedora, where we build `tinygo` from
    source, we will install that tool there too.

Perform the following:
```
$ git clone https://github.com/tinkerator/pico2ice.git
$ cd pico2ice/examples
$ tinygo flash -target=pico2-ice -scheduler tasks cram.go && tinygo monitor
Connected to /dev/ttyACM0. Press Ctrl-C to exit.
Hello, World!
Blinking the RP Tricolor LEDs randomly
FPGA program running (flashing FPGA Blue LED)
```

Once flashed, the RP2350B starts running and briefly flashes the Red
RP LED while the cram-load of the `hello.bin` FPGA bitstream is
injected into the FPGA. When this is successful, a Green LED
illuminates to indicate that the FPGA is programmed. The `cram.go`
program then lets the FPGA logic run causing it to rapidly flash its
Blue LED. After this, and indefinitely, the cram.go code randomly
flashes one color at a time of the RP2350B connected Tricolor LED.

## The pio.go file

Part of the `pico2ice` package is generated code. The input for this
generation is the collection of `.pio` files in the `pio/` directory:

- [`pio/clock.pio`](pio/clock.pio) a 2-cycle clock output (pico2: 75 MHz)
- [`pio/spi.pio`](pio/spi.pio) a simple SPI read and write transfer loop (pico2 6.25 MHz)

To rebuild the `pio.go` file (assuming that you are in the
`pico2ice/examples` directory), do the following:

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
$ piocli --src=pio/spi.pio,pio/clock.pio --name pico2ice --tinygo > pio.go
```

This shouldn't cause the file to change, but you can review if it has with:

```
$ git diff pio.go
```

## The hello.bin file

The `examples/hello.bin` file is an FPGA bitfile for the ice40 chip on
the pico2-ice board. It is built from a verilog file present in the
`v/` directory.

### Generating the hello.bin file

To convert `v/hello.v` into a `hello.bin` file we use the `yosys`
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

- Note: the `hello.pcf` file is a script to map the pins these
  `cram.go` + `hello.v` (and `comm.v`) examples use. The full FPGA
  pinout file is referenced in a comment in that `hello.pcf` file, and
  conveniently [here
  too](https://github.com/tinyvision-ai-inc/pico-ice-sdk/blob/main/rtl/pico2_ice.pcf).

If you want to replace the default build of this file with the one
made above, you can do this:

```
$ cp hello.bin ../examples
```

### comm.v

A more ambitious verilog file is `v/comm.v`. This pays attention to
runtime attempts by the `cram.go` code to control the FPGA LED from
the RP2350B microcontroller program. Namely, its attempts to set the
FPGA LED with an SPI command from the microcontroller code. To build
and install this FPGA bitfile, `cd v/` and run the following
commands:

```
$ yosys -p "synth_ice40 -top top -json hello.json" comm.v
$ nextpnr-ice40 --up5k  --package sg48 --json hello.json --pcf hello.pcf --asc hello.asc
$ icepack hello.asc hello.bin
$ cp hello.bin ../examples
```

Then, from the `../examples` directory, this command will generate
some output like this (9 wrote + read values will be displayed):

```
$ tinygo flash -target=pico2-ice -scheduler tasks cram.go && tinygo monitor
Connected to /dev/ttyACM0. Press Ctrl-C to exit.
Hello, World!
Blinking the RP Tricolor LEDs randomly
FPGA program mirrors RP LEDs
wrote: 2  read: 0
wrote: 4  read: 2
wrote: 1  read: 4
wrote: 2  read: 1
wrote: 4  read: 2
wrote: 1  read: 4
wrote: 2  read: 1
wrote: 4  read: 2
wrote: 1  read: 4
```

and will cause the FPGA to mirror the last set microcontroller LED
color. The 9 read values here should match the prior line's write
value. The 0 initially read value is just reflecting the fact that the
reset value is 0 for the 3 bit FPGA LED setting.

- **NOTE**: this example requires the FPGA logic is running 8x or
    better the speed of the SPI clock. Any slower and the SPI returned
    result will be late by 1-bit - returning results that are 2x
    smaller than expected.

### Simulating the `hello.v` logic

Typically, to simulate logic, you have some test wrapper for that
logic.  In the `hello.v` case we have included the `hello_test.v`
file. To perform that test, we use the [Icarus
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
tool. To only look at the top level signals with that text based tool
we can do the following:

```
$ twave --file dump.vcd --syms hello_test.top.clk,hello_test.reset,hello_test.red,hello_test.green,hello_test.blue
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

## Installing the developer version of tinygo

The default `tinygo` package might be very old (pico2-ice support
debuted in release v0.40.0), and not include support for the pico2-ice
board. To remedy this, you can install a fresh compile of the `tinygo`
source tree and build it yourself (see below).

On Debian, and Mac, the `tinygo` releases have some prebuilt packages
that you can install:
https://github.com/tinygo-org/tinygo/releases. For Debian, to install
the `.deb` package for your architecture, use `sudo dpkg -i
tinygo_*.deb`.

On Fedora, install some packages (the first expands to most of the
needed dependencies):

```
$ sudo dnf install tinygo clang-devel
```

Then:

```
$ git clone https://github.com/tinygo-org/tinygo.git
$ cd tinygo/
$ git submodule update --init --recursive
$ make llvm-source
$ make gen-device
```

Note, if you want even more recent code support, you can try
performing the last 3 lines above after `git checkout dev`.

The final step depends on which Linux distribution version you have:

| Linux distribution | command |
|-----------|--------|
| Fedora 41 | `$ go install --tags llvm19` |
| Fedora 42 | `$ go install --tags llvm20` |
| Fedora 43 | `$ go install --tags llvm21` |

Then, to verify that the program installed correctly:

```
$ tinygo version
```

## TODO

- Currently, the tscan sub-package uses bit banging. Might investigate
  using PIO for this sort of transfer, to support faster transfer
  speeds.

- Currently, the tinygo code cannot take advantage of the dual core
  nature of the RP2350B chip because of [tinygo bug
  4974](https://github.com/tinygo-org/tinygo/issues/4974) there has
  also [been some issue with
  llvm21](https://github.com/tinygo-org/tinygo/issues/5086). At some
  point those problems will be fixed and these instructions will drop
  the use of the `-scheduler tasks` compilation flags.

## Support

This is a personal project aiming at exploring the capabilities of the
[pico2-ice](http://pico2-ice.tinyvision.ai/) board. As such, **no
support can be expected**. That being said, if you find a bug, or want
to request a feature, use the [github pico2ice bug
tracker](https://github.com/tinkerator/pico2ice/issues).

## License information

See the [LICENSE](LICENSE) file: the same BSD 3-clause license as that
used by [golang](https://golang.org/LICENSE) itself.
