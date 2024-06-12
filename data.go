package dlt645

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Frame 表示一个DLT645协议的帧
type Frame struct {
	Start   byte   // 1
	Address []byte // 6
	Start2  byte   // 1
	Ctrl    byte   // 1
	DataLen byte   // 1 4 + len(data)
	DDDD    []byte // len(DDDD)
	Data    []byte // len(data)
	CS      byte   // 1
	End     byte   // 1
}

// NewDLT645Frame new
func NewDLT645Frame(address string, ctrl byte, DDDD, data []byte) (*Frame, error) {
	addressBcd, err := AddressToBCD(address)
	if err != nil {
		return nil, err
	}

	dddd := make([]byte, len(DDDD))
	for i := len(DDDD) - 1; i >= 0; i-- {
		dddd[len(DDDD)-i-1] = DDDD[i] + 0x33
	}

	redata := make([]byte, len(data))
	for i := len(data) - 1; i >= 0; i-- {
		redata[len(data)-i-1] = data[i] + 0x33
	}

	frame := &Frame{
		Start:   StartCtrl,
		Address: addressBcd,
		Start2:  StartCtrl,
		Ctrl:    ctrl,
		DataLen: byte(len(dddd) + len(redata)),
		DDDD:    dddd,
		Data:    redata,
		CS:      0,
		End:     EndCtrl,
	}

	return frame, nil
}

// ToData to data
func (f Frame) ToData() []byte {
	data := make([]byte, 0, 12+len(f.Data))

	data = append(data, f.Start)

	data = append(data, f.Address...)

	data = append(data, f.Start2)

	data = append(data, f.Ctrl)

	data = append(data, f.DataLen)

	data = append(data, f.DDDD...)

	data = append(data, f.Data...)

	data = append(data, 0)

	data = append(data, f.End)

	data[len(data)-2] = CalculateCS(data[:len(data)-2])

	return data
}

// AddressToBCD converts a DL/T 645 communication address (as a decimal string)
// to its BCD representation in protocol order (reversed).
func AddressToBCD(address string) ([]byte, error) {
	// Ensure the address is exactly 12 characters long.
	if len(address) > 12 {
		return nil, fmt.Errorf("address must be exactly 12 digits long")
	}
	if len(address) < 12 {
		address += strings.Repeat("0", len(address)-12)
	}
	// Convert address to BCD and reverse it according to DL/T 645 protocol.
	bcd := make([]byte, 6) // 12 digits fit into 6 BCD bytes
	for i := 0; i < 6; i++ {
		high, err := strconv.Atoi(string(address[2*i]))
		if err != nil {
			return nil, fmt.Errorf("invalid digit at position %d", 2*i)
		}
		low, err := strconv.Atoi(string(address[2*i+1]))
		if err != nil {
			return nil, fmt.Errorf("invalid digit at position %d", 2*i+1)
		}
		// Store bytes in reverse order
		bcd[5-i] = byte(high<<4 | low)
	}

	return bcd, nil
}

// BCDToAddress converts a BCD encoded byte array (in DL/T 645 order)
// back to a decimal address string.
func BCDToAddress(bcd []byte) (string, error) {
	if len(bcd) != 6 {
		return "", fmt.Errorf("BCD byte array must be exactly 6 bytes long")
	}

	address := ""
	for i := 5; i >= 0; i-- { // Reverse order
		address += strconv.Itoa(int(bcd[i] >> 4))
		address += strconv.Itoa(int(bcd[i] & 0x0F))
	}

	return address, nil
}

// CalculateCS 计算校验和
func CalculateCS(data []byte) byte {
	count := 0
	for i := 0; i < len(data); i++ {
		count += int(data[i])
	}
	return (byte)(count & 0xFF)
}

// ParseDLT645Frame 解析DLT645协议帧
func ParseDLT645Frame(data []byte) (*Frame, error) {
	if len(data) < 12 {
		return nil, errors.New("data too short")
	}

	if data[0] != StartCtrl || data[7] != StartCtrl || data[len(data)-1] != EndCtrl {
		return nil, errors.New("invalid frame")
	}

	frame := &Frame{
		Start:   data[0],
		Address: make([]byte, 6),
		Start2:  data[7],
		Ctrl:    data[8],
		DataLen: data[9],
		End:     data[len(data)-1],
	}

	copy(frame.Address[:], data[1:7])

	frame.CS = data[10+frame.DataLen]

	if CalculateCS(data[:len(data)-2]) != frame.CS {
		return nil, errors.New("checksum error")
	}

	if frame.Ctrl == ReadRepyErrCtrl || frame.Ctrl == ReadRepyErrCtrl2 {
		return nil, fmt.Errorf("read resp err: 0x%x", frame.Ctrl)
	}

	if frame.DataLen >= 4 {
		frame.DDDD = make([]byte, 4)
		frame.Data = make([]byte, frame.DataLen-4)
	}

	for i := len(frame.DDDD); i <= len(frame.DDDD) && i > 0; i-- {
		frame.DDDD[len(frame.DDDD)-i] = data[10+i-1] - 0x33
	}

	for i := int(frame.DataLen) - (len(frame.DDDD)); i <= len(frame.Data) && i > 0; i-- {
		frame.Data[len(frame.Data)-i] = data[10+len(frame.DDDD)+i-1] - 0x33
	}

	return frame, nil
}
