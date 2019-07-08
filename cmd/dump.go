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
	"io/ioutil"

	"github.com/spf13/cobra"
)

func DumpDongleInfo() {
	usb, err := unifying.NewLocalUSBDongle()
	if err != nil {
		panic(err)
	}
	defer usb.Close()

	usb.SetShowInOut(false)

	startAddr := uint16(0x0000)
	endAddr := uint16(0xffff)
	linebreakCount := 32
	linecount := 0

	rawdata := []byte{}

	for addr := startAddr; addr <= endAddr; addr++ {
		dB,eDb := usb.DumpFlashByte(addr)
		byteStr := "xx"
		if eDb == nil {
			byteStr = fmt.Sprintf("%02x", dB)
			rawdata = append(rawdata, dB)
		} else {
			rawdata = append(rawdata, 0xff)
		}

		if linecount == 0 {
			fmt.Printf("%#04x: ", addr)
		}
		linecount++
		fmt.Print(byteStr)
		if linecount == linebreakCount {
			fmt.Println()
			linecount = 0
		}

		if addr == endAddr {
			break // account for uint16 overflow
		}
	}

	set,err := usb.GetSetInfo()
	if err == nil {
		filename := fmt.Sprintf("rawdump_%02x_%02x_%02x_%02x.dump", set.Dongle.Serial[0], set.Dongle.Serial[1], set.Dongle.Serial[2], set.Dongle.Serial[3])
		eWrite := ioutil.WriteFile(filename, rawdata, 0644)
		if eWrite == nil {
			fmt.Printf("dumped data stored to file '%s'\n", filename)
		}
	}
}


// infoCmd represents the info command
var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump dongle memory utilizing secret HID++ command",
	Long: "",
	Run: func(cmd *cobra.Command, args []string) {
		DumpDongleInfo()
	},
}

func init() {
	rootCmd.AddCommand(dumpCmd)

}
