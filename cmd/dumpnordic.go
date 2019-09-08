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

func DumpDongleNordic() (err error) {
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

		if fwMaj == 0x12 || fwMaj == 0x21 {
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

	fwStart,fwEnd,_,tmpErr := usbReceiverBL.GetFirmwareMemoryInfo()
	if tmpErr != nil {
		return errors.New("Can't determin start/end offset of firmware to dump")
	}

	fwEnd = 0x7fff // read full dump, including data (not only firmware region)
	dumpbytes := make([]byte,0)
	slen := uint16(0x1c)
	for offset := fwStart; offset <= fwEnd; offset += slen {
		if (offset + slen) > fwEnd {
			slen = fwEnd - offset + 1
		}

		err,fwSlice := usbReceiverBL.ReadFirmwareSliceFromFlashNordic(offset, byte(slen))
		if err != nil {
			panic(err)
		}
		dumpbytes = append(dumpbytes, fwSlice...)
	}

	//extract firmware version string
	fwNameLen := int(dumpbytes[0x7fd0])
	fwName := string(dumpbytes[0x7fd1:0x7fd1+fwNameLen])
	botVersion := string(dumpbytes[0x7fb4:0x7fb8])
	botName := fmt.Sprintf("BOT%02x.%02x.B%02x%02x", botVersion[0], botVersion[1], botVersion[2], botVersion[3])
	filename := fmt.Sprintf("dump_%s_%s.bin", fwName, botName)

	//fmt.Printf("FwName: %s\n", fwName)
	//fmt.Printf("BotName: %s\n", botName)

	fmt.Println("========================================================================================================")
	fmt.Printf("Storing firmware dump (including device data and bootloader) to '%s'\n", filename)
	fmt.Println("========================================================================================================")
	ioutil.WriteFile(filename, dumpbytes, 0644)

	//fmt.Printf("% 02x\n", dumpbytes)

	return
}


// infoCmd represents the info command
var dumpnordicCmd = &cobra.Command{
	Use:   "dumpnordic",
	Short: "Dump dongle firmware from Nordic receivers (experimental)",
	Long: "",
	Run: func(cmd *cobra.Command, args []string) {
		DumpDongleNordic()
	},
}

func init() {
	rootCmd.AddCommand(dumpnordicCmd)

}
