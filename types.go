package main

import (
   "time"
)

type Entry struct {
   when time.Time
   host string
   msg  string
}

