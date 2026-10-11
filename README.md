# ugreen-led-go

Small Go tool to control the front LEDs of UGREEN NAS devices directly via I2C – no kernel module required.

*Kleines Go-Werkzeug zur Ansteuerung der Front-LEDs von UGREEN-NAS-Geräten direkt über I2C, ohne Kernelmodul.*

Two controller types are supported:

| Model | Controller | Command |
|---|---|---|
| **DXP480T Plus** (single power LED) | SMBus I801, address `0x26` | `ugreen-led-ctl dxp480t-plus …` |
| other DX/DXP models (power, netdev, disk LEDs) | SMBus I801, address `0x3a` | `ugreen-led-ctl probe/get/set …` |

## DXP480T Plus

On non-UGOS systems (Proxmox, Debian, TrueNAS, DSM) the power LED of the DXP480T Plus keeps blinking forever, because only UGOS tells the controller that booting has finished.

```sh
sudo modprobe i2c-dev
sudo ./ugreen-led-ctl dxp480t-plus probe        # read-only: does 0x26 answer?
sudo ./ugreen-led-ctl dxp480t-plus white        # solid white
sudo ./ugreen-led-ctl dxp480t-plus white-blink  # white, fast flashing (e.g. during shutdown)
sudo ./ugreen-led-ctl dxp480t-plus off
```

The SMBus I801 bus number is detected automatically (it is 0 on most units, sometimes 1); pass it as last argument to override. The tool refuses to run on other models unless `UGREEN_FORCE_MODEL=1` is set. Before any write it checks that the controller answers a read of register `0x50`; it only writes registers `0xa0`, `0xb1`, `0x50`, `0x51`.

Note: the controller at `0x26` returns `0x00` for the identification registers `0x5a/0x5b` that are documented for the `0x3a` controller of the other models – do not use them as a check on this model.

The register sequence was found and tested by the community ([miskcoo/ugreen_leds_controller#6](https://github.com/miskcoo/ugreen_leds_controller/issues/6), [nuvolacat/UGREEN-DXP480T-PLUS_Arc-Xpenology-LedFIX](https://github.com/nuvolacat/UGREEN-DXP480T-PLUS_Arc-Xpenology-LedFIX)). The Go implementation is unit-tested; a hardware test by the author is pending.

For a boot service, run `white` once after boot and once more ~5 s later (early boot services may overwrite it), and `off` or `white-blink` on shutdown.

## Layout

- `pkg/i2c` – minimal I2C access via `/dev/i2c-*`
- `pkg/dxp480tplus` – DXP480T Plus power LED (controller `0x26`), bus detection, model check
- `pkg/ugreen` – LED protocol of the `0x3a` controller (color, brightness, on/off, blink)
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

Prototype (started February 2026). DXP480T Plus support added October 2026. Use at your own risk – writing to the wrong I2C address can affect other devices on the bus; always run `probe` first.

## Credits

The `0x3a` LED protocol is based on the work in [miskcoo/ugreen_leds_controller](https://github.com/miskcoo/ugreen_leds_controller) (MIT License). The DXP480T Plus sequence comes from the users in issue #6 there and from [nuvolacat](https://github.com/nuvolacat/UGREEN-DXP480T-PLUS_Arc-Xpenology-LedFIX) (MIT). This repository is an independent Go implementation.

## License

MIT – see [LICENSE](LICENSE).
