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
	"log"
)

var unpairAllCmd = &cobra.Command{
	Use:   "unpairall",
	Short: "Unpair all paired devices of first receiver found on USB",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		usb, err := unifying.NewLocalUSBDongle()
		if err != nil {
			panic(err)
		}
		defer usb.Close()

		usb.SetShowInOut(false)

		set, err := usb.GetSetInfo()
		if err != nil {
			log.Fatal("Can't load devices list for dongle")
		}

		for _, devInfo := range set.ConnectedDevices {
			fmt.Printf("Remove device index %d '%s' from paired devices\n", devInfo.DeviceIndex, devInfo.Name)
			usb.Unpair(devInfo.DeviceIndex + 1)
		}
	},
}

func init() {
	rootCmd.AddCommand(unpairAllCmd)
}
