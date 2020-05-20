package message

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type HeartRateMessage AntBroadcastMessage

func (m HeartRateMessage) String() (s string) {
	s = fmt.Sprintf("[HRM] %d", m.HeartRate())
	return
}

// HeartRate represents the computed heart rate
func (m HeartRateMessage) HeartRate() (num uint8) {
	binary.Read(bytes.NewReader(AntBroadcastMessage(m).Content()[7:8]), binary.LittleEndian, &num)
	return
}
