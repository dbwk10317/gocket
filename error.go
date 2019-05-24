package gocket

import (
	"errors"
	"net"
)

type (
	ErrorCode int

	Error struct {
		code ErrorCode
		err  error
	}
)

func NewError(code ErrorCode, err error) Error {
	return Error{
		code: code,
		err:  err,
	}
}

func NewErrorFromString(code ErrorCode, errMsg string) Error {
	err := errors.New(errMsg)
	return NewError(code, err)
}

func CastErrorFromBuiltin(err error) Error {
	if e, ok := err.(Error); ok {
		return e
	}

	return NewError(BuiltinError, err)
}

func (e Error) Error() string {
	return e.err.Error()
}

func (e Error) Code() ErrorCode {
	return e.code
}

func (e Error) Timeout() bool {
	if err, ok := e.err.(net.Error); ok {
		return err.Timeout()
	}
	return false
}

func (e Error) Temporary() bool {
	if err, ok := e.err.(net.Error); ok {
		return err.Temporary()
	}
	return false
}

const (
	BuiltinError ErrorCode = 0x0100 + iota
	ServerSocketError
	SocketAddressError
	SocketError
	SessionIOError
	InvalidTypeError
	PanicError
)
