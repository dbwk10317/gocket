package gocket

import (
	"sync"
)

type (
	Pipe     func(in <-chan interface{}, done <-chan bool) (out <-chan interface{})
	Pipeline struct {
		entryPoint  Handler
		handlerList []Handler
		mutex       *sync.RWMutex
	}
)

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

func (p *Pipeline) InboundPipe(ctx HandlerContext) Pipe {
	var pipes []Pipe
	for handler := range p.iterator() {
		pipe := newPipe(ctx, handler.ProcessRead, handler.CatchError)
		pipes = append(pipes, pipe)
	}
	return p.chain(pipes...)
}

func (p *Pipeline) OutboundPipe(ctx HandlerContext) Pipe {
	var pipes []Pipe
	for handler := range p.reverseIterator() {
		pipe := newPipe(ctx, handler.ProcessWrite, handler.CatchError)
		pipes = append(pipes, pipe)
	}
	return p.chain(pipes...)
}

func (p *Pipeline) chain(pipes ...Pipe) Pipe {
	return func(in <-chan interface{}, done <-chan bool) (out <-chan interface{}) {
		ch := in
		for _, pipe := range pipes {
			ch = pipe(ch, done)
		}
		return ch
	}
}

func newPipe(ctx HandlerContext, handleFunc HandlerProcessFunc, errFunc HandlerErrorFunc) Pipe {
	catchPanic := func() {
		if err := RecoverPanic(); err != nil {
			errFunc(ctx, err)
		}
	}

	return func(in <-chan interface{}, done <-chan bool) (out <-chan interface{}) {
		c := make(chan interface{})
		go func() {
			defer catchPanic()
			defer close(c)
			for msg := range in {
				var result interface{}
				if msg != nil {
					result = handleFunc(ctx, msg)
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
