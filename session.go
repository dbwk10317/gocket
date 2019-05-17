package gocket

import (
	"sync"
)

type (
	Session interface {
		InputChannel() chan<- interface{}
		OutputChannel() <-chan interface{}
		ErrorChannel() <-chan error
	}

	ServerSession interface {
	}

	session struct {
		incoming     chan interface{}
		outgoing     chan interface{}
		failure      chan error
		notification chan interface{}
		socket       Socket
		isBindSocket bool
		isActive     bool
		mutex        *sync.RWMutex
		done         chan bool
	}

	serverSession struct {
		incoming     chan interface{}
		outgoing     chan interface{}
		failure      chan error
		notification chan interface{}
		socket       ServerSocket
		isActive     bool
		done         chan bool
	}
)

func NewSession() *session {
	s := &session{
		mutex: &sync.RWMutex{},
	}
	return s
}

func (s *session) BindSocket(sock Socket) {

}

func (s *session) ReleaseSocket() {

}

func (s *session) InputChannel() chan<- interface{} {

}

func (s *session) OutputChannel() <-chan interface{} {

}

func (s *session) ErrorChannel() <-chan error {

}

func (s *session) IsActive() bool {

}

func (s *session) processIO() {

}
