package ants

import (
	"errors"
	"time"
)

var (
	// errQueueIsFull will be returned when the worker queue is full.
	errQueueIsFull = errors.New("the queue is full")

	// errQueueIsReleased will be returned when trying to insert item to a released worker queue.
	errQueueIsReleased = errors.New("the queue length is zero")
)

type workerArray interface {
	//worker 数量；
	len() int
	//worker 数量是否为 0；
	isEmpty() bool
	//goroutine 任务执行结束后，将相应的 worker 放回workerArray中；
	insert(worker *goWorker) error
	//从workerArray中取出一个 worker；
	detach() *goWorker
	//取出所有的过期 worker；
	retrieveExpiry(duration time.Duration) []*goWorker
	//重置容器。
	reset()
}

type arrayType int

const (
	stackType arrayType = 1 << iota
	loopQueueType
)

func newWorkerArray(aType arrayType, size int) workerArray {
	switch aType {
	case stackType:
		return newWorkerStack(size)
	case loopQueueType:
		return newWorkerLoopQueue(size)
	default:
		return newWorkerStack(size)
	}
}
