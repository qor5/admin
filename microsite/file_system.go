package microsite

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type FileSystem struct {
	FileName string
	Url      string
}

func (this FileSystem) Value() (driver.Value, error) {
	return json.Marshal(this)
}

func (this *FileSystem) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), this)
	case []byte:
		return json.Unmarshal(v, this)
	default:
		return errors.New("not supported")
	}
}
