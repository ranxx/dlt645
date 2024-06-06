package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

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

func buildByte(begin byte, address string, control byte, dataLen int64, data []byte) ([]byte, error) {
	sendData := make([]byte, 0, 20)
	// 起始
	sendData = append(sendData, begin)
	// address
	addressBCD, err := AddressToBCD(address)
	if err != nil {
		return nil, err
	}
	sendData = append(sendData, addressBCD...)
	// 起始
	sendData = append(sendData, begin)
	// 控制码
	sendData = append(sendData, control)
	// 数据长度
	sendData = append(sendData, byte(dataLen))
	// 数据 + 33H
	for i := range data {
		data[i] += 0x33
	}
	sendData = append(sendData, data...)
	// 校验码
	sendData = append(sendData, byte(0))
	// 结束符
	sendData = append(sendData, 0x16)
	// 计算 cs
	sendData[len(sendData)-2] = calcCS(sendData[:len(sendData)-2])

	return sendData, nil
}

func calcCS(data []byte) byte {
	count := 0
	for i := 0; i < len(data); i++ {
		count += int(data[i])
	}
	return (byte)(count & 0xFF)
}

// // func
// func parseResult(start byte, address string, control byte, data []byte, data []byte) (interface{}, error) {

// 	// 解析
// 	// 起始帧
// 	// 地址
// 	// 起始帧
// 	// 控制码
// 	// 长度
// 	// 读取长度

// 	/*
// 		68H A0 … A5 68H D1H 01H ERR CS 16H
// 	*/

// 	if len(data) < 11 {
// 		return nil, fmt.Errorf("")
// 	}

// 	if len(data) >= 1 {
// 		if data[0] == start {

// 		}
// 	}

// 	return nil, nil

// }

// const
const (
	StartCtrl byte = 0x68
	EndCtrl   byte = 0x16

	// 读数据
	ReadCtrl             byte = 0x11
	ReadRepyCtrl         byte = 0x91
	ReadRepyUnfinishCtrl byte = 0xB1
	ReadRepyErrCtrl      byte = 0xD1
	// 读后续
	Read2Ctrl             byte = 0x12
	ReadRepy2Ctrl         byte = 0x92
	ReadRepy2UnfinishCtrl byte = 0xB2
	ReadRepyErr2Ctrl      byte = 0xD2

	// DefaultSendStartFrame           byte = 0x68
	// DefaultSendEndFrame             byte = 0x16
	// DefaultSendFrame                byte = 0x11
	// DefaultReceiveEndFrame          byte = 0x91
	// DefaultReceiveHasRemainFrame    byte = 0xB1
	// DefaultReceiveHasRemainEndFrame byte = 0x92
	// DefaultReceiveErrFrame          byte = 0xD1
)

// 变
// var (
// 	CusStartFrame  byte = DefaultStartFrame
// 	CusStart2Frame byte = DefaultStartFrame
// 	CusEndFrame    byte = DefaultEndFrame
// )

// DLT645Frame 表示一个DLT645协议的帧
type DLT645Frame struct {
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
func NewDLT645Frame(address string, ctrl byte, DDDD, data []byte) (*DLT645Frame, error) {
	addressBcd, err := AddressToBCD(address)
	if err != nil {
		return nil, err
	}
	frame := &DLT645Frame{
		Start:   StartCtrl,
		Address: addressBcd,
		Start2:  StartCtrl,
		Ctrl:    ctrl,
		DataLen: byte(len(DDDD) + len(data)),
		DDDD:    DDDD,
		Data:    data,
		End:     EndCtrl,
	}

	calcData := make([]byte, 0, int(12+frame.DataLen))
	calcData = append(calcData, frame.Start)
	calcData = append(calcData, frame.Address...)
	calcData = append(calcData, frame.Start2)
	calcData = append(calcData, frame.Ctrl)
	calcData = append(calcData, frame.DataLen)
	calcData = append(calcData, frame.DDDD...)
	calcData = append(calcData, frame.Data...)
	frame.CS = calcCS(calcData)
	return frame, nil
}

// ToData to data
func (d DLT645Frame) ToData() []byte {
	data := make([]byte, 0, 12+len(d.Data))

	data = append(data, d.Start)

	data = append(data, d.Address...)

	data = append(data, d.Start2)

	data = append(data, d.Ctrl)

	data = append(data, d.DataLen)

	data = append(data, d.DDDD...)

	data = append(data, d.Data...)

	data = append(data, 0)

	data = append(data, d.End)

	data[len(data)-2] = calcCS(data[:len(data)-2])

	return data
}

// ParseDLT645Frame 解析DLT645协议帧
func ParseDLT645Frame(data []byte) (*DLT645Frame, error) {
	if len(data) < 12 {
		return nil, errors.New("data too short")
	}

	if data[0] != StartCtrl || data[7] != StartCtrl || data[len(data)-1] != EndCtrl {
		return nil, errors.New("invalid frame")
	}

	frame := &DLT645Frame{
		Start:   data[0],
		Address: make([]byte, 6),
		Start2:  data[7],
		Ctrl:    data[8],
		DataLen: data[9],
		End:     data[len(data)-1],
	}

	copy(frame.Address[:], data[1:7])

	frame.DDDD = data[10 : 10+frame.DataLen]
	frame.CS = data[10+frame.DataLen]

	if frame.DataLen >= 4 {
		frame.Data = frame.DDDD[4:]
		frame.DDDD = frame.DDDD[:4]
	}

	if calcCS(data[:len(data)-2]) != frame.CS {
		return nil, errors.New("checksum error")
	}

	return frame, nil
}

type Client struct {
	// 变
	StartFrame  byte
	Start2Frame byte
	EndFrame    byte
}

type tcpClient struct {
}

type Request struct {

	// 	// 解析
	// 	// 起始帧
	// 	// 地址
	// 	// 起始帧
	// 	// 控制码
	// 	// 长度
	// 	// 读取长度
}

func main() {
	address := "202204080026" // Example address
	bcd, err := AddressToBCD(address)
	if err != nil {
		fmt.Println("Error converting address to BCD:", err)
		return
	}
	fmt.Printf("Original Address: %s -> BCD (reversed): %X\n", address, bcd)

	convertedAddress, err := BCDToAddress(bcd)
	if err != nil {
		fmt.Println("Error converting BCD to address:", err)
		return
	}
	fmt.Printf("BCD (reversed): %X -> Original Address: %s\n", bcd, convertedAddress)

	// frame :=

	// data, err := buildByte(0x68, address, 0x11, 4, []byte{0, 0, 0, 0})
	// data, err := buildByte(0x68, address, 0x11, 2, []byte{0x02, 0x01, 0x01, 0})
	// data, err := buildByte(0x68, address, 0x11, 4, []byte{0, 1, 1, 2})
	data, err := buildByte(0x68, address, 0x11, 4, []byte{0, 0, 1, 1})
	if err != nil {
		panic(err)
	}

	ipaddr := ""

	conn, err := net.Dial("tcp", ipaddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	n, err := conn.Write(data)
	if err != nil {
		panic(err)
	}
	log.Printf("send len: %v data: % x\n", n, data)

	resp := make([]byte, 1024)

	n, err = conn.Read(resp)
	if err != nil {
		panic(err)
	}
	log.Printf("read len: %v data: % x\n", n, resp[:n])

	/*
		68H A0 … A5 68H B1H L DI0 … DI3 N1 … Nm CS 16H
	*/

	resp = resp[:n]
	begin := false
	for i := 0; i < len(resp); i++ {
		v := resp[i]

		if !begin && v == 0x91 && i != 0 && resp[i-1] == 0x68 {
			begin = true
			i++
			continue
		}
		if begin {
			resp[i] -= 0x33
		}
	}

	// 解析
	// 起始帧
	// 地址
	// 起始帧
	// 控制码
	// 长度
	// 读取长度

	log.Printf("read-after len: %v data: % x\n", n, resp)
}
