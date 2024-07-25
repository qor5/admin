package activity

import (
	"cmp"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/theplant/testenv"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type Foo struct {
	ID   string
	Name string
}

type Bar struct {
	ID   string
	Name string
}

var db *gorm.DB

func TestMain(m *testing.M) {
	env, err := testenv.New().DBEnable(true).SetUp()
	if err != nil {
		panic(err)
	}
	defer env.TearDown()

	db = env.DB
	db.Logger = db.Logger.LogMode(logger.Info)

	if err = db.AutoMigrate(&Foo{}); err != nil {
		panic(err)
	}

	m.Run()
}

func TestTablePrefix(t *testing.T) {
	require.NoError(t, db.Create(&Foo{ID: "1", Name: "foo"}).Error)
	{
		foo := &Foo{}
		require.NoError(t, db.Where("id = ?", "1").First(foo).Error)
		require.Equal(t, "foo", foo.Name)
	}

	require.NoError(t, db.Exec(`CREATE SCHEMA IF NOT EXISTS copilot ;`).Error)

	db := db.Session(&gorm.Session{})
	db.Config.NamingStrategy = schema.NamingStrategy{
		TablePrefix:         "copilot.",
		IdentifierMaxLength: 64,
	}
	{
		foo := &Foo{}
		sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
			return tx.Where("id = ?", "1").First(foo)
		})
		require.NotContains(t, sql, "copilot") // Because the db already has an internal cache
	}
	{
		require.NoError(t, db.AutoMigrate(&Bar{}))
		require.NoError(t, db.Create(&Bar{ID: "1", Name: "bar"}).Error)

		sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
			return tx.Create(&Bar{ID: "1", Name: "bar"})
		})
		require.Contains(t, sql, "copilot") // Because the db hasn't cached the Bar yet.
	}
	// So it is not a reliable solution.
}

func TestTable(t *testing.T) {
	foo := &Foo{}
	require.Contains(t, db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Table("copilotx.foos").Where("id = ?", "1").First(foo)
	}), "copilotx")
	require.Contains(t, db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Table("copiloty.foos").Where("id = ?", "1").First(foo)
	}), "copiloty")

	require.NoError(t, db.Exec(`CREATE SCHEMA IF NOT EXISTS copilot;`).Error)
	db := db.Table("copilot.foos").Session(&gorm.Session{}) // Fixed TableName
	require.Contains(t, db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Where("id = ?", "1").First(foo)
	}), "copilot")
	require.Contains(t, db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Where("id = ?", "1").First(foo)
	}), "copilot")
}

func TestScopes(t *testing.T) {
	callCount := 0
	scopeTableName := func(tableName string) func(*gorm.DB) *gorm.DB {
		return func(db *gorm.DB) *gorm.DB {
			callCount++
			return db.Table(tableName)
		}
	}
	{
		db := db.Scopes(scopeTableName("foos"))
		require.NoError(t, db.Create(&Foo{ID: "1", Name: "foo1"}).Error)
		require.NoError(t, db.Create(&Foo{ID: "2", Name: "foo2"}).Error)
		require.Equal(t, 1, callCount) // the Scopes method is disposable
	}
	{
		db := db.Scopes(scopeTableName("foos")).Session(&gorm.Session{}) // fixed
		require.NoError(t, db.Create(&Foo{ID: "3", Name: "foo1"}).Error)
		require.NoError(t, db.Create(&Foo{ID: "4", Name: "foo2"}).Error)
		require.Equal(t, 1+2, callCount)
	}
}

func TestDynamicTablePrefix(t *testing.T) {
	getTableName := func(db *gorm.DB, tablePrefix string, model any) string {
		stmt := &gorm.Statement{DB: db}
		stmt.Parse(model)
		return tablePrefix + stmt.Schema.Table
	}

	dynamicTablePrefix := func(tablePrefix string) func(db *gorm.DB) *gorm.DB {
		return func(db *gorm.DB) *gorm.DB {
			stmt := db.Statement
			if stmt.Table != "" {
				return db
			}

			model := cmp.Or(stmt.Model, stmt.Dest)
			if model == nil {
				return db
			}

			return db.Table(getTableName(db, tablePrefix, model))
		}
	}

	type Foox struct {
		ID   string
		Name string
	}

	type Barx struct {
		ID   string
		Name string
	}

	prefix := "some_"

	db := db.Scopes(dynamicTablePrefix(prefix)).Session(&gorm.Session{})
	require.NoError(t, db.AutoMigrate(&Foox{}, &Barx{}))

	foo := &Foox{}
	require.Equal(t, "some_fooxes", getTableName(db, prefix, foo))
	require.Contains(t, db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Where("id = ?", "1").First(foo)
	}), "some_fooxes")

	bar := &Barx{}
	require.Equal(t, "some_barxes", getTableName(db, prefix, bar))
	require.Contains(t, db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Where("id = ?", "1").First(bar)
	}), "some_barxes")

	require.Contains(t, db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Table("x_barxes").Where("id = ?", "1").First(bar)
	}), "x_barxes")
}
