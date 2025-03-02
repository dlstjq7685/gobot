package i2c

import (
	"fmt"
	"strings"
	"testing"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/aio"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/gobottest"
)

// this ensures that the implementation is based on i2c.Driver, which implements the gobot.Driver
// and tests all implementations, so no further tests needed here for gobot.Driver interface
var _ gobot.Driver = (*GrovePiDriver)(nil)

// must implement the DigitalReader interface
var _ gpio.DigitalReader = (*GrovePiDriver)(nil)

// must implement the DigitalWriter interface
var _ gpio.DigitalWriter = (*GrovePiDriver)(nil)

// must implement the AnalogReader interface
var _ aio.AnalogReader = (*GrovePiDriver)(nil)

// must implement the AnalogWriter interface
var _ aio.AnalogWriter = (*GrovePiDriver)(nil)

// must implement the Adaptor interface
var _ gobot.Adaptor = (*GrovePiDriver)(nil)

func initGrovePiDriverWithStubbedAdaptor() (*GrovePiDriver, *i2cTestAdaptor) {
	adaptor := newI2cTestAdaptor()
	return NewGrovePiDriver(adaptor), adaptor
}

func TestNewGrovePiDriver(t *testing.T) {
	var di interface{} = NewGrovePiDriver(newI2cTestAdaptor())
	d, ok := di.(*GrovePiDriver)
	if !ok {
		t.Errorf("NewGrovePiDriver() should have returned a *GrovePiDriver")
	}
	gobottest.Refute(t, d.Driver, nil)
	gobottest.Assert(t, strings.HasPrefix(d.Name(), "GrovePi"), true)
	gobottest.Assert(t, d.defaultAddress, 0x04)
	gobottest.Refute(t, d.pins, nil)
}

func TestGrovePiOptions(t *testing.T) {
	// This is a general test, that options are applied in constructor by using the common WithBus() option and
	// least one of this driver. Further tests for options can also be done by call of "WithOption(val)(d)".
	d := NewGrovePiDriver(newI2cTestAdaptor(), WithBus(2))
	gobottest.Assert(t, d.GetBusOrDefault(1), 2)
}

func TestGrovePiSomeRead(t *testing.T) {
	// arrange
	var tests = map[string]struct {
		usedPin          int
		wantWritten      []uint8
		simResponse      [][]uint8
		wantErr          error
		wantCallsRead    int
		wantResult       int
		wantResultF1     float32
		wantResultF2     float32
		wantResultString string
	}{
		"DigitalRead": {
			usedPin:       2,
			wantWritten:   []uint8{commandSetPinMode, 2, 0, 0, commandReadDigital, 2, 0, 0},
			simResponse:   [][]uint8{[]uint8{0}, []uint8{commandReadDigital, 3}},
			wantCallsRead: 2,
			wantResult:    3,
		},
		"AnalogRead": {
			usedPin:     3,
			wantWritten: []uint8{commandSetPinMode, 3, 0, 0, commandReadAnalog, 3, 0, 0},
			simResponse: [][]uint8{[]uint8{0}, []uint8{commandReadAnalog, 4, 5}},
			wantResult:  1029,
		},
		"UltrasonicRead": {
			usedPin:     4,
			wantWritten: []uint8{commandSetPinMode, 4, 0, 0, commandReadUltrasonic, 4, 0, 0},
			simResponse: [][]uint8{[]uint8{0}, []uint8{commandReadUltrasonic, 5, 6}},
			wantResult:  1281,
		},
		"FirmwareVersionRead": {
			wantWritten:      []uint8{commandReadFirmwareVersion, 0, 0, 0},
			simResponse:      [][]uint8{[]uint8{commandReadFirmwareVersion, 7, 8, 9}},
			wantResultString: "7.8.9",
		},
		"DHTRead": {
			usedPin:      5,
			wantWritten:  []uint8{commandSetPinMode, 5, 0, 0, commandReadDHT, 5, 1, 0},
			simResponse:  [][]uint8{[]uint8{0}, []uint8{commandReadDHT, 164, 112, 69, 193, 20, 174, 54, 66}},
			wantResultF1: -12.34,
			wantResultF2: 45.67,
		},
		"DigitalRead_error_wrong_return_cmd": {
			usedPin:     15,
			wantWritten: []uint8{commandSetPinMode, 15, 0, 0, commandReadDigital, 15, 0, 0},
			simResponse: [][]uint8{[]uint8{0}, []uint8{0, 2}},
			wantErr:     fmt.Errorf("answer (0) was not for command (1)"),
		},
		"AnalogRead_error_wrong_return_cmd": {
			usedPin:     16,
			wantWritten: []uint8{commandSetPinMode, 16, 0, 0, commandReadAnalog, 16, 0, 0},
			simResponse: [][]uint8{[]uint8{0}, []uint8{0, 3, 4}},
			wantErr:     fmt.Errorf("answer (0) was not for command (3)"),
		},
		"UltrasonicRead_error_wrong_return_cmd": {
			usedPin:     17,
			wantWritten: []uint8{commandSetPinMode, 17, 0, 0, commandReadUltrasonic, 17, 0, 0},
			simResponse: [][]uint8{[]uint8{0}, []uint8{0, 5, 6}},
			wantErr:     fmt.Errorf("answer (0) was not for command (7)"),
		},
		"FirmwareVersionRead_error_wrong_return_cmd": {
			wantWritten: []uint8{commandReadFirmwareVersion, 0, 0, 0},
			simResponse: [][]uint8{[]uint8{0, 7, 8, 9}},
			wantErr:     fmt.Errorf("answer (0) was not for command (8)"),
		},
		"DHTRead_error_wrong_return_cmd": {
			usedPin:     18,
			wantWritten: []uint8{commandSetPinMode, 18, 0, 0, commandReadDHT, 18, 1, 0},
			simResponse: [][]uint8{[]uint8{0}, []uint8{0, 164, 112, 69, 193, 20, 174, 54, 66}},
			wantErr:     fmt.Errorf("answer (0) was not for command (40)"),
		},
		"DigitalRead_error_wrong_data_count": {
			usedPin:     28,
			wantWritten: []uint8{commandSetPinMode, 28, 0, 0, commandReadDigital, 28, 0, 0},
			simResponse: [][]uint8{[]uint8{0}, []uint8{commandReadDigital, 2, 3}},
			wantErr:     fmt.Errorf("read count mismatch (3 should be 2)"),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			g, a := initGrovePiDriverWithStubbedAdaptor()
			g.Start()
			a.written = []byte{} // reset writes of former test and start
			numCallsRead := 0
			a.i2cReadImpl = func(bytes []byte) (i int, e error) {
				numCallsRead++
				copy(bytes, tc.simResponse[numCallsRead-1])
				return len(tc.simResponse[numCallsRead-1]), nil
			}
			var got int
			var gotF1, gotF2 float32
			var gotString string
			var err error
			// act
			switch {
			case strings.Contains(name, "DigitalRead"):
				got, err = g.DigitalRead(fmt.Sprintf("%d", tc.usedPin))
			case strings.Contains(name, "AnalogRead"):
				got, err = g.AnalogRead(fmt.Sprintf("%d", tc.usedPin))
			case strings.Contains(name, "UltrasonicRead"):
				got, err = g.UltrasonicRead(fmt.Sprintf("%d", tc.usedPin), 2)
			case strings.Contains(name, "FirmwareVersionRead"):
				gotString, err = g.FirmwareVersionRead()
			case strings.Contains(name, "DHTRead"):
				gotF1, gotF2, err = g.DHTRead(fmt.Sprintf("%d", tc.usedPin), 1, 2)
			default:
				t.Errorf("unknown command %s", name)
				return
			}
			// assert
			gobottest.Assert(t, err, tc.wantErr)
			gobottest.Assert(t, a.written, tc.wantWritten)
			gobottest.Assert(t, numCallsRead, len(tc.simResponse))
			gobottest.Assert(t, got, tc.wantResult)
			gobottest.Assert(t, gotF1, tc.wantResultF1)
			gobottest.Assert(t, gotF2, tc.wantResultF2)
			gobottest.Assert(t, gotString, tc.wantResultString)
		})
	}
}

func TestGrovePiSomeWrite(t *testing.T) {
	// arrange
	var tests = map[string]struct {
		usedPin     int
		usedValue   int
		wantWritten []uint8
		simResponse []uint8
	}{
		"DigitalWrite": {
			usedPin:     2,
			usedValue:   3,
			wantWritten: []uint8{commandSetPinMode, 2, 1, 0, commandWriteDigital, 2, 3, 0},
			simResponse: []uint8{4},
		},
		"AnalogWrite": {
			usedPin:     5,
			usedValue:   6,
			wantWritten: []uint8{commandSetPinMode, 5, 1, 0, commandWriteAnalog, 5, 6, 0},
			simResponse: []uint8{7},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			g, a := initGrovePiDriverWithStubbedAdaptor()
			g.Start()
			a.written = []byte{} // reset writes of former test and start
			a.i2cReadImpl = func(bytes []byte) (i int, e error) {
				copy(bytes, tc.simResponse)
				return len(bytes), nil
			}
			var err error
			// act
			switch name {
			case "DigitalWrite":
				err = g.DigitalWrite(fmt.Sprintf("%d", tc.usedPin), byte(tc.usedValue))
			case "AnalogWrite":
				err = g.AnalogWrite(fmt.Sprintf("%d", tc.usedPin), tc.usedValue)
			default:
				t.Errorf("unknown command %s", name)
				return
			}
			// assert
			gobottest.Assert(t, err, nil)
			gobottest.Assert(t, a.written, tc.wantWritten)
		})
	}
}

func TestGrovePi_getPin(t *testing.T) {
	gobottest.Assert(t, getPin("a1"), "1")
	gobottest.Assert(t, getPin("A16"), "16")
	gobottest.Assert(t, getPin("D3"), "3")
	gobottest.Assert(t, getPin("d22"), "22")
	gobottest.Assert(t, getPin("22"), "22")
}
