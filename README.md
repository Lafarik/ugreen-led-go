# ugreen-led-go

Small Go tool to control the front LEDs of a UGREEN NAS (tested on DXP480T) directly via I2C – no kernel module required.

*Kleines Go-Werkzeug zur Ansteuerung der Front-LEDs eines UGREEN-NAS (getestet mit DXP480T) direkt über I2C, ohne Kernelmodul.*

## Layout

- `pkg/i2c` – minimal I2C access via `/dev/i2c-*`
- `pkg/ugreen` – LED protocol (color, brightness, on/off, blink)
- `main.go` – command line tool `ugreen-led-ctl`
- `cmd/test-0x26`, `cmd/test-write-0x26` – read/write tests against the LED controller at address 0x26

## Build and use

```sh
go build -o ugreen-led-ctl .      # Go >= 1.24
sudo modprobe i2c-dev
sudo ./ugreen-led-ctl probe        # read-only, safe
sudo ./ugreen-led-ctl set power color 0 0 255
sudo ./ugreen-led-ctl set disk1 blink 500 500
```

LEDs: `power`, `netdev`, `disk1` … `disk8`. Run without arguments for the full usage text.

## Status

Prototype, not actively maintained (archived February 2026). Use at your own risk – writing to the wrong I2C address can affect other devices on the bus; always run `probe` first.

## Credits

The LED protocol is based on the work in [miskcoo/ugreen_leds_controller](https://github.com/miskcoo/ugreen_leds_controller) (MIT License). This repository is an independent Go implementation.

## License

MIT – see [LICENSE](LICENSE).
