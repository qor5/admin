package examples_admin

import (
	"net/http"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	v "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

// @snippet_begin(I18nMessagesExample)
const I18nExampleKey i18n.ModuleKey = "I18nExampleKey"

type Messages struct {
	Admin   string
	Welcome string
}

var Messages_en_US = &Messages{
	Admin:   "Admin",
	Welcome: "Welcome",
}

var Messages_zh_CN = &Messages{
	Admin:   "管理系统",
	Welcome: "欢迎",
}

var Messages_ja_JP = &Messages{
	Admin:   "管理システム",
	Welcome: "ようこそ",
}

// @snippet_end

// @snippet_begin(I18nPresetsMessagesExample)

type Messages_ModelsI18nModuleKey struct {
	Homes             string
	Videos            string
	VideosName        string
	VideosDescription string
}

var Messages_zh_CN_ModelsI18nModuleKey = &Messages_ModelsI18nModuleKey{
	Homes:             "主页",
	Videos:            "视频",
	VideosName:        "视频名称",
	VideosDescription: "视频描述",
}

var Messages_ja_JP_ModelsI18nModuleKey = &Messages_ModelsI18nModuleKey{
	Homes:             "ホーム",
	Videos:            "ビデオ",
	VideosName:        "ビデオの名前",
	VideosDescription: "ビデオの説明",
}

// @snippet_end

type video struct {
	gorm.Model
	Name        string
	Description string
}

func InternationalizationExample(b *presets.Builder, db *gorm.DB) http.Handler {
	if err := db.AutoMigrate(&video{}); err != nil {
		panic(err)
	}

	b.DataOperator(gorm2op.DataOperator(db)).
		BrandFunc(func(ctx *web.EventContext) h.HTMLComponent {
			msgr := i18n.MustGetModuleMessages(ctx.R, I18nExampleKey, Messages_en_US).(*Messages)

			return v.VCardText(
				h.H1(msgr.Admin),
			).Class("pa-0")
		})

	type home struct{}
	b.Model(&home{}).URIName("home").MenuIcon("home").Listing().PageFunc(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		// @snippet_begin(I18nMustGetModuleMessages)
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nExampleKey, Messages_en_US).(*Messages)
		r.Body = v.VContainer(
			h.Div(
				h.H1(msgr.Welcome),
			).Class("text-center mt-8"),
		)
		// @snippet_end
		r.PageTitle = msgr.Welcome
		return
	}).Labels("Home")

	b.Model(&video{}).MenuIcon("movie")

	// @snippet_begin(I18nNew)
	i18nB := b.GetI18n()
	// @snippet_end

	// @snippet_begin(I18nSupportLanguages)
	i18nB.SupportLanguages(language.English, language.SimplifiedChinese, language.Japanese)
	// @snippet_end
	// @snippet_begin(I18nRegisterForModule)
	i18nB.
		RegisterForModule(language.SimplifiedChinese, presets.ModelsI18nModuleKey, Messages_zh_CN_ModelsI18nModuleKey).
		RegisterForModule(language.Japanese, presets.ModelsI18nModuleKey, Messages_ja_JP_ModelsI18nModuleKey).
		RegisterForModule(language.English, I18nExampleKey, Messages_en_US).
		RegisterForModule(language.Japanese, I18nExampleKey, Messages_ja_JP).
		RegisterForModule(language.SimplifiedChinese, I18nExampleKey, Messages_zh_CN).
		GetSupportLanguagesFromRequestFunc(func(r *http.Request) []language.Tag {
			return b.GetI18n().GetSupportLanguages()
		})
	// @snippet_end
	return b
}
