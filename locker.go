package main

import "sync"

type Locker struct {
	locks map[int64]bool
	mux   sync.Mutex
}
