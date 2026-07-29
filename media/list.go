package media

import (
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	. "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"

	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/presets"
)

const (
	mediaLibraryListField = "media-library-list"
)

func configList(b *presets.Builder, mb *Builder) {
	mm := b.Model(&media_library.MediaLibrary{}).Label("Media Library").MenuIcon("mdi-image")
	mb.mb = mm
	configScopedCRUD(mb, mm)
	oldPageFunc := mm.Listing().GetPageFunc()
	mm.Listing().PageFunc(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		var (
			filed = mediaLibraryListField
			cfg   = &media_library.MediaBoxConfig{}
			msgr  = i18n.MustGetModuleMessages(ctx.R, I18nMediaLibraryKey, Messages_en_US).(*Messages)
		)
		var pr web.PageResponse
		if pr, err = oldPageFunc(ctx); err != nil {
			return
		}
		r.PageTitle = pr.PageTitle
		ctx.WithContextValue(presets.CtxPageTitleComponent, h.Div(
			VAppBarTitle(h.Text(msgr.MediaLibrary)),
			VSpacer(),
			searchComponent(ctx, filed, cfg, true),
		).Class(W100, "d-flex align-center"))
		r.Body = h.Components(
			h.Div(
				web.Portal(
					fileChooserDialogContent(mb, filed, ctx, cfg),
				).Name(dialogContentPortalName(filed)),
			),
		)
		return
	})
}

// configScopedCRUD makes the model's generic CRUD event funcs respect the
// searcher. They address rows by primary key through the DataOperator, so
// without these guards a request could edit, delete or read back any row by id
// regardless of what the searcher allows.
func configScopedCRUD(mb *Builder, mm *presets.ModelBuilder) {
	guard := func(id string, ctx *web.EventContext) error {
		visible, err := mb.recordIsVisible(ctx, id)
		if err != nil {
			return err
		}
		if visible {
			return nil
		}
		return presets.ErrRecordNotFound
	}
	mm.Editing().
		WrapFetchFunc(func(in presets.FetchFunc) presets.FetchFunc {
			return func(obj interface{}, id string, ctx *web.EventContext) (interface{}, error) {
				if err := guard(id, ctx); err != nil {
					return nil, err
				}
				return in(obj, id, ctx)
			}
		}).
		WrapSaveFunc(func(in presets.SaveFunc) presets.SaveFunc {
			return func(obj interface{}, id string, ctx *web.EventContext) error {
				if err := guard(id, ctx); err != nil {
					return err
				}
				return in(obj, id, ctx)
			}
		}).
		WrapDeleteFunc(func(in presets.DeleteFunc) presets.DeleteFunc {
			return func(obj interface{}, id string, ctx *web.EventContext) error {
				if err := guard(id, ctx); err != nil {
					return err
				}
				return in(obj, id, ctx)
			}
		})
	// GetDetailing, not Detailing: the latter would enable detailing for the
	// model, mounting a detail route and changing listing row behavior. The
	// detailing event funcs are registered either way, so they need the guard.
	mm.GetDetailing().WrapFetchFunc(func(in presets.FetchFunc) presets.FetchFunc {
		return func(obj interface{}, id string, ctx *web.EventContext) (interface{}, error) {
			if err := guard(id, ctx); err != nil {
				return nil, err
			}
			return in(obj, id, ctx)
		}
	})
	// The listing's generic search backs ListingCompo, whose reload / paging /
	// search actions stay dispatchable even though the media library replaces the
	// listing page. Restrict it to the ids the searcher exposes; passing the
	// scoped query as a subquery keeps a Joins-based searcher intact.
	mm.Listing().WrapSearchFunc(func(in presets.SearchFunc) presets.SearchFunc {
		return func(ctx *web.EventContext, params *presets.SearchParams) (*presets.SearchResult, error) {
			if mb.searcher != nil {
				params.SQLConditions = append(params.SQLConditions, &presets.SQLCondition{
					Query: qualified("id") + " in (?)",
					Args:  []interface{}{mb.scopedDB(mb.db, ctx).Select(qualified("id"))},
				})
			}
			return in(ctx, params)
		}
	})
}
