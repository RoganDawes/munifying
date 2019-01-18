package main

import (
	"errors"
	"fmt"
	"github.com/google/gousb"
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

	USB_REPORT_TYPE_HIDPP_SHORT_LEN         = 7
	USB_REPORT_TYPE_HIDPP_LONG_LEN          = 20
	USB_REPORT_TYPE_HIDPP_SHORT_PAYLOAD_LEN = 4
	USB_REPORT_TYPE_HIDPP_LONG_PAYLOAD_LEN  = 17
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
	return fmt.Sprintf("Unknown HID++ SubID %02x", byte(t))
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
		res += fmt.Sprintf("\n\tRegister address: %s", HidPPRegister(r.Parameters[0]))
		res += fmt.Sprintf("\n\tValue: % #x", r.Parameters[1:])

		switch reg := HidPPRegister(r.Parameters[0]); reg {
		case UNIFYING_DONGLE_HIDPP_REGISTER_FIRMWARE:
			res += fmt.Sprintf("\n\tRequested register: %s", reg)
			switch r.Parameters[1] {
			case 0x01:
				res += fmt.Sprintf("\n\tFirmware version: %.2x.%.2x", r.Parameters[2], r.Parameters[3])
			case 0x02:
				res += fmt.Sprintf("\n\tFirmware build version: %#02x%02x", r.Parameters[2], r.Parameters[3])
			case 0x04:
				res += fmt.Sprintf("\n\tBootloader version: %.2x.%.2x", r.Parameters[2], r.Parameters[3])
			}
		}
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
	case UNIFYING_HIDPP_MSG_ID_RECEIVER_LOCKING_INFORMATION:
		res += fmt.Sprintf("\n\tLock open: %v", r.Parameters[0] == 0x01)
		lock_err := "undefined / reserved"
		switch r.Parameters[1] {
		case 0x00:
			lock_err = "no error"
		case 0x01:
			lock_err = "timeout"
		case 0x02:
			lock_err = "unsupported device"
		case 0x03:
			lock_err = "too many devices"
		case 0x06:
			lock_err = "connection sequence timeout"
		}
		res += fmt.Sprintf("\n\tLock error: %s", lock_err)
	case UNIFYING_HIDPP_MSG_ID_ERROR_MSG:
		res += fmt.Sprintf("\n\tError notification with parameters: % #x", r.Parameters)
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
		r.Parameters = make([]byte, USB_REPORT_TYPE_HIDPP_SHORT_PAYLOAD_LEN)
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

