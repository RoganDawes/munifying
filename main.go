package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/gousb"
	"time"
)

type UnifyingDeviceType byte
const (
	UNIFYING_DEVICE_TYPE_UNKNOWN UnifyingDeviceType = 0x00
	UNIFYING_DEVICE_TYPE_KEYBOARD UnifyingDeviceType = 0x01
	UNIFYING_DEVICE_TYPE_MOUSE UnifyingDeviceType = 0x02
	UNIFYING_DEVICE_TYPE_NUMPAD UnifyingDeviceType = 0x03
	UNIFYING_DEVICE_TYPE_PRESENTER UnifyingDeviceType = 0x04
	UNIFYING_DEVICE_TYPE_TRACKBALL UnifyingDeviceType = 0x08
	UNIFYING_DEVICE_TYPE_TOUCHPAD UnifyingDeviceType = 0x09
)
func (t UnifyingDeviceType) String() string {
	switch t {
	case UNIFYING_DEVICE_TYPE_KEYBOARD:
		return "KEYBOARD"
	case UNIFYING_DEVICE_TYPE_MOUSE:
		return "MOUSE"
	case UNIFYING_DEVICE_TYPE_NUMPAD:
		return "NUMPAD"
	case UNIFYING_DEVICE_TYPE_PRESENTER:
		return "PRESENTER"
	case UNIFYING_DEVICE_TYPE_TRACKBALL:
		return "TRACKBALL"
	case UNIFYING_DEVICE_TYPE_TOUCHPAD:
		return "TOUCHPAD"
	case UNIFYING_DEVICE_TYPE_UNKNOWN:
		return "UNKNOWN"
	default:
		return fmt.Sprintf("UNDEFINED DEVICE TYPE %02x", t)
	}

}


type Unifying struct {
	UsbCtx     *gousb.Context
	Dev        *gousb.Device
	Config     *gousb.Config
	IfaceHIDPP *gousb.Interface
	EpInHidPP  *gousb.InEndpoint

	sndQueue chan USBReport
	rcvQueue chan USBReport
	cancel   context.CancelFunc
	ctx      context.Context
}

func (u *Unifying) SendUSBReport(msg USBReport) (err error) {
	u.sndQueue <- msg
	return nil
}

func (u *Unifying) ReceiveUSBReport(timeoutMillis int) (msg USBReport, err error) {
	ctx := context.Background()
	if timeoutMillis > 0 {
		ctxNew, cancel := context.WithTimeout(ctx, time.Duration(timeoutMillis)*time.Millisecond)
		defer cancel()
		ctx = ctxNew
	}

	select {
	case rcv := <-u.rcvQueue:
		msg = rcv
	case <-ctx.Done():
		err = errors.New("timeout reached")
	}

	return
}

func (u *Unifying) rcvLoop() {
	buf := make([]byte, 32)

	for {
		n, err := u.EpInHidPP.ReadContext(u.ctx, buf)
		if err != nil {
			break
		}

		fmt.Printf("\nIn: % #x\n", buf[:n])
		switch USBReportType(buf[0]) {
		case USB_REPORT_TYPE_HIDPP_SHORT:
			fallthrough
		case USB_REPORT_TYPE_HIDPP_LONG:
			inMsg := HidPPMsg{}
			parseErr := inMsg.FromWire(buf[:n])
			if parseErr == nil {
				//fmt.Println("HID++ message")
				u.rcvQueue <- &inMsg
			} else {
				fmt.Printf("Invalid HID++ message: % x\n", buf[:n])
			}
		case USB_REPORT_TYPE_DJ_SHORT:
			fallthrough
		case USB_REPORT_TYPE_DJ_LONG:
			inMsg := DJReport{}
			parseErr := inMsg.FromWire(buf[:n])
			if parseErr == nil {
				//fmt.Println("DJ Report")
				u.rcvQueue <- &inMsg
			} else {
				fmt.Printf("Invalid DJ Report: % x\n", buf[:n])
			}
		default:
			fmt.Printf("Unknown USB input report: % x\n", buf[:n])
		}
	}

	close(u.rcvQueue)
}

func (u *Unifying) sndLoop() {
Outer:
	for {
		select {
		case <-u.ctx.Done():
			break Outer
		case outMsg := <-u.sndQueue:
			outdata, err := outMsg.ToWire()
			if err != nil {
				fmt.Println("Error processing outbound HID++ message", err)
			}

			fmt.Printf("Out: % #x\n", outdata)
			u.Dev.Control(
				0x21,                      //bit7: Host to device, bit6..5: Class: 0x1, bit4..0: Interface: 0x01
				0x09,                      //request: 0x09 SET_REPORT
				0x0200|uint16(outdata[0]), //Output: 0x02, Report ID: 0x10
				2,                         //Index 0x02
				outdata,                   //payload
			)
		}
	}

	close(u.sndQueue)
}

func (u *Unifying) Close() {
	fmt.Println("Closing Unifying dongle...")
	if u.cancel != nil {
		u.cancel()
	}

	if u.IfaceHIDPP != nil {
		u.IfaceHIDPP.Close()
	}

	if u.Config != nil {
		u.Config.Close()
	}

	if u.Dev != nil {
		u.Dev.SetAutoDetach(false)
		//u.Dev.Reset()
		u.Dev.Close()
	}

	if u.UsbCtx != nil {
		u.UsbCtx.Close()
	}
}


func NewUnifying() (res *Unifying, err error) {
	res = &Unifying{}

	res.UsbCtx = gousb.NewContext()
	res.Dev, err = res.UsbCtx.OpenDeviceWithVIDPID(UNIFYING_VID, UNIFYING_PID)
	if err != nil || res.Dev == nil {
		res.Close()
		return nil, eNoDongle
	}
	fmt.Println("Unifying dongle found", res.Dev)

	//Get device config 1
	res.Config, err = res.Dev.Config(1)
	if err != nil {
		res.Close()
		return nil, errors.New("Couldn't retrieve config 1 of Unifying dongle")
	}

	fmt.Println("Using dongle USB config:", res.Config.Desc.String())

	fmt.Println("Resetting dongle in order to release it from kernel (connected devices won't be usable)")
	//res.Dev.Reset()
	res.Dev.SetAutoDetach(true)

Outer:
	for _, ifaceDesc := range res.Config.Desc.Interfaces {
		for _, ifaceSettings := range ifaceDesc.AltSettings {
			//fmt.Printf("%+v\n", ifaceSettings.Endpoints)
			for _, epDesc := range ifaceSettings.Endpoints {
				if epDesc.MaxPacketSize == 32 && epDesc.Direction == gousb.EndpointDirectionIn {
					// This is the HID++ EP
					//fmt.Printf("EP %+v\n", epDesc.Number)
					res.IfaceHIDPP, err = res.Config.Interface(ifaceSettings.Number, ifaceSettings.Alternate)
					if err != nil {
						res.Close()
						return nil, errors.New("Couldn't access HID++ USB interface")
					} else {
						fmt.Println("HID++ interface:", res.IfaceHIDPP.String())
					}

					res.EpInHidPP, err = res.IfaceHIDPP.InEndpoint(epDesc.Number)
					if err != nil {
						res.Close()
						return nil, errors.New("Couldn't access HID++ USB interface IN endpoint")
					} else {
						fmt.Println("HID++ interface IN endpoint:", res.EpInHidPP.String())
						break Outer
					}
				}
			}
		}
	}

	if res.EpInHidPP == nil {
		res.Close()
		return nil, errors.New("Couldn't find EP for HID++ input reports")
	}

	res.sndQueue = make(chan USBReport)
	res.rcvQueue = make(chan USBReport)

	res.ctx, res.cancel = context.WithCancel(context.Background())

	go res.rcvLoop()
	go res.sndLoop()

	return
}

func (u *Unifying) HIDPP_SendAndCollectResponses(deviceID byte, id HidPPMsgSubID, parameters []byte) (responseReports []USBReport, err error) {
	params := make([]byte, USB_REPORT_TYPE_HIDPP_SHORT_PAYLOAD_LEN)
	reportType := USB_REPORT_TYPE_HIDPP_SHORT

	if len(parameters) > USB_REPORT_TYPE_HIDPP_SHORT_PAYLOAD_LEN {
		params = make([]byte, USB_REPORT_TYPE_HIDPP_LONG_PAYLOAD_LEN)
		reportType = USB_REPORT_TYPE_HIDPP_LONG
	}

	copy(params, parameters)

	hidppReq := &HidPPMsg{
		ReportID:   reportType,
		DeviceID:   deviceID,
		MsgSubID:   id,
		Parameters: params,
	}
	u.SendUSBReport(hidppReq)

	//We collect all response reports (DJ and HID++), till ...
	//  1) we receive the response matching the request
	//  2) we receive an error matching the request
	//
	// We send back an error, if USB response timeout is reached, along with reports collected so far

	for {
		rspUSB, err := u.ReceiveUSBReport(500)
		if err != nil {
			return responseReports,errors.New("USB response timeout")
		} else {
			responseReports = append(responseReports, rspUSB)

			// abort if report aligns to request
			if rspUSB.IsHIDPP() {
				rspHIDpp := rspUSB.(*HidPPMsg)

				// check if response
				// Note: we don't check parameters here, f.e. for a get register command, first param would be the register
				// and should match in the response, but receiving a response for a successive request is unlikely.
				//
				// Checking the MsgSubID against the one of the request works only, because the request IDs and response IDs
				// are the same for currently known ones (f.e. SET_REGISTER_REQ == SET_REGISTER_RSP == 0x80)
				if rspHIDpp.DeviceID == deviceID && rspHIDpp.MsgSubID == id {
					// likely final response, return
					return responseReports,nil
				}

				if rspHIDpp.DeviceID == deviceID && rspHIDpp.MsgSubID == UNIFYING_HIDPP_MSG_ID_ERROR_MSG && rspHIDpp.Parameters[0] == byte(id) {
					// likely final response, return
					return responseReports, errors.New("HID++ error response")
				}
			}
		}
	}

}

func (u *Unifying) EnablePairing(timeOutSeconds byte, devNumber byte, blockTillOff bool) (err error){
	//Enable pairing
	connectDevices := byte(0x01) //open lock
	deviceNumber := devNumber //According to specs: Same value as device index transmitted in 0x41 notification, but we haven't tx'ed anything
	openLockTimeout := timeOutSeconds
	fmt.Printf("Enable pairing for %d seconds\n", openLockTimeout)
	responses,err := u.HIDPP_SendAndCollectResponses(0xff, UNIFYING_HIDPP_MSG_ID_SET_REGISTER_REQ, []byte{byte(UNIFYING_DONGLE_HIDPP_REGISTER_PAIRING), connectDevices, deviceNumber, openLockTimeout})
	for _,r := range responses {
		fmt.Println(r.String())
	}
	fmt.Println("... Enable pairing response (should be enabled)\n")

	if !blockTillOff {
		return
	}

	//Parse successive input reports till new "receiver lock information" with lock closed occurs
	fmt.Println("Printing follow up reports ...")
	for {
		rspUSB, err := u.ReceiveUSBReport(500)
		if err == nil {

			fmt.Println(rspUSB.String())
			if rspUSB.IsHIDPP() {
				hidppRsp := rspUSB.(*HidPPMsg)
				if hidppRsp.MsgSubID == UNIFYING_HIDPP_MSG_ID_RECEIVER_LOCKING_INFORMATION && (hidppRsp.Parameters[0] & 0x01) == 0 {
					switch hidppRsp.Parameters[1] {
					case 0x00:
						return nil //"no error"
					case 0x01:
						return errors.New("timeout")
					case 0x02:
						return errors.New("unsupported device")
					case 0x03:
						return errors.New("too many devices")
					case 0x06:
						return errors.New("connection sequence timeout")
					default:
						return errors.New("pairing aborted with unknown reason")
					}


					fmt.Println("Pairing lock closed")
					return err
				}

				// device connection
				if hidppRsp.MsgSubID == UNIFYING_HIDPP_MSG_ID_DEVICE_CONNECTION {
					devIdx := hidppRsp.DeviceID
					wpid := uint16(hidppRsp.Parameters[3]) << 8 + uint16(hidppRsp.Parameters[2])
					fmt.Printf("DEVICE CONNECTION ON INDEX: %02x TYPE: %s WPID: %#04x\n", devIdx, UnifyingDeviceType(hidppRsp.Parameters[1] & 0x0F), wpid)

					//request additional information
				}
			}

		}
	}

}

func main() {
	u, err := NewUnifying()
	if err != nil {
		panic(err)
	}
	defer u.Close()

	// Request receiver data from pairing information register, p1 = 0x03 (undocumented)
	req := &HidPPMsg{
		ReportID:   USB_REPORT_TYPE_HIDPP_SHORT,
		DeviceID:   0xff,
		MsgSubID:   UNIFYING_HIDPP_MSG_ID_GET_LONG_REGISTER_REQ,
		Parameters: []byte{byte(UNIFYING_DONGLE_HIDPP_REGISTER_PAIRING_INFORMATION), 0x03}, //Param 03, no pairing info, but info on receiver itself
	}
	u.SendUSBReport(req)

	rsp, err := u.ReceiveUSBReport(500)
	//fmt.Printf("%+v\n", rsp)
	if err == nil && rsp.IsHIDPP() {
		hidppRsp := rsp.(*HidPPMsg)

		if hidppRsp.MsgSubID == UNIFYING_HIDPP_MSG_ID_GET_LONG_REGISTER_RSP &&
			hidppRsp.Parameters[0] == byte(UNIFYING_DONGLE_HIDPP_REGISTER_PAIRING_INFORMATION) &&
			hidppRsp.Parameters[1] == 0x03 {
			dongleAddr := hidppRsp.Parameters[2:6]
			unknown1 := hidppRsp.Parameters[6] //Proto ??? 0x04 == Unifying
			unknown2 := hidppRsp.Parameters[7] //max devices ??? 0x06
			unknownRest := hidppRsp.Parameters[8:]
			fmt.Printf("Dongle address: % #x, unknown data1: %#x, unknown data2: %#x, unknown rest: % #x\n", dongleAddr, unknown1, unknown2, unknownRest)
		} else {
			fmt.Println("Wrong HID++ response:", hidppRsp.String())
		}
	} else {
		fmt.Printf("Wrong resp1: %+v\n", rsp)
	}

	// Request dongle firmware version from firmware register
	fmt.Println("!!!Request dongle firmware version from firmware register")
	responses,err := u.HIDPP_SendAndCollectResponses(0xff, UNIFYING_HIDPP_MSG_ID_GET_REGISTER_REQ, []byte{byte(UNIFYING_DONGLE_HIDPP_REGISTER_FIRMWARE), 0x01, 0x00})
	for _,r := range responses {
		fmt.Println(r.String())
		}
	fmt.Println("!!!Request dongle firmware version from firmware register\n")

	// Request dongle firmware build version from firmware register
	fmt.Println("!!!Request dongle firmware build version from firmware register")
	responses,err = u.HIDPP_SendAndCollectResponses(0xff, UNIFYING_HIDPP_MSG_ID_GET_REGISTER_REQ, []byte{byte(UNIFYING_DONGLE_HIDPP_REGISTER_FIRMWARE), 0x02, 0x00})
	for _,r := range responses {
		fmt.Println(r.String())
		}
	fmt.Println("!!!Request dongle firmware build version from firmware register\n")

	// Request dongle bootloader version from firmware register
	fmt.Println("!!!Request dongle bootloader version from firmware register")
	responses,err = u.HIDPP_SendAndCollectResponses(0xff, UNIFYING_HIDPP_MSG_ID_GET_REGISTER_REQ, []byte{byte(UNIFYING_DONGLE_HIDPP_REGISTER_FIRMWARE), 0x04, 0x00})
	for _,r := range responses {
		fmt.Println(r.String())
		}
	fmt.Println("!!!Request dongle bootloader version from firmware register\n")

	//Enable wireless notifications (to be able to receive infos via device connect notify on new devices)
	fmt.Println("!!!Enable wireless notifications")
	responses,err = u.HIDPP_SendAndCollectResponses(0xff, UNIFYING_HIDPP_MSG_ID_SET_REGISTER_REQ, []byte{byte(UNIFYING_DONGLE_HIDPP_REGISTER_WIRELESS_NOTIFICATIONS), 0x00, 0x01})
	for _,r := range responses {
		fmt.Println(r.String())
	}
	fmt.Println("!!!END enable wireless notifications\n")

	fmt.Println("!!!Get connected device info")
	responses,err = u.HIDPP_SendAndCollectResponses(0xff, UNIFYING_HIDPP_MSG_ID_SET_REGISTER_REQ, []byte{byte(UNIFYING_DONGLE_HIDPP_REGISTER_CONNECTION_STATE), 0x02})
	for _,r := range responses {
		fmt.Println(r.String())
	}
	fmt.Println("!!!END get connected device info\n")


	deviceNumber := byte(0x01) //According to specs: Same value as device index transmitted in 0x41 notification, but we haven't tx'ed anything
	openLockTimeout := byte(60)
	pe := u.EnablePairing(openLockTimeout, deviceNumber,true)
	fmt.Printf("Pairing mode exited: %v\n", pe)

	//Parse successive input reports in endless loop
	fmt.Println("!!!!Parse successive input reports in endless loop...")
	for {
		rspUSB, err := u.ReceiveUSBReport(500)
		if err == nil {
			fmt.Println(rspUSB.String())
		}
	}

}
