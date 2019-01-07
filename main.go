package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/gousb"
	"os"
	"os/signal"
)

var (
	eNoDongle = errors.New("No Unifying dongle found")
	eNoHidPPMsg = errors.New("No valid HID++ 1.0 report")
	eNoHidPPReportID = errors.New("No valid HID++ 1.0 report type (set to HIDPP_TYPE_SHORT or HIDPP_TYPE_LONG)")
)

const (
	UNIFYING_VID gousb.ID = 0x046d
	UNIFYING_PID gousb.ID = 0xc52b
)

type HidPPReportType byte
const (
	HIDPP_TYPE_SHORT HidPPReportType = 0x10
	HIDPP_TYPE_LONG HidPPReportType = 0x11
)
const (
	HIDPP_TYPE_SHORT_LEN = 7
	HIDPP_TYPE_LONG_LEN = 20
	HIDPP_TYPE_SHORT_PAYLOAD_LEN = 3
	HIDPP_TYPE_LONG_PAYLOAD_LEN = 16
)

const (
	UNIFYING_MSG_ID_SET_REGISTER_REQ = 0x80
	UNIFYING_MSG_ID_SET_REGISTER_RSP = 0x80
	UNIFYING_MSG_ID_GET_REGISTER_REQ = 0x81
	UNIFYING_MSG_ID_GET_REGISTER_RSP = 0x81
	UNIFYING_MSG_ID_SET_LONG_REGISTER_REQ = 0x82
	UNIFYING_MSG_ID_SET_LONG_REGISTER_RSP = 0x82
	UNIFYING_MSG_ID_GET_LONG_REGISTER_REQ = 0x83
	UNIFYING_MSG_ID_GET_LONG_REGISTER_RSP = 0x83

	UNIFYING_MSG_ID_ERROR_MSG = 0x8f
)

const (
	UNIFYING_REGISTER_WIRELESS_NOTIFICATIONS = 0x00
	UNIFYING_REGISTER_CONNECTION_STATE = 0x02
	UNIFYING_REGISTER_PAIRING = 0xb2
	UNIFYING_REGISTER_DEVICE_ACTIVITY = 0xb3
	UNIFYING_REGISTER_PAIRING_INFORMATION = 0xb5
)

// for write/read parameters of short read/write from/to wireless notification register (0x00)
const (
	UNIYING_WIRELESS_NOTIFICATIONS_P0_BATTERY_STATUS_MASK = (1 << 4)

	UNIYING_WIRELESS_NOTIFICATIONS_P1_WIRELESS_NOTIFICATIONS_MASK = (1 << 0)
	UNIYING_WIRELESS_NOTIFICATIONS_P1_SOFTWARE_PRESENT_MASK = (1 << 3)

)

type HidPPMsg struct {
	ReportID   HidPPReportType
	DeviceID   byte
	MsgSubID   byte
	Parameters []byte
}

func (r *HidPPMsg) String() string {
	return fmt.Sprintf("ReportID: 0x%#x, DeviceID: 0x%#x, MsgSubID: 0x%#x, Params: 0x% #x", r.ReportID, r.DeviceID, r.MsgSubID, r.Parameters)
}

func (r *HidPPMsg) FromWire(payload []byte) (err error) {
	if len(payload) == HIDPP_TYPE_LONG_LEN && payload[0] == byte(HIDPP_TYPE_LONG) {
		r.ReportID = HIDPP_TYPE_LONG
		r.DeviceID = payload[1]
		r.MsgSubID = payload[2]
		r.Parameters = make([]byte,16)
		copy(r.Parameters, payload[3:])
		return
	}
	if len(payload) == HIDPP_TYPE_SHORT_LEN && payload[0] == byte(HIDPP_TYPE_SHORT) {
		r.ReportID = HIDPP_TYPE_SHORT
		r.DeviceID = payload[1]
		r.MsgSubID = payload[2]
		r.Parameters = make([]byte,3)
		copy(r.Parameters, payload[3:])
		return
	}

	return eNoHidPPMsg
}

func (r *HidPPMsg) ToWire() (payload []byte, err error) {
	if r.ReportID == HIDPP_TYPE_SHORT {
		payload := make([]byte, HIDPP_TYPE_SHORT_LEN)
		payload[0] = byte(r.ReportID)
		payload[1] = r.DeviceID
		payload[2] = r.MsgSubID
		if r.Parameters != nil {
			copy(payload[3:], r.Parameters)
		}
		return payload,nil
	}

	if r.ReportID == HIDPP_TYPE_LONG {
		payload := make([]byte, HIDPP_TYPE_LONG_LEN)
		payload[0] = byte(r.ReportID)
		payload[1] = r.DeviceID
		payload[2] = r.MsgSubID
		if r.Parameters != nil {
			copy(payload[3:], r.Parameters)
		}
		return payload,nil
	}

	err = eNoHidPPReportID
	return
}

type Unifying struct {
	UsbCtx *gousb.Context
	Dev *gousb.Device
	Config *gousb.Config
	IfaceHIDPP *gousb.Interface
	EpInHidPP *gousb.InEndpoint

	sndQueue chan []byte
	rcvQueue chan []byte
	cancel context.CancelFunc
	ctx context.Context

	reqRespQueue chan [2]*HidPPMsg
}

func (u *Unifying) SendHidPPMessage(msg HidPPMsg) (err error) {
	p,err := msg.ToWire()
	if err != nil {
		return err
	}
	u.sndQueue <- p
	return nil
}

/*
func (u *Unifying) SendRcvHidPPMessage(msg HidPPMsg, timeout time.Duration) (res *HidPPMsg, err error) {
	p,err := msg.ToWire()
	if err != nil {
		return nil,err
	}
	u.sndQueue <- p

	return nil
}
*/

func (u *Unifying) readLoop() {
	buf := make([]byte, 32)

	for {
		n,err := u.EpInHidPP.ReadContext(u.ctx, buf)
		if err != nil {
			break
		}
		rcv := make([]byte, n)
		copy(rcv, buf[:n])
		u.rcvQueue <- rcv
	}

	close(u.rcvQueue)
}

func (u *Unifying) sndRcvLoop() {
	Outer:
	for {
		select {
		case <- u.ctx.Done():
			break Outer
		case outdata := <- u.sndQueue:
			if len(outdata) == 0 {
				continue
			}
			fmt.Printf("Out: % #x\n", outdata)
			u.Dev.Control(
				0x21, //bit7: Host to device, bit6..5: Class: 0x1, bit4..0: Interface: 0x01
				0x09, //request: 0x09 SET_REPORT
				0x0200 | uint16(outdata[0]), //Output: 0x02, Report ID: 0x10
				2, //Index 0x02
				outdata, //payload
			)
		case indata := <- u.rcvQueue:
			fmt.Printf("In: % #x\n", indata)
		}
	}

	close(u.sndQueue)
	close(u.reqRespQueue)
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
	res.Dev,err = res.UsbCtx.OpenDeviceWithVIDPID(UNIFYING_VID, UNIFYING_PID)
	if err != nil || res.Dev == nil {
		res.Close()
		return nil, eNoDongle
	}
	fmt.Println("Unifying dongle found", res.Dev)

	//Get device config 1
	res.Config,err = res.Dev.Config(1)
	if err != nil {
		res.Close()
		return nil,errors.New("Couldn't retrieve config 1 of Unifying dongle")
	}

	fmt.Println("Using dongle USB config:", res.Config.Desc.String())

	fmt.Println("Resetting dongle in order to release it from kernel (connected devices won't be usable)")
	//res.Dev.Reset()
	res.Dev.SetAutoDetach(true)

	Outer:
	for _,ifaceDesc := range res.Config.Desc.Interfaces {
		for _,ifaceSettings := range ifaceDesc.AltSettings {
			//fmt.Printf("%+v\n", ifaceSettings.Endpoints)
			for _, epDesc := range ifaceSettings.Endpoints {
				if epDesc.MaxPacketSize == 32 && epDesc.Direction == gousb.EndpointDirectionIn {
					// This is the HID++ EP
					//fmt.Printf("EP %+v\n", epDesc.Number)
					res.IfaceHIDPP,err = res.Config.Interface(ifaceSettings.Number, ifaceSettings.Alternate)
					if err != nil {
						res.Close()
						return nil, errors.New("Couldn't access HID++ USB interface")
					} else {
						fmt.Println("HID++ interface:", res.IfaceHIDPP.String())
					}

					res.EpInHidPP,err = res.IfaceHIDPP.InEndpoint(epDesc.Number)
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
		return nil,errors.New("Couldn't find EP for HID++ input reports")
	}

	res.reqRespQueue = make(chan [2]*HidPPMsg)

	res.sndQueue = make(chan []byte)
	res.rcvQueue = make(chan []byte)

	res.ctx,res.cancel = context.WithCancel(context.Background())

	go res.readLoop()
	go res.sndRcvLoop()

	return
}


func main() {
	u,err := NewUnifying()
	if err != nil {
		panic(err)
	}
	defer u.Close()

	u.SendHidPPMessage(HidPPMsg{
		ReportID: HIDPP_TYPE_SHORT,
		DeviceID: 0xff,
		MsgSubID: UNIFYING_MSG_ID_GET_LONG_REGISTER_REQ,
		Parameters: []byte{0xb5, 0x03},
	})



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
