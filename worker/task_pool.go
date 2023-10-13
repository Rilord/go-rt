package worker

import (
	"context"
	"sync"
	"sync/atomic"
)

type TaskPool struct {
	Pool sync.Pool
}

var (
	countGet uint64
	countPut uint64
	countNew uint64
)

func NewTaskPool() *TaskPool {
	return &TaskPool{
		Pool: sync.Pool{
			New: func() any {
				atomic.AddUint64(&countNew, 1)
				task := new(Task)
				task.stopCh = make(chan interface{}, 1)
				task.ctx, task.cancel = context.WithCancel(context.Background())
				task.SetStateUnsafe(TaskStateNew)
				return task
			},
		},
	}
}

func (p *TaskPool) GetTask() *Task {
	atomic.AddUint64(&countGet, 1)
	task := p.Pool.Get().(*Task)
	if task.state != TaskStateNew {
		task.SetStateUnsafe(TaskStatePoolGet)
	}
	return task
}

func (p *TaskPool) putTask(task *Task) {
	if task.state == TaskStateNew || task.state == TaskStateDoneSuccess || task.state == TaskStatePoolGet {
		atomic.AddUint64(&countPut, 1)
		task.requests = nil
		task.results = nil
		task.SetState(TaskStatePoolPut)
		p.Pool.Put(task)
	}
}
