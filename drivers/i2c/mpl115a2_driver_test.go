package i2c

import (
	"errors"
	"strings"
	"testing"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/gobottest"
)

// this ensures that the implementation is based on i2c.Driver, which implements the gobot.Driver
// and tests all implementations, so no further tests needed here for gobot.Driver interface
var _ gobot.Driver = (*MPL115A2Driver)(nil)

func initTestMPL115A2DriverWithStubbedAdaptor() (*MPL115A2Driver, *i2cTestAdaptor) {
	a := newI2cTestAdaptor()
	return NewMPL115A2Driver(a), a
}

func TestNewMPL115A2Driver(t *testing.T) {
	var di interface{} = NewMPL115A2Driver(newI2cTestAdaptor())
	d, ok := di.(*MPL115A2Driver)
	if !ok {
		t.Errorf("NewMPL115A2Driver() should have returned a *MPL115A2Driver")
	}
	gobottest.Refute(t, d.Connection(), nil)
	gobottest.Assert(t, strings.HasPrefix(d.Name(), "MPL115A2"), true)
	gobottest.Assert(t, d.defaultAddress, 0x60)
}

func TestMPL115A2Options(t *testing.T) {
	// This is a general test, that options are applied in constructor by using the common WithBus() option and
	// least one of this driver. Further tests for options can also be done by call of "WithOption(val)(d)".
	d := NewMPL115A2Driver(newI2cTestAdaptor(), WithBus(2))
	gobottest.Assert(t, d.GetBusOrDefault(1), 2)
}

func TestMPL115A2ReadData(t *testing.T) {
	// sequence for read data
	// * retrieve the coefficients for temperature compensation of pressure - see test for Start()
	// * write start conversion register address (0x12)
	// * write start value - 0x00
	// * wait at least 3 ms according to data sheet (tc - conversion time)
	// * write pressure MSB register address (0x00)
	// * read pressure (16 bit, order MSB-LSB)
	// * read temperature (16 bit, order MSB-LSB)
	// * calculate temperature compensated pressure in kPa according to data sheet
	//   * shift the temperature value right for 6 bits (resolution is 10 bit)
	//   * shift the pressure value right for 6 bits (resolution is 10 bit)
	// * calculate temperature in °C according to this implementation:
	//   https://github.com/adafruit/Adafruit_MPL115A2/blob/master/Adafruit_MPL115A2.cpp
	//
	// arrange
	d, a := initTestMPL115A2DriverWithStubbedAdaptor()
	d.Start()
	a.written = []byte{}
	// arrange coefficients according the example from data sheet
	d.a0 = 2009.75
	d.b1 = -2.37585
	d.b2 = -0.92047
	d.c12 = 0.00079
	readReturnP := []byte{0x66, 0x80, 0x7E, 0xC0} // use example from data sheet
	readReturnT := []byte{0x00, 0x00, 0x7E, 0xC0} // use example from data sheet
	readCallCounter := 0
	a.i2cReadImpl = func(b []byte) (int, error) {
		readCallCounter++
		if readCallCounter == 1 {
			copy(b, readReturnP)
		}
		if readCallCounter == 2 {
			copy(b, readReturnT)
		}
		return len(b), nil
	}

	// act
	press, errP := d.Pressure()
	temp, errT := d.Temperature()
	// assert
	gobottest.Assert(t, errP, nil)
	gobottest.Assert(t, errT, nil)
	gobottest.Assert(t, readCallCounter, 2)
	gobottest.Assert(t, len(a.written), 6)
	gobottest.Assert(t, a.written[0], uint8(0x12))
	gobottest.Assert(t, a.written[1], uint8(0x00))
	gobottest.Assert(t, a.written[2], uint8(0x00))
	gobottest.Assert(t, a.written[3], uint8(0x12))
	gobottest.Assert(t, a.written[4], uint8(0x00))
	gobottest.Assert(t, a.written[5], uint8(0x00))
	gobottest.Assert(t, press, float32(96.585915))
	gobottest.Assert(t, temp, float32(23.317757))
}

func TestMPL115A2ReadDataError(t *testing.T) {
	d, a := initTestMPL115A2DriverWithStubbedAdaptor()
	d.Start()

	a.i2cWriteImpl = func([]byte) (int, error) {
		return 0, errors.New("write error")
	}
	_, err := d.Pressure()

	gobottest.Assert(t, err, errors.New("write error"))
}

func TestMPL115A2_initialization(t *testing.T) {
	// sequence for initialization the device on Start(), which calculates
	// the coefficients for temperature compensation of pressure
	// * write coefficient A0 MSB register address (0x04)
	// * read all 4 coefficients (16 bit, order MSB-LSB)
	// * write signal path reset register address (0x68)
	// * calculate A0, B1, B2, C12 according to data sheet
	//
	// arrange
	d, a := initTestMPL115A2DriverWithStubbedAdaptor()
	readCallCounter := 0
	readReturn := []byte{0x3E, 0xCE, 0xB3, 0xF9, 0xC5, 0x17, 0x33, 0xC8} // use example from data sheet
	a.i2cReadImpl = func(b []byte) (int, error) {
		readCallCounter++
		copy(b, readReturn)
		return len(b), nil
	}
	// act, assert - initialization() must be called on Start()
	err := d.Start()
	// assert
	gobottest.Assert(t, err, nil)
	gobottest.Assert(t, readCallCounter, 1)
	gobottest.Assert(t, len(a.written), 1)
	gobottest.Assert(t, a.written[0], uint8(0x04))
	gobottest.Assert(t, d.a0, float32(2009.75))
	gobottest.Assert(t, d.b1, float32(-2.3758545))
	gobottest.Assert(t, d.b2, float32(-0.9204712))
	gobottest.Assert(t, d.c12, float32(0.0007901192))
}
