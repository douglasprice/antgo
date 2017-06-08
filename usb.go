package main

import (
	"github.com/kylelemons/gousb/usb"
	"log"
	"errors"
)

type UsbDevice struct {
	vid, pid int
	context  *usb.Context
	device   *usb.Device
	in, out  usb.Endpoint
	Read chan []byte
	Write chan []byte
	stopLoop chan int
}

func (dev *UsbDevice) Open() (e error) {
	log.Println("Opening device")
	dev.Read = make(chan []byte)
	dev.Write = make(chan []byte)

	dev.context = usb.NewContext()
	dev.context.Debug(0)

	dev.device, e = dev.context.OpenDeviceWithVidPid(dev.vid, dev.pid)

	if e != nil {
		defer dev.context.Close()
		return
	}
	if dev.device == nil {
		defer dev.context.Close()
		e = errors.New("Device not found!")
		return
	}

	dev.in, e = dev.device.OpenEndpoint(1, 0, 0, uint8(1)|uint8(usb.ENDPOINT_DIR_IN))
	if e != nil {
		defer dev.context.Close()
		defer dev.Close()
		return
	}

	dev.out, e = dev.device.OpenEndpoint(1, 0, 0, uint8(1)|uint8(usb.ENDPOINT_DIR_OUT))
	if e != nil {
		defer dev.context.Close()
		defer dev.Close()
		return
	}

	go dev.loop()

	log.Println("Device opened")

	return
}

func (dev *UsbDevice) Close() {
	log.Println("Closing device")
	dev.stopLoop <- 1

	if dev.device != nil {
		dev.device.Close()
	}

	if dev.context != nil {
		dev.context.Close()
	}
	log.Println("Device closed")
}

func (dev *UsbDevice) StartRxScanMode() {
	dev.Write <- makeSystemResetMessage()
	dev.Write <- makeSetNetworkKeyMessage(0, []byte(ANTPLUS_NETWORK_KEY))
	dev.Write <- makeAssignChannelMessage(0, CHANNEL_TYPE_ONEWAY_RECEIVE)
	dev.Write <- makeSetChannelIdMessage(0)
	dev.Write <- makeSetChannelRfFrequencyMessage(0, 2457)
	dev.Write <- makeEnableExtendedMessagesMessage(true)
	dev.Write <- makeLibConfigMessage(true, true, true)
	dev.Write <- makeOpenRxScanModeMessage()
}

func (dev *UsbDevice) loop() {
	log.Println("Loop started")
	for {
		select {
		case <- dev.stopLoop:
			log.Println("Stopping loop")
			close(dev.Read)
			return
		case d := <- dev.Write:
			dev.out.Write(d)
		default:
			// Read from device
			buf := make([]byte, 64)
			i, err := dev.in.Read(buf)

			if err == nil {
				dev.Read <- buf[:i]
			}
		}
	}
}

func GetDevice(vid, pid int) *UsbDevice {
	return &UsbDevice{
		vid: vid,
		pid: pid,
		stopLoop: make(chan int),
	}
}
