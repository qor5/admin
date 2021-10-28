package exchange

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Importer struct {
	resource     interface{}
	rtResource   reflect.Type
	metas        []*Meta
	pkMetas      []*Meta
	associations []string
	validators   []func(metaValues MetaValues) error
}

func NewImporter(resource interface{}) *Importer {
	return &Importer{
		resource: resource,
	}
}

func (ip *Importer) Metas(ms ...*Meta) *Importer {
	ip.metas = ms
	return ip
}

func (ip *Importer) Associations(ts ...string) *Importer {
	ip.associations = ts
	return ip
}

func (ip *Importer) Validators(vs ...func(metaValues MetaValues) error) *Importer {
	ip.validators = vs
	return ip
}

func (ip *Importer) Exec(db *gorm.DB, r Reader) error {
	err := ip.validateAndInit()
	if err != nil {
		return err
	}

	allMetaValues := make([]url.Values, 0, r.Total())
	{
		headerIdxMetas := make(map[int]*Meta)
		header := r.Header()
		for i, _ := range ip.metas {
			m := ip.metas[i]
			for hi, h := range header {
				if h == m.columnHeader {
					headerIdxMetas[hi] = m
					break
				}
			}
		}

		for r.Next() {
			metaValues := make(url.Values)
			row, err := r.ReadRow()
			if err != nil {
				return err
			}

			for i, v := range row {
				m, ok := headerIdxMetas[i]
				if !ok {
					continue
				}
				metaValues.Set(m.field, v)
			}

			for _, vd := range ip.validators {
				err = vd(metaValues)
				if err != nil {
					return err
				}
			}

			allMetaValues = append(allMetaValues, metaValues)
		}
	}

	// primarykeys:record
	oldRecordsMap := make(map[string]interface{})
	{
		oldRecords := reflect.New(reflect.SliceOf(ip.rtResource)).Interface()
		for _, t := range ip.associations {
			db = db.Preload(t)
		}
		err = db.Find(oldRecords).Error
		if err != nil {
			return err
		}
		rv := reflect.ValueOf(oldRecords).Elem()
		for i := 0; i < rv.Len(); i++ {
			record := rv.Index(i)
			pkvs := make([]string, 0, len(ip.pkMetas))
			for _, m := range ip.pkMetas {
				pkvs = append(pkvs, fmt.Sprintf("%s", record.Elem().FieldByName(m.field).Interface()))
			}
			oldRecordsMap[strings.Join(pkvs, "$")] = record.Interface()
		}
	}

	// records := make([]interface{}, 0, r.Total())
	records := reflect.New(reflect.SliceOf(ip.rtResource)).Elem()
	for _, metaValues := range allMetaValues {
		var record interface{}
		pkvs := make([]string, 0, len(ip.pkMetas))
		for _, m := range ip.pkMetas {
			pkvs = append(pkvs, metaValues.Get(m.field))
		}
		cpkvs := strings.Join(pkvs, "$")
		if v, ok := oldRecordsMap[cpkvs]; cpkvs != "" && ok {
			record = v
		} else {
			record = reflect.New(ip.rtResource.Elem()).Interface()
		}
		for _, m := range ip.metas {
			if m.setter != nil {
				err = m.setter(record, metaValues)
				if err != nil {
					return err
				}
				continue
			}
			rfv := reflect.ValueOf(record).Elem().FieldByName(m.field)
			err = setValueFromString(rfv, metaValues.Get(m.field))
			if err != nil {
				return err
			}
		}
		records = reflect.Append(records, reflect.ValueOf(record))
	}

	ocPrimaryCols := make([]clause.Column, 0, len(ip.metas))
	for _, m := range ip.metas {
		if m.primaryKey {
			ocPrimaryCols = append(ocPrimaryCols, clause.Column{
				Name: m.snakeField,
			})
		}
	}
	return db.Clauses(clause.OnConflict{
		Columns:   ocPrimaryCols,
		UpdateAll: true,
	}).Model(ip.resource).CreateInBatches(records.Interface(), 1000).Error
}

func (ip *Importer) validateAndInit() error {
	if err := validateResourceAndMetas(ip.resource, ip.metas); err != nil {
		return err
	}

	for i, _ := range ip.metas {
		m := ip.metas[i]
		if m.primaryKey {
			ip.pkMetas = append(ip.pkMetas, m)
		}
	}

	ip.rtResource = reflect.TypeOf(ip.resource)

	return nil
}
