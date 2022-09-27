package admin

import (
	"crypto/sha256"
	"fmt"
)

type DataTableHeader struct {
	Text     string `json:"text"`
	Value    string `json:"value"`
	Width    string `json:"width"`
	Sortable bool   `json:"sortable"`
}

func getStringHash(v string, len int) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(v)))[:len]
}
