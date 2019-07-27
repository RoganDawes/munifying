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
	//fw_file := "/root/jacking/firmware/RQR39.04_B0036_G603.bin"
	//fw_file := "/root/jacking/firmware/RQR41.00_B0004_SPOTLIGHT.bin"
	//fw_file := "/root/jacking/firmware/RQR45.00_B0002_R500.bin"
	//fw_file := "/root/jacking/firmware/RQR24.07_B0030.bin"

	fw_file := "/root/jacking/firmware/RQR12.01_B0019_dump.raw"
	//fw_file := "/root/jacking/firmware/RQR12.09_B0030.bin"
	//fw_file := "/root/jacking/firmware/RQR12.07_B0029.bin"
	//fw_file := "/root/jacking/firmware/RQR12.05_B0028.bin"

	//fw_file := "/root/jacking/firmware/RQR24.06_B0030.bin"
	//fw_file := "/root/jacking/firmware/RQR39.04_B0036_G603_patch_for_BOT03.01.bin"


	//fw_sig_file := "/root/jacking/firmware/RQR24.07_B0030_sig.bin"
	fw_sig_file := "/root/jacking/firmware/RQR12.09_B0030_sig.bin"

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
			fmt.Printf("Receiver is running a firmware with uknown major version %#02x\n")
		}

		usbReceiver.GetReceiverFirmwareBuildVersion()

		fmt.Println("Try to reset dongle into bootloader mode ...")
		usbReceiver.SwitchToBootloader()


		fmt.Println("... try to re-open dongle in bootloader mode in 5 seconds...")
		time.Sleep(time.Second * 3)

	}


	//Try to open receiver in bootloader mode
	usbReceiverBL, err := unifying.NewUSBBootloaderDongle()
	if err != nil {
		fmt.Printf("can not open receiver in bootloader mode: %v\n", err)
		return
	} else {
		defer usbReceiverBL.Close()
	}
	usbReceiverBL.SetShowInOut(false)




	err = usbReceiverBL.FlashReceiver(firmware)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	} else {
		usbReceiverBL.Reboot()
	}
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
