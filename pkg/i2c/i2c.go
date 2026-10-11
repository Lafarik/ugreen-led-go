package i2c

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

const (
	I2C_SLAVE       = 0x0703
	I2C_SMBUS       = 0x0720
	I2C_SMBUS_READ  = 1
	I2C_SMBUS_WRITE = 0

	I2C_SMBUS_BYTE_DATA  = 2
	I2C_SMBUS_BLOCK_DATA = 5
)

type smbusIoctlData struct {
	readWrite uint8
	command   uint8
	size      uint32
	data      uintptr
}

type Device struct {
	file *os.File
	addr uint8
}

// Open opens an I2C device
func Open(bus int, addr uint8) (*Device, error) {
	path := fmt.Sprintf("/dev/i2c-%d", bus)
	file, err := os.OpenFile(path, os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", path, err)
	}

	// Set the I2C slave address
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), I2C_SLAVE, uintptr(addr))
	if errno != 0 {
		file.Close()
		return nil, fmt.Errorf("failed to set I2C slave address 0x%02x: %s", addr, errno)
	}

	return &Device{file: file, addr: addr}, nil
}

// Close closes the I2C device
func (d *Device) Close() error {
	if d.file != nil {
		return d.file.Close()
	}
	return nil
}

// ReadByteData reads a single byte from a register
func (d *Device) ReadByteData(reg uint8) (uint8, error) {
	var data [32]byte
	args := smbusIoctlData{
		readWrite: I2C_SMBUS_READ,
		command:   reg,
		size:      I2C_SMBUS_BYTE_DATA,
		data:      uintptr(unsafe.Pointer(&data)),
	}

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, d.file.Fd(), I2C_SMBUS, uintptr(unsafe.Pointer(&args)))
	if errno != 0 {
		return 0, fmt.Errorf("SMBus read byte failed: %s", errno)
	}

	return data[0], nil
}

// ReadBlockData reads a block of data from a register
func (d *Device) ReadBlockData(reg uint8, length int) ([]byte, error) {
	if length < 1 || length > 32 {
		return nil, errors.New("block length must be 1-32 bytes")
	}

	var data [32]byte
	data[0] = byte(length)

	args := smbusIoctlData{
		readWrite: I2C_SMBUS_READ,
		command:   reg,
		size:      I2C_SMBUS_BLOCK_DATA,
		data:      uintptr(unsafe.Pointer(&data)),
	}

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, d.file.Fd(), I2C_SMBUS, uintptr(unsafe.Pointer(&args)))
	if errno != 0 {
		return nil, fmt.Errorf("SMBus read block failed: %s", errno)
	}

	result := make([]byte, length)
	copy(result, data[1:length+1])
	return result, nil
}

// WriteBlockData writes a block of data to a register
func (d *Device) WriteBlockData(reg uint8, data []byte) error {
	length := len(data)
	if length < 1 || length > 32 {
		return errors.New("block length must be 1-32 bytes")
	}

	var buf [33]byte
	buf[0] = byte(length)
	copy(buf[1:], data)

	args := smbusIoctlData{
		readWrite: I2C_SMBUS_WRITE,
		command:   reg,
		size:      I2C_SMBUS_BLOCK_DATA,
		data:      uintptr(unsafe.Pointer(&buf)),
	}

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, d.file.Fd(), I2C_SMBUS, uintptr(unsafe.Pointer(&args)))
	if errno != 0 {
		return fmt.Errorf("SMBus write block failed: %s", errno)
	}

	return nil
}

// WriteByteData writes a single byte to a register (SMBus write byte data)
func (d *Device) WriteByteData(reg uint8, value uint8) error {
	var data [32]byte
	data[0] = value
	args := smbusIoctlData{
		readWrite: I2C_SMBUS_WRITE,
		command:   reg,
		size:      I2C_SMBUS_BYTE_DATA,
		data:      uintptr(unsafe.Pointer(&data)),
	}

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, d.file.Fd(), I2C_SMBUS, uintptr(unsafe.Pointer(&args)))
	if errno != 0 {
		return fmt.Errorf("SMBus write byte failed: %s", errno)
	}
	return nil
}
