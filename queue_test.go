package latencyqueue

import (
   "fmt"
   "testing"
   "time"
)

type Entry struct {
   when time.Time
   host string
   msg  string
}

var (
   start time.Time
)

func process(entries []Entry) error {
   for _, entry := range entries {
      fmt.Printf("process %v %v\n", entry, time.Since(start))
   }

   return nil
}

func TestEnqueue(t *testing.T) {
   lq := New(process)
   host := "test"

   lq.Enqueue(Entry{time.Now(), host, "30s max-latency message"}, 30 * time.Second)
   time.Sleep(100 * time.Millisecond)
   start = time.Now()
   lq.Enqueue(Entry{time.Now(), host, "1s max-latency message"}, 1 * time.Second)
   time.Sleep(2 * time.Second)
}
