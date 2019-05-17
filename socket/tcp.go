package socket

import (
	"bufio"
	"fmt"
	"github.com/dbwk10317/gocket"
	"net"
	"sync"
)

type (
	TCPSocket struct {
		conn        *net.TCPConn
		reader      *bufio.Reader
		writer      *bufio.Writer
		isConnected bool
		mutex       *sync.RWMutex
	}

	TCPServerSocket struct {
		listener   *net.TCPListener
		listenable bool
		socket     *TCPSocket
		mutex      *sync.RWMutex
	}
)

func NewTCPAddress(host gocket.Host, port gocket.Port) (*net.TCPAddr, error) {
	address := fmt.Sprintf("%s:%d", host, port)
	addr, err := net.ResolveTCPAddr(gocket.TCPNetwork.String(), address)
	if err != nil {
		return nil, gocket.NewError(gocket.ServerSocketError, err)
	}

	return addr, nil
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

	if conn, err := net.DialTCP(gocket.TCPNetwork.String(), nil, addr); err != nil {
		return gocket.NewError(gocket.SocketError, err)
	} else {
		s.BindConnection(conn)
	}

	return nil
}

func (s TCPSocket) BindConnection(conn *net.TCPConn) {
	s.conn = conn
	s.writer = bufio.NewWriter(conn)
	s.reader = bufio.NewReader(conn)
	s.isConnected = true
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

func NewTCPServerSocket(port gocket.Port) *TCPServerSocket {
	return &TCPServerSocket{
		listenable: false,
		socket:     NewTCPSocket(),
	}
}

func (s *TCPServerSocket) Bind(port gocket.Port) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	addr, err := NewTCPAddress("0.0.0.0", port)
	if err != nil {
		return gocket.NewError(gocket.ServerSocketError, err)
	}

	listener, err := net.ListenTCP(gocket.TCPNetwork.String(), addr)
	if err != nil {
		return gocket.NewError(gocket.ServerSocketError, err)
	}

	s.listener = listener
	s.listenable = true

	return nil
}

func (s *TCPServerSocket) Listenable() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.listenable
}

func (s *TCPServerSocket) Accept() (gocket.Socket, error) {
	if s.listenable == false {
		return nil, gocket.NewErrorFromString(gocket.ServerSocketError, "TCPListener not available")
	}

	conn, err := s.listener.AcceptTCP()
	if err != nil {
		return nil, gocket.NewError(gocket.ServerSocketError, err)
	}

	sock := NewTCPSocket()
	sock.BindConnection(conn)

	return sock, nil
}

func (s *TCPServerSocket) Read(p []byte) (n int, err error) {
	return s.socket.Read(p)
}

func (s *TCPServerSocket) Write(p []byte) (n int, err error) {
	return s.socket.Write(p)
}

func (s *TCPServerSocket) Connect(host gocket.Host, port gocket.Port) error {
	return s.socket.Connect(host, port)
}

func (s *TCPServerSocket) IsConnected() bool {
	return s.socket.IsConnected()
}

func (s *TCPServerSocket) LocalAddress() net.Addr {
	if s.listenable == false {
		return nil
	}
	return s.listener.Addr()
}

func (s *TCPServerSocket) RemoteAddress() net.Addr {
	return s.socket.RemoteAddress()
}

func (s *TCPServerSocket) Close() error {
	if err := s.socket.Close(); err != nil {
		return err
	}
	if s.listenable == false {
		return gocket.NewErrorFromString(gocket.ServerSocketError, "TCPListener not available")
	}
	return s.listener.Close()
}
