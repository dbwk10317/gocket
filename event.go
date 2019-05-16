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
)
