package worker

import (
	"context"
	"sync"
	"time"
)

type TaskState uint

const (
	TaskStateNew TaskState = iota
	TaskStatePoolGet
	TaskStatePoolPut
	TaskStateReady
	TaskStateInProgress
	TaskStateDoneSuccess
	TaskStateRecoverErr
	TaskStateTerminatedStopSignal
	TaskStateTerminatedCtxClosed
	TaskStateTerminatedTimeout
)

type Task struct {
	parentCtx context.Context
	ctx       context.Context
	cancel    context.CancelFunc
	doneCh    chan<- interface{}
	stopCh    chan interface{}
	wg        *sync.WaitGroup

	prevState TaskState
	state     TaskState

	requests []interface{}
	results  []interface{}

	err error

	duration time.Duration

	f  func(context.Context, context.Context, ...interface{}) (error, []interface{})
	mx sync.RWMutex
}

func (t *Task) SetWgUnsafe(wg *sync.WaitGroup) {
	t.wg = wg
}

func (t *Task) SetStateUnsafe(state TaskState) {
	t.prevState = t.state
	t.state = state
}

func (ts *Task) SetState(state TaskState) {
	ts.mx.Lock()
	defer ts.mx.Unlock()
	ts.SetStateUnsafe(state)
}

func (t *Task) GetError() error {
	return t.err
}

func (t *Task) GetRequests() []interface{} {
	return t.requests
}

func (t *Task) GetResults() []interface{} {
	return t.results
}

func (t *Task) Process(workerId uint) {
	if t == nil || t.f == nil {
		return
	}

	t.mx.Lock()
	defer t.mx.Unlock()

	defer func() {

		if t.doneCh != nil {
			t.doneCh <- struct{}{}
		}

		if t.wg != nil {
			t.wg.Done()
		}
	}()

	t.err, t.results = t.f(t.parentCtx, t.ctx, t.requests...)
	t.SetStateUnsafe(TaskStateDoneSuccess)
	return
}

func NewTask(parentCtx context.Context, tp *TaskPool, name string, doneCh chan<- interface{}, wg *sync.WaitGroup, id uint64, f func(context.Context, context.Context, ...interface{}) (error, []interface{}), requests ...interface{}) *Task {
	if f == nil || parentCtx == nil {
		return nil
	}

	task := tp.GetTask()

	{
		task.parentCtx = parentCtx

		task.doneCh = doneCh
		task.wg = wg

		task.requests = requests
		task.results = nil
		task.err = nil

		task.duration = 0

		task.f = f

		task.SetStateUnsafe(TaskStateReady)
	}

	return task
}

func (t *Task) Stop() {
	if t.state == TaskStateNew || t.state == TaskStateInProgress {
		if t.stopCh != nil {
			t.stopCh <- true
			close(t.stopCh)
		}
	}
}

func (t *Task) DeleteUnsafe(taskPool *TaskPool) {
	if t != nil {
		taskPool.putTask(t)
	}
}
