package pagebuilder

import (
	"fmt"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/perm"
	. "github.com/qor5/x/v3/ui/vuetify"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/publish"
)

// EventDuplicateAsNew copies the given version of a record into a brand-new record
// (new ID, draft), unlike publish.EventDuplicateVersion which adds a version under the same ID.
const EventDuplicateAsNew = "pagebuilder_EventDuplicateAsNew"

func (b *ModelBuilder) configDuplicateAsNew() {
	b.mb.RegisterEventFunc(EventDuplicateAsNew, b.duplicateAsNew)
	b.mb.Listing().RowMenu().RowMenuItem("Duplicate As New").
		ComponentFunc(func(obj interface{}, id string, ctx *web.EventContext) h.HTMLComponent {
			if !b.canDuplicateAsNew(obj, ctx) {
				return nil
			}
			msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
			return VListItem(VListItemTitle(h.Text(msgr.DuplicateAsNew))).
				PrependIcon("mdi-content-copy").
				Attr("@click", b.duplicateAsNewConfirm(ctx, id))
		})
}

func (b *ModelBuilder) canDuplicateAsNew(obj interface{}, ctx *web.EventContext) bool {
	if b.isTemplate {
		return false
	}
	if _, ok := obj.(publish.VersionInterface); !ok {
		return false
	}
	return !publish.DeniedDo(b.mb.Info().Verifier(), obj, ctx.R, presets.PermCreate)
}

func (b *ModelBuilder) duplicateAsNewConfirm(ctx *web.EventContext, slug string) string {
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
	return web.Plaid().
		EventFunc(presets.OpenConfirmDialog).
		Query(presets.ConfirmDialogConfirmEvent,
			web.Plaid().URL(b.mb.Info().ListingHref()).
				EventFunc(EventDuplicateAsNew).
				Query(presets.ParamID, slug).
				Go(),
		).
		Query(presets.ConfirmDialogPromptText, msgr.ConfirmDuplicateAsNew).
		Go()
}

func (b *ModelBuilder) duplicateAsNewButton(ctx *web.EventContext, obj interface{}, phraseHasPresetsDataChanged string) h.HTMLComponent {
	if !b.canDuplicateAsNew(obj, ctx) {
		return nil
	}
	slugger, ok := obj.(presets.SlugEncoder)
	if !ok {
		return nil
	}
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
	return VBtn(msgr.DuplicateAsNew).
		Height(36).Class("ml-2").Variant(VariantOutlined).
		Attr(":disabled", phraseHasPresetsDataChanged).
		Attr("@click", b.duplicateAsNewConfirm(ctx, slugger.PrimarySlug()))
}

func (b *ModelBuilder) duplicateAsNew(ctx *web.EventContext) (r web.EventResponse, err error) {
	defer func() {
		if err != nil {
			presets.ShowMessage(&r, err.Error(), "error")
			err = nil
		}
	}()

	obj := b.mb.NewModel()
	if obj, err = b.mb.Editing().Fetcher(obj, ctx.Param(presets.ParamID), ctx); err != nil {
		return
	}
	if !b.canDuplicateAsNew(obj, ctx) {
		return r, perm.PermissionDenied
	}

	var (
		fromID      = reflectutils.MustGet(obj, "ID").(uint)
		fromVersion = obj.(publish.VersionInterface).EmbedVersion().Version
		locale      string
	)
	if l, ok := obj.(l10n.LocaleInterface); ok {
		locale = l.EmbedLocale().LocaleCode
	}

	if err = b.resetForDuplicateAsNew(obj); err != nil {
		return
	}
	if p, ok := obj.(*Page); ok {
		if p.Slug, err = b.copySlug(p); err != nil {
			return
		}
	}

	if err = b.mb.Editing().Creating().Saver(obj, "", ctx); err != nil {
		return
	}
	toID := reflectutils.MustGet(obj, "ID").(uint)
	toVersion := obj.(publish.VersionInterface).EmbedVersion().Version

	if err = b.db.Transaction(func(tx *gorm.DB) error {
		return b.copyContainersAsNew(tx, fromID, fromVersion, toID, toVersion, locale)
	}); err != nil {
		// containers failed: drop the half-created record so no empty copy is left behind
		if delErr := b.deleteDuplicated(toID, toVersion, locale); delErr != nil {
			err = fmt.Errorf("%w (cleanup failed: %v)", err, delErr)
		}
		return
	}

	slug := obj.(presets.SlugEncoder).PrimarySlug()
	r.Emit(b.mb.NotifModelsCreated(), presets.PayloadModelsCreated{Models: []any{obj}})
	r.PushState = web.Location(nil).URL(b.mb.Info().DetailingHref(slug))
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
	presets.ShowMessage(&r, msgr.DuplicatedAsNew, "")
	return
}

// resetForDuplicateAsNew turns a fetched record into an unsaved draft with a fresh first version.
func (b *ModelBuilder) resetForDuplicateAsNew(obj interface{}) error {
	if err := reflectutils.Set(obj, "ID", uint(0)); err != nil {
		return err
	}
	version := b.db.NowFunc().Local().Format("2006-01-02") + "-v01"
	*obj.(publish.VersionInterface).EmbedVersion() = publish.Version{Version: version, VersionName: version}
	if s, ok := obj.(publish.StatusInterface); ok {
		*s.EmbedStatus() = publish.Status{Status: publish.StatusDraft}
	}
	if s, ok := obj.(publish.ScheduleInterface); ok {
		*s.EmbedSchedule() = publish.Schedule{}
	}
	for _, field := range []string{"CreatedAt", "UpdatedAt"} {
		if _, err := reflectutils.Get(obj, field); err == nil {
			if err = reflectutils.Set(obj, field, time.Time{}); err != nil {
				return err
			}
		}
	}
	return nil
}

// copySlug returns "<slug>-copy", or "<slug>-copyN" (N = 1, 2, ...) when taken,
// checked against the same publish-URL rule pageValidator uses.
func (b *ModelBuilder) copySlug(p *Page) (string, error) {
	var localePath string
	if b.builder.l10n != nil {
		localePath = b.builder.l10n.GetLocalePath(p.LocaleCode)
	}
	category, err := p.GetCategory(b.db)
	if err != nil {
		return "", err
	}

	var infos []pagePathInfo
	if err = b.db.Raw(queryLocaleCodeCategoryPathSlugSQL).Scan(&infos).Error; err != nil {
		return "", err
	}
	taken := make(map[string]bool, len(infos))
	for _, info := range infos {
		var lp string
		if b.builder.l10n != nil {
			lp = b.builder.l10n.GetLocalePath(info.LocaleCode)
		}
		taken[generatePublishUrl(lp, info.CategoryPath, info.Slug)] = true
	}

	base := strings.TrimSuffix(path.Clean("/"+p.Slug), "/")
	for i := 0; ; i++ {
		candidate := base + "-copy"
		if base == "" {
			candidate = "/copy"
		}
		if i > 0 {
			candidate = fmt.Sprintf("%s%d", candidate, i)
		}
		if !taken[generatePublishUrl(localePath, category.Path, candidate)] {
			return candidate, nil
		}
	}
}

// copyContainersAsNew deep-copies every container of the source version, including shared ones,
// which become normal containers owned by the copy so that editing either record never affects the other.
func (b *ModelBuilder) copyContainersAsNew(tx *gorm.DB, fromID uint, fromVersion string, toID uint, toVersion, locale string) error {
	var cons []*Container
	if err := tx.Order("display_order ASC").
		Find(&cons, "page_id = ? AND page_version = ? AND locale_code = ? AND page_model_name = ?", fromID, fromVersion, locale, b.name).
		Error; err != nil {
		return err
	}
	builders := b.getContainerBuilders()
	for _, c := range cons {
		idx := slices.IndexFunc(builders, func(cb *ContainerBuilder) bool { return cb.name == c.ModelName })
		if idx < 0 {
			continue
		}
		model := builders[idx].NewModel()
		if err := tx.First(model, "id = ?", c.ModelID).Error; err != nil {
			return err
		}
		if err := reflectutils.Set(model, "ID", uint(0)); err != nil {
			return err
		}
		if err := tx.Create(model).Error; err != nil {
			return err
		}
		if err := tx.Create(&Container{
			PageID:        toID,
			PageVersion:   toVersion,
			PageModelName: b.name,
			ModelName:     c.ModelName,
			ModelID:       reflectutils.MustGet(model, "ID").(uint),
			DisplayOrder:  c.DisplayOrder,
			DisplayName:   c.DisplayName,
			Hidden:        c.Hidden,
			Locale:        l10n.Locale{LocaleCode: locale},
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (b *ModelBuilder) deleteDuplicated(id uint, version, locale string) error {
	g := b.db.Unscoped().Where("id = ? AND version = ?", id, version)
	if _, ok := b.mb.NewModel().(l10n.LocaleInterface); ok {
		g = g.Where("locale_code = ?", locale)
	}
	return g.Delete(b.mb.NewModel()).Error
}
