package exchange

type Reader interface {
	Header() []string
	ReadRow() ([]string, error)
	Next() bool
	Total() uint
}

type Writer interface {
	WriteHeader([]string) error
	WriteRow([]string) error
	Flush() error
}
