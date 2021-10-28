package exchange_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/qor/qor5/exchange"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestImport(t *testing.T) {
	for _, c := range []struct {
		name          string
		metas         []*exchange.Meta
		validators    []func(metaValues exchange.MetaValues) error
		csvContent    string
		expectRecords []*TestExchangeModel
		expectError   error
	}{
		{
			name: "normal",
			metas: []*exchange.Meta{
				exchange.NewMeta("ID").PrimaryKey(true),
				exchange.NewMeta("Name").Header("Nameeee"),
				exchange.NewMeta("Age"),
				exchange.NewMeta("Birth").Setter(func(record interface{}, metaValues exchange.MetaValues) error {
					s := metaValues.Get("Birth")
					if s == "" {
						return nil
					}

					t, err := time.ParseInLocation("2006-01-02", s, time.Local)
					if err != nil {
						return err
					}

					r := record.(*TestExchangeModel)
					r.Birth = &t
					return nil
				}),
			},
			csvContent: `ID,Nameeee,Age,Birth
1,Tom,6,1939-01-01
2,Jerry,5,1940-02-10
`,
			expectRecords: []*TestExchangeModel{
				{
					ID:    1,
					Name:  "Tom",
					Age:   ptrInt(6),
					Birth: ptrTime(time.Date(1939, 1, 1, 0, 0, 0, 0, time.Local)),
				},
				{
					ID:    2,
					Name:  "Jerry",
					Age:   ptrInt(5),
					Birth: ptrTime(time.Date(1940, 2, 10, 0, 0, 0, 0, time.Local)),
				},
			},
			expectError: nil,
		},

		{
			name: "few fields",
			metas: []*exchange.Meta{
				exchange.NewMeta("ID").PrimaryKey(true),
				exchange.NewMeta("Name").Header("Nameeee"),
			},
			csvContent: `ID,Nameeee,Age,Birth
1,Tom,6,1939-01-01
2,Jerry,5,1940-02-10
`,
			expectRecords: []*TestExchangeModel{
				{
					ID:   1,
					Name: "Tom",
				},
				{
					ID:   2,
					Name: "Jerry",
				},
			},
			expectError: nil,
		},

		{
			name: "validator error",
			metas: []*exchange.Meta{
				exchange.NewMeta("ID").PrimaryKey(true),
				exchange.NewMeta("Name").Header("Nameeee"),
			},
			validators: []func(metaValues exchange.MetaValues) error{
				func(ms exchange.MetaValues) error {
					v := ms.Get("Name")
					if v == "" {
						return errors.New("name cannot be empty")
					}
					return nil
				},
			},
			csvContent: `ID,Nameeee,Age,Birth
1,Tom,6,1939-01-01
2,,5,1940-02-10
`,
			expectRecords: nil,
			expectError:   fmt.Errorf("name cannot be empty"),
		},

		{
			name: "has extra columns",
			metas: []*exchange.Meta{
				exchange.NewMeta("ID").PrimaryKey(true),
				exchange.NewMeta("Name").Header("Nameeee"),
			},
			csvContent: `ID,Nameeee,Name2,Age,Birth,Hobby
1,Tom,Tomey,6,1939-01-01,sleep
`,
			expectRecords: []*TestExchangeModel{
				{
					ID:   1,
					Name: "Tom",
				},
			},
			expectError: nil,
		},

		{
			name: "empty value",
			metas: []*exchange.Meta{
				exchange.NewMeta("ID").PrimaryKey(true),
				exchange.NewMeta("Name").Header("Nameeee"),
				exchange.NewMeta("Age"),
			},
			csvContent: `ID,Nameeee,Age,Birth
1,Tom,,1939-01-01
2,Jerry,5,1940-02-10
`,
			expectRecords: []*TestExchangeModel{
				{
					ID:   1,
					Name: "Tom",
					Age:  nil,
				},
				{
					ID:   2,
					Name: "Jerry",
					Age:  ptrInt(5),
				},
			},
			expectError: nil,
		},
	} {
		emptyTables()
		r, err := exchange.NewCSVReader(ioutil.NopCloser(strings.NewReader(c.csvContent)))
		assert.NoError(t, err, c.name)
		err = exchange.NewImporter(&TestExchangeModel{}).
			Metas(c.metas...).
			Validators(c.validators...).
			Exec(db, r)
		if err != nil {
			assert.Equal(t, c.expectError, err, c.name)
			continue
		}
		var records []*TestExchangeModel
		err = db.Order("id asc").Find(&records).Error
		assert.NoError(t, err, c.name)
		assert.Equal(t, c.expectRecords, records, c.name)
	}
}

func TestReImport(t *testing.T) {
	emptyTables()
	var err error
	importer := exchange.NewImporter(&TestExchangeModel{}).
		Metas(
			exchange.NewMeta("ID").PrimaryKey(true),
			exchange.NewMeta("Name"),
		)
	// 1st import
	r, err := exchange.NewCSVReader(ioutil.NopCloser(strings.NewReader(`ID,Name
1,Tom
2,Jerry
`)))
	assert.NoError(t, err)
	err = importer.Exec(db, r)
	assert.NoError(t, err)
	// 2nd import
	r, err = exchange.NewCSVReader(ioutil.NopCloser(strings.NewReader(`ID,Name
1,Tomey
2,
3,Spike
`)))
	assert.NoError(t, err)
	err = importer.Exec(db, r)
	assert.NoError(t, err)

	var records []*TestExchangeModel
	err = db.Order("id asc").Find(&records).Error
	assert.NoError(t, err)
	assert.Equal(t, []*TestExchangeModel{
		{
			ID:   1,
			Name: "Tomey",
		},
		{
			ID:   2,
			Name: "",
		},
		{
			ID:   3,
			Name: "Spike",
		},
	}, records)
}

func TestCompositePrimaryKey(t *testing.T) {
	emptyTables()
	var err error
	importer := exchange.NewImporter(&TestExchangeCompositePrimaryKeyModel{}).Metas(
		exchange.NewMeta("ID").PrimaryKey(true),
		exchange.NewMeta("Name").Header("Name").PrimaryKey(true),
		exchange.NewMeta("Age"),
	)
	r, err := exchange.NewCSVReader(ioutil.NopCloser(strings.NewReader(`ID,Name,Age
1,Tom,6
1,Tom2,16
2,Jerry,5
`)))
	assert.NoError(t, err)

	err = importer.Exec(db, r)
	assert.NoError(t, err)

	var records []*TestExchangeCompositePrimaryKeyModel
	err = db.Order("id asc, name asc").Find(&records).Error
	assert.NoError(t, err)
	assert.Equal(t, []*TestExchangeCompositePrimaryKeyModel{
		{
			ID:   1,
			Name: "Tom",
			Age:  ptrInt(6),
		},
		{
			ID:   1,
			Name: "Tom2",
			Age:  ptrInt(16),
		},
		{
			ID:   2,
			Name: "Jerry",
			Age:  ptrInt(5),
		},
	}, records)
}

type ExchangeUser struct {
	gorm.Model

	Number string `gorm:"uniqueIndex"`
	Name   string
	Birth  time.Time

	Hobbies []ExchangeHobby
}

type ExchangeHobby struct {
	gorm.Model

	ExchangeUserID uint

	Name      string
	OtherName string
}

func TestComplexImport(t *testing.T) {
	var err error
	err = db.AutoMigrate(&ExchangeUser{}, &ExchangeHobby{})
	if err != nil {
		panic(err)
	}

	importer := exchange.NewImporter(&ExchangeUser{}).
		Associations("Hobbies").
		Metas(
			exchange.NewMeta("Number").PrimaryKey(true),
			exchange.NewMeta("Name"),
			exchange.NewMeta("Running").Setter(func(record interface{}, metaValues exchange.MetaValues) error {
				v := strings.ToLower(metaValues.Get("Running"))
				if v != "true" {
					return nil
				}
				u := record.(*ExchangeUser)
				hs := u.Hobbies
				has := false
				for _, oh := range hs {
					if oh.Name == "Running" {
						has = true
						break
					}
				}
				if !has {
					hs = append(hs, ExchangeHobby{
						Name: "Running",
					})
				}
				u.Hobbies = hs
				return nil
			}),
			exchange.NewMeta("Swimming").Setter(func(record interface{}, metaValues exchange.MetaValues) error {
				v := strings.ToLower(metaValues.Get("Swimming"))
				if v != "true" {
					return nil
				}
				u := record.(*ExchangeUser)
				hs := u.Hobbies
				has := false
				for _, oh := range hs {
					if oh.Name == "Swimming" {
						has = true
						break
					}
				}
				if !has {
					hs = append(hs, ExchangeHobby{
						Name: "Swimming",
					})
				}
				u.Hobbies = hs
				return nil
			}),
		)
	r, err := exchange.NewCSVReader(ioutil.NopCloser(strings.NewReader(`Number,Name,Running,Swimming
100,Tom,true,true
200,Jerry,false,true
`)))
	assert.NoError(t, err)
	err = importer.Exec(db, r)
	assert.NoError(t, err)
}
