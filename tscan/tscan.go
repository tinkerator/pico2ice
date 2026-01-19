// Package tscan implements support for a simple boundary scan
// interface. This is intended to provide some handy way of scanning
// data into our out of connected logic that supports such a boundary
// scan.
package tscan

import (
	"errors"
	"machine"
	"runtime"
)

// TScan provides a basic interface to a boundary scan API.
type TScan struct {
	// en (active low enable), clk (clock), in (serial in to
	// device), out (serial out from device).
	en, clk, in, out machine.Pin
}

// New can be used to define and initiate a new test scan
// interface. The interface is essentially the jtag boundary scan
// 4-pin interface.
//
// The signals en, clk and in are RP outputs whose safe default is
// High(). The signal out is output _from_ the target device. Calling
// this function will initiate all of these signals to this safe
// state.
func New(en, clk, in, out machine.Pin) *TScan {
	clk.Configure(machine.PinConfig{Mode: machine.PinOutput})
	clk.High()
	in.Configure(machine.PinConfig{Mode: machine.PinOutput})
	in.High()
	out.Configure(machine.PinConfig{Mode: machine.PinInput})
	en.Configure(machine.PinConfig{Mode: machine.PinOutput})
	en.High()

	return &TScan{
		en:  en,
		clk: clk,
		in:  in,
		out: out,
	}
}

var ErrBadBuffer = errors.New("bad send/recv buffer combination")

// Xfer scans bits worth of the send buffer into the TScan
// interface. Note, the last bit in send is the most shallow in the
// scan chain. If recv is nil, no attempt is made to read the
// output. If the send buffer is nil, the recv is read, and cycled
// into the TScan interface - a circular scan of length bits. One of
// send and recv must not be nil.
func (ts *TScan) Xfer(bits int, send, recv []byte) error {
	bs := (bits + 7) / 8
	offset := bits % 8
	if offset == 0 {
		offset = 8
	}
	if send == nil && recv == nil {
		return ErrBadBuffer
	}
	if send != nil && len(send) < bs {
		return ErrBadBuffer
	}
	if recv != nil && len(recv) < bs {
		return ErrBadBuffer
	}
	for i := 0; i < bs; i++ {
		var u8 byte
		if send != nil {
			u8 = send[i]
		}
		var out8 byte
		for offset > 0 {
			offset--
			var bit byte
			ts.clk.Low()
			if ts.out.Get() {
				bit = 1
			}
			out8 = (out8 << 1) | bit
			if send == nil {
				if bit != 0 {
					ts.in.High()
				} else {
					ts.in.Low()
				}
			} else {
				if u8&(1<<offset) != 0 {
					ts.in.High()
				} else {
					ts.in.Low()
				}
			}
			ts.clk.High()
			runtime.Gosched()
		}
		if recv != nil {
			recv[i] = out8
		}
		offset = 8
	}
	return nil
}

// Enable engages or disengages an un-Closed() TScan interface.
func (ts *TScan) Enable(on bool) {
	ts.en.Set(!on)
}

// Close closes the interface and configures all of its pins as
// inputs. You will need to call New() again to obtain a usable TScan
// interface.
func (ts *TScan) Close() error {
	ts.Enable(false)
	ts.en.Configure(machine.PinConfig{Mode: machine.PinInput})
	ts.in.Configure(machine.PinConfig{Mode: machine.PinInput})
	ts.clk.Configure(machine.PinConfig{Mode: machine.PinInput})
	ts.en = machine.NoPin
	ts.clk = machine.NoPin
	ts.in = machine.NoPin
	ts.out = machine.NoPin
	return nil
}
