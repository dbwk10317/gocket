package gocket

import (
	"github.com/oklog/ulid"
	"github.com/sirupsen/logrus"
	"math/rand"
	"sync"
	"time"
)

type (
	Session interface {
		Unique() string
		Pipeline() *Pipeline
		Send(msg interface{})
		Logger() Logger
		AddEventListener(listener SessionEventListener)
	}

	socketSession struct {
		id                string
		writeSockChan     chan interface{}
		readSockChan      chan interface{}
		sendMsgChan       chan interface{}
		notifyChan        chan interface{}
		failureChan       chan error
		quitWriteSockChan chan bool
		quitReadSockChan  chan bool
		quitPipelineChan  chan bool
		pipeline          *Pipeline
		listenerList      []SessionEventListener
		socket            Socket
		isActive          bool
		config            SessionConfig
		mutex             *sync.RWMutex
		logger            *logrus.Entry
	}
)

var (
	uid = uidGenerator()
)

func uidGenerator() func() string {
	t := time.Now()
	randomness := rand.New(rand.NewSource(t.UnixNano()))
	entropy := ulid.Monotonic(randomness, 0)
	return func() string {
		return ulid.MustNew(ulid.Timestamp(t), entropy).String()
	}
}

func NewSocektSession() Session {
	return NewSocketSessionWithConfig(DefaultSessionConfig())
}

func NewSocketSessionWithConfig(config SessionConfig) Session {
	s := &socketSession{
		id:                uid(),
		mutex:             new(sync.RWMutex),
		isActive:          false,
		config:            config,
		pipeline:          NewPipeline(),
		quitReadSockChan:  make(chan bool, 2),
		quitWriteSockChan: make(chan bool, 2),
		quitPipelineChan:  make(chan bool, 2),
	}

	s.writeSockChan = make(chan interface{}, s.config.MaxChannelBufferSize)
	s.readSockChan = make(chan interface{}, s.config.MaxChannelBufferSize)
	s.failureChan = make(chan error, s.config.MaxChannelBufferSize)
	s.notifyChan = make(chan interface{}, s.config.MaxChannelBufferSize)
	s.sendMsgChan = make(chan interface{}, s.config.MaxChannelBufferSize)

	s.logger = logrus.WithFields(logrus.Fields{
		"sessionID": s.id,
	})

	return s
}

func (s *socketSession) Logger() Logger {
	return s.logger
}

func (s *socketSession) Unique() string {
	return s.id
}

func (s *socketSession) Pipeline() *Pipeline {
	return s.pipeline
}

func (s *socketSession) BindSocket(sock Socket) bool {
	defer s.catchPanic()
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.isActive {
		return false
	}
	if sock.IsConnected() == false {
		return false
	}

	s.socket = sock
	s.isActive = true

	go s.processRead()
	go s.processWrite()

	for _, listener := range s.listenerList {
		if listener != nil {
			listener.OnActivated(s)
		}
	}

	return true
}

func (s *socketSession) ReleaseSocket() {
	defer s.catchPanic()
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.isActive == false {
		return
	}
	if s.socket.IsConnected() {
		if err := s.socket.Close(); err != nil {
			s.failureChan <- err
		}
	}

	s.socket = nil
	s.isActive = false
	s.quitReadSockChan <- true
	s.quitWriteSockChan <- true
	s.quitPipelineChan <- true
	s.flushChannels()

	for _, listener := range s.listenerList {
		if listener != nil {
			listener.OnInactivated(s)
		}
	}
}

func (s *socketSession) Send(msg interface{}) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	s.sendMsgChan <- msg
}

func (s *socketSession) processRead() {
	defer s.catchPanic()
	buffer := make([]byte, s.config.MaxReadBufferSize)
Loop:
	for {
		select {
		case <-s.quitReadSockChan:
			break Loop
		default:
			if s.socket == nil && s.socket.IsConnected() == false {
				continue
			}
			if readCnt, err := s.socket.Read(buffer); err != nil {
				s.failureChan <- NewError(SessionIOError, err)
			} else {
				if readCnt > 0 {
					byteArr := make([]byte, readCnt)
					copy(byteArr, buffer[:readCnt])
					s.readSockChan <- byteArr
				}
			}
		}
	}
}

func (s *socketSession) processWrite() {
	defer s.catchPanic()
Loop:
	for {
		select {
		case <-s.quitWriteSockChan:
			break Loop
		case msg := <-s.writeSockChan:
			if s.socket == nil && s.socket.IsConnected() == false {
				continue
			}

			data, ok := msg.([]byte)
			if ok == false {
				s.failureChan <- NewErrorFromString(SessionIOError, "can't not convert message to byte array")
			}
			length := int32(len(data))
			bufferSize := s.config.MaxWriteBufferSize
			var err error
			for start := int32(0); start < length; start += bufferSize {
				remain := length - start
				if remain < bufferSize {
					_, err = s.socket.Write(data[start:])
				} else {
					end := start + bufferSize
					_, err = s.socket.Write(data[start:end])
				}
			}

			if err != nil {
				s.failureChan <- NewError(SessionIOError, err)
			}
		}

	}
}

func (s *socketSession) processPipeline() {
	defer s.catchPanic()

	logger := s.logger.WithFields(logrus.Fields{
		"event": "receive",
	})
	ctx := HandlerContext{
		logger: logger,
	}
	done := make(chan bool)
	pipe := s.pipeline.InboundPipe(ctx)
	result := pipe(s.readSockChan, done)

	select {
	case <-s.quitReadSockChan:
	case <-result:
	}

}

func (s *socketSession) flushChannels() {
L1:
	for {
		select {
		case <-s.writeSockChan:
		case <-s.readSockChan:
		case <-s.sendMsgChan:
		case <-s.notifyChan:
		case <-s.failureChan:
		case <-s.quitReadSockChan:
		case <-s.quitWriteSockChan:
		case <-s.quitPipelineChan:
		default:
			break L1
		}
	}
}

func (s *socketSession) catchPanic() {
	if err := RecoverPanic(); err != nil {
		if s.failureChan != nil {
			s.failureChan <- err
		} else {
			//
		}
	}
}

func (s *socketSession) IsActive() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.isActive
}

func (s *socketSession) AddEventListener(listener SessionEventListener) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.listenerList = append(s.listenerList, listener)
}
