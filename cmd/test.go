// Copyright © 2019 Marcus Mengs
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"errors"
	"fmt"
	"github.com/mame82/munifying/unifying"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"time"
)
func Test() {
	//read in firmware file
	//fw_file := "/root/jacking/firmware/RQR39.03_B0035_fake.bin"
	fw_file := "/root/jacking/firmware/RQR39.04_B0036_G603.bin"
	//fw_file := "/root/jacking/firmware/RQR41.00_B0004_SPOTLIGHT.bin"
	//fw_file := "/root/jacking/firmware/RQR45.00_B0002_R500.bin"
	//fw_file := "/root/jacking/firmware/RQR24.07_B0030.bin"

	//fw_file := "/root/jacking/firmware/RQR24.06_B0030.bin"
	//fw_file := "/root/jacking/firmware/RQR39.04_B0036_G603_patch_for_BOT03.01.bin"


	fw_sig_file := "/root/jacking/firmware/RQR24.07_B0030_sig.bin"



	firmware, err := unifying.ParseFirmware(fw_file)
	if err != nil {
		log.Fatal(err)
	}

	// add signature data
	fw_sig_bytes, err := ioutil.ReadFile(fw_sig_file)
	if err != nil {
		fmt.Printf("error reading firmware signature file: %v\n", err)
	} else {
		firmware.AddSignature(fw_sig_bytes)
	}


	fmt.Println(firmware.String())

/*
	fw_patched,_ := firmware.BaseImageDowngradeFromBL0302ToBL0301()
	fmt.Printf("%02x\n", fw_patched)
	prefix := make([]byte,0x400)
	fw_patched = append(prefix, fw_patched...)
	ioutil.WriteFile("test.raw", fw_patched, os.FileMode(440))
	return
*/
	err = FlashTIReceiver(firmware)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}
}

func FlashTIReceiver(firmware * unifying.Firmware) (err error) {
	usbReceiver, err := unifying.NewLocalUSBDongle()
	if err != nil {
		fmt.Println(err)
	} else {
		defer usbReceiver.Close()
		usbReceiver.SetShowInOut(true)
		fwMaj, _, err := usbReceiver.GetReceiverFirmwareMajorMinorVersion()
		if err != nil {
			log.Fatal(err)
		}

		if fwMaj == 0x12 {
			return errors.New("dongle has a Nordic chip, thus can not be flashed with this tool")
		}

		usbReceiver.GetReceiverFirmwareBuildVersion()

		fmt.Println("Try to reset dongle into bootloader mode ...")
		usbReceiver.SwitchToBootloader()

		fmt.Println("... try to re-open dongle in bootloader mode in 5 seconds...")
		time.Sleep(time.Second * 2)

	}

	signature_required := false


	usbReceiverBL, err := unifying.NewUSBBootloaderDongle()
	if err != nil {
		return err
	} else {
		defer usbReceiverBL.Close()
	}
	usbReceiverBL.SetShowInOut(false)


	_,BLmaj,BLmin,_,err := usbReceiverBL.GetBLVersionString()
	if err != nil {
		log.Fatal(err)
	}
	if BLmaj != 0x03 {
		return errors.New("bootloader major version hints that this is not a Texas Instruments CC2544 based Logitech dongle")
	}

	if BLmaj >= 3 && BLmin > 2 {
		fmt.Println("CAUTION: According to bootloader version, only signed firmwares are accepted!")
		signature_required = true

		if !firmware.HasSignature {
			return errors.New("provided firmware has no signature, but the bootloader requires one.")
		}
	}


	fmt.Println("Retrieving firmware memory info from bootloader...")
	fwStartAddr, fwEndAddr, fwFlashWriteBufSize, err := usbReceiverBL.GetFirmwareMemoryInfo()
	if err != nil {
		return err
	}

	fwbytes, err := firmware.BaseImage()
	if err != nil {
		return errors.New(fmt.Sprintf("error fetching firmware base image: %v", err))

	}

	intended_fw_size := fwEndAddr - fwStartAddr + 1
	if intended_fw_size != firmware.Size {
		if firmware.Size == 0x6000 && intended_fw_size == 0x6800 && BLmaj <= 3 && BLmin <= 1 {
			fmt.Println("According to the size, the provided firmware seems to be build for a Bootloader version >= 03.02 (signed)")
			fmt.Println("Target receiver's Bootloader version is <=03.01 (unsigned), try to create a downgraded firmware...")

			fmt.Println("provided firmware file has wrong size, trying to resize")

			if (signature_required) {
				return errors.New("can not resize the firmware without invalidating the signature, aborting...")
			}

			//grow firmware to needed size
			fwbytes, err = firmware.BaseImageDowngradeFromBL0302ToBL0301()
			if err != nil {
				return errors.New(fmt.Sprintf("failed to resize firmware: %v\n", err))
			}
		} else {
			return errors.New("Firmware doesn't match target bootloader's memory layout and can not be patched")
		}

	}

	//erase flash
	//ToDo: let user decide to continue
	fmt.Println("Erasing dongle flash: CAUTION the dongle will not be usable, if successive operations fail")
	err = usbReceiverBL.EraseFlashTI()
	if err != nil {
		return err
	}

	//clear RAM buffer
	fmt.Println("Clearing RAM buffer for flash write...")
	err = usbReceiverBL.EraseFlashTI()
	if err != nil {
		return err
	}

	for addr := fwStartAddr; addr <= fwEndAddr; addr += fwFlashWriteBufSize {
		chunk := fwbytes[addr-fwStartAddr : addr-fwStartAddr+fwFlashWriteBufSize]
		//fmt.Printf("%04x: %x\n", addr, chunk)

		// split flash chunk into RAM buffer chunks and upload to RAM Buffer
		for ramAddr := uint16(0x0000); ramAddr < fwFlashWriteBufSize; ramAddr += 16 {
			ram_chunk := chunk[ramAddr : ramAddr+16]
			//fmt.Printf("\tRAM buffer %04x: %x\n", ramAddr, ram_chunk)

			// Write to RAM buffer
			err = usbReceiverBL.WriteFirmwareSliceToRAMBufferTI(ramAddr, ram_chunk)
			if err != nil {
				return err
			}
		}

		// write RAM buffer to flash at proper address
		err = usbReceiverBL.StoreRAMBufferToFlashAddrTI(addr)
		if err != nil {
			return err
		}
	}

	// Write signature
	if signature_required {
		fmt.Println("Trying to write signature for firmware")
		for sig_addr := uint16(0x0000); sig_addr <= uint16(0x00ff); sig_addr += 0x10 {
			sig_chunk := firmware.Signature[sig_addr : sig_addr+0x10]

			// write signature slice
			err = usbReceiverBL.WriteSignatureSliceTI(sig_addr, sig_chunk)
			if err != nil {
				return err
			}
		}
	}

	// Check CRC
	err = usbReceiverBL.CheckFirmwareCrcAndSignatureTI()
	if err != nil {
		return err
	}

	fmt.Println("Firmware flashing SUCCEEDED")
	usbReceiverBL.Reboot()
	return nil
}

// infoCmd represents the info command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "undocumented - used for testing",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		Test()
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
