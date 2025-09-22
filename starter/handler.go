package starter

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/pkg/errors"
	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/common"
	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/media"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/admin/v3/role"
	"github.com/qor5/admin/v3/utils"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/login"
	"github.com/qor5/x/v3/oss/s3"
	"github.com/qor5/x/v3/perm"
	"github.com/theplant/inject"
	"golang.org/x/text/language"
	"gorm.io/gorm"

	_ "embed"

	plogin "github.com/qor5/admin/v3/login"
	media_oss "github.com/qor5/admin/v3/media/oss"
	v "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
)

type S3Config struct {
	Bucket   string `confx:"bucket"`
	Region   string `confx:"region"`
	Endpoint string `confx:"endpoint"`
}

// Config contains all dependencies needed for Handler
type Config struct {
	DB *gorm.DB `inject:"" confx:"-"`

	Prefix    string     `confx:"prefix" validate:"omitempty"`
	S3        S3Config   `confx:"s3"`
	S3Publish S3Config   `confx:"s3Publish"`
	Auth      AuthConfig `confx:"auth"`
}

// Handler handles admin with embedded configuration
type Handler struct {
	*Config
	*inject.Injector

	plugins     []presets.Plugin
	handlerHook common.Hook[http.Handler]
	warmupOnce  sync.Once
	handler     http.Handler
}

// NewHandler creates a new Handler with the provided configuration
func NewHandler(cfg *Config) *Handler {
	cfg.Prefix = strings.TrimRight(cfg.Prefix, "/")
	handler := &Handler{
		Config:   cfg,
		Injector: inject.New(),
	}
	_ = handler.Provide(func() *Handler { return handler })
	return handler
}

// Build initializes all components and sets up the admin interface
func (a *Handler) Build(ctx context.Context, ctors ...any) error {
	if err := a.Provide(
		a.createPresetsBuilder,
		a.createActivityBuilder,
		a.createMediaBuilder,
		a.createL10nBuilder,
		a.createRoleBuilder,
		a.createLoginBuilder,
		a.createLoginSessionBuilder,
		a.createProfileBuilder,
		a.createUserModelBuilder,
		a.createMux,
	); err != nil {
		return err
	}

	if len(ctors) > 0 {
		if err := a.Provide(ctors...); err != nil {
			return err
		}
	}

	if err := a.ApplyContext(ctx, a.Config); err != nil {
		return err
	}

	if err := a.autoMigrate(ctx); err != nil {
		return err
	}

	a.configureMediaStorage()

	if err := a.BuildContext(ctx); err != nil {
		return err
	}

	if a.Auth.InitialUserEmail != "" && a.Auth.InitialUserPassword != "" && a.Auth.InitialUserRole != "" {
		if _, err := createInitialUserIfEmpty(ctx, a.DB, &UpsertUserOptions{
			Email:    a.Auth.InitialUserEmail,
			Password: a.Auth.InitialUserPassword,
			Role:     []string{a.Auth.InitialUserRole},
		}); err != nil {
			return err
		}
	}

	return nil
}

func (a *Handler) WithHandlerHook(hooks ...common.Hook[http.Handler]) *Handler {
	a.handlerHook = common.ChainHookWith(a.handlerHook, hooks...)
	return a
}

func (a *Handler) Use(plugins ...presets.Plugin) {
	a.plugins = append(a.plugins, plugins...)
}

// autoMigrate performs database migrations
func (a *Handler) autoMigrate(ctx context.Context) error {
	db := a.DB.WithContext(ctx)
	if err := db.AutoMigrate(
		&role.Role{},
		&User{},
		&perm.DefaultDBPolicy{},
	); err != nil {
		return errors.Wrap(err, "failed to auto migrate database")
	}

	if err := createDefaultRolesIfEmpty(ctx, db); err != nil {
		return errors.Wrap(err, "failed to initialize default roles")
	}
	return nil
}

// configureMediaStorage configures S3 storage for media
func (a *Handler) configureMediaStorage() {
	media_oss.Storage = s3.New(&s3.Config{
		Bucket:   a.S3.Bucket,
		Region:   a.S3.Region,
		ACL:      string(types.ObjectCannedACLBucketOwnerFullControl),
		Endpoint: a.S3.Endpoint,
	})
}

// createActivityBuilder creates and configures the activity builder
func (a *Handler) createActivityBuilder() *activity.Builder {
	activityBuilder := activity.New(a.DB, func(ctx context.Context) (*activity.User, error) {
		u := ctx.Value(login.UserKey).(*User)
		return &activity.User{
			ID:     fmt.Sprint(u.ID),
			Name:   u.Name,
			Avatar: "",
		}, nil
	}).WrapLogModelInstall(func(in presets.ModelInstallFunc) presets.ModelInstallFunc {
		return func(pb *presets.Builder, mb *presets.ModelBuilder) (err error) {
			err = in(pb, mb)
			if err != nil {
				return
			}
			mb.Listing().WrapSearchFunc(func(in presets.SearchFunc) presets.SearchFunc {
				return func(ctx *web.EventContext, params *presets.SearchParams) (result *presets.SearchResult, err error) {
					u := GetCurrentUser(ctx.R)
					if rs := u.GetRoles(); !slices.Contains(rs, RoleAdmin) {
						params.SQLConditions = append(params.SQLConditions, &presets.SQLCondition{
							Query: "user_id = ?",
							Args:  []any{fmt.Sprint(u.ID)},
						})
					}
					return in(ctx, params)
				}
			})
			return
		}
	}).AutoMigrate()

	a.Use(activityBuilder)
	return activityBuilder
}

// createPresetsBuilder creates and configures the main presets builder
func (a *Handler) createPresetsBuilder() *presets.Builder {
	presetsBuilder := presets.New().
		URIPrefix(a.Prefix).
		DataOperator(gorm2op.DataOperator(a.DB))

	// Configure basic UI
	presetsBuilder.BrandFunc(func(_ *web.EventContext) h.HTMLComponent {
		logo := "https://qor5.com/img/qor-logo.png"
		return h.Div(
			v.VRow(
				v.VCol(h.A(h.Img(logo).Attr("width", "80")).Href("/")),
			),
		).Class("mb-n4 mt-n2")
	}).HomePageFunc(func(_ *web.EventContext) (r web.PageResponse, err error) {
		r.PageTitle = "Home"
		r.Body = h.H1("Home")
		return
	}).NotFoundPageLayoutConfig(&presets.LayoutConfig{
		NotificationCenterInvisible: true,
	}).RightDrawerWidth("700")

	// Configure permissions
	a.configurePermission(presetsBuilder)

	// Configure i18n
	a.configureI18n(presetsBuilder)

	// TODO: @molon 需要 hook ？
	presetsBuilder.MenuOrder(
		presetsBuilder.MenuGroup("Page Builder").SubItems(
			"Page",
			"shared_containers",
			"demo_containers",
			"page_templates",
			"page_categories",
		).Icon("mdi-view-quilt"),
		presetsBuilder.MenuGroup("User Management").SubItems(
			"User",
			"Role",
		).Icon("mdi-account-multiple"),
	)

	return presetsBuilder
}

// configureI18n configures i18n support
func (a *Handler) configureI18n(presetsBuilder *presets.Builder) {
	utils.Install(presetsBuilder)
	presetsBuilder.GetI18n().
		SupportLanguages(language.English, language.SimplifiedChinese, language.Japanese).
		RegisterForModule(language.English, presets.ModelsI18nModuleKey, Messages_en_US_ModelsI18nModuleKey).
		RegisterForModule(language.SimplifiedChinese, presets.ModelsI18nModuleKey, Messages_zh_CN_ModelsI18nModuleKey).
		RegisterForModule(language.Japanese, presets.ModelsI18nModuleKey, Messages_ja_JP_ModelsI18nModuleKey).
		RegisterForModule(language.English, I18nDemoKey, Messages_en_US).
		RegisterForModule(language.Japanese, I18nDemoKey, Messages_ja_JP).
		RegisterForModule(language.SimplifiedChinese, I18nDemoKey, Messages_zh_CN).
		GetSupportLanguagesFromRequestFunc(func(_ *http.Request) []language.Tag {
			return presetsBuilder.GetI18n().GetSupportLanguages()
		})
}

// createMediaBuilder creates and configures the media builder
func (a *Handler) createMediaBuilder() *media.Builder {
	mediaBuilder := media.New(a.DB).CurrentUserID(func(ctx *web.EventContext) (id uint) {
		u := GetCurrentUser(ctx.R)
		if u == nil {
			return
		}
		return u.ID
	}).Searcher(func(db *gorm.DB, ctx *web.EventContext) *gorm.DB {
		u := GetCurrentUser(ctx.R)
		if u == nil {
			return db
		}
		if rs := u.GetRoles(); !slices.Contains(rs, RoleAdmin) && !slices.Contains(rs, RoleManager) {
			return db.Where("user_id = ?", u.ID)
		}
		return db
	}).AutoMigrate()

	a.Use(mediaBuilder)
	return mediaBuilder
}

// createL10nBuilder creates and configures the localization builder
func (a *Handler) createL10nBuilder(activityBuilder *activity.Builder) *l10n.Builder {
	l10nBuilder := l10n.New(a.DB).
		Activity(activityBuilder).
		RegisterLocales("China", "cn", "China", l10n.ChinaSvg).
		RegisterLocales("Japan", "jp", "Japan", l10n.JapanSvg)

	l10nBuilder.SupportLocalesFunc(func(_ *http.Request) []string {
		return l10nBuilder.GetSupportLocaleCodes()
	})

	a.Use(l10nBuilder)
	return l10nBuilder
}

//go:embed embed/favicon.ico
var favicon []byte

func (a *Handler) createMux(presetsBuilder *presets.Builder, loginSessionBuilder *plogin.SessionBuilder) *http.ServeMux {
	mux := http.NewServeMux()
	loginSessionBuilder.Mount(mux)
	mux.Handle("/", presetsBuilder)
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(favicon)
	})
	return mux
}

// ServeHTTP implements http.Handler interface
func (a *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.warmupOnce.Do(func() {
		if _, err := a.Invoke(func(mux *http.ServeMux, presetsBuilder *presets.Builder, loginSessionBuilder *plogin.SessionBuilder) {
			presetsBuilder.Use(a.plugins...)
			presetsBuilder.Build()

			handlerHook := common.ChainHook(
				// TODO: @molon 可能要整理一下
				loginSessionBuilder.Middleware(),
				withRoles(a.DB),
				securityMiddleware(),
			)
			if a.handlerHook != nil {
				handlerHook = common.ChainHookWith(handlerHook, a.handlerHook)
			}
			a.handler = handlerHook(mux)
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	if a.handler == nil {
		http.Error(w, "handler not initialized", http.StatusInternalServerError)
		return
	}
	a.handler.ServeHTTP(w, r)
}
