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

func StoreDongleInfo() {
	usb, err := unifying.NewLocalUSBDongle()
	if err != nil {
		panic(err)
	}
	defer usb.Close()

	set,err := usb.GetSetInfo()
	if err == nil {
		fmt.Println(set.String())
		set.StoreAutoname()
	}
}


// infoCmd represents the info command
var storeCmd = &cobra.Command{
	Use:   "store",
	Short: "**not in pre-release** Store relevant information of first receiver found on USB to file (usable with 'mjackit')",
	Long: "",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("command not available in pre-release")
	},
}

func init() {
	rootCmd.AddCommand(storeCmd)

}
