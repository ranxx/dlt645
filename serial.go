package dlt645

import (
	"log"
	"sync"
	"time"

	"github.com/goburrow/serial"
)

type serialClient struct {
	Address string
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

	conn         serial.Port
	mu           sync.Mutex
	closeTimer   *time.Timer
	lastActivity time.Time
	data         []byte
}

func newSerialClient(address string) *serialClient {
	return &serialClient{
		Address: address,
	}
}

func (s *serialClient) Send(data []byte) (_ []byte, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err = s.connect(); err != nil {
		return
	}

	s.lastActivity = time.Now()
	s.startCloseTimer()

	// var timeout time.Time
	// if s.Timeout > 0 {
	// 	timeout = s.lastActivity.Add(s.Timeout)
	// }

	// if err = s.conn.SetDeadline(timeout); err != nil {
	// 	return
	// }

	// Send data
	s.logf("dlt645: sending % x", data)
	if _, err = s.conn.Write(data); err != nil {
		return
	}

	n := 0
	resp := make([]byte, 256)

	n, err = s.conn.Read(resp)
	if err != nil {
		return nil, err
	}

	s.logf("dlt645: received % x\n", resp[:n])
	return resp[:n], nil
}

func (s *serialClient) Connect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.connect()
}

func (s *serialClient) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.close()
}

func (s *serialClient) connect() error {
	if s.conn == nil {
		p, err := serial.Open(&serial.Config{
			Address:  s.Address,
			BaudRate: s.BaudRate,
			DataBits: s.DataBits,
			StopBits: s.StopBits,
			Parity:   s.Parity,
			Timeout:  s.Timeout,
			RS485:    serial.RS485Config(*s.RS485),
		})
		if err != nil {
			return err
		}
		s.conn = p
	}
	return nil
}

func (s *serialClient) close() (err error) {
	if s.conn != nil {
		err = s.conn.Close()
		s.conn = nil
	}
	if s.closeTimer != nil {
		s.closeTimer.Stop()
		s.closeTimer = nil
	}
	return
}

func (s *serialClient) startCloseTimer() {
	if s.IdleTimeout <= 0 {
		return
	}
	if s.closeTimer == nil {
		s.closeTimer = time.AfterFunc(s.IdleTimeout, s.closeIdle)
	} else {
		s.closeTimer.Reset(s.IdleTimeout)
	}
}

func (s *serialClient) closeIdle() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.IdleTimeout <= 0 {
		return
	}
	idle := time.Now().Sub(s.lastActivity)
	if idle >= s.IdleTimeout {
		s.logf("modbus: closing connection due to idle timeout: %v", idle)
		s.close()
	}
}

func (s *serialClient) logf(format string, v ...interface{}) {
	if s.Logger != nil {
		s.Logger.Printf(format, v...)
	}
}
