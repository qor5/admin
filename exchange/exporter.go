package exchange

import (
	"reflect"

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

func (ep *Exporter) Exec(db *gorm.DB, w Writer, opts ...ExporterExecOption) error {
	err := ep.validateAndInit()
	if err != nil {
		return err
	}

	maxParamsPerSQL := ep.parseOptions(opts...)

	records := reflect.New(reflect.SliceOf(ep.rtResource)).Elem()
	{
		// gorm using id to order in FindInBatches
		// var orderBy string
		// for i, m := range ep.pkMetas {
		// 	if i > 0 {
		// 		orderBy += ", "
		// 	}
		// 	orderBy += fmt.Sprintf("%s asc", m.snakeField)
		// }
		chunkRecords := reflect.New(reflect.SliceOf(ep.rtResource)).Interface()
		batchSize := maxParamsPerSQL
		if len(ep.pkMetas) > 0 {
			batchSize /= len(ep.pkMetas)
		}
		err = preloadDB(db, ep.associations).
			Model(ep.resource).
			// Order(orderBy).
			FindInBatches(chunkRecords, batchSize, func(tx *gorm.DB, batch int) error {
				records = reflect.AppendSlice(records, reflect.ValueOf(chunkRecords).Elem())
				return nil
			}).Error
		if err != nil {
			return err
		}
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
			vals[i] = cast.ToString(record.Elem().FieldByName(m.field).Interface())
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

	ep.pkMetas = []*Meta{}
	for i, _ := range ep.metas {
		m := ep.metas[i]
		if m.primaryKey {
			ep.pkMetas = append(ep.pkMetas, m)
		}
	}

	ep.rtResource = reflect.TypeOf(ep.resource)

	return nil
}

func (ep *Exporter) parseOptions(opts ...ExporterExecOption) (
	maxParamsPerSQL int,
) {
	maxParamsPerSQL = 65000
	for _, opt := range opts {
		switch v := opt.(type) {
		case *maxParamsPerSQLOption:
			maxParamsPerSQL = v.v
		}
	}

	return maxParamsPerSQL
}
