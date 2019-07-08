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
package main

import (
	"fmt"
	"github.com/mame82/munifying/unifying"
	"github.com/mame82/munifying/cmd"
)



func main() {
	cmd.Execute()
}

func mainold() {
	enterPairing := true
	unpairAll := true

	usb, err := unifying.NewLocalUSBDongle()
	if err != nil {
		panic(err)
	}
	defer usb.Close()

	usb.SetShowInOut(false)

	//usb.GetDeviceActivityCounters()

	set,err := usb.GetSetInfo()
	if err == nil {
		fmt.Println(set.String())
		set.StoreAutoname()
	}

	//usb.PrintInfoForAllConnectedDevices()

	if unpairAll {
		for i:=byte(1);i<7;i++ {
			usb.Unpair(i)
		}
	}

	//Pair new device
	if enterPairing {
		deviceNumber := byte(0x01) //According to specs: Same value as device index transmitted in 0x41 notification, but we haven't tx'ed anything
		openLockTimeout := byte(60)
		pe := usb.EnablePairing(openLockTimeout, deviceNumber,true)
		fmt.Printf("Pairing mode exitted: %v\n", pe)
	}

	return


	// Memdump using undocumented vendor command

	// Test mem dump
	startAddr := uint16(0x0000)
	endAddr := uint16(0xf000)
	linebreakCount := 20
	linecount := 0
	for addr := startAddr; addr < endAddr; addr++ {
		dB,eDb := usb.DumpFlashByte(addr)
		byteStr := "xx"
		if eDb == nil {
			byteStr = fmt.Sprintf("%02x", dB)
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
	}



	responses, err := usb.HIDPP_SendAndCollectResponses(0xff, unifying.HIDPP_MSG_ID_GET_REGISTER_REQ, []byte{byte(0xd4), 0x00,0xff})
	for _, r := range responses {
		fmt.Println(r.String())
	}


/*
	// Request dongle firmware version from firmware register
	fmt.Println("!!!Request dongle firmware version from firmware register")
	responses, err := usb.HIDPP_SendAndCollectResponses(0xff, HIDPP_MSG_ID_GET_REGISTER_REQ, []byte{byte(DONGLE_HIDPP_REGISTER_FIRMWARE), 0x01, 0x00})
	for _, r := range responses {
		fmt.Println(r.String())
	}
	fmt.Println("!!!Request dongle firmware version from firmware register\n")

	// Request dongle firmware build version from firmware register
	fmt.Println("!!!Request dongle firmware build version from firmware register")
	responses, err = usb.HIDPP_SendAndCollectResponses(0xff, HIDPP_MSG_ID_GET_REGISTER_REQ, []byte{byte(DONGLE_HIDPP_REGISTER_FIRMWARE), 0x02, 0x00})
	for _, r := range responses {
		fmt.Println(r.String())
	}
	fmt.Println("!!!Request dongle firmware build version from firmware register\n")

	// Request dongle bootloader version from firmware register
	fmt.Println("!!!Request dongle bootloader version from firmware register")
	responses, err = usb.HIDPP_SendAndCollectResponses(0xff, HIDPP_MSG_ID_GET_REGISTER_REQ, []byte{byte(DONGLE_HIDPP_REGISTER_FIRMWARE), 0x04, 0x00})
	for _, r := range responses {
		fmt.Println(r.String())
	}
	fmt.Println("!!!Request dongle bootloader version from firmware register\n")
*/

/*

	//Enable wireless notifications (to be able to receive infos via device connect notify on new devices)
	fmt.Println("!!!Enable wireless notifications")
	responses, err = usb.HIDPP_SendAndCollectResponses(0xff, HIDPP_MSG_ID_SET_REGISTER_REQ, []byte{byte(DONGLE_HIDPP_REGISTER_WIRELESS_NOTIFICATIONS), 0x00, 0x01})
	for _, r := range responses {
		fmt.Println(r.String())
	}
	fmt.Println("!!!END enable wireless notifications\n")

	fmt.Println("!!!Get connected device info")
	responses, err = usb.HIDPP_SendAndCollectResponses(0xff, HIDPP_MSG_ID_SET_REGISTER_REQ, []byte{byte(DONGLE_HIDPP_REGISTER_CONNECTION_STATE), 0x02})
	for _, r := range responses {
		fmt.Println(r.String())
	}
	fmt.Println("!!!END get connected device info\n")
*/

	//Following part is only for firmware hot-patched with illegal HID command for memdump
	//Test dump mem
	usb.SetShowInOut(false)
	for pos := 0x8000; pos < 0x8400; pos += 0x10 {
		memType := byte(0x01)
		addrH := byte((pos & 0xff00) >> 8)
		addrL := byte(pos & 0xff)
		//fmt.Printf("Reading MemType: %#02x from %#02x%02x\n", memType, addrH, addrL)
		responses, err = usb.HIDPP_SendAndCollectResponses(0xff, unifying.HIDPP_MSG_ID_GET_LONG_REGISTER_REQ, []byte{byte(unifying.DONGLE_HIDPP_REGISTER_PAIRING_INFORMATION), memType + 0x80, addrH, addrL})
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
	fmt.Println("Code")
	for pos := 0x6000; pos < 0x7400; pos += 0x10 {
		memType := byte(0x02)
		addrH := byte((pos & 0xff00) >> 8)
		addrL := byte(pos & 0xff)
		//fmt.Printf("Reading MemType: %#02x from %#02x%02x\n", memType, addrH, addrL)
		responses, err = usb.HIDPP_SendAndCollectResponses(0xff, unifying.HIDPP_MSG_ID_GET_LONG_REGISTER_REQ, []byte{byte(unifying.DONGLE_HIDPP_REGISTER_PAIRING_INFORMATION), memType + 0x80, addrH, addrL})
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

	//return
	//End test

	//Parse successive input reports in endless loop
	fmt.Println("!!!!Parse successive input reports in endless loop...")
	for {
		rspUSB, err := usb.ReceiveUSBReport(500)
		if err == nil {
			fmt.Println(rspUSB.String())
		}
	}

}
