package unifying

import (
	"errors"
	"fmt"
)

var (
	eNoHidPPMsg      = errors.New("No valid HID++ 1.0 report")
	eNoHidDJReport   = errors.New("No valid DJ report")
	eNoHidPPReportID = errors.New("No valid HID++ 1.0 report type (set to USB_REPORT_TYPE_HIDPP_SHORT or USB_REPORT_TYPE_HIDPP_LONG)")
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
	HIDPP_MSG_ID_DEVICE_DISCONNECTION         HidPPMsgSubID = 0x40
	HIDPP_MSG_ID_DEVICE_CONNECTION            HidPPMsgSubID = 0x41
	HIDPP_MSG_ID_RECEIVER_LOCKING_INFORMATION HidPPMsgSubID = 0x4a

	HIDPP_MSG_ID_SET_REGISTER_REQ      HidPPMsgSubID = 0x80
	HIDPP_MSG_ID_SET_REGISTER_RSP      HidPPMsgSubID = 0x80
	HIDPP_MSG_ID_GET_REGISTER_REQ      HidPPMsgSubID = 0x81
	HIDPP_MSG_ID_GET_REGISTER_RSP      HidPPMsgSubID = 0x81
	HIDPP_MSG_ID_SET_LONG_REGISTER_REQ HidPPMsgSubID = 0x82
	HIDPP_MSG_ID_SET_LONG_REGISTER_RSP HidPPMsgSubID = 0x82
	HIDPP_MSG_ID_GET_LONG_REGISTER_REQ HidPPMsgSubID = 0x83
	HIDPP_MSG_ID_GET_LONG_REGISTER_RSP HidPPMsgSubID = 0x83

	HIDPP_MSG_ID_ERROR_MSG HidPPMsgSubID = 0x8f
)

func (t HidPPMsgSubID) String() string {
	switch t {
	case HIDPP_MSG_ID_DEVICE_DISCONNECTION:
		return "DEVICE DISCONNECTION"
	case HIDPP_MSG_ID_DEVICE_CONNECTION:
		return "DEVICE CONNECTION"
	case HIDPP_MSG_ID_RECEIVER_LOCKING_INFORMATION:
		return "RECEIVER LOCKING INFORMATION"
	case HIDPP_MSG_ID_SET_REGISTER_REQ: //Same for response
		return "SET REGISTER SHORT"
	case HIDPP_MSG_ID_GET_REGISTER_REQ: //Same for response
		return "GET REGISTER SHORT"
	case HIDPP_MSG_ID_SET_LONG_REGISTER_REQ: //Same for response
		return "SET REGISTER LONG"
	case HIDPP_MSG_ID_GET_LONG_REGISTER_REQ: //Same for response
		return "GET REGISTER LONG"
	case HIDPP_MSG_ID_ERROR_MSG: //Same for response
		return "ERROR MESSAGE"

	}
	return fmt.Sprintf("Unknown HID++ SubID %02x", byte(t))
}

type HidPPErrorCode byte

const (
	HIDPP_ERROR_CODE_NO_ERROR         HidPPErrorCode = 0x00
	HIDPP_ERROR_CODE_UNKNOWN          HidPPErrorCode = 0x01
	HIDPP_ERROR_CODE_INVALID_ARGUMENT HidPPErrorCode = 0x02
	HIDPP_ERROR_CODE_OUT_OF_RANGE HidPPErrorCode = 0x03
	HIDPP_ERROR_CODE_HW_ERROR HidPPErrorCode = 0x04
	HIDPP_ERROR_CODE_LOGITECH_INTERNAL HidPPErrorCode = 0x05
	HIDPP_ERROR_CODE_INVALID_FEATURE_INDEX HidPPErrorCode = 0x06
	HIDPP_ERROR_CODE_INVALID_FUNCTION_ID HidPPErrorCode = 0x07
	HIDPP_ERROR_CODE_BUSY HidPPErrorCode = 0x08
	HIDPP_ERROR_CODE_UNSUPPORTED HidPPErrorCode = 0x09
)

func (t HidPPErrorCode) String() string {
	switch t {
	case HIDPP_ERROR_CODE_NO_ERROR:
		return "NO ERROR"
	case HIDPP_ERROR_CODE_UNKNOWN:
		return "UNKNOWN ERROR"
	case HIDPP_ERROR_CODE_INVALID_ARGUMENT:
		return "INVALID ARGUMENT ERROR"
	case HIDPP_ERROR_CODE_OUT_OF_RANGE:
		return "OUT OF RANGE ERROR"
	case HIDPP_ERROR_CODE_HW_ERROR:
		return "HW ERROR"
	case HIDPP_ERROR_CODE_LOGITECH_INTERNAL:
		return "LOGITECH INTERNAL ERROR"
	case HIDPP_ERROR_CODE_INVALID_FEATURE_INDEX:
		return "INVALID FEATURE INDEX ERROR"
	case HIDPP_ERROR_CODE_INVALID_FUNCTION_ID:
		return "INVALID FUNCTION ID ERROR"
	case HIDPP_ERROR_CODE_BUSY:
		return "BUSY ERROR"
	case HIDPP_ERROR_CODE_UNSUPPORTED:
		return "UNSUPPORTED ERROR"
	}
	return fmt.Sprintf("Undocumented error code %02x", byte(t))
}

type DJReportType byte

const (
	//ToDo: check https://github.com/hughsie/fwupd/blob/master/plugins/unifying/fu-unifying-hidpp.h

	DJ_REPORT_TYPE_RF_KEYBOARD          DJReportType = 0x01
	DJ_REPORT_TYPE_RF_MOUSE             DJReportType = 0x02
	DJ_REPORT_TYPE_RF_CONSUMER_CONTROL  DJReportType = 0x03
	DJ_REPORT_TYPE_RF_SYSTEM_CONTROL    DJReportType = 0x04
	DJ_REPORT_TYPE_RF_MSFT_MEDIA_CENTER DJReportType = 0x08
	DJ_REPORT_TYPE_RF_LED               DJReportType = 0x0e

	DJ_REPORT_TYPE_NOTIFICATION_DEVICE_UNPAIRED   DJReportType = 0x40
	DJ_REPORT_TYPE_NOTIFICATION_DEVICE_PAIRED     DJReportType = 0x41
	DJ_REPORT_TYPE_NOTIFICATION_CONNECTION_STATUS DJReportType = 0x42

	DJ_REPORT_TYPE_NOTIFICATION_ERROR DJReportType = 0x7f

	DJ_REPORT_TYPE_CMD_SWITCH_AND_KEEP_ALIVE DJReportType = 0x80
	DJ_REPORT_TYPE_CMD_GET_PAIRED_DEVICES    DJReportType = 0x81
)

func (t DJReportType) String() string {
	switch t {
	case DJ_REPORT_TYPE_RF_KEYBOARD:
		return "RF KEYBOARD"
	case DJ_REPORT_TYPE_RF_MOUSE:
		return "RF MOUSE"
	case DJ_REPORT_TYPE_RF_CONSUMER_CONTROL:
		return "RF CONSUMER CONTROL"
	case DJ_REPORT_TYPE_RF_SYSTEM_CONTROL:
		return "RF SYSTEM CONTROL"
	case DJ_REPORT_TYPE_RF_MSFT_MEDIA_CENTER:
		return "RF MICROSOFT MEDIA CENTER"
	case DJ_REPORT_TYPE_RF_LED:
		return "RF LED"
	case DJ_REPORT_TYPE_NOTIFICATION_DEVICE_UNPAIRED:
		return "NOTIFICATION DEVICE UNPAIRED"
	case DJ_REPORT_TYPE_NOTIFICATION_DEVICE_PAIRED:
		return "NOTIFICATION DEVICE PAIRED"
	case DJ_REPORT_TYPE_NOTIFICATION_CONNECTION_STATUS:
		return "NOTIFICATION CONNECTION STATUS"
	case DJ_REPORT_TYPE_NOTIFICATION_ERROR:
		return "NOTIFICATION ERROR"
	case DJ_REPORT_TYPE_CMD_SWITCH_AND_KEEP_ALIVE:
		return "COMMAND SWITCH AND KEEP ALIVE"
	case DJ_REPORT_TYPE_CMD_GET_PAIRED_DEVICES:
		return "COMMAND GET PAIRED DEVICES"

	}
	return fmt.Sprintf("Unknown DJ Report type %02x", t)
}

const (
	DJ_REPORT_RF_TYPE_BITFIELD_KEYBOARD         = 0x00000002
	DJ_REPORT_RF_TYPE_BITFIELD_MOUSE            = 0x00000004
	DJ_REPORT_RF_TYPE_BITFIELD_CONSUMER_CONTROL = 0x00000008
	DJ_REPORT_RF_TYPE_BITFIELD_POWER_KEYS       = 0x00000010
	DJ_REPORT_RF_TYPE_BITFIELD_MEDIA_CENTER     = 0x00000100
	DJ_REPORT_RF_TYPE_BITFIELD_KEYBOARD_LEDS    = 0x00004000
	DJ_REPORT_RF_TYPE_BITFIELD_HIDPP_SHORT      = 0x00010000 //not on USB as HID report, but in respective RF report
	DJ_REPORT_RF_TYPE_BITFIELD_HIDPP_LONG       = 0x00020000 //not on USB as HID report, but in respective RF report
)

type HidPPRegister byte

const (
	DONGLE_HIDPP_REGISTER_WIRELESS_NOTIFICATIONS HidPPRegister = 0x00
	DONGLE_HIDPP_REGISTER_CONNECTION_STATE       HidPPRegister = 0x02
	DONGLE_HIDPP_REGISTER_PAIRING                HidPPRegister = 0xb2
	DONGLE_HIDPP_REGISTER_DEVICE_ACTIVITY        HidPPRegister = 0xb3
	DONGLE_HIDPP_REGISTER_PAIRING_INFORMATION    HidPPRegister = 0xb5
	DONGLE_HIDPP_REGISTER_FIRMWARE_UPDATE        HidPPRegister = 0xf0
	DONGLE_HIDPP_REGISTER_FIRMWARE_INFO          HidPPRegister = 0xf1
	DONGLE_HIDPP_REGISTER_SECRET_MEMDUMP         HidPPRegister = 0xd4
)

func (t HidPPRegister) String() string {
	switch t {
	case DONGLE_HIDPP_REGISTER_WIRELESS_NOTIFICATIONS:
		return "REGISTER WIRELESS NOTIFICATIONS"
	case DONGLE_HIDPP_REGISTER_CONNECTION_STATE:
		return "REGISTER CONNECTION STATE"
	case DONGLE_HIDPP_REGISTER_PAIRING:
		return "REGISTER PAIRING"
	case DONGLE_HIDPP_REGISTER_DEVICE_ACTIVITY:
		return "REGISTER DEVICE ACTIVITY"
	case DONGLE_HIDPP_REGISTER_PAIRING_INFORMATION:
		return "REGISTER PAIRING INFORMATION"
	case DONGLE_HIDPP_REGISTER_FIRMWARE_INFO:
		return "REGISTER FIRMWARE"
	}
	return fmt.Sprintf("Unknown HID++ Register %02x", byte(t))
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

func (r *DJReport) String() (res string) {
	res = fmt.Sprintf("USB Report type: %s, DeviceID: %#02x, DJ Type: %s, Params: % #x", r.ReportID, r.DeviceID, r.Type, r.Parameters)

	switch r.Type {
	case DJ_REPORT_TYPE_NOTIFICATION_DEVICE_PAIRED:
		specialFunc := r.Parameters[0]
		sfMoreNotif := specialFunc&0x01 > 0
		sfOtherFieldsNotRelevant := specialFunc&0x02 > 0

		wpid := uint16(r.Parameters[2])<<8 + uint16(r.Parameters[1])

		reportTypeBitField := uint32(r.Parameters[3])<<0 + uint32(r.Parameters[4])<<8 + uint32(r.Parameters[5])<<16 + uint32(r.Parameters[6])<<24

		bKeyboard := reportTypeBitField&DJ_REPORT_RF_TYPE_BITFIELD_KEYBOARD > 0
		bMouse := reportTypeBitField&DJ_REPORT_RF_TYPE_BITFIELD_MOUSE > 0
		bConCtl := reportTypeBitField&DJ_REPORT_RF_TYPE_BITFIELD_CONSUMER_CONTROL > 0
		bSysCtl := reportTypeBitField&DJ_REPORT_RF_TYPE_BITFIELD_POWER_KEYS > 0
		bMedCtr := reportTypeBitField&DJ_REPORT_RF_TYPE_BITFIELD_MEDIA_CENTER > 0
		bLED := reportTypeBitField&DJ_REPORT_RF_TYPE_BITFIELD_KEYBOARD_LEDS > 0
		bHIDppShort := reportTypeBitField&DJ_REPORT_RF_TYPE_BITFIELD_HIDPP_SHORT > 0
		bHIDppLong := reportTypeBitField&DJ_REPORT_RF_TYPE_BITFIELD_HIDPP_LONG > 0

		res += fmt.Sprintf("\n\tSpecial func: %#02x (more Notifications: %v other fields not relevant: %v)", specialFunc, sfMoreNotif, sfOtherFieldsNotRelevant)
		res += fmt.Sprintf("\n\twpid: %#04x", wpid)
		res += fmt.Sprintf("\n\treport bitfield: %#08x", reportTypeBitField)
		res += fmt.Sprintf("\n\t\tbitfield keyboard: %v", bKeyboard)
		res += fmt.Sprintf("\n\t\tbitfield mouse: %v", bMouse)
		res += fmt.Sprintf("\n\t\tbitfield consumer control: %v", bConCtl)
		res += fmt.Sprintf("\n\t\tbitfield system control: %v", bSysCtl)
		res += fmt.Sprintf("\n\t\tbitfield media center: %v", bMedCtr)
		res += fmt.Sprintf("\n\t\tbitfield LED: %v", bLED)
		res += fmt.Sprintf("\n\t\tbitfield HID++ short: %v", bHIDppShort)
		res += fmt.Sprintf("\n\t\tbitfield HID++ long: %v", bHIDppLong)

		/*
		case HIDPP_MSG_ID_SET_LONG_REGISTER_RSP:
			fallthrough
		case HIDPP_MSG_ID_GET_LONG_REGISTER_RSP:
			fallthrough
		case HIDPP_MSG_ID_SET_REGISTER_RSP:
			res += fmt.Sprintf("\n\tRegister address: %s", HidPPRegister(r.Parameters[0]))
			res += fmt.Sprintf("\n\tValue: % #x", r.Parameters[1:])
		case HIDPP_MSG_ID_DEVICE_DISCONNECTION:
			res += fmt.Sprintf("\n\tDevice disconnected: %v", r.Parameters[0] == 0x02)
		case HIDPP_MSG_ID_DEVICE_CONNECTION:
			res += fmt.Sprintf("\n\tProtocol type: %#02x", r.Parameters[0])
			res += fmt.Sprintf("\n\tDevice type: %#02x", r.Parameters[1] & 0x0F)
			res += fmt.Sprintf("\n\tSoftware present: %v", r.Parameters[1] & 0x10 > 0)
			res += fmt.Sprintf("\n\tLink encrypted: %v", r.Parameters[1] & 0x20 > 0)
			res += fmt.Sprintf("\n\tLink established: %v", r.Parameters[1] & 0x40 == 0)
			res += fmt.Sprintf("\n\tConnection with payload: %v", r.Parameters[1] & 0x80 > 0)
			res += fmt.Sprintf("\n\tWireless PID: 0x%02x%02x", r.Parameters[3], r.Parameters[2])
		case HIDPP_MSG_ID_RECEIVER_LOCKING_INFORMATION:
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
		case HIDPP_MSG_ID_ERROR_MSG:
			res += fmt.Sprintf("\n\tError notification with parameters: % #x", r.Parameters)
		*/
	}
	return res

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
	case HIDPP_MSG_ID_GET_REGISTER_RSP:
		res += fmt.Sprintf("\n\tRegister address: %s", HidPPRegister(r.Parameters[0]))
		res += fmt.Sprintf("\n\tValue: % #x", r.Parameters[1:])

		switch reg := HidPPRegister(r.Parameters[0]); reg {
		case DONGLE_HIDPP_REGISTER_FIRMWARE_INFO:
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
	case HIDPP_MSG_ID_SET_LONG_REGISTER_RSP:
		fallthrough
	case HIDPP_MSG_ID_GET_LONG_REGISTER_RSP:
		fallthrough
	case HIDPP_MSG_ID_SET_REGISTER_RSP:
		res += fmt.Sprintf("\n\tRegister address: %s", HidPPRegister(r.Parameters[0]))
		res += fmt.Sprintf("\n\tValue: % #x", r.Parameters[1:])
	case HIDPP_MSG_ID_DEVICE_DISCONNECTION:
		res += fmt.Sprintf("\n\tDevice disconnected: %v", r.Parameters[0] == 0x02)
	case HIDPP_MSG_ID_DEVICE_CONNECTION:
		res += fmt.Sprintf("\n\tProtocol type: %#02x", r.Parameters[0])
		res += fmt.Sprintf("\n\tDevice type: %#02x", r.Parameters[1]&0x0F)
		res += fmt.Sprintf("\n\tSoftware present: %v", r.Parameters[1]&0x10 > 0)
		res += fmt.Sprintf("\n\tLink encrypted: %v", r.Parameters[1]&0x20 > 0)
		res += fmt.Sprintf("\n\tLink established: %v", r.Parameters[1]&0x40 == 0)
		res += fmt.Sprintf("\n\tConnection with payload: %v", r.Parameters[1]&0x80 > 0)
		res += fmt.Sprintf("\n\tWireless PID: 0x%02x%02x", r.Parameters[3], r.Parameters[2])
	case HIDPP_MSG_ID_RECEIVER_LOCKING_INFORMATION:
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
	case HIDPP_MSG_ID_ERROR_MSG:
		res += fmt.Sprintf("\n\tError notification with parameters: % #x", r.Parameters)
		res += fmt.Sprintf("\n\t\tparam 0 (HID++ command)  : %#02x", r.Parameters[0])
		res += fmt.Sprintf("\n\t\tparam 1 (likely register): %#02x - '%s'", r.Parameters[1], HidPPRegister(r.Parameters[1]).String())
		res += fmt.Sprintf("\n\t\tparam 2 (error)          : %#02x - '%s'", r.Parameters[2], HidPPErrorCode(r.Parameters[2]).String())

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
