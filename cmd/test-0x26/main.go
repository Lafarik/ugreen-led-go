package main

import (
	"fmt"
	"os"
	"ugreen-led-go/pkg/i2c"
)

// Test I2C device at 0x26 to see if it's a LED controller
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: test-0x26 <read|scan>")
		os.Exit(1)
	}

	dev, err := i2c.Open(0, 0x26)
	if err != nil {
		fmt.Printf("Failed to open I2C device: %v\n", err)
		os.Exit(1)
	}
	defer dev.Close()

	cmd := os.Args[1]

	switch cmd {
	case "read":
		// Read first 16 registers
		fmt.Println("Reading registers 0x00-0x0F from device 0x26:")
		for reg := uint8(0x00); reg < 0x10; reg++ {
			data, err := dev.ReadByteData(reg)
			if err != nil {
				fmt.Printf("  0x%02x: ERROR - %v\n", reg, err)
			} else {
				fmt.Printf("  0x%02x: 0x%02x (%d)\n", reg, data, data)
			}
		}

	case "scan":
		// Try to detect register patterns
		fmt.Println("Scanning for responsive registers on device 0x26:")
		for reg := uint8(0x00); reg <= 0xFF; reg++ {
			data, err := dev.ReadByteData(reg)
			if err == nil {
				fmt.Printf("  0x%02x: 0x%02x\n", reg, data)
			}
			if reg == 0xFF {
				break
			}
		}

	default:
		fmt.Println("Unknown command. Use 'read' or 'scan'")
		os.Exit(1)
	}
}
