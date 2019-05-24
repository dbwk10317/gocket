package gocket

import "net"

type (
	SocketEventListener interface {
		OnConnect(socket Socket, localAddr net.Addr, remoteAddr net.Addr)
		OnClose(socket Socket)
		OnError(socket Socket, err error)
	}

	ServerSocketEventListener interface {
		OnBind(socket ServerSocket, localAddr net.Addr)
		OnAccept(serverSocket ServerSocket, socket Socket)
	}

	SessionEventListener interface {
		OnActivated(session Session)
		OnInactivated(session Session)
	}

	FutureListener interface {
		OnDone()
		OnSuccess(result interface{})
		OnFail(err error)
	}

	FutureListenerAdapter struct {
		done    func()
		success func(result interface{})
		fail    func(err error)
	}
)

func (f FutureListenerAdapter) OnDone() {
	if f.done != nil {
		f.done()
	}
}

func (f FutureListenerAdapter) OnSuccess(result interface{}) {
	if f.success != nil {
		f.success(result)
	}
}

func (f FutureListenerAdapter) OnFail(err error) {
	if f.fail != nil {
		f.success(err)
	}
}
