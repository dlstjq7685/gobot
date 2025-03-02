package system

import (
	"syscall"
	"unsafe"
)

// mockSyscall represents the mock Syscall used for unit tests
type mockSyscall struct {
	lastTrap   uintptr
	lastFile   File
	lastSignal uintptr
	devAddress uintptr
	smbus      *i2cSmbusIoctlData
	sliceSize  uint8
	dataSlice  []byte
	Impl       func(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno)
}

// Syscall calls the user defined implementation, used for tests, implements the SystemCaller interface
func (sys *mockSyscall) syscall(trap uintptr, f File, signal uintptr, payload unsafe.Pointer) (r1, r2 uintptr, err syscall.Errno) {
	sys.lastTrap = trap     // points to the used syscall (e.g. "SYS_IOCTL")
	sys.lastFile = f        // a character device file (e.g. file to path "/dev/i2c-1")
	sys.lastSignal = signal // points to used function type (e.g. I2C_SMBUS, I2C_RDWR)

	if signal == I2C_SLAVE {
		// in this case the uintptr corresponds the address
		sys.devAddress = uintptr(payload)
	}

	if signal == I2C_SMBUS {
		// set the I2C smbus data object reference to payload and fill with some data
		sys.smbus = (*i2cSmbusIoctlData)(payload)
		if sys.smbus.data != nil {
			sys.sliceSize = sys.retrieveSliceSize()

			if sys.smbus.readWrite == I2C_SMBUS_WRITE {
				// get the data object payload as byte slice
				sys.dataSlice = unsafe.Slice((*byte)(unsafe.Pointer(sys.smbus.data)), sys.sliceSize)
			}

			if sys.smbus.readWrite == I2C_SMBUS_READ {
				// fill data object with data from given slice to simulate reading
				if sys.dataSlice != nil {
					slc := unsafe.Slice((*byte)(unsafe.Pointer(sys.smbus.data)), sys.sliceSize)
					if sys.smbus.protocol == I2C_SMBUS_BLOCK_DATA || sys.smbus.protocol == I2C_SMBUS_I2C_BLOCK_DATA {
						copy(slc[1:], sys.dataSlice)
					} else {
						copy(slc, sys.dataSlice)
					}
				}
			}
		}
	}

	// call mock implementation
	if sys.Impl != nil {
		return sys.Impl(trap, f.Fd(), signal, uintptr(unsafe.Pointer(payload)))
	}
	return 0, 0, 0
}

func (sys *mockSyscall) retrieveSliceSize() uint8 {
	switch sys.smbus.protocol {
	case I2C_SMBUS_BYTE:
		return 1
	case I2C_SMBUS_BYTE_DATA:
		return 1
	case I2C_SMBUS_WORD_DATA:
		return 2
	default:
		// for I2C_SMBUS_BLOCK_DATA, I2C_SMBUS_I2C_BLOCK_DATA
		return *(*byte)(unsafe.Pointer(sys.smbus.data)) + 1 // first data element contains data size
	}
}
