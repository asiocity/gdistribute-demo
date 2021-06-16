package library

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func RegisterHandlers() {
	handler := new(libraryHandler)
	// 表示多个资源
	http.Handle("/library", handler)
	// 表示后面还有参数, 对应一个资源
	http.Handle("/library/", handler)
}

type libraryHandler struct{}

// /library 获取图书馆信息
// /library/book 获取全部书籍信息
// /library/book/{id} 获取书籍信息
// /library/takeout 获取全部借书证信息
// /library/takeout/{id} 获取借书证信息
func (lh libraryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pathSegments := strings.Split(r.URL.Path, "/")
	// 暂时不支持 takeout 查询
	switch len(pathSegments) {
	case 2:
		lh.getLibraryStatus(w, r)
	case 3:
		lh.getBooksStatus(w, r)
	case 4:
		id, err := strconv.Atoi(pathSegments[3])
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		lh.getBook(w, r, id)
	default:
		w.WriteHeader(http.StatusForbidden)
	}
}

func (lh *libraryHandler) getLibraryStatus(w http.ResponseWriter, r *http.Request) {
	data, err := lh.toJSON(library)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(data)
}

func (lh *libraryHandler) toJSON(obj interface{}) ([]byte, error) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	if err := enc.Encode(obj); err != nil {
		return nil, fmt.Errorf("Failed to serialize students: %v", err)
	}
	return b.Bytes(), nil
}

func (lh *libraryHandler) getBooksStatus(w http.ResponseWriter, r *http.Request) {
	data, err := lh.toJSON(library.books)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(data)
}

func (lh *libraryHandler) getBook(w http.ResponseWriter, r *http.Request, id int) {
	book := library.GetBookByID(uint64(id))
	if book != nil {
		data, err := lh.toJSON(book)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.Write(data)
	}
}
