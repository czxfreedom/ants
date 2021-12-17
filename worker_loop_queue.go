package ants

import "time"

//循环队列
type loopQueue struct {
	items  []*goWorker
	expiry []*goWorker
	head   int
	tail   int
	size   int
	isFull bool
}

func newWorkerLoopQueue(size int) *loopQueue {
	return &loopQueue{
		items: make([]*goWorker, size),
		size:  size,
	}
}

func (wq *loopQueue) len() int {
	if wq.size == 0 {
		return 0
	}
	if wq.head == wq.tail {
		if wq.isFull {
			return wq.size
		}
		return 0
	}

	if wq.tail > wq.head {
		return wq.tail - wq.head
	}

	return wq.size - wq.head + wq.tail
}

func (wq *loopQueue) isEmpty() bool {
	return wq.head == wq.tail && !wq.isFull
}

//goroutine完成任务后,会将对应的worker对象放回loopQueue
func (wq *loopQueue) insert(worker *goWorker) error {
	//队列为空
	if wq.size == 0 {
		return errQueueIsReleased
	}
	//队列满
	if wq.isFull {
		return errQueueIsFull
	}

	wq.items[wq.tail] = worker
	wq.tail++
	//循环队列,所以这边指向开始处
	if wq.tail == wq.size {
		wq.tail = 0
	}
	//俩个相等的时候说明队列满了
	if wq.tail == wq.head {
		wq.isFull = true
	}

	return nil
}

//新任务到来调用loopQueeue.detach()方法获取一个空闲的 worker 结构
func (wq *loopQueue) detach() *goWorker {
	if wq.isEmpty() {
		return nil
	}

	w := wq.items[wq.head]
	wq.items[wq.head] = nil
	wq.head++
	if wq.head == wq.size {
		wq.head = 0
	}
	//每次出队后，队列肯定不满了，isFull要重置为false
	wq.isFull = false

	return w
}

func (wq *loopQueue) retrieveExpiry(duration time.Duration) []*goWorker {
	if wq.isEmpty() {
		return nil
	}

	wq.expiry = wq.expiry[:0]
	expiryTime := time.Now().Add(-duration)

	for !wq.isEmpty() {
		if expiryTime.Before(wq.items[wq.head].recycleTime) {
			break
		}
		wq.expiry = append(wq.expiry, wq.items[wq.head])
		wq.items[wq.head] = nil
		wq.head++
		if wq.head == wq.size {
			wq.head = 0
		}
		wq.isFull = false
	}

	return wq.expiry
}

func (wq *loopQueue) reset() {
	if wq.isEmpty() {
		return
	}

Releasing:
	if w := wq.detach(); w != nil {
		w.task <- nil
		goto Releasing
	}
	wq.items = wq.items[:0]
	wq.size = 0
	wq.head = 0
	wq.tail = 0
}
