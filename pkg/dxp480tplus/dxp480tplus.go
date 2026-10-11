// Package dxp480tplus drives the single power LED of the UGREEN NASync DXP480T Plus.
//
// Unlike the other DXP models (multi-LED controller at 0x3a, see package ugreen),
// the DXP480T Plus has a simple controller at address 0x26 on the SMBus I801
// adapter. Register meaning and the sequences below come from community testing:
// miskcoo/ugreen_leds_controller issue #6 and
// nuvolacat/UGREEN-DXP480T-PLUS_Arc-Xpenology-LedFIX.
//
//	0xa0 = 1 / 2   red off / white off
//	0xb1 = 2       white on
//	0x51 = 0/1     slow flash off/on
//	0x50 = 0/1     solid / fast flash
package dxp480tplus

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Addr is the I2C address of the LED controller.
const Addr = 0x26

// RegMode is read to check that the controller answers.
const RegMode = 0x50

// Step is one register write.
type Step struct {
	Reg   uint8
	Value uint8
}

// Sequences known to work on the DXP480T Plus.
var (
	// White: solid white (as set by UGOS after boot).
	White = []Step{{0xa0, 1}, {0xa0, 2}, {0x51, 0}, {0xb1, 2}, {0x50, 0}}
	// WhiteBlink: white, fast flashing (boot/shutdown indication).
	WhiteBlink = []Step{{0xa0, 1}, {0xa0, 2}, {0x51, 0}, {0xb1, 2}, {0x50, 1}}
	// Off: all LEDs off, no flashing.
	Off = []Step{{0xa0, 1}, {0xa0, 2}, {0x51, 0}, {0x50, 0}}
)

// Sequences maps command names to sequences.
var Sequences = map[string][]Step{
	"white":       White,
	"white-blink": WhiteBlink,
	"off":         Off,
}

// Bus is the minimal register access needed (implemented by *i2c.Device).
type Bus interface {
	ReadByteData(reg uint8) (uint8, error)
	WriteByteData(reg uint8, value uint8) error
}

// Probe reads the mode register; an error means the controller did not answer.
func Probe(b Bus) (uint8, error) {
	v, err := b.ReadByteData(RegMode)
	if err != nil {
		return 0, fmt.Errorf("controller 0x%02x does not answer: %w", Addr, err)
	}
	return v, nil
}

// Apply probes the controller and then writes the steps in order.
// It stops at the first failing write.
func Apply(b Bus, steps []Step) error {
	if _, err := Probe(b); err != nil {
		return err
	}
	for i, s := range steps {
		if err := b.WriteByteData(s.Reg, s.Value); err != nil {
			return fmt.Errorf("step %d (reg 0x%02x = %d): %w", i+1, s.Reg, s.Value, err)
		}
	}
	return nil
}

// FindI801Bus returns the number of the single "SMBus I801" adapter below
// sysfsRoot (normally "/sys"). The number can change between boots.
func FindI801Bus(sysfsRoot string) (int, error) {
	names, _ := filepath.Glob(filepath.Join(sysfsRoot, "bus/i2c/devices/i2c-*/name"))
	sort.Strings(names)
	var found []int
	for _, n := range names {
		raw, err := os.ReadFile(n)
		if err != nil || !strings.HasPrefix(strings.TrimSpace(string(raw)), "SMBus I801") {
			continue
		}
		num, err := strconv.Atoi(strings.TrimPrefix(filepath.Base(filepath.Dir(n)), "i2c-"))
		if err == nil {
			found = append(found, num)
		}
	}
	switch len(found) {
	case 0:
		return 0, errors.New("no SMBus I801 adapter found (is i2c-dev loaded?)")
	case 1:
		return found[0], nil
	default:
		return 0, fmt.Errorf("more than one SMBus I801 adapter found: %v", found)
	}
}

// IsDXP480TPlus checks the DMI product name below sysfsRoot.
func IsDXP480TPlus(sysfsRoot string) bool {
	vendor, _ := os.ReadFile(filepath.Join(sysfsRoot, "class/dmi/id/sys_vendor"))
	product, _ := os.ReadFile(filepath.Join(sysfsRoot, "class/dmi/id/product_name"))
	return strings.TrimSpace(string(vendor)) == "UGREEN" && strings.TrimSpace(string(product)) == "DXP480T Plus"
}
