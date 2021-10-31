package exchange

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/mohae/deepcopy"
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
	if len(ip.pkMetas) > 0 {
		oldRecords := reflect.New(reflect.SliceOf(ip.rtResource)).Interface()
		err = preloadDB(db, ip.associations).Find(oldRecords).Error
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

	records := reflect.New(reflect.SliceOf(ip.rtResource)).Elem()
	for _, metaValues := range allMetaValues {
		var record reflect.Value
		var oldRecord reflect.Value
		{
			pkvs := make([]string, 0, len(ip.pkMetas))
			for _, m := range ip.pkMetas {
				pkvs = append(pkvs, metaValues.Get(m.field))
			}
			cpkvs := strings.Join(pkvs, "$")
			if v, ok := oldRecordsMap[cpkvs]; cpkvs != "" && ok {
				record = reflect.ValueOf(v)
				oldRecord = reflect.ValueOf(deepcopy.Copy(v))
			} else {
				record = reflect.New(ip.rtResource.Elem())
			}
		}
		for _, m := range ip.metas {
			if m.setter != nil {
				err = m.setter(record.Interface(), metaValues.Get(m.field), metaValues)
				if err != nil {
					return err
				}
				continue
			}
			fv := record.Elem().FieldByName(m.field)
			err = setValueFromString(fv, metaValues.Get(m.field))
			if err != nil {
				return err
			}
		}
		if oldRecord.IsValid() {
			for _, a := range ip.associations {
				newV := record.Elem().FieldByName(a)
				oldV := oldRecord.Elem().FieldByName(a)
				if !cmp.Equal(newV.Interface(), oldV.Interface()) {
					if newV.IsZero() {
						err = db.Model(record.Interface()).Association(a).Clear()
					} else {
						ft, _ := ip.rtResource.Elem().FieldByName(a)
						if !strings.Contains(ft.Tag.Get("gorm"), "many2many") {
							ip.clearPrimaryKeyValueForAssociation(newV)
						}
						err = db.Model(record.Interface()).Association(a).Replace(newV.Interface())
					}
					if err != nil {
						return err
					}
				}
				newV.Set(reflect.New(newV.Type()).Elem())
			}
		}
		records = reflect.Append(records, record)
	}

	ocPrimaryCols := make([]clause.Column, 0, len(ip.metas))
	for _, m := range ip.metas {
		if m.primaryKey {
			ocPrimaryCols = append(ocPrimaryCols, clause.Column{
				Name: m.snakeField,
			})
		}
	}
	// .Session(&gorm.Session{FullSaveAssociations: true}) cannot auto delete associations and not work with many-to-many
	return db.Clauses(clause.OnConflict{
		Columns:   ocPrimaryCols,
		UpdateAll: true,
	}).Model(ip.resource).CreateInBatches(records.Interface(), 1000).Error
}

func (ip *Importer) clearPrimaryKeyValueForAssociation(v reflect.Value) {
	rv := getIndirect(v)
	rt := rv.Type()
	switch rt.Kind() {
	case reflect.Struct:
		clearPrimaryKeyValue(rv)
	case reflect.Slice:
		for i := 0; i < rv.Len(); i++ {
			clearPrimaryKeyValue(getIndirect(rv.Index(i)))
		}
	}
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
