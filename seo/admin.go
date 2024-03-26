package seo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/qor5/admin/l10n"
	"github.com/qor5/admin/media"
	"github.com/qor5/admin/media/media_library"
	"github.com/qor5/admin/media/views"
	"github.com/qor5/admin/presets"
	. "github.com/qor5/ui/vuetify"
	"github.com/qor5/web"
	"github.com/qor5/x/i18n"
	"github.com/qor5/x/perm"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
)

const (
	I18nSeoKey i18n.ModuleKey = "I18nSeoKey"
)

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

func (b *Builder) Configure(pb *presets.Builder) (seoModel *presets.ModelBuilder) {
	// The registration of FieldDefaults for writing Setting here
	// must be executed before `pb.Model(&QorSEOSetting{})...`,
	pb.FieldDefaults(presets.WRITE).
		FieldType(Setting{}).
		ComponentFunc(b.EditingComponentFunc).
		SetterFunc(EditSetterFunc)

	seoModel = pb.Model(&QorSEOSetting{}).PrimaryField("Name").
		Label("SEO").
		RightDrawerWidth("1000").
		LayoutConfig(&presets.LayoutConfig{
			SearchBoxInvisible:          true,
			NotificationCenterInvisible: true,
		})

	// Configure Listing Page
	b.configListing(seoModel)
	// Configure Editing Page
	b.configEditing(seoModel)

	pb.I18n().
		RegisterForModule(language.English, I18nSeoKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nSeoKey, Messages_zh_CN)

	pb.ExtraAsset("/vue-seo.js", "text/javascript", SeoJSComponentsPack())
	permVerifier = perm.NewVerifier("seo", pb.GetPermission())
	return
}

func (b *Builder) configListing(seoModel *presets.ModelBuilder) {
	listing := seoModel.Listing("Name").Title("SEO")
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
			icon := "folder"
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

	oldSearcher := listing.Searcher
	listing.SearchFunc(func(model interface{}, params *presets.SearchParams, ctx *web.EventContext) (r interface{}, totalCount int, err error) {
		locale, _ := l10n.IsLocalizableFromCtx(ctx.R.Context())
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
		r, totalCount, err = oldSearcher(model, params, ctx)
		if totalCount == 0 {
			panic("The localization of SEO is not configured correctly. " +
				"Please check if you correctly configured the `WithLocales` option when initializing the SEO Builder.")
		}
		b.SortSEOs(r.([]*QorSEOSetting))
		return
	})
}

func (b *Builder) configEditing(seoModel *presets.ModelBuilder) {
	editing := seoModel.Editing("Variables", "Setting")

	// Customize the Saver to trigger the invocation of the `afterSave` hook function (if available)
	// when updating the global seo.
	editing.Saver = func(obj interface{}, id string, ctx *web.EventContext) (err error) {
		seoSetting := obj.(*QorSEOSetting)
		if err = b.db.Updates(obj).Error; err != nil {
			return err
		}
		if b.afterSave != nil {
			if err = b.afterSave(ctx.R.Context(), seoSetting.Name, seoSetting.LocaleCode); err != nil {
				return err
			}
		}
		return nil
	}

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
							Attr(web.VField(fmt.Sprintf("%s.%s", formKeyForVariablesField, varName), seoSetting.Variables[varName])).
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
			return b.vseo("Setting", b.GetSEO(seoSetting.Name), &seoSetting.Setting, ctx.R)
		},
	)
}

func EditSetterFunc(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
	var setting Setting
	var mediaBox = media_library.MediaBox{}
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

func (b *Builder) EditingComponentFunc(obj interface{}, _ *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
	var (
		msgr        = i18n.MustGetModuleMessages(ctx.R, I18nSeoKey, Messages_en_US).(*Messages)
		fieldPrefix string
		setting     Setting
		db          = b.db
		locale, _   = l10n.IsLocalizableFromCtx(ctx.R.Context())
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

	hideActions := false
	if ctx.R.FormValue("hideActionsIconForSEOForm") == "true" {
		hideActions = true
	}
	openCustomizePanel := 1
	if setting.EnabledCustomize {
		openCustomizePanel = 0
	}

	return web.Scope(
		h.Div(
			h.Label(msgr.Seo).Class("v-label theme--light"),
			VExpansionPanels(
				VExpansionPanel(

					VExpansionPanelHeader(
						VSwitch().
							Label(msgr.Customize).Attr("ref", "switchComp").
							Bind("input-value", "locals.enabledCustomize").
							Attr(web.VField(fmt.Sprintf("%s.%s", fieldPrefix, "EnabledCustomize"), setting.EnabledCustomize)...),
					).
						Attr("style", "padding: 0px 24px;").HideActions(hideActions).
						Attr("@click", "locals.enabledCustomize=!locals.enabledCustomize;$refs.switchComp.$emit('change', locals.enabledCustomize)"),
					VExpansionPanelContent(
						VCardText(
							b.vseo(fieldPrefix, seo, &setting, ctx.R),
						),
					).Eager(true),
				),
			).Flat(true).Attr("v-model", "locals.openCustomizePanel"),
		).Class("pb-4"),
	).Init(fmt.Sprintf(`{enabledCustomize: %t, openCustomizePanel: %d}`, setting.EnabledCustomize, openCustomizePanel)).
		VSlot("{ locals }")
}

func (b *Builder) vseo(fieldPrefix string, seo *SEO, setting *Setting, req *http.Request) h.HTMLComponent {
	var (
		msgr = i18n.MustGetModuleMessages(req, I18nSeoKey, Messages_en_US).(*Messages)
		db   = b.db
	)

	var varComps []h.HTMLComponent
	for varName := range seo.getAvailableVars() {
		varComps = append(varComps,
			VChip(
				VIcon("add_box").Class("mr-2"),
				h.Text(i18n.PT(req, presets.ModelsI18nModuleKey, "Seo Variable", varName)),
			).Attr("@click", fmt.Sprintf("$refs.seo.addTags('%s')", varName)).Label(true).Variant(VariantOutlined),
		)
	}
	var variablesEle []h.HTMLComponent
	variablesEle = append(variablesEle, VChipGroup(varComps...).Column(true).Class("ma-4"))

	image := &setting.OpenGraphImageFromMediaLibrary
	if image.ID.String() == "0" {
		image.ID = json.Number("")
	}
	refPrefix := strings.ReplaceAll(strings.ToLower(fieldPrefix), " ", "_")
	return VSeo(
		h.H4(msgr.Basic).Style("margin-top:15px;font-weight: 500"),
		VRow(
			variablesEle...,
		),
		VCard(
			VCardText(
				VTextField().Attr("counter", true).FieldName(fmt.Sprintf("%s.%s", fieldPrefix, "Title")).Label(msgr.Title).Value(setting.Title).Attr("@focus", fmt.Sprintf("$refs.seo.tagInputsFocus($refs.%s)", fmt.Sprintf("%s_title", refPrefix))).Attr("ref", fmt.Sprintf("%s_title", refPrefix)),
				VTextField().Attr("counter", true).FieldName(fmt.Sprintf("%s.%s", fieldPrefix, "Description")).Label(msgr.Description).Value(setting.Description).Attr("@focus", fmt.Sprintf("$refs.seo.tagInputsFocus($refs.%s)", fmt.Sprintf("%s_description", refPrefix))).Attr("ref", fmt.Sprintf("%s_description", refPrefix)),
				VTextarea().Attr("counter", true).Rows(2).AutoGrow(true).FieldName(fmt.Sprintf("%s.%s", fieldPrefix, "Keywords")).Label(msgr.Keywords).Value(setting.Keywords).Attr("@focus", fmt.Sprintf("$refs.seo.tagInputsFocus($refs.%s)", fmt.Sprintf("%s_keywords", refPrefix))).Attr("ref", fmt.Sprintf("%s_keywords", refPrefix)),
			),
		).Variant(VariantOutlined).Flat(true),

		h.H4(msgr.OpenGraphInformation).Style("margin-top:15px;margin-bottom:15px;font-weight: 500"),
		VCard(
			VCardText(
				VRow(
					VCol(VTextField().FieldName(fmt.Sprintf("%s.%s", fieldPrefix, "OpenGraphTitle")).Label(msgr.OpenGraphTitle).Value(setting.OpenGraphTitle).Attr("@focus", fmt.Sprintf("$refs.seo.tagInputsFocus($refs.%s)", fmt.Sprintf("%s_og_title", refPrefix))).Attr("ref", fmt.Sprintf("%s_og_title", refPrefix))).Cols(6),
					VCol(VTextField().FieldName(fmt.Sprintf("%s.%s", fieldPrefix, "OpenGraphDescription")).Label(msgr.OpenGraphDescription).Value(setting.OpenGraphDescription).Attr("@focus", fmt.Sprintf("$refs.seo.tagInputsFocus($refs.%s)", fmt.Sprintf("%s_og_description", refPrefix))).Attr("ref", fmt.Sprintf("%s_og_description", refPrefix))).Cols(6),
				),
				VRow(
					VCol(VTextField().FieldName(fmt.Sprintf("%s.%s", fieldPrefix, "OpenGraphURL")).Label(msgr.OpenGraphURL).Value(setting.OpenGraphURL).Attr("@focus", fmt.Sprintf("$refs.seo.tagInputsFocus($refs.%s)", fmt.Sprintf("%s_og_url", refPrefix))).Attr("ref", fmt.Sprintf("%s_og_url", refPrefix))).Cols(6),
					VCol(VTextField().FieldName(fmt.Sprintf("%s.%s", fieldPrefix, "OpenGraphType")).Label(msgr.OpenGraphType).Value(setting.OpenGraphType).Attr("@focus", fmt.Sprintf("$refs.seo.tagInputsFocus($refs.%s)", fmt.Sprintf("%s_og_type", refPrefix))).Attr("ref", fmt.Sprintf("%s_og_type", refPrefix))).Cols(6),
				),
				VRow(
					VCol(VTextField().FieldName(fmt.Sprintf("%s.%s", fieldPrefix, "OpenGraphImageURL")).Label(msgr.OpenGraphImageURL).Value(setting.OpenGraphImageURL).Attr("@focus", fmt.Sprintf("$refs.seo.tagInputsFocus($refs.%s)", fmt.Sprintf("%s_og_imageurl", refPrefix))).Attr("ref", fmt.Sprintf("%s_og_imageurl", refPrefix))).Cols(12),
				),
				VRow(
					VCol(views.QMediaBox(db).Label(msgr.OpenGraphImage).
						FieldName(fmt.Sprintf("%s.%s", fieldPrefix, "OpenGraphImageFromMediaLibrary")).
						Value(image).
						Config(&media_library.MediaBoxConfig{
							AllowType: "image",
							Sizes: map[string]*media.Size{
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
				VRow(
					VCol(VTextarea().FieldName(fmt.Sprintf("%s.%s", fieldPrefix, "OpenGraphMetadataString")).Label(msgr.OpenGraphMetadata).Value(GetOpenGraphMetadataString(setting.OpenGraphMetadata)).Attr("@focus", fmt.Sprintf("$refs.seo.tagInputsFocus($refs.%s)", fmt.Sprintf("%s_og_metadata", refPrefix))).Attr("ref", fmt.Sprintf("%s_og_metadata", refPrefix))).Cols(12),
				),
			),
		).Variant(VariantOutlined).Flat(true),
	).Attr("ref", "seo")
}
