// Package pico2ice supports programming the pico2-ice FPGA using the
// CRAM mechanism and also driving the FPGA clock pin at 75 MHz.
package pico2ice

import (
	"context"
	"errors"
	"fmt"
	"machine"
	"runtime"
	"time"

	pio "github.com/tinygo-org/pio/rp2-pio"
)

// ErrTimeout is used to terminate an upload operation.
var ErrTimeout = errors.New("context done (timeout)")

// e is our wrapper for pio.PIO0.
var e *Engine

// clk and spi are the supported state machines.
var clk, ice *StateMachine

func Init() {
	var err error
	e, err = Assign(pio.PIO0)
	if err != nil {
		panic(fmt.Sprint("failed to assign PIO0: ", err))
	}

	clk, err = e.ConfigureClock(machine.ICE35_G0)
	if err != nil {
		panic(fmt.Sprint("failed to create clk for FPGA: ", err))
	}
	clk.Start()
	clk.Activate(true)

	ice, err = e.ConfigureSpi(machine.ICE_SO /* FPGA's output pin (RP's input) */, machine.ICE_SI /* FPGA's input pin (RP's output) */, machine.ICE_CK)
	if err != nil {
		panic(fmt.Sprint("failed to create spi for FPGA: ", err))
	}
	ice.Start()
	ice.Activate(true)
}

// put writes a value to the ice.SM.TxFifo. It blocks until the write
// can complete.
func put(value uint32) {
	for {
		if !ice.SM.IsTxFIFOFull() {
			ice.SM.TxPut(value)
			return
		}
		runtime.Gosched()
	}
}

// spiWR sends data over the spi bus to the FPGA to send and
// (optionally, recv != nil) receive a byte stream. The slices send
// and recv can alias the same slice, in which case the read bytes
// overwrite the send buffer.
func spiWR(send, recv []byte) {
	go func() {
		bits := uint32(len(send)*8) - 1
		// Start by writing the number of bits to be sent
		put(bits)
		for _, b := range send {
			put(uint32(b) << 24)
		}
	}()
	keep := len(send) == len(recv)
	for i := 0; i < len(send); {
		if !ice.SM.IsRxFIFOEmpty() {
			b := ice.SM.RxGet()
			if keep {
				recv[i] = byte(b)
			}
			i++
			continue
		}
		runtime.Gosched()
	}
}

// SPIxF activates the SPI interface to the FPGA by lowering
// chip-select and performs an SPI transfer with the FPGA.
func SPIxF(send, recv []byte) {
	fpgaCS := machine.ICE_SSN
	fpgaCS.Configure(machine.PinConfig{Mode: machine.PinOutput})
	fpgaCS.Low()
	spiWR(send, recv)
	fpgaCS.High()
}

// CramFPGA resets the FPGA and uploads the data bit file.
func CramFPGA(ctx context.Context, data []byte) error {
	fpgaReset := machine.FPGA_RSTN
	fpgaCS := machine.ICE_SSN
	iceDone := machine.ICE_DONE

	StartFPGA()

	fpgaReset.Configure(machine.PinConfig{Mode: machine.PinOutput})
	fpgaCS.Configure(machine.PinConfig{Mode: machine.PinOutput})
	iceDone.Configure(machine.PinConfig{Mode: machine.PinInput})

	// Reset FPGA and set fpgaCS low.
	fpgaReset.Low()
	fpgaCS.Low()
	time.Sleep(2 * time.Microsecond)
	fpgaReset.High()

	// At least 1200us for FPGA to fully reset.
	time.Sleep(1300 * time.Microsecond)

	// 8 SCLK cycles with fpgaCS high
	var junk [7]byte
	fpgaCS.High()
	spiWR(junk[0:1], nil)

	fpgaCS.Low()
	spiWR(data, nil)
	fpgaCS.High()

	// Spin waiting for FPGA programming to be done.
	for {
		if iceDone.Get() {
			break
		}
		select {
		case <-ctx.Done():
			return ErrTimeout
		case <-time.After(1 * time.Millisecond):
		}
	}

	// 49 SCLK after iceDone goes high
	spiWR(junk[0:7], nil)

	return nil
}

// StartFPGA begins the FPGA logic clock.
func StartFPGA() {
	clk.Activate(true)
}

// StartFPGA stops the FPGA logic clock.
func StopFPGA() {
	clk.Activate(false)
}
