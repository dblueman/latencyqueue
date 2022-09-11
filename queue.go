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
      process: process,
   }
   go lq.waiter()

   return &lq
}

func (lq *LatencyQueue) waiter() {
   var left time.Duration

   for {
      select {
      case left = <-lq.deadlineChan:
         continue
      case <-time.After(left):
         lq.lock.Lock()
         err := lq.process(lq.pending)
         if err == nil {
            lq.pending = []Entry{}
         }
         lq.lock.Unlock()
      }
   }
}

func (lq *LatencyQueue) Enqueue(entry Entry, maxLatency time.Duration) {
   lq.lock.Lock()
   defer lq.lock.Unlock()

   lq.pending = append(lq.pending, entry)

   if lq.deadline.IsZero() || maxLatency < time.Until(lq.deadline) {
     lq.deadline = time.Now().Add(maxLatency)
     lq.deadlineChan <- maxLatency
   }
}
