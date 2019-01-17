package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/gousb"
	"os"
	"os/signal"
	"time"
)

var (
	eNoDongle        = errors.New("No Unifying dongle found")
	eNoHidPPMsg      = errors.New("No valid HID++ 1.0 report")
	eNoHidDJReport   = errors.New("No valid DJ report")
	eNoHidPPReportID = errors.New("No valid HID++ 1.0 report type (set to USB_REPORT_TYPE_HIDPP_SHORT or USB_REPORT_TYPE_HIDPP_LONG)")
)

const (
	UNIFYING_VID gousb.ID = 0x046d
	UNIFYING_PID gousb.ID = 0xc52b
)

const (
	USB_REPORT_TYPE_DJ_SHORT_LEN         = 15
	USB_REPORT_TYPE_DJ_LONG_LEN          = 32
	USB_REPORT_TYPE_DJ_SHORT_PAYLOAD_LEN = 12
	USB_REPORT_TYPE_DJ_LONG_PAYLOAD_LEN  = 29

	USB_REPORT_TYPE_HIDPP_SHORT_LEN        = 7
	USB_REPORT_TYPE_HIDPP_LONG_LEN         = 20
	USB_REPORT_TYPE_HIDPP_PAYLOAD_LEN      = 4
	USB_REPORT_TYPE_HIDPP_LONG_PAYLOAD_LEN = 17
)

type USBReportType byte

const (
	USB_REPORT_TYPE_DJ_SHORT    USBReportType = 0x20
	USB_REPORT_TYPE_DJ_LONG     USBReportType = 0x21
	USB_REPORT_TYPE_HIDPP_SHORT USBReportType = 0x10
	USB_REPORT_TYPE_HIDPP_LONG  USBReportType = 0x11
)

func (t USBReportType) String() string {
	switch t {
	case USB_REPORT_TYPE_HIDPP_SHORT:
		return "HID++ short message"
	case USB_REPORT_TYPE_HIDPP_LONG:
		return "HID++ long message"
	case USB_REPORT_TYPE_DJ_SHORT:
		return "DJ Report short"
	case USB_REPORT_TYPE_DJ_LONG:
		return "DJ Report long"
	}
	return fmt.Sprintf("Unknown USB report type %02x", t)
}

type USBReport interface {
	FromWire(payload []byte) (err error)
	ToWire() (payload []byte, err error)
	IsHIDPP() bool
	IsDJ() bool

	fmt.Stringer
	//String(9 should be present to, but we don't force Stringer interface
}

type HidPPMsgSubID byte

const (
	UNIFYING_HIDPP_MSG_ID_DEVICE_DISCONNECTION         HidPPMsgSubID = 0x40
	UNIFYING_HIDPP_MSG_ID_DEVICE_CONNECTION            HidPPMsgSubID = 0x41
	UNIFYING_HIDPP_MSG_ID_RECEIVER_LOCKING_INFORMATION HidPPMsgSubID = 0x4a

	UNIFYING_HIDPP_MSG_ID_SET_REGISTER_REQ      HidPPMsgSubID = 0x80
	UNIFYING_HIDPP_MSG_ID_SET_REGISTER_RSP      HidPPMsgSubID = 0x80
	UNIFYING_HIDPP_MSG_ID_GET_REGISTER_REQ      HidPPMsgSubID = 0x81
	UNIFYING_HIDPP_MSG_ID_GET_REGISTER_RSP      HidPPMsgSubID = 0x81
	UNIFYING_HIDPP_MSG_ID_SET_LONG_REGISTER_REQ HidPPMsgSubID = 0x82
	UNIFYING_HIDPP_MSG_ID_SET_LONG_REGISTER_RSP HidPPMsgSubID = 0x82
	UNIFYING_HIDPP_MSG_ID_GET_LONG_REGISTER_REQ HidPPMsgSubID = 0x83
	UNIFYING_HIDPP_MSG_ID_GET_LONG_REGISTER_RSP HidPPMsgSubID = 0x83

	UNIFYING_HIDPP_MSG_ID_ERROR_MSG HidPPMsgSubID = 0x8f
)

func (t HidPPMsgSubID) String() string {
	switch t {
	case UNIFYING_HIDPP_MSG_ID_DEVICE_DISCONNECTION:
		return "DEVICE DISCONNECTION"
	case UNIFYING_HIDPP_MSG_ID_DEVICE_CONNECTION:
		return "DEVICE CONNECTION"
	case UNIFYING_HIDPP_MSG_ID_RECEIVER_LOCKING_INFORMATION:
		return "RECEIVER LOCKING INFORMATION"
	case UNIFYING_HIDPP_MSG_ID_SET_REGISTER_REQ: //Same for response
		return "SET REGISTER SHORT"
	case UNIFYING_HIDPP_MSG_ID_GET_REGISTER_REQ: //Same for response
		return "GET REGISTER SHORT"
	case UNIFYING_HIDPP_MSG_ID_SET_LONG_REGISTER_REQ: //Same for response
		return "SET REGISTER LONG"
	case UNIFYING_HIDPP_MSG_ID_GET_LONG_REGISTER_REQ: //Same for response
		return "GET REGISTER LONG"

	}
	return fmt.Sprintf("Unknown HID++ SubID %02x", t)
}

type DJReportType byte

const (
	UNIFYING_DJ_REPORT_TYPE_RF_KEYBOARD          DJReportType = 0x01
	UNIFYING_DJ_REPORT_TYPE_RF_MOUSE             DJReportType = 0x02
	UNIFYING_DJ_REPORT_TYPE_RF_CONSUMER_CONTROL  DJReportType = 0x03
	UNIFYING_DJ_REPORT_TYPE_RF_SYSTEM_CONTROL    DJReportType = 0x04
	UNIFYING_DJ_REPORT_TYPE_RF_MSFT_MEDIA_CENTER DJReportType = 0x08
	UNIFYING_DJ_REPORT_TYPE_RF_LED               DJReportType = 0x0e

	UNIFYING_DJ_REPORT_TYPE_NOTIFICATION_DEVICE_UNPAIRED   DJReportType = 0x40
	UNIFYING_DJ_REPORT_TYPE_NOTIFICATION_DEVICE_PAIRED     DJReportType = 0x41
	UNIFYING_DJ_REPORT_TYPE_NOTIFICATION_CONNECTION_STATUS DJReportType = 0x42

	UNIFYING_DJ_REPORT_TYPE_NOTIFICATION_ERROR DJReportType = 0x7f

	UNIFYING_DJ_REPORT_TYPE_CMD_SWITCH_AND_KEEP_ALIVE DJReportType = 0x80
	UNIFYING_DJ_REPORT_TYPE_CMD_GET_PAIRED_DEVICES    DJReportType = 0x81
)

func (t DJReportType) String() string {
	switch t {
	case UNIFYING_DJ_REPORT_TYPE_RF_KEYBOARD:
		return "RF KEYBOARD"
	case UNIFYING_DJ_REPORT_TYPE_RF_MOUSE:
		return "RF MOUSE"
	case UNIFYING_DJ_REPORT_TYPE_RF_CONSUMER_CONTROL:
		return "RF CONSUMER CONTROL"
	case UNIFYING_DJ_REPORT_TYPE_RF_SYSTEM_CONTROL:
		return "RF SYSTEM CONTROL"
	case UNIFYING_DJ_REPORT_TYPE_RF_MSFT_MEDIA_CENTER:
		return "RF MICROSOFT MEDIA CENTER"
	case UNIFYING_DJ_REPORT_TYPE_RF_LED:
		return "RF LED"
	case UNIFYING_DJ_REPORT_TYPE_NOTIFICATION_DEVICE_UNPAIRED:
		return "NOTIFICATION DEVICE UNPAIRED"
	case UNIFYING_DJ_REPORT_TYPE_NOTIFICATION_DEVICE_PAIRED:
		return "NOTIFICATION DEVICE PAIRED"
	case UNIFYING_DJ_REPORT_TYPE_NOTIFICATION_CONNECTION_STATUS:
		return "NOTIFICATION CONNECTION STATUS"
	case UNIFYING_DJ_REPORT_TYPE_NOTIFICATION_ERROR:
		return "NOTIFICATION ERROR"
	case UNIFYING_DJ_REPORT_TYPE_CMD_SWITCH_AND_KEEP_ALIVE:
		return "COMMAND SWITCH AND KEEP ALIVE"
	case UNIFYING_DJ_REPORT_TYPE_CMD_GET_PAIRED_DEVICES:
		return "COMMAND GET PAIRED DEVICES"

	}
	return fmt.Sprintf("Unknown DJ Report type %02x", t)
}

const (
	UNIFYING_DJ_REPORT_RF_TYPE_BITFIELD_KEYBOARD         = 0x00000002
	UNIFYING_DJ_REPORT_RF_TYPE_BITFIELD_MOUSE            = 0x00000004
	UNIFYING_DJ_REPORT_RF_TYPE_BITFIELD_CONSUMER_CONTROL = 0x00000008
	UNIFYING_DJ_REPORT_RF_TYPE_BITFIELD_POWER_KEYS       = 0x00000010
	UNIFYING_DJ_REPORT_RF_TYPE_BITFIELD_MEDIA_CENTER     = 0x00000100
	UNIFYING_DJ_REPORT_RF_TYPE_BITFIELD_KEYBOARD_LEDS    = 0x00004000
	UNIFYING_DJ_REPORT_RF_TYPE_BITFIELD_HIDPP_SHORT      = 0x00010000 //not on USB as HID report, but in respective RF report
	UNIFYING_DJ_REPORT_RF_TYPE_BITFIELD_HIDPP_LONG       = 0x00020000 //not on USB as HID report, but in respective RF report
)

type HidPPRegister byte
const (
	UNIFYING_DONGLE_HIDPP_REGISTER_WIRELESS_NOTIFICATIONS HidPPRegister = 0x00
	UNIFYING_DONGLE_HIDPP_REGISTER_CONNECTION_STATE       HidPPRegister = 0x02
	UNIFYING_DONGLE_HIDPP_REGISTER_PAIRING                HidPPRegister = 0xb2
	UNIFYING_DONGLE_HIDPP_REGISTER_DEVICE_ACTIVITY        HidPPRegister = 0xb3
	UNIFYING_DONGLE_HIDPP_REGISTER_PAIRING_INFORMATION    HidPPRegister = 0xb5
	UNIFYING_DONGLE_HIDPP_REGISTER_FIRMWARE               HidPPRegister = 0xf1
)
func (t HidPPRegister) String() string {
	switch t {
	case UNIFYING_DONGLE_HIDPP_REGISTER_WIRELESS_NOTIFICATIONS:
		return "REGISTER WIRELESS NOTIFICATIONS"
	case UNIFYING_DONGLE_HIDPP_REGISTER_CONNECTION_STATE:
		return "REGISTER CONNECTION STATE"
	case UNIFYING_DONGLE_HIDPP_REGISTER_PAIRING:
		return "REGISTER PAIRING"
	case UNIFYING_DONGLE_HIDPP_REGISTER_DEVICE_ACTIVITY:
		return "REGISTER DEVICE ACTIVITY"
	case UNIFYING_DONGLE_HIDPP_REGISTER_PAIRING_INFORMATION:
		return "REGISTER PAIRING INFORMATION"
	case UNIFYING_DONGLE_HIDPP_REGISTER_FIRMWARE:
		return "REGISTER FIRMWARE"
	}
	return fmt.Sprintf("Unknown HID++ Register %02x", t)
}


// for write/read parameters of short read/write from/to wireless notification register (0x00)
const (
	UNIYING_WIRELESS_NOTIFICATIONS_P0_BATTERY_STATUS_MASK = (1 << 4)

	UNIYING_WIRELESS_NOTIFICATIONS_P1_WIRELESS_NOTIFICATIONS_MASK = (1 << 0)
	UNIYING_WIRELESS_NOTIFICATIONS_P1_SOFTWARE_PRESENT_MASK       = (1 << 3)
)

type DJReport struct {
	ReportID   USBReportType
	DeviceID   byte
	Type       DJReportType
	Parameters []byte
}

func (r *DJReport) IsHIDPP() bool {
	return r.ReportID == USB_REPORT_TYPE_HIDPP_LONG || r.ReportID == USB_REPORT_TYPE_HIDPP_SHORT
}

func (r *DJReport) IsDJ() bool {
	return r.ReportID == USB_REPORT_TYPE_DJ_LONG || r.ReportID == USB_REPORT_TYPE_DJ_SHORT
}

func (r *DJReport) String() string {
	return fmt.Sprintf("USB Report type: %s, DeviceID: %#02x, DJ Type: %s, Params: % #x", r.ReportID, r.DeviceID, r.Type, r.Parameters)
}

func (r *DJReport) IsRFReport() bool {
	return r.Type < 0x40
}

func (r *DJReport) IsNotification() bool {
	return r.Type > 0x3f && r.Type < 0x80
}

func (r *DJReport) IsCommand() bool {
	return r.Type > 0x7f
}

func (r *DJReport) FromWire(payload []byte) (err error) {
	if len(payload) == USB_REPORT_TYPE_DJ_LONG_LEN && payload[0] == byte(USB_REPORT_TYPE_DJ_LONG) {
		r.ReportID = USB_REPORT_TYPE_DJ_LONG
		r.DeviceID = payload[1]
		r.Type = DJReportType(payload[2])
		r.Parameters = make([]byte, USB_REPORT_TYPE_DJ_LONG_PAYLOAD_LEN)
		copy(r.Parameters, payload[3:])
		return
	}
	if len(payload) == USB_REPORT_TYPE_DJ_SHORT_LEN && payload[0] == byte(USB_REPORT_TYPE_DJ_SHORT) {
		r.ReportID = USB_REPORT_TYPE_DJ_SHORT
		r.DeviceID = payload[1]
		r.Type = DJReportType(payload[2])
		r.Parameters = make([]byte, USB_REPORT_TYPE_DJ_SHORT_PAYLOAD_LEN)
		copy(r.Parameters, payload[3:])
		return
	}

	return eNoHidDJReport
}

func (r *DJReport) ToWire() (payload []byte, err error) {
	if r.ReportID == USB_REPORT_TYPE_DJ_SHORT {
		payload := make([]byte, USB_REPORT_TYPE_DJ_SHORT_LEN)
		payload[0] = byte(r.ReportID)
		payload[1] = r.DeviceID
		payload[2] = byte(r.Type)
		if r.Parameters != nil {
			copy(payload[3:], r.Parameters)
		}
		return payload, nil
	}

	if r.ReportID == USB_REPORT_TYPE_DJ_LONG {
		payload := make([]byte, USB_REPORT_TYPE_DJ_LONG_LEN)
		payload[0] = byte(r.ReportID)
		payload[1] = r.DeviceID
		payload[2] = byte(r.Type)
		if r.Parameters != nil {
			copy(payload[3:], r.Parameters)
		}
		return payload, nil
	}

	err = eNoHidDJReport
	return
}

type HidPPMsg struct {
	ReportID   USBReportType
	DeviceID   byte
	MsgSubID   HidPPMsgSubID
	Parameters []byte
}

func (r *HidPPMsg) String() (res string) {
	res = fmt.Sprintf("USB Report type: %s, DeviceID: %#02x, SubID: %s, Params: % #x", r.ReportID, r.DeviceID, r.MsgSubID, r.Parameters)
	switch r.MsgSubID {
	case UNIFYING_HIDPP_MSG_ID_GET_REGISTER_RSP:
		fallthrough
	case UNIFYING_HIDPP_MSG_ID_SET_LONG_REGISTER_RSP:
		fallthrough
	case UNIFYING_HIDPP_MSG_ID_GET_LONG_REGISTER_RSP:
		fallthrough
	case UNIFYING_HIDPP_MSG_ID_SET_REGISTER_RSP:
		res += fmt.Sprintf("\n\tRegister address: %s", HidPPRegister(r.Parameters[0]))
		res += fmt.Sprintf("\n\tValue: % #x", r.Parameters[1:])
	case UNIFYING_HIDPP_MSG_ID_DEVICE_DISCONNECTION:
		res += fmt.Sprintf("\n\tDevice disconnected: %v", r.Parameters[0] == 0x02)
	case UNIFYING_HIDPP_MSG_ID_DEVICE_CONNECTION:
		res += fmt.Sprintf("\n\tProtocol type: %#02x", r.Parameters[0])
		res += fmt.Sprintf("\n\tDevice type: %#02x", r.Parameters[1] & 0x0F)
		res += fmt.Sprintf("\n\tSoftware present: %v", r.Parameters[1] & 0x10 > 0)
		res += fmt.Sprintf("\n\tLink encrypted: %v", r.Parameters[1] & 0x20 > 0)
		res += fmt.Sprintf("\n\tLink established: %v", r.Parameters[1] & 0x40 == 0)
		res += fmt.Sprintf("\n\tConnection with payload: %v", r.Parameters[1] & 0x80 > 0)
		res += fmt.Sprintf("\n\tWireless PID: 0x%02x%02x", r.Parameters[3], r.Parameters[2])

/*
	case UNIFYING_HIDPP_MSG_ID_RECEIVER_LOCKING_INFORMATION:
		fmt.Printf("Receiver locking information notification with parameters: % #x\n", rsp.Parameters)
	case UNIFYING_HIDPP_MSG_ID_ERROR_MSG:
		fmt.Printf("Receiver error message notification with parameters: % #x\n", rsp.Parameters)
*/
	}
	return res
}

func (r *HidPPMsg) FromWire(payload []byte) (err error) {
	if len(payload) == USB_REPORT_TYPE_HIDPP_LONG_LEN && payload[0] == byte(USB_REPORT_TYPE_HIDPP_LONG) {
		r.ReportID = USB_REPORT_TYPE_HIDPP_LONG
		r.DeviceID = payload[1]
		r.MsgSubID = HidPPMsgSubID(payload[2])
		r.Parameters = make([]byte, USB_REPORT_TYPE_HIDPP_LONG_PAYLOAD_LEN)
		copy(r.Parameters, payload[3:])
		return
	}
	if len(payload) == USB_REPORT_TYPE_HIDPP_SHORT_LEN && payload[0] == byte(USB_REPORT_TYPE_HIDPP_SHORT) {
		r.ReportID = USB_REPORT_TYPE_HIDPP_SHORT
		r.DeviceID = payload[1]
		r.MsgSubID = HidPPMsgSubID(payload[2])
		r.Parameters = make([]byte, USB_REPORT_TYPE_HIDPP_PAYLOAD_LEN)
		copy(r.Parameters, payload[3:])
		return
	}

	return eNoHidPPMsg
}

func (r *HidPPMsg) ToWire() (payload []byte, err error) {
	if r.ReportID == USB_REPORT_TYPE_HIDPP_SHORT {
		payload := make([]byte, USB_REPORT_TYPE_HIDPP_SHORT_LEN)
		payload[0] = byte(r.ReportID)
		payload[1] = r.DeviceID
		payload[2] = byte(r.MsgSubID)
		if r.Parameters != nil {
			copy(payload[3:], r.Parameters)
		}
		return payload, nil
	}

	if r.ReportID == USB_REPORT_TYPE_HIDPP_LONG {
		payload := make([]byte, USB_REPORT_TYPE_HIDPP_LONG_LEN)
		payload[0] = byte(r.ReportID)
		payload[1] = r.DeviceID
		payload[2] = byte(r.MsgSubID)
		if r.Parameters != nil {
			copy(payload[3:], r.Parameters)
		}
		return payload, nil
	}

	err = eNoHidPPReportID
	return
}

func (r *HidPPMsg) IsHIDPP() bool {
	return r.ReportID == USB_REPORT_TYPE_HIDPP_LONG || r.ReportID == USB_REPORT_TYPE_HIDPP_SHORT
}

func (r *HidPPMsg) IsDJ() bool {
	return r.ReportID == USB_REPORT_TYPE_DJ_LONG || r.ReportID == USB_REPORT_TYPE_DJ_SHORT
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

func (u *Unifying) readLoop() {
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

		/*
					if pay[0] == 0x20 || pay[0] == 0x21 {

						fmt.Println("DJ message")

						if pay[0] == 0x20 {
							fmt.Printf("\tDJ short message: % #x\n", pay)
						} else if pay[0] == 0x21 {
							fmt.Printf("\tDJ long message: % #x\n", pay)
						}

						fmt.Printf("\tDJ message device index: %#02x\n", pay[1])

						if pay[2] < 0x40 {
							fmt.Println("\tRF report")
						} else if pay[1] < 0x80 {
							fmt.Println("\tDJ notification")
						} else {
							fmt.Println("\tDJ command")
						}
					}
				}
		*/
	}

	close(u.rcvQueue)
}

func (u *Unifying) sndRcvLoop() {
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
			/*
			case indata := <- u.rcvQueue:
				fmt.Printf("In: % #x\n", indata)
			*/
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

	go res.readLoop()
	go res.sndRcvLoop()

	return
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

	// Request major/minor dongle level of firmware from firmware register
	req = &HidPPMsg{
		ReportID:   USB_REPORT_TYPE_HIDPP_SHORT,
		DeviceID:   0xff,
		MsgSubID:   UNIFYING_HIDPP_MSG_ID_GET_REGISTER_REQ,
		Parameters: []byte{byte(UNIFYING_DONGLE_HIDPP_REGISTER_FIRMWARE), 0x01}, //Param 03, no pairing info, but info on receiver itself
	}
	u.SendUSBReport(req)

	rsp, err = u.ReceiveUSBReport(500)
	if err == nil && rsp.IsHIDPP() {
		hidppRsp := rsp.(*HidPPMsg)
		if hidppRsp.MsgSubID == UNIFYING_HIDPP_MSG_ID_GET_REGISTER_RSP &&
			hidppRsp.Parameters[0] == byte(UNIFYING_DONGLE_HIDPP_REGISTER_FIRMWARE) &&
			hidppRsp.Parameters[1] == 0x01 {
			major := hidppRsp.Parameters[2]
			minor := hidppRsp.Parameters[3]
			fmt.Printf("Firmware maj.min: %.2x.%.2x\n", major, minor)
		} else {
			fmt.Println("Wrong HID++ response:", hidppRsp.String())
		}
	} else {
		fmt.Printf("Wrong resp2: %+v\n", rsp)
	}

	// Request dongle firmware patch version from firmware register
	req = &HidPPMsg{
		ReportID:   USB_REPORT_TYPE_HIDPP_SHORT,
		DeviceID:   0xff,
		MsgSubID:   UNIFYING_HIDPP_MSG_ID_GET_REGISTER_REQ,
		Parameters: []byte{byte(UNIFYING_DONGLE_HIDPP_REGISTER_FIRMWARE), 0x02}, //Param 03, no pairing info, but info on receiver itself
	}
	u.SendUSBReport(req)

	rsp, err = u.ReceiveUSBReport(500)
	if err == nil && rsp.IsHIDPP() {
		hidppRsp := rsp.(*HidPPMsg)
		if hidppRsp.MsgSubID == UNIFYING_HIDPP_MSG_ID_GET_REGISTER_RSP &&
			hidppRsp.Parameters[0] == byte(UNIFYING_DONGLE_HIDPP_REGISTER_FIRMWARE) &&
			hidppRsp.Parameters[1] == 0x02 {
			patchMSB := hidppRsp.Parameters[2]
			patchLSB := hidppRsp.Parameters[3]
			fmt.Printf("Firmware patch: %.2x%.2x\n", patchMSB, patchLSB)
		} else {
			fmt.Println("Wrong HID++ response:", hidppRsp.String())
		}
	} else {
		fmt.Printf("Wrong resp3: %+v\n", rsp)
	}

	// Request dongle bootloader patch version from firmware register
	req = &HidPPMsg{
		ReportID:   USB_REPORT_TYPE_HIDPP_SHORT,
		DeviceID:   0xff,
		MsgSubID:   UNIFYING_HIDPP_MSG_ID_GET_REGISTER_REQ,
		Parameters: []byte{byte(UNIFYING_DONGLE_HIDPP_REGISTER_FIRMWARE), 0x04}, //Param 03, no pairing info, but info on receiver itself
	}
	u.SendUSBReport(req)

	rsp, err = u.ReceiveUSBReport(500)
	if err == nil && rsp.IsHIDPP() {
		hidppRsp := rsp.(*HidPPMsg)
		if hidppRsp.MsgSubID == UNIFYING_HIDPP_MSG_ID_GET_REGISTER_RSP &&
			hidppRsp.Parameters[0] == byte(UNIFYING_DONGLE_HIDPP_REGISTER_FIRMWARE) &&
			hidppRsp.Parameters[1] == 0x04 {
			major := hidppRsp.Parameters[2]
			minor := hidppRsp.Parameters[3]
			fmt.Printf("Bootloader maj.min: %.2x.%.2x\n", major, minor)
		} else {
			fmt.Println("Wrong HID++ response:", hidppRsp.String())
		}
	} else {

		fmt.Printf("Wrong resp4: %+v\n", rsp)
	}

	//Enable wireless notifications (to be able to receive infos via device connect notify on new devices)
	req = &HidPPMsg{
		ReportID:   USB_REPORT_TYPE_HIDPP_SHORT,
		DeviceID:   0xff,
		MsgSubID:   UNIFYING_HIDPP_MSG_ID_SET_REGISTER_REQ,
		Parameters: []byte{byte(UNIFYING_DONGLE_HIDPP_REGISTER_WIRELESS_NOTIFICATIONS), 0x00, 0x01, 0x00},
	}
	u.SendUSBReport(req)

	rsp, err = u.ReceiveUSBReport(500)
	rsp, err = u.ReceiveUSBReport(500)
	if err == nil && rsp.IsHIDPP() {
		hidppRsp := rsp.(*HidPPMsg)
		if hidppRsp.MsgSubID == UNIFYING_HIDPP_MSG_ID_SET_REGISTER_RSP &&
			hidppRsp.Parameters[0] == byte(UNIFYING_DONGLE_HIDPP_REGISTER_WIRELESS_NOTIFICATIONS) &&
			hidppRsp.Parameters[1] == 0x00 {

			fmt.Printf("Wireless notifications enabled\n")
		}
	} else {
		fmt.Printf("Wrong resp4: %+v\n", rsp)
	}

	//Get connected device' info
	fmt.Println("Get connected device info")
	req = &HidPPMsg{
		ReportID:   USB_REPORT_TYPE_HIDPP_SHORT,
		DeviceID:   0xff,
		MsgSubID:   UNIFYING_HIDPP_MSG_ID_SET_REGISTER_REQ,
		Parameters: []byte{byte(UNIFYING_DONGLE_HIDPP_REGISTER_CONNECTION_STATE), 0x02, 0x00, 0x00},
	}
	u.SendUSBReport(req)

GetDevInfo:
	for {
		rsp, err = u.ReceiveUSBReport(500)

		if err == nil && rsp.IsDJ() {
			djRsp := rsp.(*DJReport)
			djRsp.String()

		} else if err == nil && rsp.IsHIDPP() {
			hidppRsp := rsp.(*HidPPMsg)
			switch hidppRsp.MsgSubID {
			case UNIFYING_HIDPP_MSG_ID_SET_REGISTER_RSP:
				if hidppRsp.Parameters[0] == byte(UNIFYING_DONGLE_HIDPP_REGISTER_CONNECTION_STATE) && hidppRsp.Parameters[1] == 0x00 {
					fmt.Println("Get dev info done")
					break GetDevInfo
				} else {
					fmt.Printf("unexpected SET_REGISTER_RESPONSE: %+v\n", hidppRsp)
				}
			case UNIFYING_HIDPP_MSG_ID_DEVICE_CONNECTION:
				fmt.Printf("Device connected notification with parameters: % #x\n", hidppRsp.Parameters)
			case UNIFYING_HIDPP_MSG_ID_DEVICE_DISCONNECTION:
				fmt.Printf("Device disconnected notification with parameters: % #x\n", hidppRsp.Parameters)
			case UNIFYING_HIDPP_MSG_ID_RECEIVER_LOCKING_INFORMATION:
				fmt.Printf("Receiver locking information notification with parameters: % #x\n", hidppRsp.Parameters)
			case UNIFYING_HIDPP_MSG_ID_ERROR_MSG:
				fmt.Printf("Receiver error message notification with parameters: % #x\n", hidppRsp.Parameters)
			}
		} else {
			fmt.Printf("Unknown USB Report type: % #x\n", rsp)
		}

		if err == nil {

		}
	}
	/*
		> 10ff80b201123c // enable pairing (p0: 0x01 - Open Lock, p1: 0x12 - Device Number ??, p2: 0x3c == 60 sec) //timeout of 0 would be default of 30sec
		< 10ff4a01000000 //Notif Lock open (pairing on)
		< 10ff80b2000000 //SetRegResp for enable pairing
	*/

	//Enable pairing
	pairingTimeout := byte(60)
	fmt.Printf("Enable pairing for %d seconds\n", pairingTimeout)
	req = &HidPPMsg{
		ReportID:   USB_REPORT_TYPE_HIDPP_SHORT,
		DeviceID:   0xff,
		MsgSubID:   UNIFYING_HIDPP_MSG_ID_SET_REGISTER_REQ,
		Parameters: []byte{byte(UNIFYING_DONGLE_HIDPP_REGISTER_PAIRING), 0x01, 0x04, pairingTimeout}, //param1: lock open, param2: device index (the unused one which will be announced ??), param3: timeout
	}
	u.SendUSBReport(req)

	//PairingLoop:
	for {
		rspUSB, err := u.ReceiveUSBReport(500)
		if err == nil {
			fmt.Println(rspUSB.String())
		}
	}

	go func() {
		for {
			fmt.Println(u.ReceiveUSBReport(0))
		}

	}()

	//Catch signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for sig := range c {
		// sig is a ^C, handle it
		fmt.Println("Received signal", sig)
		u.Close()
		os.Exit(0)
		return
	}

}
