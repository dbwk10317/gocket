package gocket

import "sync"

type (
	Pipe     func(in <-chan interface{}) (out <-chan interface{})
	Pipeline struct {
		entryPoint  Handler
		handlerList []Handler
		mutex       *sync.RWMutex
	}

	entryPointHandler struct {
	}
)

func NewPipeline() *Pipeline {
	return NewPipelineWithConfig(DefaultPipelineConfig())
}

func NewPipelineWithConfig(config PipelineConfig) *Pipeline {
	return &Pipeline{
		mutex: new(sync.RWMutex),
	}
}

func (p *Pipeline) AddHandler(handler Handler) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.handlerList = append(p.handlerList, handler)
}

func (p *Pipeline) Inbound() Pipe {

}

func (p *Pipeline) wrap(handler Handler) (inPipe Pipe, outPipe Pipe) {

}

func (p *Pipeline) chain(pipes ...Pipe) Pipe {
	return func(in <-chan interface{}) (out <-chan interface{}) {
		ch := in
		for _, pipe := range pipes {
			ch = pipe(ch)
		}
		return ch
	}
}

func (p *Pipeline) iterator() (next func() Handler) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	cursor := 0
	length := len(p.handlerList)
	return func() Handler {
		var handler Handler
		if cursor < length {
			handler = p.handlerList[cursor]
			cursor++
		}
		return handler
	}
}

func (p *Pipeline) reverseIterator() (next func() Handler) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	cursor := len(p.handlerList) - 1
	return func() Handler {
		var handler Handler
		if cursor >= 0 {
			handler = p.handlerList[cursor]
			cursor--
		}
		return handler
	}
}
