package unifying

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/gousb"
	log "github.com/sirupsen/logrus"
	"time"
)

var (
	eNoDongle        = errors.New("No Logitech Receiver dongle found")
)

const (
	VID          gousb.ID = 0x046d
	PID_UNIFYING gousb.ID = 0xc52b //cu0007, cu0008, cu0012
	PID_CU0016   gousb.ID = 0xc540 //cu0016
	PID_RR0011   gousb.ID = 0xc53e //R-R0011
)


type LocalUSBDongle struct {
	UsbCtx     *gousb.Context
	Dev        *gousb.Device
	Config     *gousb.Config
	IfaceHIDPP *gousb.Interface
	EpInHidPP  *gousb.InEndpoint

	sndQueue chan USBReport
	rcvQueue chan USBReport
	cancel   context.CancelFunc
	ctx      context.Context

	showInOut                  bool
}

func (u *LocalUSBDongle) SendUSBReport(msg USBReport) (err error) {
	u.sndQueue <- msg
	return nil
}

func (u *LocalUSBDongle) ReceiveUSBReport(timeoutMillis int) (msg USBReport, err error) {
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

func (u *LocalUSBDongle) rcvLoop() {
	buf := make([]byte, 32)

	for {
		n, err := u.EpInHidPP.ReadContext(u.ctx, buf)
		if err != nil {
			break
		}

		if u.showInOut {
			fmt.Printf("\nIn: % #x\n", buf[:n])
		}
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

func (u *LocalUSBDongle) sndLoop() {
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

			if u.showInOut {
				fmt.Printf("Out: % #x\n", outdata)
			}
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

func (u *LocalUSBDongle) SetShowInOut(show bool) {
	u.showInOut = show
	return
}

func (u *LocalUSBDongle) Close() {
	fmt.Println("Closing Logitech Receiver dongle...")
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


func (u *LocalUSBDongle) HIDPP_SendAndCollectResponses(deviceID byte, id HidPPMsgSubID, parameters []byte) (responseReports []USBReport, err error) {
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
			return responseReports, errors.New("USB response timeout")
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
					return responseReports, nil
				}

				if rspHIDpp.DeviceID == deviceID && rspHIDpp.MsgSubID == HIDPP_MSG_ID_ERROR_MSG && rspHIDpp.Parameters[0] == byte(id) {
					// likely final response, return
					return responseReports, errors.New("HID++ error response")
				}
			}
		}
	}

}

func (u *LocalUSBDongle) HIDPP_Send(deviceID byte, id HidPPMsgSubID, parameters []byte) (err error) {
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
	return u.SendUSBReport(hidppReq)
}

func (u *LocalUSBDongle) EnablePairing(timeOutSeconds byte, devNumber byte, blockTillOff bool) (err error) {
	//Enable pairing
	connectDevices := byte(0x01) //open lock
	deviceNumber := devNumber    //According to specs: Same value as device index transmitted in 0x41 notification, but we haven't tx'ed anything
	openLockTimeout := timeOutSeconds
	fmt.Printf("Enable pairing for %d seconds\n", openLockTimeout)

/*
	if !blockTillOff {
		return u.HIDPP_Send(0xff, HIDPP_MSG_ID_SET_REGISTER_REQ, []byte{byte(DONGLE_HIDPP_REGISTER_PAIRING), connectDevices, deviceNumber, openLockTimeout})
	}
*/
	responses, err := u.HIDPP_SendAndCollectResponses(0xff, HIDPP_MSG_ID_SET_REGISTER_REQ, []byte{byte(DONGLE_HIDPP_REGISTER_PAIRING), connectDevices, deviceNumber, openLockTimeout})
	for _, r := range responses {
		fmt.Println(r.String())
	}
	if err != nil {
		return
	}

	fmt.Println("... Enable pairing response (should be enabled)\n")

	if !blockTillOff {
		return nil
	}


	//Parse successive input reports till new "receiver lock information" with lock closed occurs
	fmt.Println("Printing follow up reports ...")
	for {
		rspUSB, err := u.ReceiveUSBReport(500)
		if err == nil {

			fmt.Println(rspUSB.String())
			if rspUSB.IsHIDPP() {
				hidppRsp := rspUSB.(*HidPPMsg)
				if hidppRsp.MsgSubID == HIDPP_MSG_ID_RECEIVER_LOCKING_INFORMATION && (hidppRsp.Parameters[0]&0x01) == 0 {
					switch hidppRsp.Parameters[1] {
					case 0x00:
						return nil //"no error"
					case 0x01:
						return errors.New("pairing timeout or interrupted")
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
				if hidppRsp.MsgSubID == HIDPP_MSG_ID_DEVICE_CONNECTION {
					devIdx := hidppRsp.DeviceID
					wpid := uint16(hidppRsp.Parameters[3])<<8 + uint16(hidppRsp.Parameters[2])
					fmt.Printf("DEVICE CONNECTION ON INDEX: %02x TYPE: %s WPID: %#04x\n", devIdx, DeviceType(hidppRsp.Parameters[1]&0x0F), wpid)

					//request additional information
				}
			}

		}
	}

}

func (u *LocalUSBDongle) DisablePairing() (err error) {
	//Enable pairing
	connectDevices := byte(0x02) //close lock

	_,err = u.HIDPP_SendAndCollectResponses(0xff, HIDPP_MSG_ID_SET_REGISTER_REQ, []byte{byte(DONGLE_HIDPP_REGISTER_PAIRING), connectDevices, 0, 0})
	return err
}

func (u *LocalUSBDongle) Unpair(deviceIndex byte) (err error) {
	//Enable pairing
	connectDevices := byte(0x03) //unpair
	deviceNumber := deviceIndex  //According to specs: Same value as device index transmitted in 0x41 notification, but we haven't tx'ed anything
	responses, err := u.HIDPP_SendAndCollectResponses(0xff, HIDPP_MSG_ID_SET_REGISTER_REQ, []byte{byte(DONGLE_HIDPP_REGISTER_PAIRING), connectDevices, deviceNumber})
	for _, r := range responses {
		fmt.Println(r.String())
	}
	return
}

func (u *LocalUSBDongle) GetNumPairedDevices() (numPairedDevices byte, err error) {
	//fmt.Println("GetPairedDevices")
	responses, err := u.HIDPP_SendAndCollectResponses(0xff, HIDPP_MSG_ID_GET_REGISTER_REQ, []byte{byte(DONGLE_HIDPP_REGISTER_CONNECTION_STATE)})

	var connStateResp *HidPPMsg = nil
	for _, r := range responses {
		if r.IsHIDPP() {
			hppmsg := r.(*HidPPMsg)
			if hppmsg.MsgSubID == HIDPP_MSG_ID_GET_REGISTER_RSP && len(hppmsg.Parameters) == 4 && hppmsg.Parameters[0] == byte(DONGLE_HIDPP_REGISTER_CONNECTION_STATE) {
				//Connection state response
				connStateResp = hppmsg
				break
			}
		}
		fmt.Println(r.String())
	}
	if connStateResp == nil {
		err = errors.New("couldn't determine count of paired devices")
		return
	}

	numPairedDevices = connStateResp.Parameters[2]
	//fmt.Println("Num paired devices:", numPairedDevices)
	return
}

func (u *LocalUSBDongle) GetDeviceActivityCounters() (activityCounters []byte, err error) {
	//fmt.Println("GetDeviceActivityCounters")
	responses, err := u.HIDPP_SendAndCollectResponses(0xff, HIDPP_MSG_ID_GET_LONG_REGISTER_REQ, []byte{byte(DONGLE_HIDPP_REGISTER_DEVICE_ACTIVITY)})

	var devActivityResp *HidPPMsg = nil
	for _, r := range responses {
		if r.IsHIDPP() {
			hppmsg := r.(*HidPPMsg)
			if hppmsg.MsgSubID == HIDPP_MSG_ID_GET_LONG_REGISTER_RSP && len(hppmsg.Parameters) == USB_REPORT_TYPE_HIDPP_LONG_PAYLOAD_LEN && hppmsg.Parameters[0] == byte(DONGLE_HIDPP_REGISTER_DEVICE_ACTIVITY) {
				//Connection state response
				devActivityResp = hppmsg
				break
			}
		}
		fmt.Println(r.String())
	}
	if devActivityResp == nil {
		err = errors.New("couldn't read device activity register")
		return
	}

	activityCounters = devActivityResp.Parameters[1:7]
	return
}



func (u *LocalUSBDongle) GetDevicePairingInfo(deviceID byte) (res DeviceInfo, err error) {
	if deviceID < 0 || deviceID > 6 {
		err = errors.New("invalid device ID")
		return
	}

	infoType := byte(0x20) //Pairing Info
	//fmt.Printf("GetDevicePairingInfo devIdx %d, infoType %02x\n", deviceID, infoType)
	responses, err := u.HIDPP_SendAndCollectResponses(0xff, HIDPP_MSG_ID_GET_LONG_REGISTER_REQ, []byte{byte(DONGLE_HIDPP_REGISTER_PAIRING_INFORMATION), deviceID + infoType})

	var devPairingInfo *HidPPMsg = nil
	for _, r := range responses {
		if r.IsHIDPP() {
			hppmsg := r.(*HidPPMsg)
			if hppmsg.MsgSubID == HIDPP_MSG_ID_GET_LONG_REGISTER_RSP && len(hppmsg.Parameters) == USB_REPORT_TYPE_HIDPP_LONG_PAYLOAD_LEN && hppmsg.Parameters[0] == byte(DONGLE_HIDPP_REGISTER_PAIRING_INFORMATION) && hppmsg.Parameters[1] == deviceID+infoType {
				//Connection state response
				devPairingInfo = hppmsg
				break
			}
		}
		fmt.Println(r.String())
	}
	if devPairingInfo == nil {
		err = errors.New("couldn't read device pairing info")
		return
	}

	res.DeviceIndex = deviceID
	res.DestinationID = devPairingInfo.Parameters[2]
	res.DefaultReportInterval = time.Duration(devPairingInfo.Parameters[3]) * time.Millisecond;
	res.WPID = devPairingInfo.Parameters[4:6]
	res.DeviceType = DeviceType(devPairingInfo.Parameters[8])

	res.Caps = LogitechDeviceCapabilities(devPairingInfo.Parameters[9])

	infoType = byte(0x30) //extended pairing Info
	//fmt.Printf("GetDevicePairingInfo devIdx %d, infoType %02x\n", deviceID, infoType)
	responses, err = u.HIDPP_SendAndCollectResponses(0xff, HIDPP_MSG_ID_GET_LONG_REGISTER_REQ, []byte{byte(DONGLE_HIDPP_REGISTER_PAIRING_INFORMATION), deviceID + infoType})

	var devExtPairingInfo *HidPPMsg = nil
	for _, r := range responses {
		if r.IsHIDPP() {
			hppmsg := r.(*HidPPMsg)
			if hppmsg.MsgSubID == HIDPP_MSG_ID_GET_LONG_REGISTER_RSP && len(hppmsg.Parameters) == USB_REPORT_TYPE_HIDPP_LONG_PAYLOAD_LEN && hppmsg.Parameters[0] == byte(DONGLE_HIDPP_REGISTER_PAIRING_INFORMATION) && hppmsg.Parameters[1] == deviceID+infoType {
				//Connection state response
				devExtPairingInfo = hppmsg
				break
			}
		}
		fmt.Println(r.String())
	}
	if devExtPairingInfo == nil {
		err = errors.New("couldn't read device extended pairing info")
		return
	}
	res.Serial = devExtPairingInfo.Parameters[2:6]
	res.ReportTypes = ReportTypes(0)
	res.ReportTypes.FromSlice(devExtPairingInfo.Parameters[6:10])
	res.UsabilityInfo = UsabilityInfo(devExtPairingInfo.Parameters[10])

	infoType = byte(0x40) //device name
	//fmt.Printf("GetDevicePairingInfo devIdx %d, infoType %02x\n", deviceID, infoType)
	responses, err = u.HIDPP_SendAndCollectResponses(0xff, HIDPP_MSG_ID_GET_LONG_REGISTER_REQ, []byte{byte(DONGLE_HIDPP_REGISTER_PAIRING_INFORMATION), deviceID + infoType})

	var devName *HidPPMsg = nil
	for _, r := range responses {
		//fmt.Println(r.String())
		if r.IsHIDPP() {
			hppmsg := r.(*HidPPMsg)
			if hppmsg.MsgSubID == HIDPP_MSG_ID_GET_LONG_REGISTER_RSP && len(hppmsg.Parameters) == USB_REPORT_TYPE_HIDPP_LONG_PAYLOAD_LEN && hppmsg.Parameters[0] == byte(DONGLE_HIDPP_REGISTER_PAIRING_INFORMATION) && hppmsg.Parameters[1] == deviceID+infoType {
				//Connection state response
				devName = hppmsg
				break
			}
		}
	}
	if devName == nil {
		err = errors.New("couldn't read device name")
		return
	}
	res.Name = string(devName.Parameters[3 : 3+devName.Parameters[2]])

	res.RawKeyData,_ = u.DumpRawKeyData(deviceID) //we ingnore errors, seems only to apply to dongles with WPID 0x8808 (not 0x8802)
	//fmt.Printf("Rawkey: % 02x\n", res.RawKeyData)
	if len(res.RawKeyData) > 0 {
		res.Key,_ = KeyData2Key(res.RawKeyData) //Ignore errors
	}


	res.RFAddr = make([]byte,5)

	return
}

/*
func (u *LocalUSBDongle) PrintInfoForAllConnectedDevices() (err error) {
	numPaired, err := u.GetNumPairedDevices()
	if err != nil {
		return
	}

	for devIdx := byte(0); devIdx < numPaired; devIdx++ {
		pi, ePi := u.GetDevicePairingInfo(devIdx)
		if ePi == nil {
			fmt.Println(pi.String())

		} else {
			fmt.Printf("Error for device index %d: %v\n", devIdx, ePi)
		}

	}
	return nil
}
*/

func (u *LocalUSBDongle) GetAllConnectedDevices() (devices []DeviceInfo, err error) {
	numPaired, err := u.GetNumPairedDevices()
	if err != nil {
		return
	}
	devices = make([]DeviceInfo,0)

	for devIdx := byte(0); devIdx < numPaired; devIdx++ {
		pi, ePi := u.GetDevicePairingInfo(devIdx)
		if ePi == nil {
			//fmt.Println(pi.String())
			devices = append(devices, pi)

		} else {
			fmt.Printf("Error for device index %d: %v\n", devIdx, ePi)
		}

	}
	return
}


func (u *LocalUSBDongle) GetDongleInfo() (res DongleInfo, err error) {
	//fmt.Printf("GetDevicePairingInfo devIdx %d, infoType %02x\n", deviceID, infoType)
	responses, err := u.HIDPP_SendAndCollectResponses(0xff, HIDPP_MSG_ID_GET_LONG_REGISTER_REQ, []byte{byte(DONGLE_HIDPP_REGISTER_PAIRING_INFORMATION), 0x02})

	var dongleInfo1 *HidPPMsg = nil
	for _, r := range responses {
		if r.IsHIDPP() {
			hppmsg := r.(*HidPPMsg)
			if hppmsg.MsgSubID == HIDPP_MSG_ID_GET_LONG_REGISTER_RSP && len(hppmsg.Parameters) == USB_REPORT_TYPE_HIDPP_LONG_PAYLOAD_LEN && hppmsg.Parameters[0] == byte(DONGLE_HIDPP_REGISTER_PAIRING_INFORMATION) && hppmsg.Parameters[1] == 0x02 {
				//Connection state response
				dongleInfo1 = hppmsg
				break
			}
		}
		fmt.Println(r.String())
	}
	if dongleInfo1 == nil {
		err = errors.New("couldn't read dongle info")
		return
	}

	res.FwMajor = dongleInfo1.Parameters[2]
	res.FwMinor = dongleInfo1.Parameters[3]
	res.FwBuild = uint16(dongleInfo1.Parameters[4])<<8 + uint16(dongleInfo1.Parameters[5])
	res.WPID = dongleInfo1.Parameters[6:8]
	res.LikelyProto = dongleInfo1.Parameters[8]

	responses, err = u.HIDPP_SendAndCollectResponses(0xff, HIDPP_MSG_ID_GET_LONG_REGISTER_REQ, []byte{byte(DONGLE_HIDPP_REGISTER_PAIRING_INFORMATION), 0x03}) //Note 0x03 flash table entry exists per device, we only grab the first one
	var dongleInfo2 *HidPPMsg = nil
	for _, r := range responses {
		if r.IsHIDPP() {
			hppmsg := r.(*HidPPMsg)
			if hppmsg.MsgSubID == HIDPP_MSG_ID_GET_LONG_REGISTER_RSP && len(hppmsg.Parameters) == USB_REPORT_TYPE_HIDPP_LONG_PAYLOAD_LEN && hppmsg.Parameters[0] == byte(DONGLE_HIDPP_REGISTER_PAIRING_INFORMATION) && hppmsg.Parameters[1] == 0x03 {
				//Connection state response
				dongleInfo2 = hppmsg
				break
			}
		}
		fmt.Println(r.String())
	}
	if dongleInfo1 == nil {
		err = errors.New("couldn't read dongle info")
		return
	}

	res.Serial = dongleInfo2.Parameters[2:6]

	responses, err = u.HIDPP_SendAndCollectResponses(0xff, HIDPP_MSG_ID_GET_REGISTER_REQ, []byte{byte(DONGLE_HIDPP_REGISTER_FIRMWARE), 0x04, 0x00})
	for _, r := range responses {
		if r.IsHIDPP() {
			hppmsg := r.(*HidPPMsg)
			if hppmsg.MsgSubID == HIDPP_MSG_ID_GET_REGISTER_RSP && len(hppmsg.Parameters) == USB_REPORT_TYPE_HIDPP_SHORT_PAYLOAD_LEN && hppmsg.Parameters[0] == byte(DONGLE_HIDPP_REGISTER_FIRMWARE) && hppmsg.Parameters[1] == 0x04 {
				//Connection state response
				res.BootloaderMajor = hppmsg.Parameters[2]
				res.BootloaderMinor = hppmsg.Parameters[3]
				break
			}
		}
	}


	//Bootloader version

	return
}

func (u *LocalUSBDongle) GetSetInfo() (set SetInfo, err error) {
	di, eDi := u.GetDongleInfo()
	if eDi == nil {
		//Create new set
		set = SetInfo{
			Dongle: di,
		}
		devs,eDevs := u.GetAllConnectedDevices()
		if eDevs == nil {
			for _,d := range devs {
				set.AddDevice(d)
			}
		}


		set.Dongle.NumConnectedDevices = byte(len(set.ConnectedDevices))
	} else {
		return set,eDi
	}
	return
}

// Dumps memory from flash / flash info page using undocumented register 0xd4
func (u *LocalUSBDongle) DumpFlashByte(addr uint16) (res byte, err error) {
	//reg 0xd4 reads an arbitrary byte from flash (xdata) at address given by r2 (MSB, r1(LSB)
	// only accessible flash pages are valid, on cu0007:
	// - 0x0000..0x000f (maps to active flashpage + 0x30..0x3f)
	// - 0x6c00..0x6fff (one to one mapping)
	// - 0xfe00..0xffff maps to flash info page 0x00..0x1ff (see table 117 of nRF24LU1+ specs)
	//
	// The active flash page holds pairing info, extended pairing info and names of connected devices + some dongle data.
	// Most of this data is accessible via pairing info register 0xb5, but the for the device RF address entries
	// (starting with 0x03 in active flash page) only the first entry could be fetched utilizing the pairing info reg
	//
	// On newer dongles, reading arbitrary addresses always produces a result (no error), but regions are zeroed out
	// on CU0012 with RQR24.07.0030 device data is layed out differently and contains per device entries which contain
	// the raw key material before substitution (4 byte dongle serial, 2 byte device WPID, 2 byte dongle WPID, 4 byte
	// device nonce, 4 byte dongle nonce). Those entries are prepended with a marker in form 0x6nffffff (n is device idx
	// between 0 and 5). It is likely that those table entries could be dumped using the pairing info register 0xb5, too!
	//
	// Example of flash region 0xe800 from mentioned dongle dumped using reg 0xd4:
	// 3fffffff
	// 02ffffff 24070030 88080401 00000000 00000000
	// 03ffffff e2c794f2 01064000 00000000 00000000
	// 7cffffff 00180724 d02a8da2 11000000 00000000
	// 60ffffff e2c794f2 404d8808 9393ffcb 273b052e <- key data dev 1
	// 03ffffff e2c794f2 02064100 00000000 00000000
	// 20ffffff 4108404d 04020147 00000000 00000000
	// 30ffffff 2d9a9fe1 1e400000 09000000 00000000
	// 40ffffff 094b3430 3020506c 75730000 00000000
	// 61ffffff e2c794f2 40048808 75d3e48e e88aa760 <-- key data dev 2
	// 03ffffff e2c794f2 0e064200 00000000 00000000
	// 21ffffff 42144004 0402010d 00000000 00000000
	// 31ffffff 0d63a9c2 1a400000 02000000 00000000
	// 41ffffff 044b3336 30000000 00000000 00000000

	addrMSB := byte (addr >> 8)
	addrLSB := byte(addr & 0xff)


	responses, _ := u.HIDPP_SendAndCollectResponses(0xff, HIDPP_MSG_ID_GET_REGISTER_REQ, []byte{byte(DONGLE_HIDPP_REGISTER_SECRET_MEMDUMP), addrLSB, addrMSB})
	for _, r := range responses {
		if r.IsHIDPP() {
			h := r.(*HidPPMsg)
			if h.Parameters[0] == byte(DONGLE_HIDPP_REGISTER_SECRET_MEMDUMP) {
				res = h.Parameters[3]
				return
			}
		}
	}
	err = errors.New("can not read given offset")
	return
}


func (u *LocalUSBDongle) DumpRawKeyData(devID byte) (res []byte, err error) {
	//find flash page with device data
	flashPagesToConsider := []uint16{0xe400, 0xe800, 0xec00}

	activePageAddr := uint16(0)
	for _,pageAddr := range flashPagesToConsider {
		dB,eDb := u.DumpFlashByte(pageAddr)
		if eDb == nil && dB == 0x3f {
			activePageAddr = pageAddr
			break
		}
	}
	if activePageAddr == 0x00 {
		err = errors.New("could not find active flash page")
		return
	}

	activePageAddr += 0x04 //offset to first entry
	marker := byte(0x60 + devID)
	keyAddr := uint16(0)
	stepsize := 0x14
	maxSteps := 3 + 5*6 //3 entries for dongle, max 5 entries per device
	for step := 0; step < maxSteps; step++ {
		checkAddress := activePageAddr + uint16(stepsize * step)
		dB,eDb := u.DumpFlashByte(checkAddress)
		if eDb == nil && dB == marker {
			keyAddr = checkAddress
			break
		}
	}
	if keyAddr == 0x00 {
		err = errors.New("could not find entry with key data")
		return
	}
	keyAddr += 4
	res = make([]byte, 16)
	for idx,_ := range  res {
		res[idx],_ = u.DumpFlashByte(keyAddr + uint16(idx))
	}
	return

}

func NewLocalUSBDongle() (res *LocalUSBDongle, err error) {
	res = &LocalUSBDongle{}
	res.showInOut = true

	res.UsbCtx = gousb.NewContext()

	/*
	res.Dev, err = res.UsbCtx.OpenDeviceWithVIDPID(VID, PID_UNIFYING)
	if err != nil || res.Dev == nil {

		log.Println("No Unifying dongle found, try to find dongle for R500 presentation clicker")

		// try R500 dongle CU0016
		res.Dev, err = res.UsbCtx.OpenDeviceWithVIDPID(VID, PID_RR0011)
		if err != nil || res.Dev == nil {
			res.Close()
			log.Println("No valid Unifying or presentation clicker dongle found")
			return nil, eNoDongle
		}
		log.Println("Found RR0011 for SPOTLIGHT presentation clicker")
	}
	fmt.Println("LocalUSBDongle dongle found", res.Dev)
	*/

	if res.Dev, err = res.UsbCtx.OpenDeviceWithVIDPID(VID, PID_UNIFYING); err == nil && res.Dev != nil {
		fmt.Println("Logitech Unifying dongle found")
	} else if res.Dev, err = res.UsbCtx.OpenDeviceWithVIDPID(VID, PID_RR0011); err == nil && res.Dev != nil {
		fmt.Println("Found CU0016 Dongle for Logitech SPOTLIGHT presentation clicker")
	} else if res.Dev, err = res.UsbCtx.OpenDeviceWithVIDPID(VID, PID_CU0016); err == nil && res.Dev != nil {
		fmt.Println("Found CU0016 Dongle for R500 presentation clicker")
	} else {
		res.Close()
		log.Fatal("No known dongle found")

		return nil, eNoDongle
	}

	//Get device config 1
	res.Config, err = res.Dev.Config(1)
	if err != nil {
		res.Close()
		return nil, errors.New("Couldn't retrieve config 1 of LocalUSBDongle dongle")
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
				//fmt.Printf("EP descr: %+v\n", epDesc.String())
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