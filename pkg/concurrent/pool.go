package concurrent

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"
)

type PoolState uint

const (
	PoolStateNew PoolState = iota
	PoolStateOnlineRunning
	PoolStateOnlineDone
	PoolStateIncompleteDone
	PoolStateRecoverErr
	PoolStateBgRunning
	PoolStateShuttingDown
	PoolStateTerminate
	PoolStateShutdown
)

type PoolConfig struct {
	TaskQueueSize     int
	WorkerConcurrency int
}

type Pool struct {
	parentCtx context.Context
	ctx       context.Context
	cancel    context.CancelFunc

	state             PoolState
	stopCh            chan interface{}
	workers           map[int]*Worker
	workerErrCh       chan error
	workerConcurrency int
	isBackground      bool

	taskQueueCh   chan *Task
	taskQueueSize int

	mx sync.RWMutex
}

func (p *Pool) setState(state PoolState) {
	p.mx.Lock()
	defer p.mx.Unlock()
	p.setStateUnsafe(state)
}

func (p *Pool) GetState() PoolState {
	p.mx.RLock()
	defer p.mx.RUnlock()
	return p.getStateUnsafe()
}

func (p *Pool) getStateUnsafe() PoolState {
	return p.state
}

func (p *Pool) setStateUnsafe(state PoolState) {
	p.state = state
}

func (p *Pool) RunBackGround(extId uint) (err error) {
	if p == nil {
		return errors.New("pool is nil")
	}

	p.mx.Lock()

	if p.state == PoolStateNew {
		p.setStateUnsafe(PoolStateBgRunning)
		p.isBackground = true
	} else {
		p.mx.Unlock()
		return errors.New("Pool incorrect state")
	}

	p.ctx, p.cancel = context.WithCancel(context.Background())
	p.workers = make(map[int]*Worker, p.workerConcurrency)
	p.workerErrCh = make(chan error, p.workerConcurrency) // достаточно по одной ошибке на worker
	p.taskQueueCh = make(chan *Task, p.taskQueueSize)     // Канал-очередь task
	p.stopCh = make(chan interface{}, 1)

	defer func() {
		if r := recover(); r != nil {
			p.mx.Lock()
			defer p.mx.Unlock()
			p.setStateUnsafe(PoolStateRecoverErr)
			_ = p.shutdownUnsafe(PoolShutdownHard)
		}

		if p.cancel != nil {
			p.cancel()
		}
	}()

	for workerId := 1; workerId <= p.workerConcurrency; workerId++ {
		worker := NewWorker(p.ctx, p, p.taskQueueCh, uint(workerId), p.workerErrCh)

		p.workers[workerId] = worker

		go worker.Run(nil)
	}

	p.mx.Unlock()

	for {
		select {
		case _, ok := <-p.workerErrCh:
			if ok {
			} else {
				return nil
			}
		case <-p.stopCh:
			return nil
		case <-p.parentCtx.Done():
			p.mx.Lock()
			err = p.shutdownUnsafe(PoolShutdownHard)
			p.mx.Unlock()
			return err
		}
	}
}

func NewPool(parentCtx context.Context, externalId uint64, name string, cfg *PoolConfig) (*Pool, error) {

	if parentCtx == nil {
		return nil, errors.New("parent ctx is Nil")
	}

	var pool = Pool{
		parentCtx:         parentCtx,
		workerConcurrency: cfg.WorkerConcurrency,
		taskQueueSize:     cfg.TaskQueueSize,
	}

	pool.setState(PoolStateNew)

	if pool.workerConcurrency == 0 {
		pool.workerConcurrency = runtime.NumCPU() * 2
	}

	if pool.taskQueueSize == 0 {
		pool.taskQueueSize = pool.workerConcurrency * 1000
	}

	return &pool, nil
}

func (p *Pool) AddTask(task *Task) (err error) {
	if p == nil {
		return errors.New("Nil pool pointer")
	}

	if task == nil {
		return nil
	}

	p.mx.RLock()

	if p.state != PoolStateBgRunning {
		err = errors.New(fmt.Sprintf("Incorrect state: %d", p.state))
		p.mx.RUnlock()
		return err
	}

	p.mx.RUnlock()

	defer func() {
		if r := recover(); r != nil {
		}
	}()

	if task.wg != nil {
		task.wg.Add(1)
	}
	p.taskQueueCh <- task
	return nil
}

type PoolShutdownMode uint

const (
	PoolShutDownLight PoolShutdownMode = iota
	PoolShutdownSoft
	PoolShutdownHard
)

func (p *Pool) shutdownUnsafe(shutdownMode PoolShutdownMode) (err error) {

	if p.state != PoolStateShutdown && p.state != PoolStateShuttingDown {
		defer func() {
			if r := recover(); r != nil {
			}

			p.setStateUnsafe(PoolStateShutdown)
		}()

		p.setStateUnsafe(PoolStateShuttingDown)
		if p.isBackground {
			close(p.taskQueueCh)
		}

		if shutdownMode == PoolShutdownHard || shutdownMode == PoolShutdownSoft {
			for task := range p.taskQueueCh {
				if task != nil {
					task.Stop()
				}
			}
		}

		p.stopWorkersUnsafe(shutdownMode)

		close(p.workerErrCh)
		if len(p.workerErrCh) != 0 {
		} else {
		}
	}

	return err
}

func (p *Pool) stopWorkersUnsafe(shutdownMode PoolShutdownMode) {

	for _, worker := range p.workers {
		worker.Stop()
	}

	for {

		anyNonStoppedWorker := false
		for _, worker := range p.workers {
			if worker.state != WorkerStateTerminated {
				anyNonStoppedWorker = true
			}
		}

		if !anyNonStoppedWorker {
			return
		}

		time.Sleep(time.Millisecond + 10)
	}
}

func (p *Pool) Stop(shutdownMode PoolShutdownMode) error {
	if p == nil {
		return errors.New("")
	}

	p.mx.Lock()
	defer p.mx.Unlock()

	if p.state != PoolStateShutdown && p.state != PoolStateShuttingDown {
		err := p.shutdownUnsafe(shutdownMode)

		if p.stopCh != nil {
			p.stopCh <- true
			close(p.stopCh)
		}
		return err
	}

	return nil
}
