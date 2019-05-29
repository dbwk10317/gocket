package gocket

import (
	"github.com/oklog/ulid"
	"math/rand"
	"sync"
	"time"
)

type (
	Session interface {
		Unique() string
		Pipeline() *Pipeline
		Send(msg interface{})
		AddEventListener(listener SessionEventListener)
	}

	socketSession struct {
		id                string
		writeSockChan     chan []byte
		readSockChan      chan []byte
		sendMsgChan       chan interface{}
		notifyChan        chan interface{}
		failureChan       chan error
		quitWriteSockChan chan bool
		quitReadSockChan  chan bool
		pipeline          *Pipeline
		listenerList      []SessionEventListener
		socket            Socket
		isActive          bool
		config            SessionConfig
		mutex             *sync.RWMutex
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

func catchPanic(errChan chan error) {
	if err := RecoverPanic(); err != nil {
		errChan <- err
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
	}

	s.writeSockChan = make(chan []byte, s.config.MaxChannelBufferSize)
	s.readSockChan = make(chan []byte, s.config.MaxChannelBufferSize)
	s.failureChan = make(chan error, s.config.MaxChannelBufferSize)
	s.notifyChan = make(chan interface{}, s.config.MaxChannelBufferSize)
	s.sendMsgChan = make(chan interface{}, s.config.MaxChannelBufferSize)

	return s
}

func (s *socketSession) Unique() string {
	return s.id
}

func (s *socketSession) Pipeline() *Pipeline {
	return s.pipeline
}

func (s *socketSession) BindSocket(sock Socket) bool {
	defer catchPanic(s.failureChan)
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
	defer catchPanic(s.failureChan)
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
	defer catchPanic(s.failureChan)
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
	defer catchPanic(s.failureChan)
Loop:
	for {
		select {
		case <-s.quitWriteSockChan:
			break Loop
		case data := <-s.writeSockChan:
			if s.socket == nil && s.socket.IsConnected() == false {
				continue
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
	defer catchPanic(s.failureChan)

Loop:
	for {
		select {
		case readSockData := <-s.readSockChan:
		case sendMsgData := <-s.sendMsgChan:
		}
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
		default:
			break L1
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
