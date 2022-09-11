package main

import (
   "sync"
   "time"
)

type LatencyQueue struct {
   deadlineChan chan time.Duration
   process      func([]Entry) error

   lock         sync.Mutex
   pending      []Entry
   deadline     time.Time
}

func New(process func([]Entry) error) *LatencyQueue {
   lq := LatencyQueue{
      deadlineChan: make(chan time.Duration),
      process:      process,
      deadline:     time.Unix(1 << 62, 0),
   }
   go lq.waiter()

   return &lq
}

func (lq *LatencyQueue) dequeue() {
   lq.lock.Lock()
   defer lq.lock.Unlock()

   err := lq.process(lq.pending)
   if err == nil {
      lq.pending = []Entry{}
   }
}

func (lq *LatencyQueue) waiter() {
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

func (lq *LatencyQueue) Enqueue(entry Entry, maxLatency time.Duration) {
   lq.lock.Lock()
   defer lq.lock.Unlock()

   lq.pending = append(lq.pending, entry)

   if maxLatency >= time.Until(lq.deadline) {
      return
   }

   lq.deadline = time.Now().Add(maxLatency)
   lq.deadlineChan <- maxLatency
}
