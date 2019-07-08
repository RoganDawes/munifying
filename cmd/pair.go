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
	"os"
	"os/signal"
)


var pairCmd = &cobra.Command{
	Use:   "pair",
	Short: "Pair new devices to first receiver found on USB",
	Long: "",
	Run: func(cmd *cobra.Command, args []string) {
		usb, err := unifying.NewLocalUSBDongle()
		if err != nil {
			panic(err)
		}
		defer usb.Close()

		usb.SetShowInOut(false) //don't show raw HID traffic

		//Pair new device
		deviceNumber := byte(0x01) //According to specs: Same value as device index transmitted in 0x41 notification, but we haven't tx'ed anything
		openLockTimeout := byte(60)
		err = usb.EnablePairing(openLockTimeout, deviceNumber,false)
		if err != nil {
			fmt.Println(err)
			return
		}

		pairingAborted := false
		go func() {
			cleanupDone := make(chan struct{})
			signalChan := make(chan os.Signal, 1)

			signal.Notify(signalChan, os.Interrupt)
			go func() {
				<-signalChan
				fmt.Println("\nReceived an interrupt, exit pairing mode...\n")

				usb.DisablePairing()
				close(cleanupDone)
				pairingAborted = true

			}()
			// wait for pairing to be disabled
			<-cleanupDone

			return
		}()

		//Parse successive input reports till new "receiver lock information" with lock closed occurs
		fmt.Println("Printing follow up reports ...")
		for {
			rspUSB, err := usb.ReceiveUSBReport(500)
			if err == nil {

				//fmt.Println(rspUSB.String())
				if rspUSB.IsHIDPP() {
					hidppRsp := rspUSB.(*unifying.HidPPMsg)
					if hidppRsp.MsgSubID == unifying.HIDPP_MSG_ID_RECEIVER_LOCKING_INFORMATION && (hidppRsp.Parameters[0]&0x01) == 0 {
						switch hidppRsp.Parameters[1] {
						case 0x00:
							fmt.Println("Device paired successfully") //"no error"
						case 0x01:
							fmt.Println("timeout")
						case 0x02:
							fmt.Println("unsupported device")
						case 0x03:
							fmt.Println("too many devices")
						case 0x06:
							fmt.Println("connection sequence timeout")
						default:
							fmt.Println("pairing aborted with unknown reason")
						}

						fmt.Println("Pairing lock closed")
						return
					}

					// device connection
					if hidppRsp.MsgSubID == unifying.HIDPP_MSG_ID_DEVICE_CONNECTION {
						devIdx := hidppRsp.DeviceID
						wpid := uint16(hidppRsp.Parameters[3])<<8 + uint16(hidppRsp.Parameters[2])
						encrypted := false
						if (hidppRsp.Parameters[1] & (1 << 5)) > 0 {
							encrypted = true
						}
						link := true
						if (hidppRsp.Parameters[1] & (1 << 6)) > 0 {
							link = false
						}
						fmt.Printf("DEVICE CONNECTION ON INDEX: %02x TYPE: %s WPID: %#04x ENCRYPTED: %v CONNECTED: %v\n", devIdx, unifying.DeviceType(hidppRsp.Parameters[1]&0x0F), wpid, encrypted, link)

						//request additional information
					}
				}

				if rspUSB.IsDJ() {
					djRsp := rspUSB.(*unifying.DJReport)
					//fmt.Println(djRsp.String())
					if (djRsp.Type == unifying.DJ_REPORT_TYPE_NOTIFICATION_DEVICE_PAIRED) {
						fmt.Println("New device paired")
						break //device paired, exit read loop
					}
				}

			} else {
				if pairingAborted {
					fmt.Println("Pairing aborted")
					break
				}
			}
		}

		// re-evaluate paired devices and print

		set,err := usb.GetSetInfo()
		if err == nil {
			fmt.Println(set.String())
		}

	},
}

func init() {
	rootCmd.AddCommand(pairCmd)
}
