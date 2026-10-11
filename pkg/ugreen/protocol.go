package ugreen

import (
	"errors"
	"fmt"
	"ugreen-led-go/pkg/i2c"
)

const (
	I2C_ADDR = 0x3a
)

// LED IDs
type LEDType uint8

const (
	LEDPower  LEDType = 0
	LEDNetdev LEDType = 1
	LEDDisk1  LEDType = 2
	LEDDisk2  LEDType = 3
	LEDDisk3  LEDType = 4
	LEDDisk4  LEDType = 5
	LEDDisk5  LEDType = 6
	LEDDisk6  LEDType = 7
	LEDDisk7  LEDType = 8
	LEDDisk8  LEDType = 9
)

// LED States
type LEDState uint8

const (
	StateOff    LEDState = 0
	StateOn     LEDState = 1
	StateBlink  LEDState = 2
	StateBreath LEDState = 3
)

// LED Status structure
type LEDStatus struct {
	Available  bool
	State      LEDState
	Brightness uint8
	R, G, B    uint8
	TOn        uint16 // milliseconds on
	TOff       uint16 // milliseconds off
}

type Controller struct {
	dev *i2c.Device
}

// NewController creates a new LED controller
func NewController(bus int) (*Controller, error) {
	dev, err := i2c.Open(bus, I2C_ADDR)
	if err != nil {
		return nil, fmt.Errorf("failed to open I2C device: %w", err)
	}
	return &Controller{dev: dev}, nil
}

// Close closes the controller
func (c *Controller) Close() error {
	return c.dev.Close()
}

// computeChecksum calculates the checksum for the given data
func computeChecksum(data []byte) uint16 {
	sum := uint16(0)
	for _, b := range data {
		sum += uint16(b)
	}
	return sum
}

// verifyChecksum verifies the checksum of received data
func verifyChecksum(data []byte) bool {
	if len(data) < 2 {
		return false
	}
	dataLen := len(data)
	sum := computeChecksum(data[:dataLen-2])
	checksum := uint16(data[dataLen-2])<<8 | uint16(data[dataLen-1])
	return sum != 0 && sum == checksum
}

// GetStatus reads the status of an LED
func (c *Controller) GetStatus(led LEDType) (*LEDStatus, error) {
	// Read 11 bytes from address 0x81 + LED_ID
	data, err := c.dev.ReadBlockData(0x81+uint8(led), 11)
	if err != nil {
		return nil, fmt.Errorf("failed to read LED status: %w", err)
	}

	if len(data) != 11 {
		return nil, fmt.Errorf("invalid data length: got %d, expected 11", len(data))
	}

	if !verifyChecksum(data) {
		return nil, errors.New("checksum verification failed")
	}

	status := &LEDStatus{
		Available:  true,
		Brightness: data[1],
		R:          data[2],
		G:          data[3],
		B:          data[4],
	}

	// Parse state
	switch data[0] {
	case 0:
		status.State = StateOff
	case 1:
		status.State = StateOn
	case 2:
		status.State = StateBlink
	case 3:
		status.State = StateBreath
	default:
		return nil, fmt.Errorf("unknown state: %d", data[0])
	}

	// Parse timing
	tHigh := uint16(data[5])<<8 | uint16(data[6])
	tLow := uint16(data[7])<<8 | uint16(data[8])
	status.TOn = tLow
	status.TOff = tHigh - tLow

	return status, nil
}

// changeState sends a command to change LED state
func (c *Controller) changeState(led LEDType, command uint8, params [4]uint8) error {
	// Construct the 12-byte write buffer
	buf := []byte{
		uint8(led),
		0xa0, 0x01, 0x00, 0x00,
		command,
		params[0], params[1], params[2], params[3],
	}

	// Compute checksum of bytes 1-9 (indices 1-9)
	checksum := computeChecksum(buf[1:])
	buf = append(buf, byte(checksum>>8), byte(checksum&0xff))

	// Write to address 0x00 + LED_ID
	return c.dev.WriteBlockData(uint8(led), buf)
}

// SetOnOff turns an LED on or off
func (c *Controller) SetOnOff(led LEDType, on bool) error {
	var status uint8
	if on {
		status = 1
	}
	return c.changeState(led, 0x03, [4]uint8{status, 0, 0, 0})
}

// SetColor sets the RGB color of an LED
func (c *Controller) SetColor(led LEDType, r, g, b uint8) error {
	return c.changeState(led, 0x02, [4]uint8{r, g, b, 0})
}

// SetBrightness sets the brightness of an LED
func (c *Controller) SetBrightness(led LEDType, brightness uint8) error {
	return c.changeState(led, 0x01, [4]uint8{brightness, 0, 0, 0})
}

// SetBlink sets the LED to blink mode
func (c *Controller) SetBlink(led LEDType, tOn, tOff uint16) error {
	tCycle := tOn + tOff
	return c.changeState(led, 0x04, [4]uint8{
		byte(tCycle >> 8),
		byte(tCycle & 0xff),
		byte(tOn >> 8),
		byte(tOn & 0xff),
	})
}

// SetBreath sets the LED to breath mode
func (c *Controller) SetBreath(led LEDType, tOn, tOff uint16) error {
	tCycle := tOn + tOff
	return c.changeState(led, 0x05, [4]uint8{
		byte(tCycle >> 8),
		byte(tCycle & 0xff),
		byte(tOn >> 8),
		byte(tOn & 0xff),
	})
}

// CheckLastCommand checks if the last command was successful
func (c *Controller) CheckLastCommand() (bool, error) {
	status, err := c.dev.ReadByteData(0x80)
	if err != nil {
		return false, err
	}
	return status == 1, nil
}
