package gocket

import "sync"

type (
	FutureTask func(param interface{}) (result interface{}, err error)

	Future interface {
		Run(param interface{}) Future
		Sync() Future
		AddFutureListener(listener FutureListener)
	}

	future struct {
		task         FutureTask
		listenerList []FutureListener
		waitGroup    *sync.WaitGroup
		done         <-chan struct{}
	}
)

func NewFuture(task FutureTask) Future {
	return &future{
		task:      task,
		waitGroup: &sync.WaitGroup{},
	}
}

func (f *future) AddFutureListener(listener FutureListener) {
	f.listenerList = append(f.listenerList, listener)
}

func (f *future) Run(param interface{}) Future {
	go func() {
		defer f.waitGroup.Done()
		result, err := f.task(param)
		for _, listener := range f.listenerList {
			if result != nil {
				listener.OnSuccess(result)
			}
			if err != nil {
				listener.OnFail(err)
			}
			listener.OnDone()
		}
	}()
	return f
}

func (f *future) Sync() Future {
	f.waitGroup.Add(1)
	f.waitGroup.Wait()
	return f
}
