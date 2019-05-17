package gocket

import (
	"github.com/sirupsen/logrus"
	"sync"
)

type (
	Bootstrap struct {
		logger *logrus.Logger
		mutex  *sync.RWMutex
	}

	ServerBootstrap struct {
		bootstrap   *Bootstrap
		session     ServerSession
		isListening bool
	}
)

func NewBootstrap() *Bootstrap {
	b := &Bootstrap{
		logger: logrus.New(),
		mutex:  &sync.RWMutex{},
	}

	return b
}

func NewServerBootstrap() *ServerBootstrap {
	b := &ServerBootstrap{
		bootstrap: NewBootstrap(),
	}

	return b
}

func (b *Bootstrap) Logger() Logger {
	return b.logger
}

func (b *ServerBootstrap) Logger() Logger {
	return b.bootstrap.Logger()
}

func (b *ServerBootstrap) Bind(port Port) (Future, error) {
	m := b.bootstrap.mutex
	m.Lock()
	defer m.Unlock()

}
