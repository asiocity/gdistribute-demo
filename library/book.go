package library

type Book struct {
	ID       uint64 `json:"id"`
	Title    string `json:"title"`
	Author   string `json:"author"`
	Category []Type `json:"category"`
}
