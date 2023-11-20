package concurrent

import (
	"context"
	"errors"
	"sync"
)

type WorkerPool struct {
	ctx    context.Context
	cancel context.CancelFunc
	errCh  chan<- error
	stopCh chan struct{}
	name   string

	cfg  *PoolConfig
	pool *Pool
}

func NewWorkerPool(parentCtx context.Context, name string, errCh chan<- error, cfg *PoolConfig) (*WorkerPool, error) {

	var err error

	workerPool := &WorkerPool{
		name:   name,
		errCh:  errCh,
		cfg:    cfg,
		stopCh: make(chan struct{}, 1),
	}

	if parentCtx == nil {
		workerPool.ctx, workerPool.cancel = context.WithCancel(context.Background())
	} else {
		workerPool.ctx, workerPool.cancel = context.WithCancel(parentCtx)
	}

	workerPool.pool, err = NewPool(parentCtx, 1, workerPool.name, cfg)

	if err != nil {
		return nil, err
	}

	return workerPool, nil
}

func (wp *WorkerPool) AddTask(t *Task) error {
	if t != nil {
		err := wp.pool.AddTask(t)
		if err != nil {
			return errors.New("")
		}
		return nil
	}
	return errors.New("")
}

func (wp *WorkerPool) RunTasksInWG(tasks []*Task, taskGroupName string) error {
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	var wg sync.WaitGroup

	for _, task := range tasks {
		if task != nil {
			task.SetWgUnsafe(&wg)

			if err := wp.AddTask(task); err != nil {
				return err
			}
		}

	}

	wg.Wait()

	return nil
}

func (wp *WorkerPool) Run() error {
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	return wp.pool.RunBackGround(0)

}

func (wp *WorkerPool) Shutdown(hardShutdown bool) error {
	defer wp.cancel()

	err := wp.pool.Stop(PoolShutdownHard)

	return err
}
