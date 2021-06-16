package library

func init() {
	library = NewLibrary()
	var types []Type
	types = append(types, TypeNovel)

	library.AddBook("Hamlet", "William Shakespeare", types)
	library.AddBook("Othello", "William Shakespeare", types)
}
