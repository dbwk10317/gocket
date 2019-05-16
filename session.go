package gocket

type (
	Session struct {
		incoming chan interface{}
		outgoing chan interface{}
		failure  chan interface{}
		socket   Socket
	}
)
