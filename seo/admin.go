package seo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/qor5/web/v3"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"

	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/perm"
	. "github.com/qor5/x/v3/ui/vuetify"
	. "github.com/qor5/x/v3/ui/vuetifyx"

	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/media"
	"github.com/qor5/admin/v3/media/base"
	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/presets"
)

const I18nSeoKey i18n.ModuleKey = "I18nSeoKey"

const SeoDetailFieldName = "SEO"

var permVerifier *perm.Verifier

type myTd struct {
	td    *h.HTMLTagBuilder
	child h.MutableAttrHTMLComponent
}

func (mtd *myTd) SetAttr(k string, v interface{}) {
	mtd.td.SetAttr(k, v)
	mtd.child.SetAttr(k, v)
}

func (mtd *myTd) MarshalHTML(ctx context.Context) ([]byte, error) {
	mtd.td.Children(mtd.child)
	return mtd.td.MarshalHTML(ctx)
}

func (b *Builder) Install(pb *presets.Builder) error {
	// NOTE: do not replace b.seoRoot.name with defaultGlobalSEOName.
	// because the name of global seo may be changed by user through WithGlobalSEOName option.
	if err := insertIfNotExists(b.db, b.seoRoot.name, b.locales); err != nil {
		return err
	}
	// The registration of FieldDefaults for writing Setting here
	// must be executed before `pb.Model(&QorSEOSetting{})...`,
	pb.FieldDefaults(presets.WRITE).
		FieldType(Setting{}).
		ComponentFunc(b.EditingComponentFunc).
		SetterFunc(EditSetterFunc)

	seoModel := pb.Model(&QorSEOSetting{}).PrimaryField("Name").
		Label("SEO").
		RightDrawerWidth("1000").
		LayoutConfig(&presets.LayoutConfig{
			NotificationCenterInvisible: true,
		})
	b.mb = seoModel

	// Configure Listing Page
	b.configListing(seoModel)
	// Configure Editing Page
	b.configEditing(seoModel)
	// b.ConfigDetailing(pb)

	pb.GetI18n().
		RegisterForModule(language.English, I18nSeoKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nSeoKey, Messages_zh_CN).
		RegisterForModule(language.Japanese, I18nSeoKey, Messages_ja_JP)

	permVerifier = perm.NewVerifier("seo", pb.GetPermission())
	return nil
}

func (b *Builder) configListing(seoModel *presets.ModelBuilder) {
	listing := seoModel.Listing("Name").Title(func(evCtx *web.EventContext, style presets.ListingStyle, currentTitle string) (title string, titleCompo h.HTMLComponent, err error) {
		return "SEO", nil, nil
	})
	// disable new btn globally, no one can add new SEO record after the server start up.
	listing.NewButtonFunc(func(ctx *web.EventContext) h.HTMLComponent {
		return nil
	})
	listing.DisablePagination(true)

	// Remove the row menu from each row
	listing.RowMenu().Empty()

	// Configure the indentation for Name field to display hierarchy.
	listing.Field("Name").ComponentFunc(
		func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			seoSetting := obj.(*QorSEOSetting)
			icon := "mdi-folder"
			priority := b.GetSEOPriority(seoSetting.Name)
			return &myTd{
				td: h.Td(),
				child: h.Div(
					VIcon(icon).Size(SizeSmall).Class("mb-1"),
					h.Text(seoSetting.Name),
				).Style(fmt.Sprintf("padding-left: %dpx;", 32*(priority-1))),
			}
		},
	)

	listing.WrapSearchFunc(func(in presets.SearchFunc) presets.SearchFunc {
		return func(ctx *web.EventContext, params *presets.SearchParams) (result *presets.SearchResult, err error) {
			locale, _ := l10n.IsLocalizableFromContext(ctx.R.Context())
			var seoNames []string
			for name := range b.registeredSEO {
				if name, ok := name.(string); ok {
					seoNames = append(seoNames, name)
				}
			}
			cond := presets.SQLCondition{
				Query: "locale_code = ? and name in (?)",
				Args:  []interface{}{locale, seoNames},
			}

			params.SQLConditions = append(params.SQLConditions, &cond)
			result, err = in(ctx, params)
			b.SortSEOs(result.Nodes.([]*QorSEOSetting))
			return
		}
	})
}

func (b *Builder) configEditing(seoModel *presets.ModelBuilder) {
	editing := seoModel.Editing("Variables", "Setting")

	// Customize the Saver to trigger the invocation of the `afterSave` hook function (if available)
	// when updating the global seo.
	editing.WrapSaveFunc(func(in presets.SaveFunc) presets.SaveFunc {
		return func(obj interface{}, id string, ctx *web.EventContext) (err error) {
			if err = in(obj, id, ctx); err != nil {
				return
			}
			seoSetting := obj.(*QorSEOSetting)

			if b.afterSave != nil {
				if err = b.afterSave(ctx.R.Context(), seoSetting.Name, seoSetting.LocaleCode); err != nil {
					return err
				}
			}
			return
		}
	})

	// configure variables field
	{
		const formKeyForVariablesField = "Variables"
		editing.Field("Variables").ComponentFunc(
			func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
				seoSetting := obj.(*QorSEOSetting)
				msgr := i18n.MustGetModuleMessages(ctx.R, I18nSeoKey, Messages_en_US).(*Messages)
				settingVars := b.GetSEO(seoSetting.Name).settingVars
				var variablesComps h.HTMLComponents
				if len(settingVars) > 0 {
					variablesComps = append(variablesComps, h.H3(msgr.Variable).Style("margin-top:15px;font-weight: 500"))
					for varName := range settingVars {
						fieldComp := VTextField().
							Attr(web.VField(fmt.Sprintf("%s.%s", formKeyForVariablesField, varName), seoSetting.Variables[varName])...).
							Label(i18n.PT(ctx.R, presets.ModelsI18nModuleKey, "Seo Variable", varName))
						variablesComps = append(variablesComps, fieldComp)
					}
				}
				return variablesComps
			},
		)
		// Because the Variables type is of map type, you need to configure the setter func by yourself.
		// If not configured, it will cause the updated valued to not be written to the database.
		editing.Field("Variables").SetterFunc(
			func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
				seoSetting := obj.(*QorSEOSetting)
				if seoSetting.Variables == nil {
					seoSetting.Variables = make(Variables)
				}
				for fieldName := range ctx.R.Form {
					if strings.HasPrefix(fieldName, formKeyForVariablesField) {
						varName := strings.TrimPrefix(fieldName, formKeyForVariablesField+".")
						val := ctx.R.Form[fieldName][0]
						seoSetting.Variables[varName] = val
					}
				}
				return nil
			},
		)
	}

	editing.Field("Setting").ComponentFunc(
		func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			seoSetting := obj.(*QorSEOSetting)
			return b.vseo("Setting", field, b.GetSEO(seoSetting.Name), &seoSetting.Setting, ctx.R)
		},
	)
}

func EditSetterFunc(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
	var setting Setting
	mediaBox := media_library.MediaBox{}
	for fieldWithPrefix := range ctx.R.Form {
		// make sure OpenGraphImageFromMediaLibrary.Description set after OpenGraphImageFromMediaLibrary.Values
		if fieldWithPrefix == fmt.Sprintf("%s.%s", field.Name, "OpenGraphImageFromMediaLibrary.Values") {
			err = mediaBox.Scan(ctx.R.FormValue(fieldWithPrefix))
			if err != nil {
				return
			}
			break
		}
	}
	for fieldWithPrefix := range ctx.R.Form {
		if strings.HasPrefix(fieldWithPrefix, fmt.Sprintf("%s.%s", field.Name, "OpenGraphImageFromMediaLibrary")) {
			if fieldWithPrefix == fmt.Sprintf("%s.%s", field.Name, "OpenGraphImageFromMediaLibrary.Description") {
				mediaBox.Description = ctx.R.Form.Get(fieldWithPrefix)
				setting.OpenGraphImageFromMediaLibrary = mediaBox
			}
			continue
		}
		if fieldWithPrefix == fmt.Sprintf("%s.%s", field.Name, "OpenGraphMetadataString") {
			metadata := GetOpenGraphMetadata(ctx.R.Form.Get(fieldWithPrefix))
			setting.OpenGraphMetadata = metadata
			continue
		}
		if strings.HasPrefix(fieldWithPrefix, fmt.Sprintf("%s.", field.Name)) {
			reflectutils.Set(&setting, strings.TrimPrefix(fieldWithPrefix, fmt.Sprintf("%s.", field.Name)), ctx.R.Form.Get(fieldWithPrefix))
		}
	}
	return reflectutils.Set(obj, field.Name, setting)
}

func (b *Builder) EditingComponentFunc(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
	var (
		msgr        = i18n.MustGetModuleMessages(ctx.R, I18nSeoKey, Messages_en_US).(*Messages)
		fieldPrefix string
		setting     Setting
		db          = b.db
		locale, _   = l10n.IsLocalizableFromContext(ctx.R.Context())
	)
	seo := b.GetSEO(obj)
	if seo == nil {
		return h.Div()
	}

	value := reflect.Indirect(reflect.ValueOf(obj))
	for i := 0; i < value.NumField(); i++ {
		if s, ok := value.Field(i).Interface().(Setting); ok {
			setting = s
			fieldPrefix = value.Type().Field(i).Name
		}
	}
	if !setting.EnabledCustomize && setting.IsEmpty() {
		modelSetting := &QorSEOSetting{}
		db.Where("name = ? AND locale_code = ?", seo.name, locale).First(modelSetting)
		setting = modelSetting.Setting
	}
	customizeForm := fmt.Sprintf("%s.%s", fieldPrefix, "EnabledCustomize")
	return web.Scope(
		h.Div(
			VSwitch().
				Disabled(field.Disabled).
				Label(msgr.Customize).Color("primary").
				Attr(web.VField(customizeForm, setting.EnabledCustomize)...).
				Attr("@update:model-value", "locals.enabledCustomize=$event"),
			h.Div(
				b.vseo(fieldPrefix, field, seo, &setting, ctx.R),
			).Attr("v-show", "locals.enabledCustomize"),
		).Class("pb-4"),
	).Init(fmt.Sprintf(`{enabledCustomize: %t}`, setting.EnabledCustomize)).
		VSlot("{ locals }")
}

func (b *Builder) vseo(fieldPrefix string, field *presets.FieldContext, seo *SEO, setting *Setting, req *http.Request) h.HTMLComponent {
	var (
		msgr = i18n.MustGetModuleMessages(req, I18nSeoKey, Messages_en_US).(*Messages)
		db   = b.db
	)

	var varComps []h.HTMLComponent
	for varName := range seo.getAvailableVars() {
		varComps = append(varComps,
			VBtn(
				i18n.PT(req, presets.ModelsI18nModuleKey, "Seo Variable", varName)).
				PrependIcon("mdi-plus").
				Attr("@click", fmt.Sprintf("$refs.seo.addTags('%s')", varName)).
				Variant(VariantTonal).
				Size(SizeSmall).
				Disabled(field.Disabled).
				Color(ColorPrimary).Class("mr-2"),
		)
	}

	image := &setting.OpenGraphImageFromMediaLibrary
	if image.ID.String() == "0" {
		image.ID = json.Number("")
	}
	refPrefix := strings.ReplaceAll(strings.ToLower(fieldPrefix), " ", "_")
	return VXSendVariables(
		h.Div(
			h.Span(msgr.Basic).Class("text-subtitle-1 px-2 py-1 rounded", "bg-"+ColorGreyLighten3),
		),
		h.Div(
			VRow(
				VCol(
					varComps...,
				),
			),
		).Class("mt-4 mb-4"),

		VXField().Disabled(field.Disabled).Attr("counter", true).Attr(web.VField(fmt.Sprintf("%s.%s", fieldPrefix, "Title"), setting.Title)...).Label(msgr.Title).Attr("@focus", fmt.Sprintf("$refs.seo.tagInputsFocus($refs.%s)", fmt.Sprintf("%s_title", refPrefix))).Attr("ref", fmt.Sprintf("%s_title", refPrefix)),
		VXField().Disabled(field.Disabled).Attr("counter", true).Attr(web.VField(fmt.Sprintf("%s.%s", fieldPrefix, "Description"), setting.Description)...).Label(msgr.Description).Attr("@focus", fmt.Sprintf("$refs.seo.tagInputsFocus($refs.%s)", fmt.Sprintf("%s_description", refPrefix))).Attr("ref", fmt.Sprintf("%s_description", refPrefix)),
		VXField().Disabled(field.Disabled).Type("textarea").Attr("counter", true).Attr(web.VField(fmt.Sprintf("%s.%s", fieldPrefix, "Keywords"), setting.Keywords)...).Label(msgr.Keywords).Attr("@focus", fmt.Sprintf("$refs.seo.tagInputsFocus($refs.%s)", fmt.Sprintf("%s_keywords", refPrefix))).Attr("ref", fmt.Sprintf("%s_keywords", refPrefix)),

		h.Div(
			h.Span(msgr.OpenGraphInformation).Class("text-subtitle-1 px-2 py-1 rounded", "bg-"+ColorGreyLighten3),
		).Class("mb-6 mt-6"),
		VXField().Disabled(field.Disabled).Attr(web.VField(fmt.Sprintf("%s.%s", fieldPrefix, "OpenGraphTitle"), setting.OpenGraphTitle)...).Label(msgr.OpenGraphTitle).Attr("@focus", fmt.Sprintf("$refs.seo.tagInputsFocus($refs.%s)", fmt.Sprintf("%s_og_title", refPrefix))).Attr("ref", fmt.Sprintf("%s_og_title", refPrefix)),
		VXField().Disabled(field.Disabled).Attr(web.VField(fmt.Sprintf("%s.%s", fieldPrefix, "OpenGraphDescription"), setting.OpenGraphDescription)...).Label(msgr.OpenGraphDescription).Attr("@focus", fmt.Sprintf("$refs.seo.tagInputsFocus($refs.%s)", fmt.Sprintf("%s_og_description", refPrefix))).Attr("ref", fmt.Sprintf("%s_og_description", refPrefix)),
		VXField().Disabled(field.Disabled).Attr(web.VField(fmt.Sprintf("%s.%s", fieldPrefix, "OpenGraphURL"), setting.OpenGraphURL)...).Label(msgr.OpenGraphURL).Attr("@focus", fmt.Sprintf("$refs.seo.tagInputsFocus($refs.%s)", fmt.Sprintf("%s_og_url", refPrefix))).Attr("ref", fmt.Sprintf("%s_og_url", refPrefix)),
		VXField().Disabled(field.Disabled).Attr(web.VField(fmt.Sprintf("%s.%s", fieldPrefix, "OpenGraphType"), setting.OpenGraphType)...).Label(msgr.OpenGraphType).Attr("@focus", fmt.Sprintf("$refs.seo.tagInputsFocus($refs.%s)", fmt.Sprintf("%s_og_type", refPrefix))).Attr("ref", fmt.Sprintf("%s_og_type", refPrefix)),
		VXField().Disabled(field.Disabled).Attr(web.VField(fmt.Sprintf("%s.%s", fieldPrefix, "OpenGraphImageURL"), setting.OpenGraphImageURL)...).Label(msgr.OpenGraphImageURL).Attr("@focus", fmt.Sprintf("$refs.seo.tagInputsFocus($refs.%s)", fmt.Sprintf("%s_og_imageurl", refPrefix))).Attr("ref", fmt.Sprintf("%s_og_imageurl", refPrefix)),
		media.QMediaBox(db).Disabled(field.Disabled).Label(msgr.OpenGraphImage).
			FieldName(fmt.Sprintf("%s.%s", fieldPrefix, "OpenGraphImageFromMediaLibrary")).
			Value(image).
			Config(&media_library.MediaBoxConfig{
				AllowType: "image",
				Sizes: map[string]*base.Size{
					"og": {
						Width:  1200,
						Height: 630,
					},
					"twitter-large": {
						Width:  1200,
						Height: 600,
					},
					"twitter-small": {
						Width:  630,
						Height: 630,
					},
				},
			}),
		VXField().Disabled(field.Disabled).Type("textarea").Attr(web.VField(fmt.Sprintf("%s.%s", fieldPrefix, "OpenGraphMetadataString"), GetOpenGraphMetadataString(setting.OpenGraphMetadata))...).Label(msgr.OpenGraphMetadata).Attr("@focus", fmt.Sprintf("$refs.seo.tagInputsFocus($refs.%s)", fmt.Sprintf("%s_og_metadata", refPrefix))).Attr("ref", fmt.Sprintf("%s_og_metadata", refPrefix)),
	).Attr("ref", "seo")
}

func (b *Builder) vSeoReadonly(obj interface{}, fieldPrefix, locale string, seo *SEO, setting *Setting, req *http.Request) h.HTMLComponent {
	var (
		msgr = i18n.MustGetModuleMessages(req, I18nSeoKey, Messages_en_US).(*Messages)
		db   = b.db
	)
	image := &setting.OpenGraphImageFromMediaLibrary
	if image.ID.String() == "0" {
		image.ID = json.Number("")
	}
	localeFinalSeoSetting := seo.getLocaleFinalQorSEOSetting(locale, b.db)
	variables := localeFinalSeoSetting.Variables
	finalContextVars := seo.getFinalContextVars()
	// execute function for context var
	for varName, varFunc := range finalContextVars {
		variables[varName] = varFunc(obj, setting, req)
	}
	*setting = replaceVariables(*setting, variables)

	return h.Components(
		h.Div(
			h.Span(msgr.Basic).Class("text-subtitle-1 px-2 py-1 rounded", "bg-"+ColorGreyLighten3),
		),
		seoFieldPortal(msgr.Title, setting.Title),
		seoFieldPortal(msgr.Description, setting.Description),
		seoFieldPortal(msgr.Keywords, setting.Keywords),
		h.Div(
			h.Span(msgr.OpenGraphInformation).Class("text-subtitle-1 px-2 py-1 rounded", "bg-"+ColorGreyLighten3),
		).Class("mt-10"),
		seoFieldPortal(msgr.OpenGraphTitle, setting.OpenGraphTitle),
		seoFieldPortal(msgr.OpenGraphDescription, setting.OpenGraphDescription),
		seoFieldPortal(msgr.OpenGraphURL, setting.OpenGraphURL),
		seoFieldPortal(msgr.OpenGraphImageURL, setting.OpenGraphImageURL),
		h.Div(
			h.Span(msgr.OpenGraphImage).Class("text-subtitle-1 px-2 py-1 rounded", "bg-"+ColorGreyLighten3),
		).Class("mt-10 mb-2"),
		VContainer(
			VRow(
				VCol(media.QMediaBox(db).
					Readonly(true).
					FieldName(fmt.Sprintf("%s.%s", fieldPrefix, "OpenGraphImageFromMediaLibrary")).
					Value(image).
					Config(&media_library.MediaBoxConfig{
						AllowType:   "image",
						DisableCrop: true,
						Sizes: map[string]*base.Size{
							"og": {
								Width:  1200,
								Height: 630,
							},
							"twitter-large": {
								Width:  1200,
								Height: 600,
							},
							"twitter-small": {
								Width:  630,
								Height: 630,
							},
						},
					})).Cols(12)),
		).Class("pl-0 pt-2"),
		h.Div(
			h.Span(msgr.OpenGraphMetadata).Class("text-subtitle-1 px-2 py-1 rounded", "bg-"+ColorGreyLighten3),
		).Class("mt-1"),
		h.Div(
			h.Pre(
				GetOpenGraphMetadataString(setting.OpenGraphMetadata),
			).Style("margin: 0; font-family: inherit;"),
		).Class("mt-4 px-3"),
	)
}

func (b *Builder) ModelInstall(pb *presets.Builder, mb *presets.ModelBuilder) error {
	b.configDetailing(mb)
	return nil
}

func (b *Builder) configDetailing(mb *presets.ModelBuilder) {
	pd := mb.Detailing()
	fb := pd.GetField(SeoDetailFieldName)
	if fb != nil && fb.GetCompFunc() == nil {
		seoSection := presets.NewSectionBuilder(mb, "SEO").
			Editing("SEO").
			SetterFunc(b.detailSaver).
			ViewComponentFunc(b.detailShowComponent).
			EditComponentFunc(b.EditingComponentFunc)
		pd.Section(seoSection)
	}
}

func (b *Builder) detailShowComponent(obj interface{}, _ *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
	var (
		fieldPrefix string
		setting     Setting
		db          = b.db
		locale, _   = l10n.IsLocalizableFromContext(ctx.R.Context())
	)

	seo := b.GetSEO(obj)
	if seo == nil {
		return h.Div()
	}

	value := reflect.Indirect(reflect.ValueOf(obj))
	for i := 0; i < value.NumField(); i++ {
		if s, ok := value.Field(i).Interface().(Setting); ok {
			setting = s
			fieldPrefix = value.Type().Field(i).Name
		}
	}
	if !setting.EnabledCustomize {
		modelSetting := &QorSEOSetting{}
		db.Where("name = ? AND locale_code = ?", seo.name, locale).First(modelSetting)
		setting = modelSetting.Setting
	}

	return h.Div(
		b.vSeoReadonly(obj, fieldPrefix, locale, seo, &setting, ctx.R),
	).Class("pb-4")
}

func (b *Builder) detailSaver(obj interface{}, ctx *web.EventContext) (err error) {
	if err = EditSetterFunc(obj, &presets.FieldContext{Name: SeoDetailFieldName}, ctx); err != nil {
		return
	}
	return
}

func seoFieldPortal(label string, value string) h.HTMLComponent {
	return h.Div(
		VXLabel(
			h.Span(label).
				Style("line-height:20px; font-size:14px; font-weight:500;"),
		),
		h.Div(
			h.Span(value),
		).Class("pa-2 px-3 d-flex align-center gap-1"),
	).Class("mt-4")
}
