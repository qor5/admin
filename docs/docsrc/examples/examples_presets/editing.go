package examples_presets

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/microcosm-cc/bluemonday"
	"github.com/qor5/web/v3"
	v "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/media"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/admin/v3/tiptap"
)

// @snippet_begin(PresetsEditingCustomizationDescriptionSample)
//
//go:embed assets
var assets embed.FS

func PresetsEditingCustomizationDescription(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	mb, cl, ce, _ = PresetsListingCustomizationBulkActions(b, db)

	ce.Only("Name", "Email", "CompanyID", "Description", "HTMLSanitizerPolicyTiptapInput", "HTMLSanitizerPolicyUGCInput", "HTMLSanitizerPolicyStrictInput", "HTMLSanitizerPolicyCustomInput")

	ce.ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		customer := obj.(*Customer)
		if customer.Name == "" {
			err.FieldError("Name", "name must not be empty")
		}
		if customer.Email == "" {
			err.FieldError("Email", "email must not be empty")
		}
		if customer.Description == "" {
			err.FieldError("Description", "description must not be empty")
		}
		if customer.HTMLSanitizerPolicyTiptapInput == "" {
			err.FieldError("HTMLSanitizerPolicyTiptapInput", "HTMLSanitizerPolicyTiptapInput must not be empty")
		}
		if customer.HTMLSanitizerPolicyUGCInput == "" {
			err.FieldError("HTMLSanitizerPolicyUGCInput", "HTMLSanitizerPolicyUGCInput must not be empty")
		}
		if customer.HTMLSanitizerPolicyStrictInput == "" {
			err.FieldError("HTMLSanitizerPolicyStrictInput", "HTMLSanitizerPolicyStrictInput must not be empty")
		}
		if customer.HTMLSanitizerPolicyCustomInput == "" {
			err.FieldError("HTMLSanitizerPolicyCustomInput", "HTMLSanitizerPolicyCustomInput must not be empty")
		}
		return
	})

	ce.Field("Description").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		extensions := vx.TiptapSlackLikeExtensions()
		return tiptap.TiptapEditor(db, field.FormKey).
			Extensions(extensions).
			Attr(web.VField(field.FormKey, fmt.Sprint(reflectutils.MustGet(obj, field.Name)))...).
			Label(field.Label).
			Disabled(field.Disabled).
			ErrorMessages(field.Errors...)
	})

	ce.Field("HTMLSanitizerPolicyTiptapInput").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		extensions := vx.TiptapSlackLikeExtensions()
		return tiptap.TiptapEditor(db, field.FormKey).
			Extensions(extensions).
			Attr(web.VField(field.FormKey, fmt.Sprint(reflectutils.MustGet(obj, field.Name)))...).
			Label(field.Label).
			Disabled(field.Disabled).
			ErrorMessages(field.Errors...)
	}).SetterFunc(presets.CreateHTMLSanitizer(&presets.HTMLSanitizerConfig{
		Policy: presets.CreateHTMLSanitizerPolicy(presets.HTMLSanitizerPolicyTiptap),
	}))

	ce.Field("HTMLSanitizerPolicyUGCInput").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		extensions := vx.TiptapSlackLikeExtensions()
		return tiptap.TiptapEditor(db, field.FormKey).
			Extensions(extensions).
			Attr(web.VField(field.FormKey, fmt.Sprint(reflectutils.MustGet(obj, field.Name)))...).
			Label(field.Label).
			Disabled(field.Disabled).
			ErrorMessages(field.Errors...)
	}).SetterFunc(presets.CreateHTMLSanitizer(&presets.HTMLSanitizerConfig{
		Policy: presets.CreateHTMLSanitizerPolicy(presets.HTMLSanitizerPolicyUGC),
	}))

	ce.Field("HTMLSanitizerPolicyStrictInput").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		extensions := vx.TiptapSlackLikeExtensions()
		return tiptap.TiptapEditor(db, field.FormKey).
			Extensions(extensions).
			Attr(web.VField(field.FormKey, fmt.Sprint(reflectutils.MustGet(obj, field.Name)))...).
			Label(field.Label).
			Disabled(field.Disabled).
			ErrorMessages(field.Errors...)
	}).SetterFunc(presets.CreateHTMLSanitizer(&presets.HTMLSanitizerConfig{
		Policy: presets.CreateHTMLSanitizerPolicy(presets.HTMLSanitizerPolicyStrict),
	}))

	policy := bluemonday.NewPolicy()

	p := policy.AllowElements("video", "audio")
	p.AllowAttrs("src", "controls").OnElements("video", "audio")

	ce.Field("HTMLSanitizerPolicyCustomInput").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		extensions := vx.TiptapSlackLikeExtensions()
		return tiptap.TiptapEditor(db, field.FormKey).
			Extensions(extensions).
			Attr(web.VField(field.FormKey, fmt.Sprint(reflectutils.MustGet(obj, field.Name)))...).
			Label(field.Label).
			Disabled(field.Disabled).
			ErrorMessages(field.Errors...)
	}).SetterFunc(presets.CreateHTMLSanitizer(&presets.HTMLSanitizerConfig{
		Policy: p,
	}))

	// If you just want to specify the label to be displayed
	wrapper := presets.WrapperFieldLabel(func(evCtx *web.EventContext, obj interface{}, field *presets.FieldContext) (name2label map[string]string, err error) {
		return map[string]string{
			"Name":  "Customer Name",
			"Email": "Customer Email",
		}, nil
	})
	ce.Field("Name").LazyWrapComponentFunc(wrapper)
	ce.Field("Email").LazyWrapComponentFunc(wrapper)
	return
}

// @snippet_end

func PresetsEditingTiptap(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	mb, cl, ce, dp = PresetsEditingCustomizationDescription(b, db)

	mediaBuilder := media.New(db)
	defer func() {
		b.Use(mediaBuilder)
	}()

	b.ExtraAsset("/tiptap.css", "text/css", tiptap.ThemeGithubCSSComponentsPack())
	ce.Field("Description").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		// extensions := vx.TiptapSlackLikeExtensions()
		// extensions = append(extensions,
		// 	&vx.VXTiptapEditorExtension{Name: "ImageGlue"}, // Do not use Image, please use ImageGlue to integrate the media library
		// 	&vx.VXTiptapEditorExtension{Name: "Video"},
		// )
		extensions := tiptap.TiptapExtensions()
		return tiptap.TiptapEditor(db, field.FormKey).
			Extensions(extensions).
			MarkdownTheme("github"). // Match tiptap.ThemeGithubCSSComponentsPack
			Attr(web.VField(field.FormKey, fmt.Sprint(reflectutils.MustGet(obj, field.Name)))...).
			Label(field.Label).
			Disabled(field.Disabled).
			ErrorMessages(field.Errors...)
	})

	return
}

// @snippet_begin(PresetsEditingSingletonNestedSample)

// SingletonNestedItem is the element type for nested slice field in singleton demo.
type SingletonNestedItem struct {
	Name  string
	Value string
}

// SingletonWithNested demonstrates a singleton model that contains a nested slice field.
// Items is persisted by GORM as a JSONB column in the database.
type SingletonWithNested struct {
	ID    uint
	Title string
	Items datatypes.JSONSlice[*SingletonNestedItem] `gorm:"default:'[]'"`
}

// PresetsEditingSingletonNested installs a singleton page with a nested list field
// to reproduce the Nested interactions losing ParamID on singleton editing page.
func PresetsEditingSingletonNested(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	b.DataOperator(gorm2op.DataOperator(db))
	_ = db.AutoMigrate(&SingletonWithNested{})

	mb = b.Model(&SingletonWithNested{}).Singleton(true).Label("Singleton Nested Demo")
	// Configure editing fields: a simple title and a nested list field "Items"
	itemFB := b.NewFieldsBuilder(presets.WRITE).Model(&SingletonNestedItem{}).Only("Name", "Value")
	ce = mb.Editing().Only("Title", "Items")
	ce.ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		singleton := obj.(*SingletonWithNested)
		if len(singleton.Title) > 10 {
			err.FieldError("Title", "title must not be longer than 10 characters")
		}
		return
	})
	ce.Field("Items").Nested(itemFB)

	ce.EditingTitleFunc(func(obj interface{}, defaultTitle string, ctx *web.EventContext) h.HTMLComponent {
		singleton := obj.(*SingletonWithNested)
		if singleton.Title != "" {
			return h.Text(fmt.Sprintf("Custom Title: %s", singleton.Title))
		}
		return h.Text(defaultTitle)
	})

	return
}

// @snippet_end

type MyFile string

type Product struct {
	ID        int
	Title     string
	MainImage MyFile
}

func PresetsEditingCustomizationFileType(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	mb, cl, ce, dp = PresetsEditingCustomizationDescription(b, db)
	err := db.AutoMigrate(&Product{})
	if err != nil {
		panic(err)
	}

	b.FieldDefaults(presets.WRITE).
		FieldType(MyFile("")).
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			val := field.Value(obj).(MyFile)
			var img h.HTMLComponent
			if len(string(val)) > 0 {
				img = h.Img(string(val))
			}
			var er h.HTMLComponent
			if len(field.Errors) > 0 {
				er = h.Div().Text(field.Errors[0]).Style("color:red")
			}
			return h.Div(
				img,
				er,
				h.Input("").Type("file").Attr("@change", fmt.Sprintf("form.%s_NewFile = $event.target.files[0]", field.Name)),
			)
		}).
		SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
			ff, _, _ := ctx.R.FormFile(fmt.Sprintf("%s_NewFile", field.Name))

			if ff == nil {
				return
			}
			var req *http.Request
			req, err = http.NewRequest("PUT", "https://transfer.sh/myfile.png", ff)
			if err != nil {
				return
			}
			var res *http.Response
			res, err = http.DefaultClient.Do(req)
			if err != nil {
				panic(err)
			}
			var b []byte
			b, err = io.ReadAll(res.Body)
			if err != nil {
				return
			}
			if res.StatusCode == 500 {
				err = fmt.Errorf("%s", string(b))
				return
			}
			err = reflectutils.Set(obj, field.Name, MyFile(b))
			return
		})

	pmb := b.Model(&Product{})
	pmb.Editing("Title", "MainImage")
	return
}

// @snippet_begin(PresetsEditingCustomizationValidationSample)

func PresetsEditingCustomizationValidation(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	mb, cl, ce, _ = PresetsEditingCustomizationDescription(b, db)

	ce.ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		cus := obj.(*Customer)
		if len(cus.Name) < 10 {
			err.FieldError("Name", "name is too short")
		}
		return
	})
	return
}

// @snippet_end

// @snippet_begin(PresetsEditingCustomizationTabsSample)

func PresetsEditingCustomizationTabs(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	b.DataOperator(gorm2op.DataOperator(db))
	mb = b.Model(&Company{})
	mb.Listing("ID", "Name")
	mb.Editing().AppendTabsPanelFunc(func(obj interface{}, ctx *web.EventContext) (tab, content h.HTMLComponent) {
		c := obj.(*Company)
		tab = v.VTab(h.Text("New Tab")).Value("2")
		content = v.VTabsWindowItem(
			v.VListItemTitle(h.Text(fmt.Sprintf("Name: %s", c.Name))),
		).Value("2").Class("pa-4")
		return
	})
	return
}

// @snippet_end

func PresetsEditingValidate(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	b.DataOperator(gorm2op.DataOperator(db))
	db.AutoMigrate(&Company{})
	mb = b.Model(&Company{})
	mb.Listing("ID", "Name")
	mb.Editing().ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		company := obj.(*Company)
		if company.Name == "" {
			err.GlobalError("name must not be empty")
		}
		return
	})

	return
}

func PresetsEditingSetter(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	b.DataOperator(gorm2op.DataOperator(db))
	db.AutoMigrate(&Company{})
	mb = b.Model(&Company{})
	mb.Listing("ID", "Name")
	eb := mb.Editing("Name")
	eb.Field("Name").SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
		c := obj.(*Company)
		if c.Name == "" {
			return errors.New("name must not be empty")
		}
		if c.Name == "global" {
			return web.ValidationGlobalError(errors.New(`You can not use global as name`))
		}

		return
	})
	eb.Field("Name").LazyWrapSetterFunc(func(in presets.FieldSetterFunc) presets.FieldSetterFunc {
		return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
			c := obj.(*Company)
			if c.Name == "system" {
				return errors.New(`You can not use "system" as name`)
			}
			return in(obj, field, ctx)
		}
	})

	return
}

func PresetsEditingSection(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	b.DataOperator(gorm2op.DataOperator(db))
	db.AutoMigrate(&Company{})
	mb = b.Model(&Company{})
	section := presets.NewSectionBuilder(mb, "section1").
		Editing("Name").Viewing("Name")

	detail := mb.Detailing("section1").Drawer(true)
	detail.Section(section)

	edit := mb.Editing("section1").ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		c := obj.(*Company)
		if len(c.Name) > 10 {
			err.GlobalError("too long name")
		}
		return
	})

	edit.Section(section.Clone())
	return
}

func PresetsEditingSaverValidation(b *presets.Builder, db *gorm.DB) (mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	b.DataOperator(gorm2op.DataOperator(db))
	db.AutoMigrate(&Company{})
	mb = b.Model(&Company{})
	mb.Editing().SaveFunc(func(obj interface{}, id string, ctx *web.EventContext) (err error) {
		ve := web.ValidationErrors{}
		if obj.(*Company).Name == "system" {
			ve.FieldError("Name", "You can not use system as name")
		}
		if ve.HaveErrors() {
			return &ve
		}
		return nil
	})
	return
}
