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

	SocketListener interface {
		Bind(port Port) error
		Accept() (Socket, error)
		LocalAddress() net.Addr
		Close() error
	}

	Socket interface {
		io.Reader
		io.Writer
		Connect(host Host, port Port) error
		RemoteAddress() net.Addr
		LocalAddress() net.Addr
		IsConnected() bool
		Close() error
		Option()
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
