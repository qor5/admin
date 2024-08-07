package examples_autocompelte

import (
	"fmt"
	"github.com/qor5/admin/v3/autocomplete"

	docsexamples "github.com/qor5/admin/v3/docs/docsrc/examples"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3/examples"
	"gorm.io/gorm"
)

func SamplesHandler(mux examples.Muxer, ab *autocomplete.Builder) {
	db := docsexamples.ExampleDB()
	ab.DB(db)
	addExample(mux, db, ab, PresetsAutoCompleteBasicFilter)
	return
}

type exampleFunc func(b *presets.Builder, ab *autocomplete.Builder, db *gorm.DB) (
	cust *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
)

func addExample(mux examples.Muxer, db *gorm.DB, ab *autocomplete.Builder, f exampleFunc) {
	path := examples.URLPathByFunc(f)
	p := presets.New().URIPrefix(path)
	f(p, ab, db)
	fmt.Println("Example mounting at: ", path)
	mux.Handle(path, p)
}
