package examples_admin

import (
	"net/http"
	"reflect"
	"time"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3"
	v "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
	"github.com/theplant/osenv"
	"github.com/theplant/relay"
	"github.com/theplant/relay/gormrelay"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

var (
	exampleDB      *gorm.DB
	dbParamsString = osenv.Get("DB_PARAMS", "admin example database connection string", "user=docs password=docs dbname=docs sslmode=disable host=localhost port=6532 TimeZone=Asia/Tokyo")
)

func ExampleDB() (db *gorm.DB) {
	if exampleDB != nil {
		return exampleDB
	}
	var err error
	db, err = gorm.Open(postgres.Open(dbParamsString), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db.Logger.LogMode(logger.Info)
	err = db.AutoMigrate(
		&Post{},
		&Category{},
		&WithPublishProduct{},
	)
	if err != nil {
		panic(err)
	}
	return
}

// @snippet_begin(PresetsListingSample)

type Post struct {
	ID        uint
	Title     string
	Body      string
	UpdatedAt time.Time
	CreatedAt time.Time
	Disabled  bool

	Status string

	CategoryID uint
}

type Category struct {
	ID   uint
	Name string

	UpdatedAt time.Time
	CreatedAt time.Time
}

func ListingExample(b *presets.Builder, db *gorm.DB) http.Handler {
	return listingExample(b, db, nil)
}

func listingExample(b *presets.Builder, db *gorm.DB, customize func(mb *presets.ModelBuilder)) http.Handler {
	db.AutoMigrate(&Post{}, &Category{})

	// Setup the project name, ORM and Homepage
	b.DataOperator(gorm2op.DataOperator(db))

	// Register Post into the builder
	// Use m to customize the model, Or config more models here.
	postModelBuilder := b.Model(&Post{})
	postModelBuilder.Listing("ID", "Title", "Body", "CategoryID", "VirtualField")

	postModelBuilder.Listing().SearchFunc(func(ctx *web.EventContext, params *presets.SearchParams) (result *presets.SearchResult, err error) {
		qdb := db.Where("disabled != true")
		return gorm2op.DataOperator(qdb).Search(ctx, params)
	})

	rmn := postModelBuilder.Listing().RowMenu()
	rmn.RowMenuItem("Show").
		ComponentFunc(func(obj interface{}, id string, ctx *web.EventContext) h.HTMLComponent {
			return v.VListItem(
				web.Slot(
					v.VIcon("mdi-menu"),
				).Name("prepend"),
				v.VListItemTitle(
					h.Text("Show"),
				),
			)
		})

	postModelBuilder.Editing().Field("CategoryID").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		categories := []Category{}
		if err := db.Find(&categories).Error; err != nil {
			// ignore err for now
		}

		return v.VAutocomplete().
			Chips(true).
			Attr(web.VField(field.Name, field.Value(obj))...).Label(field.Label).
			Items(categories).
			ItemTitle("Name").
			ItemValue("ID")
	})

	postModelBuilder.Listing().Field("CategoryID").Label("Category").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		c := Category{}
		cid, _ := field.Value(obj).(uint)
		if err := db.Where("id = ?", cid).Find(&c).Error; err != nil {
			// ignore err in the example
		}
		return h.Td(h.Text(c.Name))
	})

	postModelBuilder.Listing().Field("VirtualField").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return h.Td(h.Text("virtual field"))
	})

	if customize != nil {
		customize(postModelBuilder)
	}

	b.Model(&Category{})
	// Use m to customize the model, Or config more models here.
	return b
}

// @snippet_end

type PostWithCategory struct {
	Post
	Category *Category `gorm:"foreignKey:CategoryID;references:ID"`
}

func (p *PostWithCategory) TableName() string {
	return "posts"
}

func ListingWithJoinsExample(b *presets.Builder, db *gorm.DB) http.Handler {
	return listingPostWithCategory(b, db, nil)
}

func listingPostWithCategory(b *presets.Builder, db *gorm.DB, customize func(mb *presets.ModelBuilder)) http.Handler {
	db.AutoMigrate(&Post{}, &Category{})

	b.DataOperator(gorm2op.DataOperator(db))

	mb := b.Model(&PostWithCategory{})

	eb := mb.Editing("Title", "Body", "Disabled", "Status", "CategoryID")
	eb.Field("CategoryID").Label("Category").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		categories := []Category{}
		if err := db.Find(&categories).Error; err != nil {
			panic(err)
		}
		fieldValue := field.Value(obj)
		if reflect.ValueOf(fieldValue).IsZero() && len(categories) > 0 {
			fieldValue = categories[0].ID
		}
		return v.VAutocomplete().
			Chips(true).
			Attr(web.VField(field.Name, fieldValue)...).Label(field.Label).
			Items(categories).
			ItemTitle("Name").
			ItemValue("ID")
	})

	lb := mb.Listing("ID", "Title", "Body", "Disabled", "Status", "CategoryName")

	lb.RelayPagination(gorm2op.KeysetBasedPagination(true)).
		OrderableFields([]*presets.OrderableField{
			{FieldName: "ID"},
			{FieldName: "CategoryName"},
		}).
		DefaultOrderBy(relay.Order{Field: "ID", Direction: relay.OrderDirectionDesc})

	lb.WrapSearchFunc(func(in presets.SearchFunc) presets.SearchFunc {
		return func(ctx *web.EventContext, params *presets.SearchParams) (result *presets.SearchResult, err error) {
			oldR := ctx.R
			defer func() {
				ctx.R = oldR // restore the original request context
			}()
			ctx = gorm2op.EventContextWithHook(ctx, func(db *gorm.DB) *gorm.DB {
				return db.Joins("Category")
			})
			ctx = gorm2op.EventContextWithRelayComputedHook(ctx, func(computed *gormrelay.Computed[any]) *gormrelay.Computed[any] {
				computed.Columns["CategoryName"] = clause.Column{
					Name: `("Category"."name")`,
					Raw:  true,
				}
				return computed
			})
			return in(ctx, params)
		}
	})

	lb.Field("CategoryName").Label("Category").LazyWrapComponentFunc(func(in presets.FieldComponentFunc) presets.FieldComponentFunc {
		return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			v := obj.(*PostWithCategory)
			if v.Category == nil {
				return h.Td(h.Text("-"))
			}
			return h.Td(h.Text(v.Category.Name))
		}
	})

	if customize != nil {
		customize(mb)
	}

	b.Model(&Category{})
	// Use m to customize the model, Or config more models here.
	return b
}
