package latencyqueue

import (
   "sync"
   "time"
)

type LatencyQueue[T any] struct {
   deadlineChan chan time.Duration
   process      func([]T) error

   lock         sync.Mutex
   pending      []T
   deadline     time.Time
}

func New[T any](process func([]T) error) *LatencyQueue[T] {
   lq := LatencyQueue[T]{
      deadlineChan: make(chan time.Duration),
      process:      process,
      deadline:     time.Unix(1 << 62, 0),
   }
   go lq.waiter()

   return &lq
}

func (lq *LatencyQueue[T]) dequeue() {
   lq.lock.Lock()
   defer lq.lock.Unlock()

   err := lq.process(lq.pending)
   if err == nil {
      lq.pending = lq.pending[:0]
   }
}

func (lq *LatencyQueue[T]) waiter() {
   var left time.Duration

   for {
      select {
      case left = <-lq.deadlineChan:
         continue
      case <-time.After(left):
         lq.dequeue()
      }
   }
}

func (lq *LatencyQueue[T]) Enqueue(entry T, maxLatency time.Duration) {
   lq.lock.Lock()
   defer lq.lock.Unlock()

   lq.pending = append(lq.pending, entry)

   if maxLatency >= time.Until(lq.deadline) {
      return
   }

   lq.deadline = time.Now().Add(maxLatency)
   lq.deadlineChan <- maxLatency
}
