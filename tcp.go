package dlt645

import (
	"log"
	"net"
	"sync"
	"time"
)

type tcpClient struct {
	Address string
	// Connect & Read timeout
	Timeout time.Duration
	// Idle timeout to close the connection
	IdleTimeout time.Duration
	// Transmission logger
	Logger *log.Logger

	conn         net.Conn
	mu           sync.Mutex
	closeTimer   *time.Timer
	lastActivity time.Time
	data         []byte
}

func newTCPClient(address string) *tcpClient {
	return &tcpClient{
		Address: address,
		data:    make([]byte, 256),
	}
}

func (t *tcpClient) Send(data []byte) (_ []byte, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if err = t.connect(); err != nil {
		return
	}

	t.lastActivity = time.Now()
	t.startCloseTimer()

	var timeout time.Time
	if t.Timeout > 0 {
		timeout = t.lastActivity.Add(t.Timeout)
	}
	if err = t.conn.SetDeadline(timeout); err != nil {
		return
	}

	// start := data[0]
	// end := data[len(data)-1]

	// Send data
	t.logf("dlt645: sending % x", data[:])
	if _, err = t.conn.Write(data); err != nil {
		return
	}

	n := 0
	resp := make([]byte, 256)

	n, err = t.conn.Read(resp)
	if err != nil {
		return nil, err
	}

	t.logf("dlt645: received % x\n", resp[:n])
	return resp[:n], nil
}

func (t *tcpClient) Connect() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.connect()
}

// Close closes current connection.
func (t *tcpClient) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.close()
}

func (t *tcpClient) close() (err error) {
	if t.conn != nil {
		err = t.conn.Close()
		t.conn = nil
	}
	if t.closeTimer != nil {
		t.closeTimer.Stop()
		t.closeTimer = nil
	}
	return
}

func (t *tcpClient) connect() error {
	if t.conn == nil {
		dialer := net.Dialer{Timeout: t.Timeout}
		conn, err := dialer.Dial("tcp", t.Address)
		if err != nil {
			return err
		}
		t.conn = conn
	}
	return nil
}

func (t *tcpClient) startCloseTimer() {
	if t.IdleTimeout <= 0 {
		return
	}
	if t.closeTimer == nil {
		t.closeTimer = time.AfterFunc(t.IdleTimeout, t.closeIdle)
	} else {
		t.closeTimer.Reset(t.IdleTimeout)
	}
}

// closeIdle closes the connection if last activity is passed behind IdleTimeout.
func (t *tcpClient) closeIdle() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.IdleTimeout <= 0 {
		return
	}
	idle := time.Now().Sub(t.lastActivity)
	if idle >= t.IdleTimeout {
		t.logf("modbus: closing connection due to idle timeout: %v", idle)
		t.close()
	}
}

func (t *tcpClient) logf(format string, v ...interface{}) {
	if t.Logger != nil {
		t.Logger.Printf(format, v...)
	}
}
