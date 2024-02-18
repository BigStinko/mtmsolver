package priorityqueue

import (
	"container/heap"
	"sync"
)

type SafePriorityQueue struct {
	pq PriorityQueue
	mux *sync.Mutex
}

func NewSafePQ() SafePriorityQueue {
	return SafePriorityQueue{
		pq: PriorityQueue{},
		mux: &sync.Mutex{},
	}
}

func (pq *SafePriorityQueue) Push(value, priority int) {
	pq.mux.Lock()
	defer pq.mux.Unlock()
	item := Item{
		value: value,
		priority: priority,
	}
	heap.Push(&pq.pq, &item)
}

func (pq *SafePriorityQueue) Pop() int {
	pq.mux.Lock()
	defer pq.mux.Unlock()
	item := heap.Pop(&pq.pq).(*Item)
	return item.value
}

func (pq *SafePriorityQueue) Len() int {
	pq.mux.Lock()
	defer pq.mux.Unlock()
	return pq.pq.Len()
}
