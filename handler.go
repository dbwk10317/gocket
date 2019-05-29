package gocket

//noinspection SpellCheckingInspection
type (
	Skipper interface{}

	Handler interface {
		CatchError(ctx HandlerContext, err error)
		ProcessRead(ctx HandlerContext, msg interface{}) (result interface{})
		ProcessWrite(ctx HandlerContext, msg interface{}) (result interface{})
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

func IsSkipper(msg interface{}) bool {
	if msg != nil {
		switch msg.(type) {
		case Skipper:
			return true
		default:
			return false
		}
	}
	return false
}

func (c HandlerContext) Session() Session {
	return c.session
}

func (c HandlerContext) Pipeline() *Pipeline {
	return c.session.Pipeline()
}

func (c HandlerContext) Attribute(key string) interface{} {
	return c.attributeMap[key]
}

func (c HandlerContext) SetAttribute(key string, value interface{}) {
	c.attributeMap[key] = value
}

func (c HandlerContext) Send(msg interface{}) {
	c.session.Send(msg)
}

func (c HandlerContext) Throw(err error) {
	c.handler.CatchError(c, err)
}
