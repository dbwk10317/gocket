package gocket

import (
	"sync"
)

type (
	Pipe     func(in <-chan interface{}, done <-chan struct{}) (out <-chan interface{})
	Pipeline struct {
		entryPoint  Handler
		handlerList []Handler
		mutex       *sync.RWMutex
	}
)

func protectInboundHandle(ctx HandlerContext, handler Handler) func(msg interface{}) interface{} {
	return func(msg interface{}) interface{} {
		defer func() {
			if err := RecoverPanic(); err != nil {
				handler.CatchError(ctx, err)
			}
		}()
		return handler.ProcessRead(ctx, msg)
	}
}

func protectOutboundHandle(ctx HandlerContext, handler Handler) func(msg interface{}) interface{} {
	return func(msg interface{}) interface{} {
		defer func() {
			if err := RecoverPanic(); err != nil {
				handler.CatchError(ctx, err)
			}
		}()
		return handler.ProcessRead(ctx, msg)
	}
}

func makePipe(protectedHandler func(msg interface{}) interface{}) Pipe {
	return func(in <-chan interface{}, done <-chan struct{}) (out <-chan interface{}) {
		c := make(chan interface{})
		go func() {
			defer close(c)
			for msg := range in {
				var result interface{}
				if msg != nil {
					result = protectedHandler(msg)
					if IsSkipper(result) {
						result = msg
					}
				}
				select {
				case c <- result:
				case <-done:
					return
				}
			}
		}()
		return c
	}
}

func NewPipeline() *Pipeline {
	return &Pipeline{
		mutex: new(sync.RWMutex),
	}
}

func (p *Pipeline) AddHandler(handler Handler) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.handlerList = append(p.handlerList, handler)
}

func (p *Pipeline) Inbound(ctx HandlerContext) Pipe {
	var pipes []Pipe
	for handler := range p.iterator() {
		pipe := makePipe(protectInboundHandle(ctx, handler))
		pipes = append(pipes, pipe)
	}
	return p.chain(pipes...)
}

func (p *Pipeline) Outbound(ctx HandlerContext) Pipe {
	var pipes []Pipe
	for handler := range p.reverseIterator() {
		pipe := makePipe(protectOutboundHandle(ctx, handler))
		pipes = append(pipes, pipe)
	}
	return p.chain(pipes...)
}

func (p *Pipeline) chain(pipes ...Pipe) Pipe {
	return func(in <-chan interface{}, done <-chan struct{}) (out <-chan interface{}) {
		ch := in
		for _, pipe := range pipes {
			ch = pipe(ch, done)
		}
		return ch
	}
}

func (p *Pipeline) iterator() <-chan Handler {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	c := make(chan Handler)
	length := len(p.handlerList)
	list := make([]Handler, length)
	copy(list, p.handlerList)
	go func() {
		defer close(c)
		for _, handler := range list {
			c <- handler
		}
	}()

	return c
}

func (p *Pipeline) reverseIterator() <-chan Handler {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	c := make(chan Handler)
	length := len(p.handlerList)
	list := make([]Handler, length)
	copy(list, p.handlerList)
	go func() {
		defer close(c)
		for index := range list {
			c <- list[length-index-1]
		}
	}()

	return c
}
