package concurrent

import (
	"context"
	"errors"
	"sync"
)

type WorkerState uint

const (
	WorkerStateNew WorkerState = iota
	WorkerStateWorking
	WorkerStateIdle
	WorkerStateTerminatingParentCtxClosed
	WorkerStateTerminatingStopSignal
	WorkerStateTerminatingTaskChClosed
	WorkerStateTerminated
	WorkerStateRecoverErr
)

type Worker struct {
	pool      *Pool
	parentCtx context.Context

	id uint

	stopCh chan interface{}
	errCh  chan<- error
	state  WorkerState

	taskQueueCh   <-chan *Task
	taskInProcess *Task
	mx            sync.RWMutex
}

func (wr *Worker) setState(state WorkerState) {
	wr.mx.Lock()
	defer wr.mx.Unlock()
	wr.setStateUnsafe(state)
}

func (wr *Worker) setStateUnsafe(state WorkerState) {
	wr.state = state
}

func (wr *Worker) GetState() WorkerState {
	wr.mx.RLock()
	defer wr.mx.RUnlock()
	return wr.GetStateUnsafe()
}

func (wr *Worker) GetStateUnsafe() WorkerState {
	return wr.state
}

func NewWorker(parentCtx context.Context, pool *Pool, taskQueueCh <-chan *Task, id uint, errCh chan<- error) *Worker {
	var worker = Worker{
		parentCtx:   parentCtx,
		id:          id,
		pool:        pool,
		taskQueueCh: taskQueueCh,
		errCh:       errCh,
	}

	worker.setState(WorkerStateNew)

	return &worker
}

func (w *Worker) Run(wg *sync.WaitGroup) {
	w.mx.Lock()
	defer w.mx.Unlock()

	if w.state == WorkerStateNew || w.state == WorkerStateRecoverErr {
		w.setStateUnsafe(WorkerStateIdle)
	} else {
		w.errCh <- errors.New("Wrong state")
	}

	w.stopCh = make(chan interface{}, 1)

	defer func() {
		if r := recover(); r != nil {
		} else {
			w.setStateUnsafe(WorkerStateTerminated)
		}

		if wg != nil {
			wg.Done()
		}

	}()

	for {
		select {
		case _, ok := <-w.stopCh:
			if ok {
			} else {
			}
		case <-w.parentCtx.Done():
			return
		default:
		}

		select {
		case task, ok := <-w.taskQueueCh:
			if ok {
				if task != nil {
					w.setStateUnsafe(WorkerStateWorking)
					w.taskInProcess = task
					task.Process(w.id)
					w.setStateUnsafe(WorkerStateIdle)
				}
			} else {
				w.setStateUnsafe(WorkerStateTerminatingTaskChClosed)
				return
			}
		case _, ok := <-w.stopCh:
			if ok {
				w.setStateUnsafe(WorkerStateTerminatingStopSignal)
			} else {
			}
		case <-w.parentCtx.Done():
			return
		}
	}

}

func (w *Worker) Stop() {
	if w == nil {
		return
	}

	if w.state == WorkerStateNew || w.state == WorkerStateWorking || w.state == WorkerStateIdle {
		if w.stopCh != nil {
			w.stopCh <- true
			close(w.stopCh)
		}
	}
}
