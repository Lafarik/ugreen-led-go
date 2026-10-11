package main

import (
	"fmt"
	"os"
	"strconv"
	"ugreen-led-go/pkg/i2c"
)

// CAREFUL test writes to I2C device 0x26
// This is EXPERIMENTAL and could affect hardware
func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: test-write-0x26 <register> <value> <confirm>")
		fmt.Println("Example: test-write-0x26 0x00 0x01 YES")
		fmt.Println("")
		fmt.Println("WARNING: This writes to I2C device 0x26")
		fmt.Println("Only use if you understand the risks!")
		os.Exit(1)
	}

	if os.Args[3] != "YES" {
		fmt.Println("Safety check failed. Use 'YES' as third argument to confirm.")
		os.Exit(1)
	}

	reg, err := strconv.ParseUint(os.Args[1], 0, 8)
	if err != nil {
		fmt.Printf("Invalid register: %v\n", err)
		os.Exit(1)
	}

	val, err := strconv.ParseUint(os.Args[2], 0, 8)
	if err != nil {
		fmt.Printf("Invalid value: %v\n", err)
		os.Exit(1)
	}

	dev, err := i2c.Open(0, 0x26)
	if err != nil {
		fmt.Printf("Failed to open I2C device: %v\n", err)
		os.Exit(1)
	}
	defer dev.Close()

	// Read current value
	oldVal, err := dev.ReadByteData(uint8(reg))
	if err != nil {
		fmt.Printf("Failed to read register 0x%02x: %v\n", reg, err)
		os.Exit(1)
	}

	fmt.Printf("Current value at 0x%02x: 0x%02x\n", reg, oldVal)
	fmt.Printf("Writing value 0x%02x to register 0x%02x...\n", val, reg)

	// Write via i2cset equivalent
	// We'll use WriteBlockData with single byte
	data := []byte{uint8(val)}
	err = dev.WriteBlockData(uint8(reg), data)
	if err != nil {
		fmt.Printf("Write failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Write completed successfully!")

	// Read back
	newVal, err := dev.ReadByteData(uint8(reg))
	if err != nil {
		fmt.Printf("Failed to read back: %v\n", err)
	} else {
		fmt.Printf("New value at 0x%02x: 0x%02x\n", reg, newVal)
	}
}
