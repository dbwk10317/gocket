package gocket

import (
	"io"
	"net"
)

type (
	SocketType  int
	NetworkType string
	Port        int32
	Host        string

	Socket interface {
		io.Reader
		io.Writer
		Connect(host Host, port Port) error
		RemoteAddress() net.Addr
		LocalAddress() net.Addr
		IsConnected() bool
		Close() error
	}

	ServerSocket interface {
		Socket
		Bind(port Port) error
		Accept() (Socket, error)
		Listenable() bool
	}
)

const (
	TcpSocket SocketType = 0x1000 + iota
)

const (
	TCPNetwork  NetworkType = "tcp"
	TCPNetwork4 NetworkType = "tcp4"
	TCPNetwork6 NetworkType = "tcp6"
)

func (n NetworkType) String() string {
	return string(n)
}
