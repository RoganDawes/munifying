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
)

func PatchdumpDongle() {
	usb, eDongle := unifying.NewLocalUSBDongle()
	if eDongle != nil {
		panic(eDongle)
	}
	defer usb.Close()

	//Following part is only for firmware hot-patched with illegal HID command for memdump
	//Test dump mem
	usb.SetShowInOut(false)
	for pos := 0x8000; pos < 0x8400; pos += 0x10 {
		memType := byte(0x01)
		addrH := byte((pos & 0xff00) >> 8)
		addrL := byte(pos & 0xff)
		//fmt.Printf("Reading MemType: %#02x from %#02x%02x\n", memType, addrH, addrL)
		responses, _ := usb.HIDPP_SendAndCollectResponses(0xff, unifying.HIDPP_MSG_ID_GET_LONG_REGISTER_REQ, []byte{byte(unifying.DONGLE_HIDPP_REGISTER_PAIRING_INFORMATION), memType + 0x80, addrH, addrL})
		for _, r := range responses {
			if r.IsHIDPP() {
				rpp := r.(*unifying.HidPPMsg)
				if rpp.MsgSubID == unifying.HIDPP_MSG_ID_GET_LONG_REGISTER_RSP && rpp.Parameters[0] == byte(unifying.DONGLE_HIDPP_REGISTER_PAIRING_INFORMATION) {
					fmt.Printf("MemType %#02x addr %#02x%02x: % 02x\n", memType, addrH, addrL, rpp.Parameters[1:])
				}
			}
			//fmt.Println(r.String())
		}
		//fmt.Printf("Done reading MemType: %#02x from %#02x%02x\n", memType, addrH, addrL)
	}
	usb.SetShowInOut(true)

	usb.SetShowInOut(false)
	//fmt.Println("Code")
	for pos := 0x6000; pos < 0x7400; pos += 0x10 {
		memType := byte(0x02)
		addrH := byte((pos & 0xff00) >> 8)
		addrL := byte(pos & 0xff)
		//fmt.Printf("Reading MemType: %#02x from %#02x%02x\n", memType, addrH, addrL)
		responses, _ := usb.HIDPP_SendAndCollectResponses(0xff, unifying.HIDPP_MSG_ID_GET_LONG_REGISTER_REQ, []byte{byte(unifying.DONGLE_HIDPP_REGISTER_PAIRING_INFORMATION), memType + 0x80, addrH, addrL})
		for _, r := range responses {
			if r.IsHIDPP() {
				rpp := r.(*unifying.HidPPMsg)
				if rpp.MsgSubID == unifying.HIDPP_MSG_ID_GET_LONG_REGISTER_RSP && rpp.Parameters[0] == byte(unifying.DONGLE_HIDPP_REGISTER_PAIRING_INFORMATION) {
					fmt.Printf("MemType %#02x addr %#02x%02x: % 02x\n", memType, addrH, addrL, rpp.Parameters[1:])
				}
			}
			//fmt.Println(r.String())
		}
		//fmt.Printf("Done reading MemType: %#02x from %#02x%02x\n", memType, addrH, addrL)
	}
	usb.SetShowInOut(true)
}


// infoCmd represents the info command
var patchdumpCmd = &cobra.Command{
	Use:   "patchdump",
	Short: "Dumps RAM using firmwaremod for CU0007 (not published)",
	Long: "",
	Run: func(cmd *cobra.Command, args []string) {
		PatchdumpDongle()
	},
}

func init() {
	rootCmd.AddCommand(patchdumpCmd)

}
