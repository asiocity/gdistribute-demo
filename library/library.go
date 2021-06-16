package library

import (
	"fmt"
	"sync"
	"sync/atomic"
)

var library *Library

type Library struct {
	books       []*Book    `json:"books"`        // 书籍
	bookSize    *uint64    `json:"book_size"`    // 书籍数量
	takeouts    []*Takeout `json:"takeouts"`     // 借书证
	takeoutSize *uint64    `json:"takeout_size"` // 借书证数量
	lock        *sync.RWMutex
}

func NewLibrary() *Library {
	library := &Library{
		bookSize:    new(uint64),
		takeoutSize: new(uint64),
		lock:        new(sync.RWMutex),
	}
	return library
}

func (l *Library) AddBook(title, author string, catagory []Type) {
	l.lock.Lock()
	defer l.lock.Unlock()

	book := &Book{
		ID:       atomic.AddUint64(l.bookSize, 1),
		Title:    title,
		Author:   author,
		Category: catagory,
	}
	l.books = append(l.books, book)
}

func (l *Library) RemoveBook(book *Book) error {
	l.lock.Lock()
	defer l.lock.Unlock()

	for i := range l.books {
		if book == l.books[i] {
			l.books = append(l.books[:i], l.books[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("Book not in this library")
}

func (l *Library) Resign(name string) *Takeout {
	l.lock.Lock()
	defer l.lock.Unlock()

	takeout := &Takeout{
		id:   int(atomic.AddUint64(l.takeoutSize, 1)),
		name: name,
	}
	l.takeouts = append(l.takeouts, takeout)

	return takeout
}

func (l *Library) Borrow(title string, takeout *Takeout) error {
	takeout.lock.Lock()
	defer takeout.lock.Unlock()
	if len(takeout.books) > 3 {
		return fmt.Errorf("borrow too many books")
	}

	l.lock.Lock()
	defer l.lock.Unlock()
	for i := range l.books {
		if title == l.books[i].Title {
			takeout.books = append(takeout.books, l.books[i])
			return nil
		}
	}
	return fmt.Errorf("book %v not exist", title)
}

func (l *Library) Return(book *Book, takeout *Takeout) error {
	takeout.lock.Lock()
	defer takeout.lock.Unlock()

	for i := range takeout.books {
		if takeout.books[i] == book {
			takeout.books = append(takeout.books[:i], takeout.books[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("not borrow this book")
}

func (l *Library) GetBookByID(id uint64) *Book {
	l.lock.Lock()
	defer l.lock.Unlock()

	for i := range l.books {
		if l.books[i].ID == id {
			return l.books[i]
		}
	}
	return nil
}
