package socket

import (
	"bufio"
	"fmt"
	"github.com/dbwk10317/gocket"
	"net"
	"sync"
)

type (
	TCPListener struct {
		listener   *net.TCPListener
		listenable bool
	}

	TCPSocket struct {
		conn        *net.TCPConn
		reader      *bufio.Reader
		writer      *bufio.Writer
		isConnected bool
		mutex       *sync.RWMutex
	}
)

func NewTCPAddress(host gocket.Host, port gocket.Port) (*net.TCPAddr, error) {
	address := fmt.Sprintf("%s:%d", host, port)
	addr, err := net.ResolveTCPAddr(gocket.TCPNetwork.String(), address)
	if err != nil {
		return nil, gocket.NewError(gocket.SocketListenerError, err)
	}

	return addr, nil
}

func NewTCPListener(port gocket.Port) *TCPListener {
	return &TCPListener{
		listenable: false,
	}
}

func (l *TCPListener) Bind(port gocket.Port) error {
	addr, err := NewTCPAddress("0.0.0.0", port)
	if err != nil {
		return gocket.NewError(gocket.SocketListenerError, err)
	}

	listener, err := net.ListenTCP(gocket.TCPNetwork.String(), addr)
	if err != nil {
		return gocket.NewError(gocket.SocketListenerError, err)
	}

	l.listener = listener
	l.listenable = true

	return nil
}

func (l *TCPListener) Accept() (gocket.Socket, error) {
	if l.listenable == false {
		return nil, gocket.NewErrorFromString(gocket.SocketListenerError, "TCPListener not available")
	}

	conn, err := l.listener.AcceptTCP()
	if err != nil {
		return nil, gocket.NewError(gocket.SocketListenerError, err)
	}

	return NewTCPSocketWithConnection(conn), nil
}

func (l *TCPListener) LocalAddress() net.Addr {
	if l.listenable == false {
		return nil
	}
	return l.listener.Addr()
}

func (l *TCPListener) Close() error {
	if l.listenable == false {
		return gocket.NewErrorFromString(gocket.SocketListenerError, "TCPListener not available")
	}
	return l.listener.Close()
}

func NewTCPSocketWithConnection(conn *net.TCPConn) *TCPSocket {
	s := NewTCPSocket()
	s.conn = conn
	s.reader = bufio.NewReader(conn)
	s.writer = bufio.NewWriter(conn)
	s.isConnected = true

	return s
}

func NewTCPSocket() *TCPSocket {
	return &TCPSocket{
		isConnected: false,
		mutex:       &sync.RWMutex{},
	}
}

func (s *TCPSocket) Connect(host gocket.Host, port gocket.Port) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.isConnected {
		return gocket.NewErrorFromString(gocket.SocketError, "socket already connected")
	}

	addr, err := NewTCPAddress(host, port)
	if err != nil {
		return gocket.NewError(gocket.SocketAddressError, err)
	}

	conn, err := net.DialTCP(gocket.TCPNetwork.String(), nil, addr)
	if err != nil {
		return gocket.NewError(gocket.SocketError, err)
	}

	s.conn = conn
	s.isConnected = true

	return nil
}

func (s *TCPSocket) Read(p []byte) (n int, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.isConnected == false {
		return 0, gocket.NewErrorFromString(gocket.SocketError, "socket already closed")
	}

	return s.reader.Read(p)
}

func (s *TCPSocket) Write(p []byte) (n int, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.isConnected == false {
		return 0, gocket.NewErrorFromString(gocket.SocketError, "socket already closed")
	}

	return s.writer.Write(p)
}

func (s *TCPSocket) RemoteAddress() net.Addr {
	return s.conn.RemoteAddr()
}

func (s *TCPSocket) LocalAddress() net.Addr {
	return s.conn.LocalAddr()
}

func (s *TCPSocket) IsConnected() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.isConnected
}

func (s *TCPSocket) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.isConnected == false {
		return gocket.NewErrorFromString(gocket.SocketError, "socket already closed")
	}

	s.isConnected = false
	return s.conn.Close()
}

func (s *TCPSocket) Option() {

}
