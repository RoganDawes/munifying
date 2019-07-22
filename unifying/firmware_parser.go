package unifying

import (
	"errors"
	"fmt"
	"github.com/sigurn/crc16"
	"io/ioutil"
	"strings"
)

type Firmware struct {
	RawData      []byte
	Size         uint16
	StartOffset  uint16
	LastOffset   uint16
	HasBL        bool
	CRC          uint16
	TailPos      uint16
	Signature    [256]byte
	HasSignature bool
}

func (f *Firmware) AddSignature(sig []byte) (err error) {
	fmt.Printf("signature length length: %#x (%d) bytes\n", len(sig), len(sig))

	if len(sig) != 256 {
		f.HasSignature = false
		return errors.New("wrong size of firmware signature")
	}
	copy(f.Signature[:], sig)
	f.HasSignature = true
	return
}

func (f *Firmware) BaseImage() (img []byte, err error) {
	img = make([]byte, f.Size)
	copy(img, f.RawData[f.StartOffset:f.StartOffset+f.Size])
	return
}

// shrinks or enlargens the image as needed and recalculates CRC (of course this doesn't work for signed bootloaders >BOT03.02_Bxxxx)
func (f *Firmware) BaseImageResized(size uint16) (img []byte, err error) {
	if size == f.Size {
		return f.BaseImage()
	}

	if (size < f.Size) {
		//shrink, check if we don't end up in the middle of code (not 0xff)
		if f.RawData[f.StartOffset+size-1] != 0xff {
			return nil, errors.New("can't shrink the firmware image, because code would be truncated")
		}
		img = make([]byte, size)
		copy(img, f.RawData[f.StartOffset:])
		return img, nil
	} else {
		// copy in image without tail (CRC and magic numbers)
		size_to_tail := f.TailPos - f.StartOffset
		img = make([]byte, size_to_tail)
		copy(img, f.RawData[f.StartOffset:f.StartOffset+size_to_tail])

		// build part to append
		bytes_to_append := make([]byte, size-size_to_tail)
		for i := 0; i < len(bytes_to_append)-4; i++ {
			bytes_to_append[i] = 0xff
		}
		//replace last 4 bytes with magic number
		copy(bytes_to_append[:len(bytes_to_append)-4], []byte{0xfe, 0xc0, 0xad, 0xde})

		//concat
		img = append(img, bytes_to_append...)

		//recalculate CRC
		calculated_crc := crc16.Checksum(img[:len(img)-6], crc16.MakeTable(crc16.CRC16_CCITT_FALSE))
		img[len(img)-6] = byte(calculated_crc & 0x00ff)
		img[len(img)-5] = byte(calculated_crc >> 8)

		return img, nil
	}

	return
}

func (f *Firmware) String() string {
	res := ""
	res += fmt.Sprintf("Size %#04x start: %#04x end %#04x CRC %#04x\n", f.Size, f.StartOffset, f.LastOffset, f.CRC)
	return res
}

func ParseFirmware(filepath string) (f *Firmware, err error) {
	fmt.Println("Parsing raw firmware blob ...")
	f = &Firmware{}
	f.RawData, err = ioutil.ReadFile(filepath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error reading firmware file: %v", err))
	}

	// if a bootloader is present the following data is present
	// - 0x03f8 uint16, USB VID (LE)
	// - 0x03fa uint16, USB PID (LE)
	// - 0x03fc byte, BL major
	// - 0x03fd byte, BL minor
	// - 0x03fe uint16, BL Build number
	assumed_bootloader := f.RawData[:0x0400]

	// check USB VID in order to determine if a BL is prepended to the firmware blob (Logitech VID is 0x046d)
	if (assumed_bootloader[0x3f8] == 0x6d && assumed_bootloader[0x3f9] == 0x04) {
		f.HasBL = true
		f.StartOffset = 0x400
		fmt.Println("...firmware blob has a bootloader prepended")
	} else {
		f.HasBL = false
		f.StartOffset = 0x0000
		fmt.Println("...firmware blob has no bootloader prepended")
	}

	// ToDo: The firmware type could be determined from bootloader PID
	if pos := strings.Index(string(f.RawData[f.StartOffset:]), "\xfe\xc0\xad\xde"); pos < 0 {
		//can't find magic bytes
		return nil, errors.New("seems to be no valid Logitech firmware for TI, magic bytes missing")
	} else {
		f.Size = uint16(pos) + 4
		f.LastOffset = f.Size + f.StartOffset - 1
		f.TailPos = f.StartOffset + f.Size - 6
	}

	fmt.Println(f.String())

	// extract CRC
	f.CRC = uint16(f.RawData[f.TailPos+1])<<8 | uint16(f.RawData[f.TailPos])

	// check CRC
	calculated_crc := crc16.Checksum(f.RawData[f.StartOffset:f.StartOffset+f.Size-6], crc16.MakeTable(crc16.CRC16_CCITT_FALSE))
	if calculated_crc != f.CRC {
		return nil, errors.New(fmt.Sprintf("Firmware has wrong CRC (inteded %#04x, found %#04x)", calculated_crc, f.CRC))
	}
	fmt.Printf("...firmware CRC correct: %04x\n", calculated_crc)

	return f, nil
}
