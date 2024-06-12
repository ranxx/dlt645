package dlt645

import (
	"log"
	"time"
)

// Config is common configuration for serial port.
type Config struct {
	// Connect & Read timeout
	Timeout time.Duration
	// Idle timeout to close the connection
	IdleTimeout time.Duration
	// Transmission logger
	Logger *log.Logger
	// Baud rate (default 19200)
	BaudRate int
	// Data bits: 5, 6, 7 or 8 (default 8)
	DataBits int
	// Stop bits: 1 or 2 (default 1)
	StopBits int
	// Parity: N - None, E - Even, O - Odd (default E)
	// (The use of no parity requires 2 stop bits.)
	Parity string
	// Configuration related to RS485
	RS485 *RS485Config
}

// RS485Config platform independent RS485 config. Thie structure is ignored unless Enable is true.
type RS485Config struct {
	// Enable RS485 support
	Enabled bool
	// Delay RTS prior to send
	DelayRtsBeforeSend time.Duration
	// Delay RTS after send
	DelayRtsAfterSend time.Duration
	// Set RTS high during send
	RtsHighDuringSend bool
	// Set RTS high after send
	RtsHighAfterSend bool
	// Rx during Tx
	RxDuringTx bool
}

func defaultConfig() Config {
	return Config{
		Timeout:     time.Second * 5,
		IdleTimeout: time.Second * 60 * 5,
		BaudRate:    19200,
		DataBits:    8,
		StopBits:    1,
		Parity:      "E",
		// RS485:    RS485Config{},
	}
}

func mergeConfig(cf Config, cfs ...*Config) *Config {
	for _, v := range cfs {
		if v.Timeout > 0 {
			cf.Timeout = v.Timeout
		}
		if v.IdleTimeout > 0 {
			cf.IdleTimeout = v.IdleTimeout
		}
		if v.BaudRate > 0 {
			cf.BaudRate = v.BaudRate
		}
		if v.DataBits > 0 {
			cf.DataBits = v.DataBits
		}
		if v.StopBits > 0 {
			cf.StopBits = v.StopBits
		}
		if len(v.Parity) > 0 {
			cf.Parity = v.Parity
		}
		if v.RS485 != nil {
			cf.RS485 = v.RS485
		}
		if v.Logger != nil {
			cf.Logger = v.Logger
		}
	}
	return &cf
}
