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
	"github.com/mame82/munifying/helper"
	"github.com/mame82/munifying/unifying"
	"github.com/spf13/cobra"
	"log"
)


func SelectPaired(usb *unifying.LocalUSBDongle) (devInfo unifying.DeviceInfo, err error) {

	usb.SetShowInOut(false)
	si,err := usb.GetSetInfo()
	if err != nil {
		log.Fatal("Can't load devices list for dongle")
	}


	fmt.Println(si.String())

	if si.Dongle.NumConnectedDevices == 0 {
		fmt.Println("no device paired to receiver")
		return devInfo, errors.New("no device paired to receiver")
	}

	devToUse := si.ConnectedDevices[0]

	if si.Dongle.NumConnectedDevices > 1 {
		fmt.Println("Multiple devices connected to target dongle, select device to use...")

		options := make([]string, si.Dongle.NumConnectedDevices)
		for i, d := range si.ConnectedDevices {
			options[i] = fmt.Sprintf("%02x:%02x:%02x:%02x:%02x %s '%s')", d.RFAddr[0], d.RFAddr[1], d.RFAddr[2], d.RFAddr[3], d.RFAddr[4], d.DeviceType.String(), d.Name)
		}

		var selected int
		for {
			s, eS := helper.Select("choose device to sniff: ", options)
			if eS != nil {
				fmt.Println(eS)
			} else {
				selected = s
				break
			}
		}

		devToUse = si.ConnectedDevices[selected]

	}

	return devToUse, nil
}


var unpairCmd = &cobra.Command{
	Use:   "unpair",
	Short: "Unpair devices of first receiver found on USB",
	Long: "",
	Run: func(cmd *cobra.Command, args []string) {
		usb, err := unifying.NewLocalUSBDongle()
		if err != nil {
			panic(err)
		}
		defer usb.Close()

		di,err := SelectPaired(usb)
		if err == nil {
			fmt.Printf("Remove device number %d '%s' from paired devices", di.DeviceIndex, di.Name)
			usb.Unpair(di.DeviceIndex+1)
		}
	},
}


func init() {
	rootCmd.AddCommand(unpairCmd)
}
