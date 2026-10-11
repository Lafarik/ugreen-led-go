package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"ugreen-led-go/pkg/ugreen"
)

func usage() {
	fmt.Fprintf(os.Stderr, `UGREEN LED Controller

Usage:
  %s probe [bus]           - Probe I2C bus for LEDs (read-only, safe)
  %s get <led>             - Get LED status
  %s set <led> <command>   - Set LED state

LEDs: power, netdev, disk1-disk8
Commands for 'set':
  on                       - Turn LED on
  off                      - Turn LED off
  color R G B              - Set color (0-255 each)
  brightness N             - Set brightness (0-255)
  blink T_ON T_OFF         - Blink mode (milliseconds)

Examples:
  %s probe                 - Probe default bus 0
  %s probe 1               - Probe bus 1
  %s get power             - Get power LED status
  %s set power on          - Turn power LED on
  %s set power color 255 0 0 - Set power LED to red
  %s set disk1 blink 500 500 - Blink disk1 LED

`, os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0])
	os.Exit(1)
}

var ledNames = map[string]ugreen.LEDType{
	"power":  ugreen.LEDPower,
	"netdev": ugreen.LEDNetdev,
	"disk1":  ugreen.LEDDisk1,
	"disk2":  ugreen.LEDDisk2,
	"disk3":  ugreen.LEDDisk3,
	"disk4":  ugreen.LEDDisk4,
	"disk5":  ugreen.LEDDisk5,
	"disk6":  ugreen.LEDDisk6,
	"disk7":  ugreen.LEDDisk7,
	"disk8":  ugreen.LEDDisk8,
}

func probeBus(bus int) error {
	fmt.Printf("Probing I2C bus %d at address 0x%02x...\n", bus, ugreen.I2C_ADDR)
	
	ctrl, err := ugreen.NewController(bus)
	if err != nil {
		return fmt.Errorf("failed to open controller: %w", err)
	}
	defer ctrl.Close()

	results := make(map[string]interface{})
	
	for name, ledType := range ledNames {
		status, err := ctrl.GetStatus(ledType)
		if err != nil {
			results[name] = map[string]string{"error": err.Error()}
			continue
		}
		
		stateStr := "unknown"
		switch status.State {
		case ugreen.StateOff:
			stateStr = "off"
		case ugreen.StateOn:
			stateStr = "on"
		case ugreen.StateBlink:
			stateStr = "blink"
		case ugreen.StateBreath:
			stateStr = "breath"
		}
		
		results[name] = map[string]interface{}{
			"state":      stateStr,
			"brightness": status.Brightness,
			"color":      fmt.Sprintf("RGB(%d,%d,%d)", status.R, status.G, status.B),
			"t_on_ms":    status.TOn,
			"t_off_ms":   status.TOff,
		}
	}
	
	output, _ := json.MarshalIndent(results, "", "  ")
	fmt.Println(string(output))
	
	return nil
}

func getLED(ctrl *ugreen.Controller, ledName string) error {
	ledType, ok := ledNames[ledName]
	if !ok {
		return fmt.Errorf("unknown LED: %s", ledName)
	}
	
	status, err := ctrl.GetStatus(ledType)
	if err != nil {
		return err
	}
	
	stateStr := "unknown"
	switch status.State {
	case ugreen.StateOff:
		stateStr = "off"
	case ugreen.StateOn:
		stateStr = "on"
	case ugreen.StateBlink:
		stateStr = "blink"
	case ugreen.StateBreath:
		stateStr = "breath"
	}
	
	fmt.Printf("%s:\n", ledName)
	fmt.Printf("  State:      %s\n", stateStr)
	fmt.Printf("  Brightness: %d\n", status.Brightness)
	fmt.Printf("  Color:      RGB(%d, %d, %d)\n", status.R, status.G, status.B)
	if status.State == ugreen.StateBlink || status.State == ugreen.StateBreath {
		fmt.Printf("  Timing:     %dms on, %dms off\n", status.TOn, status.TOff)
	}
	
	return nil
}

func setLED(ctrl *ugreen.Controller, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("not enough arguments for set command")
	}
	
	ledName := args[0]
	ledType, ok := ledNames[ledName]
	if !ok {
		return fmt.Errorf("unknown LED: %s", ledName)
	}
	
	command := args[1]
	
	switch command {
	case "on":
		return ctrl.SetOnOff(ledType, true)
		
	case "off":
		return ctrl.SetOnOff(ledType, false)
		
	case "color":
		if len(args) < 5 {
			return fmt.Errorf("color requires 3 arguments: R G B")
		}
		r, _ := strconv.Atoi(args[2])
		g, _ := strconv.Atoi(args[3])
		b, _ := strconv.Atoi(args[4])
		
		if r < 0 || r > 255 || g < 0 || g > 255 || b < 0 || b > 255 {
			return fmt.Errorf("color values must be 0-255")
		}
		
		return ctrl.SetColor(ledType, uint8(r), uint8(g), uint8(b))
		
	case "brightness":
		if len(args) < 3 {
			return fmt.Errorf("brightness requires 1 argument")
		}
		brightness, _ := strconv.Atoi(args[2])
		if brightness < 0 || brightness > 255 {
			return fmt.Errorf("brightness must be 0-255")
		}
		
		return ctrl.SetBrightness(ledType, uint8(brightness))
		
	case "blink":
		if len(args) < 4 {
			return fmt.Errorf("blink requires 2 arguments: T_ON T_OFF (in milliseconds)")
		}
		tOn, _ := strconv.Atoi(args[2])
		tOff, _ := strconv.Atoi(args[3])
		
		if tOn < 0 || tOn > 65535 || tOff < 0 || tOff > 65535 {
			return fmt.Errorf("timing values must be 0-65535")
		}
		
		return ctrl.SetBlink(ledType, uint16(tOn), uint16(tOff))
		
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	
	cmd := strings.ToLower(os.Args[1])
	
	switch cmd {
	case "probe":
		bus := 0
		if len(os.Args) > 2 {
			var err error
			bus, err = strconv.Atoi(os.Args[2])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid bus number: %v\n", err)
				os.Exit(1)
			}
		}
		
		if err := probeBus(bus); err != nil {
			fmt.Fprintf(os.Stderr, "Probe failed: %v\n", err)
			os.Exit(1)
		}
		
	case "get":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "get requires LED name\n")
			usage()
		}
		
		ctrl, err := ugreen.NewController(0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open controller: %v\n", err)
			os.Exit(1)
		}
		defer ctrl.Close()
		
		if err := getLED(ctrl, os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get LED status: %v\n", err)
			os.Exit(1)
		}
		
	case "set":
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "set requires LED name and command\n")
			usage()
		}
		
		ctrl, err := ugreen.NewController(0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open controller: %v\n", err)
			os.Exit(1)
		}
		defer ctrl.Close()
		
		if err := setLED(ctrl, os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to set LED: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Println("✓ LED updated successfully")
		
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		usage()
	}
}
