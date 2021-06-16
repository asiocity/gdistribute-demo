package library

import "sync"

type Takeout struct {
	id    int
	name  string
	books []*Book
	lock  sync.Mutex
}
