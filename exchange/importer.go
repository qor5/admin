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
	resource           interface{}
	resourceParams     int
	rtResource         reflect.Type
	metas              []*Meta
	pkMetas            []*Meta
	associations       []string
	associationsParams []int
	validators         []func(metaValues MetaValues) error
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

func (ip *Importer) Exec(db *gorm.DB, r Reader, opts ...ImporterExecOption) error {
	err := ip.validateAndInit()
	if err != nil {
		return err
	}

	maxParamsPerSQL := ip.parseOptions(opts...)

	fullPrimaryKeyValues := make([][]string, 0, r.Total())
	allMetaValues := make([]url.Values, 0, r.Total())
	{
		headerIdxMetas := make(map[int]*Meta)
		header := r.Header()
		for i, _ := range ip.metas {
			m := ip.metas[i]
			hasCol := false
			for hi, h := range header {
				if h == m.columnHeader {
					hasCol = true
					headerIdxMetas[hi] = m
					break
				}
			}
			if !hasCol {
				return fmt.Errorf("column %s not found", m.columnHeader)
			}
		}

		for r.Next() {
			metaValues := make(url.Values)
			row, err := r.ReadRow()
			if err != nil {
				return err
			}

			notEmptyPrimaryKeyValues := make([]string, 0, len(ip.pkMetas))
			for i, v := range row {
				m, ok := headerIdxMetas[i]
				if !ok {
					continue
				}
				metaValues.Set(m.field, v)

				if m.primaryKey && v != "" && m.setter == nil {
					notEmptyPrimaryKeyValues = append(notEmptyPrimaryKeyValues, v)
				}
			}
			if len(ip.pkMetas) > 0 && len(notEmptyPrimaryKeyValues) == len(ip.pkMetas) {
				fullPrimaryKeyValues = append(fullPrimaryKeyValues, notEmptyPrimaryKeyValues)
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
	if len(fullPrimaryKeyValues) > 0 {
		oldRecords := reflect.New(reflect.SliceOf(ip.rtResource)).Elem()
		tx := preloadDB(db, ip.associations)
		searchInKeys := false
		{
			var total int64
			err = db.Model(ip.resource).Count(&total).Error
			if err != nil {
				return err
			}
			if int64(len(fullPrimaryKeyValues))*100 < total {
				searchInKeys = true
			}
		}
		if searchInKeys {
			pkvsGroups := splitStringSliceSlice(fullPrimaryKeyValues, maxParamsPerSQL/len(ip.pkMetas))
			for _, g := range pkvsGroups {
				if len(ip.pkMetas) == 1 {
					vs := make([]string, 0, len(g))
					for _, pkvs := range g {
						vs = append(vs, pkvs[0])
					}
					tx = tx.Where(fmt.Sprintf("%s in (?)", ip.pkMetas[0].snakeField), vs)
				} else {
					var pks []string
					for _, m := range ip.pkMetas {
						pks = append(pks, m.snakeField)
					}
					// only test this on Postgres, not sure if this is valid for other databases
					tx = tx.Where(fmt.Sprintf("(%s) in (?)", strings.Join(pks, ",")), g)
				}

				chunkRecords := reflect.New(reflect.SliceOf(ip.rtResource)).Interface()
				err = tx.Find(chunkRecords).Error
				oldRecords = reflect.AppendSlice(oldRecords, reflect.ValueOf(chunkRecords).Elem())
			}
		} else {
			chunkRecords := reflect.New(reflect.SliceOf(ip.rtResource)).Interface()
			err = tx.FindInBatches(chunkRecords, maxParamsPerSQL/len(ip.pkMetas), func(tx *gorm.DB, batch int) error {
				oldRecords = reflect.AppendSlice(oldRecords, reflect.ValueOf(chunkRecords).Elem())
				return nil
			}).Error
		}
		if err != nil {
			return err
		}
		for i := 0; i < oldRecords.Len(); i++ {
			record := oldRecords.Index(i)
			pkvs := make([]string, 0, len(ip.pkMetas))
			for _, m := range ip.pkMetas {
				pkvs = append(pkvs, fmt.Sprintf("%v", record.Elem().FieldByName(m.field).Interface()))
			}
			oldRecordsMap[strings.Join(pkvs, "$")] = record.Interface()
		}
	}

	records := reflect.New(reflect.SliceOf(ip.rtResource)).Elem()
	// key is association
	recordsToClearAssociations := make(map[string]reflect.Value)
	// key is association
	recordsToReplaceAssociations := make(map[string]reflect.Value)
	for _, a := range ip.associations {
		recordsToClearAssociations[a] = reflect.New(reflect.SliceOf(ip.rtResource)).Elem()
		recordsToReplaceAssociations[a] = reflect.New(reflect.SliceOf(ip.rtResource)).Elem()
	}
	// key is association
	toReplaceAssociations := make(map[string][]interface{})
	maxAssociationsRecordsLen := make(map[string]int)
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
						recordsToClearAssociations[a] = reflect.Append(recordsToClearAssociations[a], record)
					} else {
						ft, _ := ip.rtResource.Elem().FieldByName(a)
						if !strings.Contains(ft.Tag.Get("gorm"), "many2many") {
							ip.clearPrimaryKeyValueForAssociation(newV)
						}
						recordsToReplaceAssociations[a] = reflect.Append(recordsToReplaceAssociations[a], record)
						iNewV := newV.Interface()
						if newV.Kind() == reflect.Struct {
							v := reflect.New(newV.Type()).Elem()
							v.Set(reflect.ValueOf(deepcopy.Copy(iNewV)))
							iNewV = v.Addr().Interface()
						}
						toReplaceAssociations[a] = append(toReplaceAssociations[a], iNewV)
					}
					if err != nil {
						return err
					}
				}
				oldV.Set(reflect.New(oldV.Type()).Elem())
				newV.Set(reflect.New(newV.Type()).Elem())
			}
			if cmp.Equal(record.Interface(), oldRecord.Interface()) {
				continue
			}
		}
		for _, a := range ip.associations {
			afv := record.Elem().FieldByName(a)
			if afv.Type().Kind() == reflect.Slice {
				maxAssociationsRecordsLen[a] = afv.Len()
			}
		}
		records = reflect.Append(records, record)
	}

	batchSize := 10000
	{
		max := maxParamsPerSQL / ip.resourceParams
		for i, a := range ip.associations {
			l, ok := maxAssociationsRecordsLen[a]
			if ok {
				if l > 0 {
					if v := maxParamsPerSQL / (l * ip.associationsParams[i]); max > v {
						max = v
					}
				}
			} else {
				if v := maxParamsPerSQL / ip.associationsParams[i]; max > v {
					max = v
				}
			}
		}
		if batchSize > max {
			batchSize = max
		}
	}
	return db.Transaction(func(tx *gorm.DB) error {
		for _, a := range ip.associations {
			if recordsToClearAssociations[a].Len() > 0 {
				rgs := splitReflectSliceValue(recordsToClearAssociations[a], maxParamsPerSQL/len(ip.pkMetas))
				for _, g := range rgs {
					err = db.Model(g.Interface()).Association(a).Clear()
					if err != nil {
						return err
					}
				}
			}
			if recordsToReplaceAssociations[a].Len() > 0 {
				// TODO: limit batch size
				// TODO: it seems not updated in batch from the gorm log
				err = db.Model(recordsToReplaceAssociations[a].Interface()).Association(a).Replace(toReplaceAssociations[a]...)
				if err != nil {
					return err
				}
			}
		}

		if records.Len() == 0 {
			return nil
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
		}).Model(ip.resource).CreateInBatches(records.Interface(), batchSize).Error
	})
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

	ip.pkMetas = []*Meta{}
	for i, _ := range ip.metas {
		m := ip.metas[i]
		if m.primaryKey {
			ip.pkMetas = append(ip.pkMetas, m)
		}
	}

	ip.rtResource = reflect.TypeOf(ip.resource)
	getParamsNumbers(&ip.resourceParams, ip.rtResource.Elem(), ip.associations)
	for _, a := range ip.associations {
		n := 0
		fv, _ := ip.rtResource.Elem().FieldByName(a)
		getParamsNumbers(&n, getIndirectStruct(fv.Type), nil)
		ip.associationsParams = append(ip.associationsParams, n)
	}

	return nil
}

func (ip *Importer) parseOptions(opts ...ImporterExecOption) (
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
