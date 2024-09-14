package examples_presets

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/qor5/admin/v3/media"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/admin/v3/richeditor"
	"github.com/qor5/admin/v3/tiptap"
	"github.com/qor5/web/v3"
	v "github.com/qor5/x/v3/ui/vuetify"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
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
	js, _ := assets.ReadFile("assets/fontcolor.min.js")
	richeditor.Plugins = []string{"alignment", "table", "video", "imageinsert", "fontcolor"}
	richeditor.PluginsJS = [][]byte{js}
	b.ExtraAsset("/redactor.js", "text/javascript", richeditor.JSComponentsPack())
	b.ExtraAsset("/redactor.css", "text/css", richeditor.CSSComponentsPack())

	mb, cl, ce, _ = PresetsListingCustomizationBulkActions(b, db)

	ce.Only("Name", "Email", "CompanyID", "Description")

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
		return
	})

	ce.Field("Description").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return richeditor.RichEditor(db, "Body").
			Plugins([]string{"alignment", "video", "imageinsert", "fontcolor"}).
			Value(obj.(*Customer).Description).
			Label(field.Label).
			Disabled(field.Disabled).
			ErrorMessages(field.Errors...)
	})

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
		// 	&vx.VXTiptapEditorExtension{Name: "ImageGlue"},
		// 	&vx.VXTiptapEditorExtension{Name: "Video"},
		// )
		extensions := tiptap.TiptapExtensions()
		return tiptap.TiptapEditor(db, "Body").
			Extensions(extensions).
			MarkdownTheme("github"). // match tiptap.ThemeGithubCSSComponentsPack
			Attr(web.VField(field.FormKey, fmt.Sprint(reflectutils.MustGet(obj, field.Name)))...).
			Disabled(field.Disabled)
	})

	return
}

// @snippet_begin(PresetsEditingCustomizationFileTypeSample)

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

// @snippet_end

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
