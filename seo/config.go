package seo

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/go-playground/form"
	"github.com/goplaid/web"
	"github.com/goplaid/x/i18n"
	"github.com/goplaid/x/perm"
	"github.com/goplaid/x/presets"
	. "github.com/goplaid/x/vuetify"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

const (
	saveCollectionEvent                = "seo_save_collection"
	I18nSeoKey          i18n.ModuleKey = "I18nSeoKey"
)

var permVerifier *perm.Verifier

func (collection *Collection) Configure(b *presets.Builder, db *gorm.DB) {
	err := db.AutoMigrate(
		collection.settingModel,
	)

	if err != nil {
		panic(err)
	}

	b.GetWebBuilder().RegisterEventFunc(saveCollectionEvent, saveCollection(collection, db))
	b.Model(collection.settingModel).PrimaryField("Name").Label("SEO").Listing().PageFunc(collection.pageFunc(db))

	b.FieldDefaults(presets.WRITE).
		FieldType(Setting{}).
		ComponentFunc(SeoEditingComponentFunc(collection))

	b.I18n().
		RegisterForModule(language.English, I18nSeoKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nSeoKey, Messages_zh_CN)

	permVerifier = perm.NewVerifier("seo", b.GetPermission())
}

func SeoEditingComponentFunc(collection *Collection) presets.FieldComponentFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nSeoKey, Messages_en_US).(*Messages)

		return collection.settingComponent(msgr, obj)
	}
}

func (collection *Collection) pageFunc(db *gorm.DB) web.PageFunc {
	return func(ctx *web.EventContext) (r web.PageResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nSeoKey, Messages_en_US).(*Messages)

		r.PageTitle = msgr.PageTitle
		r.Body = h.If(editIsAllowed(ctx.R) == nil, VContainer(
			VSnackbar(h.Text(msgr.SavedSuccessfully)).
				Attr("v-model", "vars.seoSnackbarShow").
				Top(true).
				Color("primary").
				Timeout(2000),
			VRow(
				VCol(
					VContainer(
						h.H3(msgr.SiteWideTitle).Style("font-weight: 500"),
						h.P().Text(msgr.SiteWideDescription)),
				).Cols(3),
				VCol(
					VCard(
						VForm(
							collection.renderGlobalSection(msgr, db),
						),
					).Outlined(true).Elevation(2),
				).Cols(9),
			),
			VRow(
				VCol(
					VContainer(
						h.H3(msgr.PageMetadataTitle).Attr("style", "font-weight: 500"),
						h.P().Text(msgr.PageMetadataDescription)),
				).Cols(3),
				VCol(
					VExpansionPanels(
						collection.renderSeoSections(msgr, db),
					).Focusable(true),
				).Cols(9),
			),
		).Attr("style", "background-color: #f5f5f5;max-width:100%").Attr(web.InitContextVars, `{seoSnackbarShow: false}`))

		return
	}
}

func (collection *Collection) renderGlobalSection(msgr *Messages, db *gorm.DB) h.HTMLComponent {
	setting := reflect.New(reflect.Indirect(reflect.ValueOf(collection.settingModel)).Type()).Interface().(QorSEOSettingInterface)
	err := db.Where("is_global_seo = ? AND name = ?", true, collection.Name).First(setting).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		setting.SetName(collection.Name)
		setting.SetSEOType(collection.Name)
		setting.SetIsGlobalSEO(true)
		db.Save(setting)
	}

	value := reflect.Indirect(reflect.ValueOf(collection.globalSetting))
	settingValue := setting.GetGlobalSetting()
	for i := 0; i < value.NumField(); i++ {
		fieldName := value.Type().Field(i).Name
		if settingValue[fieldName] != "" {
			value.Field(i).Set(reflect.ValueOf(settingValue[fieldName]))
		}
	}

	var comps h.HTMLComponents
	for i := 0; i < value.Type().NumField(); i++ {
		filed := value.Type().Field(i)
		comps = append(comps, VTextField().FieldName(fmt.Sprintf("%s.%s", collection.Name, filed.Name)).Label(msgr.DynamicMessage[filed.Name]).Value(value.Field(i).String()).Clearable(true))
	}

	return VForm(
		VCardText(
			comps,
		),

		VCardActions(
			VSpacer(),
			VBtn(msgr.Save).Bind("loading", "vars.global_seo_loading").Color("primary").Large(true).Attr("@click", web.Plaid().EventFunc(saveCollectionEvent, collection.Name, "global_seo_loading").BeforeScript(`vars.global_seo_loading = true;`).Go()),
		).Attr(web.InitContextVars, `{global_seo_loading: false}`),
	)
}

func (collection *Collection) renderSeoSections(msgr *Messages, db *gorm.DB) h.HTMLComponents {
	var comps h.HTMLComponents
	for _, seo := range collection.registeredSEO {
		setting := reflect.New(reflect.Indirect(reflect.ValueOf(collection.settingModel)).Type()).Interface().(QorSEOSettingInterface)
		err := db.Where("name = ?", seo.Name).First(setting).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			setting.SetName(seo.Name)
			setting.SetSEOType(seo.Name)
			db.Save(setting)
		}
		setting.SetCollection(collection)

		loadingName := strings.ToLower(seo.Name)
		loadingName = strings.ReplaceAll(loadingName, " ", "_")
		comp := VExpansionPanel(
			VExpansionPanelHeader(h.H4(msgr.DynamicMessage[seo.Name]).Style("font-weight: 500;")),
			VExpansionPanelContent(
				VCardText(
					collection.settingComponent(msgr, setting),
				),
				VCardActions(
					VSpacer(),
					VBtn(msgr.Save).Bind("loading", fmt.Sprintf("vars.%s", loadingName)).Color("primary").Large(true).Attr("@click", web.Plaid().EventFunc(saveCollectionEvent, seo.Name, loadingName).BeforeScript(fmt.Sprintf("vars.%s = true;", loadingName)).Go()),
				).Attr(web.InitContextVars, fmt.Sprintf(`{%s: false}`, loadingName)),
			),
		)

		comps = append(comps, comp)
	}

	return comps
}

func (collection *Collection) settingComponent(msgr *Messages, obj interface{}) h.HTMLComponents {
	var (
		fieldPrefix string
		isModel     bool
		seo         *SEO
		setting     Setting
	)

	if qorSeoSetting, ok := obj.(QorSEOSettingInterface); ok {
		fieldPrefix = qorSeoSetting.GetName()
		seo = collection.GetSEO(fieldPrefix)
		setting = qorSeoSetting.GetSEOSetting()
	} else {
		if seoInterface, ok := obj.(interface{ GetSEO() *SEO }); ok {
			isModel = true
			seo = seoInterface.GetSEO()
			value := reflect.Indirect(reflect.ValueOf(obj))
			for i := 0; i < value.NumField(); i++ {
				if s, ok := value.Field(i).Interface().(Setting); ok {
					setting = s
					fieldPrefix = value.Type().Field(i).Name
				}
			}
		} else {
			return nil
		}
	}

	var variables []string
	value := reflect.Indirect(reflect.ValueOf(collection.globalSetting)).Type()
	for i := 0; i < value.NumField(); i++ {
		fieldName := value.Field(i).Name
		variables = append(variables, fieldName)
	}
	variables = append(variables, seo.Variables...)

	var variablesEle []h.HTMLComponent
	for _, variable := range variables {
		variablesEle = append(variablesEle, VCol(
			VBtn("").Width(100).Icon(true).Children(VIcon("add_box"), h.Text(msgr.DynamicMessage[variable])),
		).Cols(2))
	}

	commonSettingComponent := h.Components(
		VRow(
			variablesEle...,
		),
		h.H6(msgr.Basic).Style("margin-top:15px;margin-bottom:15px;"),
		VCard(
			VCardText(
				VTextField().Counter(65).FieldName(fmt.Sprintf("%s.%s", fieldPrefix, "Title")).Label(msgr.Title).Value(setting.Title).Clearable(true),
				VTextField().Counter(150).FieldName(fmt.Sprintf("%s.%s", fieldPrefix, "Description")).Label(msgr.Description).Value(setting.Description).Clearable(true),
				VTextarea().Counter(255).Rows(2).AutoGrow(true).FieldName(fmt.Sprintf("%s.%s", fieldPrefix, "Keywords")).Label(msgr.Keywords).Value(setting.Keywords).Clearable(true),
			),
		).Color("#f5f5f5").Elevation(0),
		h.H6(msgr.OpenGraphInformation).Style("margin-top:15px;margin-bottom:15px;"),
		VCard(
			VCardText(
				VRow(
					VCol(VTextField().FieldName(fmt.Sprintf("%s.%s", fieldPrefix, "OpenGraphURL")).Label(msgr.OpenGraphURL).Value(setting.OpenGraphURL).Clearable(true)).Cols(6),
					VCol(VTextField().FieldName(fmt.Sprintf("%s.%s", fieldPrefix, "OpenGraphType")).Label(msgr.OpenGraphType).Value(setting.OpenGraphType).Clearable(true)).Cols(6),
				),
				VRow(
					VCol(VTextField().FieldName(fmt.Sprintf("%s.%s", fieldPrefix, "OpenGraphImageURL")).Label(msgr.OpenGraphImageURL).Value(setting.OpenGraphImageURL).Clearable(true)).Cols(6),
					VCol(VTextField().FieldName(fmt.Sprintf("%s.%s", fieldPrefix, "OpenGraphImageFromMediaLibrary")).Label(msgr.OpenGraphImage).Clearable(true)).Cols(6),
				),
			),
		).Color("#f5f5f5").Elevation(0),
	)

	if !isModel {
		return commonSettingComponent
	}

	return h.HTMLComponents{
		h.Label(msgr.Seo).Class("v-label theme--light"),
		VCard(
			VCardText(
				VSwitch().Label(msgr.UseDefaults).Attr("v-model", "locals.userDefaults").On("change", "locals.enabledCustomize = !locals.userDefaults;$refs.customize.$emit('change', locals.enabledCustomize)"),
				VSwitch().FieldName(fmt.Sprintf("%s.%s", fieldPrefix, "EnabledCustomize")).Label("EnabledCustomize").Attr(":input-value", "locals.enabledCustomize").Attr("ref", "customize").Attr("style", "display:none;"),
				h.Div(commonSettingComponent).Attr("v-show", "locals.userDefaults == false"),
			).Attr("style", "padding-bottom: 0px;padding-top: 1px;"),
		).Attr(web.InitContextLocals, fmt.Sprintf(`{enabledCustomize: %t, userDefaults: %t}`, setting.EnabledCustomize, !setting.EnabledCustomize)),
	}
}

func saveCollection(collection *Collection, db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		if len(ctx.Event.Params) != 2 {
			return
		}

		prefix := ctx.Event.Params[0]

		setting := reflect.New(reflect.Indirect(reflect.ValueOf(collection.settingModel)).Type()).Interface().(QorSEOSettingInterface)
		err = db.Where("name = ?", prefix).First(setting).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return
		}

		if setting.GetIsGlobalSEO() {
			globalSetting := make(map[string]string)
			for fieldWithPrefix := range ctx.R.Form {
				if strings.HasPrefix(fieldWithPrefix, prefix) {
					field := strings.Replace(fieldWithPrefix, fmt.Sprintf("%s.", prefix), "", -1)
					globalSetting[field] = ctx.R.Form.Get(fieldWithPrefix)
				}
			}
			setting.SetGlobalSetting(globalSetting)
		} else {
			values := url.Values{}
			for fieldWithPrefix := range ctx.R.Form {
				if strings.HasPrefix(fieldWithPrefix, prefix) {
					field := strings.Replace(fieldWithPrefix, fmt.Sprintf("%s.", prefix), "", -1)
					values[field] = []string{ctx.R.Form.Get(fieldWithPrefix)}
				}
			}
			s := setting.GetSEOSetting()
			decoder := form.NewDecoder()
			decoder.Decode(&s, values)
			setting.SetSEOSetting(s)
		}

		err = db.Save(setting).Error
		if err != nil {
			return
		}

		r.VarsScript = fmt.Sprintf(`vars.seoSnackbarShow = true;vars.%s = false;`, ctx.Event.Params[1])
		return
	}
}
