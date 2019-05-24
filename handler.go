package gocket

//noinspection SpellCheckingInspection
type (
	Handler interface {
		CatchError(ctx HandlerContext, err error)
	}

	InboundHandler interface {
		Handler
		ReadMessage(ctx HandlerContext, msg interface{})
	}

	OutboundHandler interface {
		Handler
		WriteMessage(ctx HandlerContext, msg interface{})
	}

	HandlerInvoker interface {
		NextReadMessage(msg interface{})
		NextWriteMessage(msg interface{})
	}

	HandlerInitializer interface {
		Initialize(session Session)
	}

	InboundHandlerAdapter struct {
		catch func(ctx HandlerContext, err error)
	}

	HandlerContext struct {
		session      Session
		attributeMap map[string]interface{}
		handler      Handler
		logger       Logger
	}
)

func (c HandlerContext) Session() Session {
	return c.session
}

func (c HandlerContext) Attribute(key string) interface{} {
	return c.attributeMap[key]
}

func (c HandlerContext) Send(msg interface{}) {
	c.session.Send(msg)
}

func (c HandlerContext) Throw(err error) {
	c.handler.CatchError(c, err)
}

func (c HandlerContext) NextReadMessage(msg interface{}) {

}
