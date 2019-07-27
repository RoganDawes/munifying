package unifying

import "fmt"

/*
USB HID request to bootloader:
guint8		cmd;
guint16		addr;
guint8		len;
guint8 		data[28];
 */

type BootloaderCommand byte

const (
	BOOTLOADER_COMMAND_ERROR                            BootloaderCommand = 0x01
	BOOTLOADER_COMMAND_NORDIC_READ                      BootloaderCommand = 0x10
	BOOTLOADER_COMMAND_NORDIC_WRITE                        BootloaderCommand = 0x20
	BOOTLOADER_COMMAND_NORDIC_ERASE_PAGE                   BootloaderCommand = 0x30
	BOOTLOADER_COMMAND_REBOOT                              BootloaderCommand = 0x70
	BOOTLOADER_COMMAND_GET_MEMORY_INFO                     BootloaderCommand = 0x80
	BOOTLOADER_COMMAND_GET_BOOTLOADER_VERSION_STRING       BootloaderCommand = 0x90
	BOOTLOADER_COMMAND_FLASH_READ_SIGNATURE                BootloaderCommand = 0xb0
	BOOTLOADER_COMMAND_TI_WRITE_TO_RAM_BUFFER              BootloaderCommand = 0xc0
	BOOTLOADER_COMMAND_TI_WRITE_TO_RAM_BUFFER_INVALID_ADDR BootloaderCommand = 0xc1
	BOOTLOADER_COMMAND_TI_WRITE_TO_RAM_BUFFER_OVERFLOW     BootloaderCommand = 0xc2
	BOOTLOADER_COMMAND_TI_FLASH                            BootloaderCommand = 0xd0
	BOOTLOADER_COMMAND_FLASH_INVALID_ADDR                  BootloaderCommand = 0xd1
	BOOTLOADER_COMMAND_FLASH_WRONG_CRC                     BootloaderCommand = 0xd2
	BOOTLOADER_COMMAND_FLASH_PAGE0_INVALID                 BootloaderCommand = 0xd3
	BOOTLOADER_COMMAND_FLASH_RAM_INVALID_ORDER             BootloaderCommand = 0xd4
	BOOTLOADER_COMMAND_TI_FLASH_WRITE_SIGNATURE            BootloaderCommand = 0xe0
)

type BootloaderSubCommandFlash byte

const (
	BOOTLOADER_SUB_COMMAND_FLASH_ERASE_ALL        BootloaderSubCommandFlash = 0x00
	BOOTLOADER_SUB_COMMAND_FLASH_WRITE_RAM_BUFFER BootloaderSubCommandFlash = 0x01
	BOOTLOADER_SUB_COMMAND_FLASH_CLEAR_RAM_BUFFER BootloaderSubCommandFlash = 0x02
	BOOTLOADER_SUB_COMMAND_FLASH_CHECK_CRC        BootloaderSubCommandFlash = 0x03
)

type BootloaderReport struct {
	Cmd  BootloaderCommand
	Addr uint16
	Len  byte
	Data [28]byte
}

func (r *BootloaderReport) String() (res string) {
	res = fmt.Sprintf("Bootloader Report cmd: %02x, Addr: %#04x, len: %d, data: % x", r.Cmd, r.Addr, r.Len, r.Data[:r.Len])

	return res
}

func (r *BootloaderReport) FromWire(payload []byte) (err error) {
	r.Cmd = BootloaderCommand(payload[0])
	r.Addr = uint16(payload[1]) << 8
	r.Addr += uint16(payload[2])
	r.Len = payload[3]

	copy(r.Data[:], payload[4:])
	return nil
}

func (r *BootloaderReport) ToWire() (payload []byte, err error) {
	payload = make([]byte, 32)
	payload[0] = byte(r.Cmd)
	payload[1] = byte(r.Addr >> 8)
	payload[2] = byte(r.Addr & 0x00ff)
	payload[3] = r.Len
	copy(payload[4:], r.Data[:])
	return payload, nil
}
