package exchange

import (
	"fmt"
	"reflect"
	"time"

	"github.com/spf13/cast"
	"gorm.io/gorm"
)

type Exporter struct {
	resource     interface{}
	rtResource   reflect.Type
	metas        []*Meta
	pkMetas      []*Meta
	associations []string
}

func NewExporter(resource interface{}) *Exporter {
	return &Exporter{
		resource: resource,
	}
}

func (ep *Exporter) Metas(ms ...*Meta) *Exporter {
	ep.metas = ms
	return ep
}

func (ep *Exporter) Associations(ts ...string) *Exporter {
	ep.associations = ts
	return ep
}

func (ep *Exporter) Exec(db *gorm.DB, w Writer) error {
	err := ep.validateAndInit()
	if err != nil {
		return err
	}

	var records reflect.Value
	{
		iRecords := reflect.New(reflect.SliceOf(ep.rtResource)).Interface()
		for _, t := range ep.associations {
			db = db.Preload(t)
		}
		var orderBy string
		for i, m := range ep.pkMetas {
			if i > 0 {
				orderBy += ", "
			}
			orderBy += fmt.Sprintf("%s asc", m.snakeField)
		}
		err = db.Model(ep.resource).
			Order(orderBy).
			Find(iRecords).
			Error
		if err != nil {
			return err
		}
		records = reflect.ValueOf(iRecords).Elem()
	}

	headers := make([]string, 0, len(ep.metas))
	for _, m := range ep.metas {
		headers = append(headers, m.columnHeader)
	}
	err = w.WriteHeader(headers)
	if err != nil {
		return err
	}

	vals := make([]string, len(ep.metas))
	for i := 0; i < records.Len(); i++ {
		record := records.Index(i)
		for i, m := range ep.metas {
			if m.valuer != nil {
				v, err := m.valuer(record.Interface())
				if err != nil {
					return err
				}
				vals[i] = v
				continue
			}
			iv := record.Elem().FieldByName(m.field).Interface()
			switch v := iv.(type) {
			case time.Time:
				vals[i] = v.Format(time.RFC3339Nano)
			case *time.Time:
				if v != nil {
					vals[i] = v.Format(time.RFC3339Nano)
				}
			default:
				vals[i] = cast.ToString(iv)
			}
		}
		err = w.WriteRow(vals)
		if err != nil {
			return err
		}
	}

	return w.Flush()
}

func (ep *Exporter) validateAndInit() error {
	if err := validateResourceAndMetas(ep.resource, ep.metas); err != nil {
		return err
	}

	for i, _ := range ep.metas {
		m := ep.metas[i]
		if m.primaryKey {
			ep.pkMetas = append(ep.pkMetas, m)
		}
	}

	ep.rtResource = reflect.TypeOf(ep.resource)

	return nil
}
