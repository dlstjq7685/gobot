package i2c

import (
	"errors"
	"strings"
	"testing"

	"gobot.io/x/gobot/gobottest"
)

func initTestADS1115DriverWithStubbedAdaptor() (*ADS1x15Driver, *i2cTestAdaptor) {
	a := newI2cTestAdaptor()
	d := NewADS1115Driver(a)
	if err := d.Start(); err != nil {
		panic(err)
	}
	return d, a
}

func TestNewADS1115Driver(t *testing.T) {
	var di interface{} = NewADS1115Driver(newI2cTestAdaptor())
	d, ok := di.(*ADS1x15Driver)
	if !ok {
		t.Errorf("NewADS1115Driver() should have returned a *ADS1x15Driver")
	}
	gobottest.Refute(t, d.Driver, nil)
	gobottest.Assert(t, strings.HasPrefix(d.Name(), "ADS1115"), true)
	for i := 0; i <= 3; i++ {
		gobottest.Assert(t, d.channelCfgs[i].gain, 1)
		gobottest.Assert(t, d.channelCfgs[i].dataRate, 128)
	}
}

func TestADS1115Options(t *testing.T) {
	// This is a general test, that options are applied in constructor by using the common WithBus() option and
	// least one of this driver. Further tests for options can also be done by call of "WithOption(val)(d)".
	d := NewADS1115Driver(newI2cTestAdaptor(), WithBus(2), WithADS1x15Gain(2), WithADS1x15DataRate(860))
	gobottest.Assert(t, d.GetBusOrDefault(1), 2)
	for i := 0; i <= 3; i++ {
		gobottest.Assert(t, d.channelCfgs[i].gain, 2)
		gobottest.Assert(t, d.channelCfgs[i].dataRate, 860)
	}
}

func TestADS1115WithADS1x15BestGainForVoltage(t *testing.T) {
	d, _ := initTestADS1115DriverWithStubbedAdaptor()
	WithADS1x15BestGainForVoltage(1.01)(d)
	for i := 0; i <= 3; i++ {
		gobottest.Assert(t, d.channelCfgs[i].gain, 3)
	}
}

func TestADS1115WithADS1x15ChannelBestGainForVoltage(t *testing.T) {
	d, _ := initTestADS1115DriverWithStubbedAdaptor()
	WithADS1x15ChannelBestGainForVoltage(0, 1.0)(d)
	WithADS1x15ChannelBestGainForVoltage(1, 2.5)(d)
	WithADS1x15ChannelBestGainForVoltage(2, 3.3)(d)
	WithADS1x15ChannelBestGainForVoltage(3, 5.0)(d)
	gobottest.Assert(t, d.channelCfgs[0].gain, 3)
	gobottest.Assert(t, d.channelCfgs[1].gain, 1)
	gobottest.Assert(t, d.channelCfgs[2].gain, 1)
	gobottest.Assert(t, d.channelCfgs[3].gain, 0)
}

func TestADS1115AnalogRead(t *testing.T) {
	d, a := initTestADS1115DriverWithStubbedAdaptor()
	WithADS1x15WaitSingleCycle()(d)

	a.i2cReadImpl = func(b []byte) (int, error) {
		copy(b, []byte{0x7F, 0xFF})
		return 2, nil
	}

	val, err := d.AnalogRead("0")
	gobottest.Assert(t, val, 32767)
	gobottest.Assert(t, err, nil)

	val, err = d.AnalogRead("1")
	gobottest.Assert(t, val, 32767)
	gobottest.Assert(t, err, nil)

	val, err = d.AnalogRead("2")
	gobottest.Assert(t, val, 32767)
	gobottest.Assert(t, err, nil)

	val, err = d.AnalogRead("3")
	gobottest.Assert(t, val, 32767)
	gobottest.Assert(t, err, nil)

	val, err = d.AnalogRead("0-1")
	gobottest.Assert(t, val, 32767)
	gobottest.Assert(t, err, nil)

	val, err = d.AnalogRead("0-3")
	gobottest.Assert(t, val, 32767)
	gobottest.Assert(t, err, nil)

	val, err = d.AnalogRead("1-3")
	gobottest.Assert(t, val, 32767)
	gobottest.Assert(t, err, nil)

	val, err = d.AnalogRead("2-3")
	gobottest.Assert(t, val, 32767)
	gobottest.Assert(t, err, nil)

	val, err = d.AnalogRead("3-2")
	gobottest.Refute(t, err.Error(), nil)
}

func TestADS1115AnalogReadError(t *testing.T) {
	d, a := initTestADS1115DriverWithStubbedAdaptor()

	a.i2cReadImpl = func(b []byte) (int, error) {
		return 0, errors.New("read error")
	}

	_, err := d.AnalogRead("0")
	gobottest.Assert(t, err, errors.New("read error"))
}

func TestADS1115AnalogReadInvalidPin(t *testing.T) {
	d, _ := initTestADS1115DriverWithStubbedAdaptor()

	_, err := d.AnalogRead("98")
	gobottest.Assert(t, err, errors.New("Invalid channel (98), must be between 0 and 3"))
}

func TestADS1115AnalogReadWriteError(t *testing.T) {
	d, a := initTestADS1115DriverWithStubbedAdaptor()

	a.i2cWriteImpl = func([]byte) (int, error) {
		return 0, errors.New("write error")
	}

	_, err := d.AnalogRead("0")
	gobottest.Assert(t, err, errors.New("write error"))

	_, err = d.AnalogRead("0-1")
	gobottest.Assert(t, err, errors.New("write error"))

	_, err = d.AnalogRead("2-3")
	gobottest.Assert(t, err, errors.New("write error"))
}

func TestADS1115ReadInvalidChannel(t *testing.T) {
	d, _ := initTestADS1115DriverWithStubbedAdaptor()

	_, err := d.Read(7, 1, 1600)
	gobottest.Assert(t, err, errors.New("Invalid channel (7), must be between 0 and 3"))
}

func TestADS1115ReadInvalidGain(t *testing.T) {
	d, _ := initTestADS1115DriverWithStubbedAdaptor()

	_, err := d.Read(0, 21, 1600)
	gobottest.Assert(t, err, errors.New("Gain (21) must be one of: [0 1 2 3 4 5 6 7]"))
}

func TestADS1115ReadInvalidDataRate(t *testing.T) {
	d, _ := initTestADS1115DriverWithStubbedAdaptor()

	_, err := d.Read(0, 1, 678)
	gobottest.Assert(t, err, errors.New("Invalid data rate (678). Accepted values: [8 16 32 64 128 250 475 860]"))
}

func TestADS1115ReadDifferenceInvalidChannel(t *testing.T) {
	d, _ := initTestADS1115DriverWithStubbedAdaptor()

	_, err := d.ReadDifference(5, 1, 1600)
	gobottest.Assert(t, err, errors.New("Invalid channel (5), must be between 0 and 3"))
}

func TestADS1115_rawRead(t *testing.T) {
	// sequence to read:
	// * prepare config register content (mode, input, gain, data rate, comparator)
	// * write config register (16 bit, MSByte first)
	// * read config register (16 bit, MSByte first) and wait for bit 15 is set
	// * read conversion register (16 bit, MSByte first) for the value
	// * apply two's complement converter, relates to one digit resolution (1/2^15), voltage multiplier
	var tests = map[string]struct {
		input      []uint8
		gain       int
		dataRate   int
		want       int
		wantConfig []uint8
	}{
		"+FS": {
			input:      []uint8{0x7F, 0xFF},
			gain:       0,
			dataRate:   8,
			want:       (1<<15 - 1),
			wantConfig: []uint8{0x91, 0x03},
		},
		"+1": {
			input:      []uint8{0x00, 0x01},
			gain:       0,
			dataRate:   16,
			want:       1,
			wantConfig: []uint8{0x91, 0x23},
		},
		"+-0": {
			input:      []uint8{0x00, 0x00},
			gain:       0,
			dataRate:   32,
			want:       0,
			wantConfig: []uint8{0x91, 0x43},
		},
		"-1": {
			input:      []uint8{0xFF, 0xFF},
			gain:       0,
			dataRate:   64,
			want:       -1,
			wantConfig: []uint8{0x91, 0x63},
		},
		"-FS": {
			input:      []uint8{0x80, 0x00},
			gain:       0,
			dataRate:   128,
			want:       -(1 << 15),
			wantConfig: []uint8{0x91, 0x83},
		},
		"+FS gain 1": {
			input:      []uint8{0x7F, 0xFF},
			gain:       1,
			dataRate:   250,
			want:       (1<<15 - 1),
			wantConfig: []uint8{0x93, 0xA3},
		},
		"+FS gain 3": {
			input:      []uint8{0x7F, 0xFF},
			gain:       3,
			dataRate:   475,
			want:       (1<<15 - 1),
			wantConfig: []uint8{0x97, 0xC3},
		},
		"+FS gain 5": {
			input:      []uint8{0x7F, 0xFF},
			gain:       5,
			dataRate:   860,
			want:       (1<<15 - 1),
			wantConfig: []uint8{0x9B, 0xE3},
		},
		"+FS gain 7": {
			input:      []uint8{0x7F, 0xFF},
			gain:       7,
			dataRate:   128,
			want:       (1<<15 - 1),
			wantConfig: []uint8{0x9F, 0x83},
		},
	}
	d, a := initTestADS1115DriverWithStubbedAdaptor()
	// arrange
	channel := 0
	channelOffset := 1
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			a.written = []byte{} // reset writes of Start() and former test
			// arrange reads
			conversion := []uint8{0x00, 0x00}   // a conversion is in progress
			noConversion := []uint8{0x80, 0x00} // no conversion in progress
			returnRead := [3][]uint8{conversion, noConversion, tt.input}
			numCallsRead := 0
			a.i2cReadImpl = func(b []byte) (int, error) {
				numCallsRead++
				retRead := returnRead[numCallsRead-1]
				copy(b, retRead)
				return len(b), nil
			}
			// act
			got, err := d.rawRead(channel, channelOffset, tt.gain, tt.dataRate)
			// assert
			gobottest.Assert(t, err, nil)
			gobottest.Assert(t, got, tt.want)
			gobottest.Assert(t, numCallsRead, 3)
			gobottest.Assert(t, len(a.written), 6)
			gobottest.Assert(t, a.written[0], uint8(ads1x15PointerConfig))
			gobottest.Assert(t, a.written[1], tt.wantConfig[0])            // MSByte: OS, MUX, PGA, MODE
			gobottest.Assert(t, a.written[2], tt.wantConfig[1])            // LSByte: DR, COMP_*
			gobottest.Assert(t, a.written[3], uint8(ads1x15PointerConfig)) // first check for no conversion
			gobottest.Assert(t, a.written[4], uint8(ads1x15PointerConfig)) // second check for no conversion
			gobottest.Assert(t, a.written[5], uint8(ads1x15PointerConversion))
		})
	}
}
