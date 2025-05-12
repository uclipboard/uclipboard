package msgqueue

import (
	"sync"
)

type Subscriber chan any

type existing struct{}

type MsgQueue struct {
	subscribers map[Subscriber]existing
	lock        sync.RWMutex
}

func NewMsgQueue() *MsgQueue {
	return &MsgQueue{
		subscribers: make(map[Subscriber]existing),
	}
}

func (mq *MsgQueue) Subscribe() Subscriber {
	ch := make(Subscriber)
	mq.lock.Lock()
	mq.subscribers[ch] = existing{}
	mq.lock.Unlock()
	return ch
}

func (mq *MsgQueue) Unsubscribe(sub Subscriber) {
	mq.lock.Lock()
	delete(mq.subscribers, sub)
	close(sub)
	mq.lock.Unlock()
}

func (mq *MsgQueue) Push(msg any) {
	mq.lock.RLock()
	defer mq.lock.RUnlock()
	for sub := range mq.subscribers {
		select {
		case sub <- msg:
		default:
			// Drop message if subscriber is not ready
		}
	}
}
