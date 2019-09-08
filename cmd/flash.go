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

var (
	tmpFirmwarePathRaw  = ""
	tmpFirmwarePathHex  = ""
	tmpSignaturePathRaw = ""
)

func FlashFirmwareFromHexFile(fw_hex_file string, fw_sig_file string) {
	fw, err := unifying.ParseFirmwareHex(fw_hex_file)
	if err == nil {
		fmt.Println(fw.String())
	} else {
		fmt.Println(err)
		return
	}

	// add signature data
	if len(fw_sig_file) > 0 {
		fw_sig_bytes, err := ioutil.ReadFile(fw_sig_file)
		if err != nil {
			fmt.Printf("error reading firmware signature file, %v\n", err)
			fmt.Println("...continue without signature")
		} else {
			if fw.HasSignature {
				fmt.Println("WARNING: The firmware file already has a signature included, but the provided signature")
				fmt.Println("file will be used instead.")
			}
			fw.AddSignature(fw_sig_bytes)
		}
	}


	if err := FlashFirmware(fw); err != nil {
		fmt.Println("Error", err)
	}
}

func FlashFirmwareFromRawFiles(fw_file string, fw_sig_file string) {

	fw_bin, err := ioutil.ReadFile(fw_file)
	if err != nil {
		fmt.Printf("error reading firmware file: %v", err)
	} else {
		fmt.Printf("Opened firmware blob '%s'\n", fw_file)
	}

	firmware, err := unifying.ParseFirmwareBin(fw_bin)
	if err != nil {
		log.Fatal(err)
	}

	// add signature data
	if len(fw_sig_file) > 0 {
		fw_sig_bytes, err := ioutil.ReadFile(fw_sig_file)
		if err != nil {
			fmt.Printf("error reading firmware signature file, %v\n", err)
			fmt.Println("...continue without signature")
		} else {
			firmware.AddSignature(fw_sig_bytes)
		}
	}

	if err := FlashFirmware(firmware); err != nil {
		fmt.Println("Error", err)
	}
}

func FlashFirmware(firmware *unifying.Firmware) (err error) {
	fmt.Println("trying to flash firmware...")
	fmt.Println(firmware.String())

	/*
		fw_patched,_ := firmware.BaseImageDowngradeFromBL0302ToBL0301()
		fmt.Printf("%02x\n", fw_patched)
		prefix := make([]byte,0x400)
		fw_patched = append(prefix, fw_patched...)
		ioutil.WriteFile("test.raw", fw_patched, os.FileMode(440))
		return
	*/

	// Access receiver to obtain info on running firmware and reset to bootloader mode
	usbReceiver, err := unifying.NewLocalUSBDongle()
	if err != nil {
		fmt.Println(err)
	} else {
		defer usbReceiver.Close()
		usbReceiver.SetShowInOut(false)
		fwMaj, _, err := usbReceiver.GetReceiverFirmwareMajorMinorVersion()
		if err != nil {
			log.Fatal(err)
		}

		if fwMaj == 0x12 {
			fmt.Println("Receiver is running a Nordic firmware")
			//return errors.New("dongle has a Nordic chip, thus can not be flashed with this tool")
		} else if fwMaj == 0x24 {
			fmt.Println("Receiver is running a Texas Instruments firmware")
			//return errors.New("dongle has a Nordic chip, thus can not be flashed with this tool")
		} else {
			fmt.Printf("Receiver is running a firmware with uknown major version RQR%02x\n", byte(fwMaj))
		}

		usbReceiver.GetReceiverFirmwareBuildVersion()

		fmt.Println("Try to reset dongle into bootloader mode ...")
		usbReceiver.SwitchToBootloader()

		fmt.Println("... try to re-open dongle in bootloader mode in 3 seconds...")
		time.Sleep(time.Second * 3)

	}

	//Try to open receiver in bootloader mode
	usbReceiverBL, err := unifying.NewUSBBootloaderDongle()
	if err != nil {
		return errors.New(fmt.Sprintf("can not open receiver in bootloader mode: %v\n", err))
	} else {
		defer usbReceiverBL.Close()
	}
	usbReceiverBL.SetShowInOut(false)

	err = usbReceiverBL.FlashReceiver(firmware)
	if err != nil {
		return err
	} else {
		usbReceiverBL.Reboot()
	}

	return nil
}

// infoCmd represents the info command
var flashCmd = &cobra.Command{
	Use:   "flash",
	Short: "Flash a firmware to a receiver (experimental)",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		if len(tmpFirmwarePathHex) > 0 {
			fmt.Printf("Trying to flash hex file '%s'\n", tmpFirmwarePathHex)
			FlashFirmwareFromHexFile(tmpFirmwarePathHex, tmpSignaturePathRaw)
		} else if len(tmpFirmwarePathRaw) > 0 {
			fmt.Printf("Trying to flash raw file '%s'\n", tmpFirmwarePathRaw)
			FlashFirmwareFromRawFiles(tmpFirmwarePathRaw, tmpSignaturePathRaw)
		} else {
			fmt.Println("Error: no firmware file given for flashing")
			fmt.Println()
			fmt.Println("A firmware file could either be provided as hex/shex file with the `-f` flag")
			fmt.Println("or as raw binary using the `-r` flag")
			fmt.Println()
			fmt.Println("If the receiver uses a secure bootloader, the firmwyare has to be signed.")
			fmt.Println("For `shex` firmware files the signature should already be included (in contrast")
			fmt.Println("to `hex` files).")
			fmt.Println()
			fmt.Println("If no signature is included in the firmware file (because the firmware is given")
			fmt.Println("as raw binary or `hex` file) it could be provided as raw 256byte blob in an")
			fmt.Println("additional file using the `-s` flag.")
			cmd.Usage()
		}
		//Test()
	},
}

func init() {
	rootCmd.AddCommand(flashCmd)
	// -h --hex, -r --raw, -s --sig
	flashCmd.Flags().StringVarP(&tmpFirmwarePathHex, "hexfile", "f", "", "path to firmware file in Logitech hex/shex format")
	flashCmd.Flags().StringVarP(&tmpFirmwarePathRaw, "rawfile", "r", "", "path to firmware file in raw binary format")
	flashCmd.Flags().StringVarP(&tmpSignaturePathRaw, "sigfile", "s", "", "path to firmware signature file, if not included in firmware file")
}
