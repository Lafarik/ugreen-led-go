package dxp480tplus

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type fakeBus struct {
	readErr  error
	failAt   int
	writes   []Step
	readRegs []uint8
}

func (f *fakeBus) ReadByteData(reg uint8) (uint8, error) {
	f.readRegs = append(f.readRegs, reg)
	return 0, f.readErr
}

func (f *fakeBus) WriteByteData(reg, value uint8) error {
	if f.failAt > 0 && len(f.writes)+1 == f.failAt {
		return errors.New("nack")
	}
	f.writes = append(f.writes, Step{reg, value})
	return nil
}

func TestApplyWhiteWritesExactSequence(t *testing.T) {
	b := &fakeBus{}
	if err := Apply(b, White); err != nil {
		t.Fatal(err)
	}
	if len(b.readRegs) != 1 || b.readRegs[0] != RegMode {
		t.Fatalf("expected one read of 0x50, got %v", b.readRegs)
	}
	if len(b.writes) != len(White) {
		t.Fatalf("writes %v", b.writes)
	}
	for i := range White {
		if b.writes[i] != White[i] {
			t.Fatalf("step %d: got %v want %v", i, b.writes[i], White[i])
		}
	}
}

func TestNoWritesWhenControllerSilent(t *testing.T) {
	b := &fakeBus{readErr: errors.New("timeout")}
	if err := Apply(b, Off); err == nil {
		t.Fatal("expected error")
	}
	if len(b.writes) != 0 {
		t.Fatalf("unexpected writes %v", b.writes)
	}
}

func TestStopsAtFirstFailedWrite(t *testing.T) {
	b := &fakeBus{failAt: 3}
	if err := Apply(b, White); err == nil {
		t.Fatal("expected error")
	}
	if len(b.writes) != 2 {
		t.Fatalf("expected 2 successful writes, got %d", len(b.writes))
	}
}

func TestSequencesNeverUseUnknownRegisters(t *testing.T) {
	allowed := map[uint8]bool{0xa0: true, 0xb1: true, 0x50: true, 0x51: true}
	for name, seq := range Sequences {
		for _, s := range seq {
			if !allowed[s.Reg] {
				t.Fatalf("%s uses register 0x%02x", name, s.Reg)
			}
		}
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFindI801Bus(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "bus/i2c/devices/i2c-0/name"), "Synopsys DesignWare I2C adapter\n")
	writeFile(t, filepath.Join(root, "bus/i2c/devices/i2c-1/name"), "SMBus I801 adapter at efa0\n")
	bus, err := FindI801Bus(root)
	if err != nil || bus != 1 {
		t.Fatalf("got %d, %v", bus, err)
	}
	writeFile(t, filepath.Join(root, "bus/i2c/devices/i2c-3/name"), "SMBus I801 adapter at f000\n")
	if _, err := FindI801Bus(root); err == nil {
		t.Fatal("expected error for two adapters")
	}
}

func TestIsDXP480TPlus(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "class/dmi/id/sys_vendor"), "UGREEN\n")
	writeFile(t, filepath.Join(root, "class/dmi/id/product_name"), "DXP4800 Plus\n")
	if IsDXP480TPlus(root) {
		t.Fatal("DXP4800 Plus must not match")
	}
	writeFile(t, filepath.Join(root, "class/dmi/id/product_name"), "DXP480T Plus\n")
	if !IsDXP480TPlus(root) {
		t.Fatal("DXP480T Plus must match")
	}
}
