package exchange

import (
	"encoding/csv"
	"errors"
	"io"
	"strings"
)

func NewCSVReader(rc io.ReadCloser) (Reader, error) {
	if rc == nil {
		return nil, errors.New("reader is nil")
	}
	defer rc.Close()

	r := csv.NewReader(rc)
	r.TrimLeadingSpace = true

	header, err := r.Read()
	if err != nil {
		return nil, err
	}
	header = trimStringSliceSpace(header)
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	for i, r := range records {
		records[i] = trimStringSliceSpace(r)
	}

	return &csvReader{
		header:  header,
		records: records,
		curr:    0,
		total:   len(records),
	}, nil
}

type csvReader struct {
	header  []string
	records [][]string
	curr    int
	total   int
}

var _ Reader = (*csvReader)(nil)

func (c *csvReader) Header() []string {
	return c.header
}

func (c *csvReader) ReadRow() ([]string, error) {
	if c.total < c.curr {
		return nil, errors.New("no more row")
	}

	return c.records[c.curr-1], nil
}

func (c *csvReader) Next() bool {
	if c.total <= c.curr {
		return false
	}

	c.curr++
	return true
}

func (c *csvReader) Total() uint {
	return uint(c.total)
}

func trimStringSliceSpace(rs []string) []string {
	nrs := make([]string, 0, len(rs))
	for _, r := range rs {
		nrs = append(nrs, strings.TrimSpace(r))
	}
	return nrs
}

func NewCSVWriter(w io.Writer) (Writer, error) {
	if w == nil {
		return nil, errors.New("writer is nil")
	}

	return &csvWriter{
		w: csv.NewWriter(w),
	}, nil
}

type csvWriter struct {
	w *csv.Writer
}

func (c *csvWriter) WriteHeader(h []string) error {
	return c.w.Write(h)
}

func (c *csvWriter) WriteRow(r []string) error {
	return c.w.Write(r)
}

func (c *csvWriter) Flush() error {
	c.w.Flush()
	return nil
}
