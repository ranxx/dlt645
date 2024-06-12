package dlt645

import (
	"fmt"
)

// Type type
type Type int

// const type
const (
	Serial Type = iota
	TCP
)

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
	ReadCtrl2             byte = 0x12
	ReadRepyCtrl2         byte = 0x92
	ReadRepyUnfinishCtrl2 byte = 0xB2
	ReadRepyErrCtrl2      byte = 0xD2
)

// Sender sender
type Sender interface {
	Send([]byte) ([]byte, error)
}

// Client client
type Client struct {
	serialC *serialClient
	tcpC    *tcpClient

	DeviceAddress string

	config *Config
}

// NewClient new client
func NewClient(address string, t Type, deviceAddress string, cfs ...*Config) (*Client, error) {
	cf := mergeConfig(defaultConfig(), cfs...)
	switch t {
	case Serial:
		return newSerial(address, deviceAddress, cf)
	case TCP:
		return newTCP(address, deviceAddress, cf)
	}
	return nil, fmt.Errorf("不支持的type")
}

// Read read
func (c *Client) Read(ctrl byte, DDDD, data []byte) (*Frame, error) {
	frame, err := NewDLT645Frame(c.DeviceAddress, ctrl, DDDD, data)
	if err != nil {
		return nil, err
	}
	sendData := frame.ToData()

	var sender Sender
	if c.serialC != nil {
		sender = c.serialC
	} else if c.tcpC != nil {
		sender = c.tcpC
	}
	if sender == nil {
		return nil, fmt.Errorf("不支持的type")
	}

	respData, err := sender.Send(sendData)
	if err != nil {
		return nil, err
	}

	f, err := ParseDLT645Frame(respData)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// Close close
func (c *Client) Close() {
	if c.serialC != nil {
		c.serialC.Close()
	}
	if c.tcpC != nil {
		c.tcpC.Close()
	}
}

func newSerial(address, deviceAddress string, cf *Config) (*Client, error) {
	h := newSerialClient(address)
	if cf.Timeout > 0 {
		h.Timeout = cf.Timeout
	}
	if h.IdleTimeout > 0 {
		h.IdleTimeout = cf.IdleTimeout
	}
	if cf.BaudRate > 0 {
		h.BaudRate = cf.BaudRate
	}
	if cf.DataBits > 0 {
		h.DataBits = cf.DataBits
	}
	if len(cf.Parity) > 0 {
		h.Parity = cf.Parity
	}
	if cf.StopBits > 0 {
		h.StopBits = cf.StopBits
	}
	if cf.Logger != nil {
		h.Logger = cf.Logger
	}
	if err := h.Connect(); err != nil {
		return nil, err
	}
	return &Client{serialC: h, DeviceAddress: deviceAddress, config: cf}, nil
}

func newTCP(address, deviceAddress string, cf *Config) (*Client, error) {
	h := newTCPClient(address)
	if cf.Timeout > 0 {
		h.Timeout = cf.Timeout
	}
	if cf.IdleTimeout > 0 {
		h.IdleTimeout = cf.IdleTimeout
	}
	if cf.Logger != nil {
		h.Logger = cf.Logger
	}
	if err := h.Connect(); err != nil {
		return nil, err
	}
	return &Client{tcpC: h, DeviceAddress: deviceAddress, config: cf}, nil
}
