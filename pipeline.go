package gocket

import "sync"

type (
	Pipeline struct {
		entryPoint  Handler
		handlerList []Handler
		inbound     chan interface{}
		outbound    chan interface{}
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
		inbound:  make(chan interface{}, config.MaxChannelBufferSize),
		outbound: make(chan interface{}, config.MaxChannelBufferSize),
		mutex:    new(sync.RWMutex),
	}
}

func (p *Pipeline) AddHandler(handler Handler) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.handlerList = append(p.handlerList, handler)
}
