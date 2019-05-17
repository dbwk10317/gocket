package gocket

type (
	SocketEventListener interface {
		OnConnect()
		OnClose()
		OnError()
	}

	ServerSocketEventListener interface {
		SocketEventListener
		OnBind()
		OnAccept()
	}

	SessionEventListener interface {
		OnActivated()
		OnInactivated()
	}

	HandlerEventListener interface {
		OnHandleComplete()
		OnHandleError()
	}

	PipelineEventListener interface {
		OnHandlerAdded()
		OnHandlerRemoved()
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
