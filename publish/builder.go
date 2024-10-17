package publish

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/iancoleman/strcase"
	"github.com/qor/oss"
	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/utils"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type (
	PublishFunc   func(ctx context.Context, record any) error
	UnPublishFunc func(ctx context.Context, record any) error
)

type Builder struct {
	db                *gorm.DB
	storage           oss.StorageInterface
	ab                *activity.Builder
	ctxValueProviders []ContextValueFunc
	afterInstallFuncs []func()
	autoSchedule      bool

	publish   PublishFunc
	unpublish UnPublishFunc
}

type ContextValueFunc func(ctx context.Context) context.Context

func New(db *gorm.DB, storage oss.StorageInterface) *Builder {
	b := &Builder{
		db:      db,
		storage: storage,
	}
	b.publish = b.defaultPublish
	b.unpublish = b.defaultUnPublish
	return b
}

func (b *Builder) Activity(v *activity.Builder) (r *Builder) {
	b.ab = v
	return b
}

func (b *Builder) AutoSchedule(v bool) (r *Builder) {
	b.autoSchedule = v
	return b
}

func (b *Builder) AfterInstall(f func()) *Builder {
	b.afterInstallFuncs = append(b.afterInstallFuncs, f)
	return b
}

func (b *Builder) ModelInstall(pb *presets.Builder, m *presets.ModelBuilder) error {
	db := b.db

	obj := m.NewModel()
	_ = obj.(presets.SlugEncoder)
	_ = obj.(presets.SlugDecoder)

	if model, ok := obj.(VersionInterface); ok {
		if schedulePublishModel, ok := model.(ScheduleInterface); ok {
			VersionPublishModels[m.Info().URIName()] = reflect.ValueOf(schedulePublishModel).Elem().Interface()
		}

		b.configVersionAndPublish(pb, m, db)
	} else {
		if schedulePublishModel, ok := obj.(ScheduleInterface); ok {
			NonVersionPublishModels[m.Info().URIName()] = reflect.ValueOf(schedulePublishModel).Elem().Interface()
		}
	}

	if model, ok := obj.(ListInterface); ok {
		if schedulePublishModel, ok := model.(ScheduleInterface); ok {
			ListPublishModels[m.Info().URIName()] = reflect.ValueOf(schedulePublishModel).Elem().Interface()
		}
	}

	if _, ok := obj.(StatusInterface); ok {
		m.Editing().WrapSaveFunc(func(in presets.SaveFunc) presets.SaveFunc {
			return func(obj interface{}, id string, ctx *web.EventContext) (err error) {
				if status := EmbedStatus(obj); status.Status == "" {
					status.Status = StatusDraft
				}
				return in(obj, id, ctx)
			}
		})
		if m.HasDetailing() {
			detailFields := m.Detailing().FieldsBuilder
			for _, detailName := range detailFields.FieldNames() {
				if _, ok := m.Detailing().GetField(detailName.(string)).GetComponent().(*presets.SectionBuilder); !ok {
					continue
				}
				detailField := m.Detailing().GetField(detailName.(string)).GetComponent().(*presets.SectionBuilder)
				wrapper := func(in presets.ObjectBoolFunc) presets.ObjectBoolFunc {
					return func(obj interface{}, ctx *web.EventContext) bool {
						return in(obj, ctx) && EmbedStatus(obj).Status == StatusDraft
					}
				}
				detailField.WrapComponentEditBtnFunc(wrapper)
				detailField.WrapComponentHoverFunc(wrapper)
				detailField.WrapElementEditBtnFunc(wrapper)
				detailField.WrapElementHoverFunc(wrapper)
			}
		}
	}

	registerEventFuncsForResource(db, m, b)
	return nil
}

func (b *Builder) configVersionAndPublish(pb *presets.Builder, mb *presets.ModelBuilder, db *gorm.DB) {
	eb := mb.Editing()
	eb.Creating().Except(VersionsPublishBar)
	// On demand, currently only supported dp
	// var fb *presets.FieldBuilder
	// if !mb.HasDetailing() {
	// 	fb = eb.GetField(VersionsPublishBar)
	// } else {
	// 	fb = mb.Detailing().GetField(VersionsPublishBar)
	// }
	var dp *presets.DetailingBuilder
	if !mb.HasDetailing() {
		dp = mb.Detailing().Drawer(true)
	} else {
		dp = mb.Detailing()
	}

	fb := dp.GetField(VersionsPublishBar)
	if fb != nil && fb.GetCompFunc() == nil {
		fb.ComponentFunc(DefaultVersionComponentFunc(mb))
	}

	lb := mb.Listing()
	lb.WrapSearchFunc(makeSearchFunc(mb, db))
	lb.RowMenu().RowMenuItem("Delete").ComponentFunc(func(obj interface{}, id string, ctx *web.EventContext) h.HTMLComponent {
		// DeleteRowMenu should be disabled when using the version interface
		return nil
	})

	setter := makeSetVersionSetterFunc(db)
	eb.WrapSetterFunc(setter)

	lb.Field(ListingFieldDraftCount).ComponentFunc(draftCountFunc(mb, db))
	lb.Field(ListingFieldLive).ComponentFunc(liveFunc(db))
	lb.WrapColumns(presets.CustomizeColumnLabel(func(evCtx *web.EventContext) (map[string]string, error) {
		msgr := i18n.MustGetModuleMessages(evCtx.R, I18nPublishKey, Messages_en_US).(*Messages)
		return map[string]string{
			ListingFieldDraftCount: msgr.HeaderDraftCount,
			ListingFieldLive:       msgr.HeaderLive,
		}, nil
	}))
	lb.WrapColumns(presets.CustomizeColumnHeader(func(evCtx *web.EventContext, col *presets.Column, th h.MutableAttrHTMLComponent) (h.MutableAttrHTMLComponent, error) {
		th.SetAttr("style", "min-width: 200px;")
		return th, nil
	}, ListingFieldLive))

	slugDecoder := mb.NewModel().(presets.SlugDecoder)
	uniqueWithoutVersion := func(id string) string {
		var kvsWithoutVersion []string
		for k, v := range slugDecoder.PrimaryColumnValuesBySlug(id) {
			if k == SlugVersion {
				continue
			}
			kvsWithoutVersion = append(kvsWithoutVersion, fmt.Sprintf(`%s:%s`, k, v))
		}
		sort.Strings(kvsWithoutVersion)
		return strings.Join(kvsWithoutVersion, ",")
	}
	wrapIdCurrentActive := func(in presets.IdCurrentActiveProcessor) presets.IdCurrentActiveProcessor {
		return func(ctx *web.EventContext, current string) (string, error) {
			current, err := in(ctx, current)
			if err != nil {
				return "", err
			}
			if current == "" {
				return "", nil
			}
			return uniqueWithoutVersion(current), nil
		}
	}
	eb.WrapIdCurrentActive(wrapIdCurrentActive)
	eb.Creating().WrapIdCurrentActive(wrapIdCurrentActive)
	dp.WrapIdCurrentActive(wrapIdCurrentActive)
	lb.WrapRow(func(in presets.RowProcessor) presets.RowProcessor {
		return func(evCtx *web.EventContext, row h.MutableAttrHTMLComponent, id string, obj any) (h.MutableAttrHTMLComponent, error) {
			unique := uniqueWithoutVersion(id)
			if unique == "" {
				return in(evCtx, row, id, obj)
			}
			row.SetAttr(":class", fmt.Sprintf(`{ %q: vars.%s === %q }`,
				presets.ListingCompo_CurrentActiveClass, presets.ListingCompo_GetVarCurrentActive(evCtx), unique,
			))
			return in(evCtx, row, id, obj)
		}
	})
	configureVersionListDialog(db, b, pb, mb)
}

func makeSearchFunc(mb *presets.ModelBuilder, db *gorm.DB) func(searcher presets.SearchFunc) presets.SearchFunc {
	return func(searcher presets.SearchFunc) presets.SearchFunc {
		return func(ctx *web.EventContext, params *presets.SearchParams) (result *presets.SearchResult, err error) {
			stmt := &gorm.Statement{DB: db}
			stmt.Parse(params.Model)
			tn := stmt.Schema.Table

			var pks []string
			condition := ""
			for _, f := range stmt.Schema.Fields {
				if f.Name == "DeletedAt" {
					condition = "WHERE deleted_at IS NULL"
				}
			}
			for _, f := range stmt.Schema.PrimaryFields {
				if f.Name != "Version" {
					pks = append(pks, f.DBName)
				}
			}
			pkc := strings.Join(pks, ",")

			sql := fmt.Sprintf(`
			(%s, version) IN (
				SELECT %s, version
				FROM (
					SELECT %s, version,
						ROW_NUMBER() OVER (PARTITION BY %s ORDER BY CASE WHEN status = '%s' THEN 0 ELSE 1 END, version DESC) as rn
					FROM %s %s
				) subquery
				WHERE subquery.rn = 1
			)`, pkc, pkc, pkc, pkc, StatusOnline, tn, condition)

			if _, ok := mb.NewModel().(ScheduleInterface); ok {
				// Also need to get the most recent planned to publish
				sql = fmt.Sprintf(`
				(%s, version) IN (
					SELECT %s, version
					FROM (
						SELECT %s, version,
							ROW_NUMBER() OVER (
								PARTITION BY %s 
								ORDER BY 
									CASE WHEN status = '%s' THEN 0 ELSE 1 END,
									CASE 
										WHEN scheduled_start_at >= now() THEN scheduled_start_at
										ELSE NULL 
									END,
									version DESC
							) as rn
						FROM %s %s
					) subquery
					WHERE subquery.rn = 1
				)`, pkc, pkc, pkc, pkc, StatusOnline, tn, condition)
			}

			con := presets.SQLCondition{
				Query: sql,
			}
			params.SQLConditions = append(params.SQLConditions, &con)

			return searcher(ctx, params)
		}
	}
}

func makeSetVersionSetterFunc(db *gorm.DB) func(presets.SetterFunc) presets.SetterFunc {
	return func(in presets.SetterFunc) presets.SetterFunc {
		return func(obj interface{}, ctx *web.EventContext) {
			if ctx.Param(presets.ParamID) == "" {
				version := fmt.Sprintf("%s-v01", db.NowFunc().Local().Format("2006-01-02"))
				if err := reflectutils.Set(obj, "Version.Version", version); err != nil {
					return
				}
				if err := reflectutils.Set(obj, "Version.VersionName", version); err != nil {
					return
				}
			}
			if in != nil {
				in(obj, ctx)
			}
		}
	}
}

func (b *Builder) Install(pb *presets.Builder) error {
	if b.autoSchedule {
		defer func() {
			go RunPublisher(context.Background(), b.db, b.storage, b)
		}()
	}
	pb.FieldDefaults(presets.LIST).
		FieldType(Status{}).
		ComponentFunc(StatusListFunc())

	pb.GetI18n().
		RegisterForModule(language.English, I18nPublishKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nPublishKey, Messages_zh_CN).
		RegisterForModule(language.Japanese, I18nPublishKey, Messages_ja_JP)

	utils.Install(pb)
	for _, f := range b.afterInstallFuncs {
		f()
	}
	return nil
}

func (b *Builder) ContextValueFuncs(vs ...ContextValueFunc) *Builder {
	b.ctxValueProviders = append(b.ctxValueProviders, vs...)
	return b
}

func (b *Builder) WithContextValues(ctx context.Context) context.Context {
	for _, v := range b.ctxValueProviders {
		ctx = v(ctx)
	}
	return ctx
}

func (b *Builder) getPublishContent(ctx context.Context, obj interface{}) (r string, err error) {
	var (
		mb PreviewBuilderInterface
		ok bool
	)
	builder := ctx.Value(utils.GetObjectName(obj))
	mb, ok = builder.(PreviewBuilderInterface)
	if !ok {
		return
	}
	r = mb.PreviewHTML(obj)
	return
}

func (b *Builder) getObjectLiveUrl(ctx context.Context, db *gorm.DB, obj interface{}) (url string) {
	builder := ctx.Value(utils.GetObjectName(obj))
	mb, ok := builder.(PreviewBuilderInterface)
	if !ok {
		return
	}
	newRecord := reflect.New(reflect.TypeOf(obj).Elem()).Interface()
	id := reflectutils.MustGet(obj, "ID")
	lrdb := db.Model(newRecord).Select("online_url").Where("id = ? AND status = ?", id, StatusOnline)
	if mb.ExistedL10n() {
		localeCode := reflectutils.MustGet(obj, "LocaleCode")
		lrdb.Where("locale_code = ?", localeCode)
	}
	lrdb.Scan(&url)
	return
}

func (b *Builder) defaultPublishActions(ctx context.Context, _ *gorm.DB, _ oss.StorageInterface, obj interface{}) (actions []*PublishAction, err error) {
	var (
		content string
		p       PublishModelInterface
		ok      bool
	)
	if content, err = b.getPublishContent(ctx, obj); err != nil {
		return
	}

	p, ok = obj.(PublishModelInterface)
	if !ok {
		return nil, errors.New("wrong PublishModelInterface")
	}
	if publishUrl := p.PublishUrl(b.db, ctx, b.storage); publishUrl != "" {
		actions = append(actions, &PublishAction{
			Url:      publishUrl,
			Content:  content,
			IsDelete: false,
		})
		if liveUrl := b.getObjectLiveUrl(ctx, b.db, obj); liveUrl != "" && liveUrl != publishUrl {
			actions = append(actions, &PublishAction{
				Url:      liveUrl,
				IsDelete: true,
			})
		}
	}
	return
}

func (b *Builder) getPublishActions(ctx context.Context, obj interface{}) (actions []*PublishAction, err error) {
	var (
		p  PublishInterface
		m  WrapPublishInterface
		ok bool
	)
	p, ok = obj.(PublishInterface)
	if ok {
		return p.GetPublishActions(ctx, b.db, b.storage)
	}
	if m, ok = obj.(WrapPublishInterface); ok {
		return m.WrapPublishActions(b.defaultPublishActions)(ctx, b.db, b.storage, obj)
	}
	return b.defaultPublishActions(ctx, b.db, b.storage, obj)
}

func (b *Builder) defaultUnPublishActions(ctx context.Context, _ *gorm.DB, _ oss.StorageInterface, obj interface{}) (actions []*PublishAction, err error) {
	if liveUrl := b.getObjectLiveUrl(ctx, b.db, obj); liveUrl != "" {
		actions = append(actions, &PublishAction{
			Url:      liveUrl,
			IsDelete: true,
		})
	}
	return
}

func (b *Builder) getUnPublishActions(ctx context.Context, obj interface{}) (actions []*PublishAction, err error) {
	var (
		p  UnPublishInterface
		m  WrapUnPublishInterface
		ok bool
	)

	p, ok = obj.(UnPublishInterface)
	if ok {
		return p.GetUnPublishActions(ctx, b.db, b.storage)
	}
	if m, ok = obj.(WrapUnPublishInterface); ok {
		return m.WrapUnPublishActions(b.defaultUnPublishActions)(ctx, b.db, b.storage, obj)
	}
	return b.defaultUnPublishActions(ctx, b.db, b.storage, obj)
}

func (b *Builder) WrapPublish(w func(in PublishFunc) PublishFunc) *Builder {
	b.publish = w(b.publish)
	return b
}

func (b *Builder) Publish(ctx context.Context, record any) (err error) {
	return b.publish(ctx, record)
}

// 幂等
func (b *Builder) defaultPublish(ctx context.Context, record any) (err error) {
	err = utils.Transact(b.db, func(tx *gorm.DB) (err error) {
		// publish content
		var objs []*PublishAction
		if objs, err = b.getPublishActions(ctx, record); err != nil {
			return
		}
		if err = UploadOrDelete(objs, b.storage); err != nil {
			return
		}
		// update status
		if r, ok := record.(StatusInterface); ok {
			now := b.db.NowFunc()
			if version, ok := record.(VersionInterface); ok {
				var modelSchema *schema.Schema
				modelSchema, err = schema.Parse(record, &sync.Map{}, b.db.NamingStrategy)
				if err != nil {
					return
				}
				scope := setPrimaryKeysConditionWithoutVersion(b.db.Model(reflect.New(modelSchema.ModelType).Interface()), record, modelSchema).Where("version <> ? AND status = ?", version.EmbedVersion().Version, StatusOnline)

				oldVersionUpdateMap := make(map[string]interface{})
				if _, ok := record.(ScheduleInterface); ok {
					oldVersionUpdateMap["scheduled_end_at"] = nil
					oldVersionUpdateMap["actual_end_at"] = &now
				}
				if _, ok := record.(ListInterface); ok {
					oldVersionUpdateMap["list_deleted"] = true
				}
				oldVersionUpdateMap["status"] = StatusOffline
				if err = scope.Updates(oldVersionUpdateMap).Error; err != nil {
					return
				}
			}
			updateMap := make(map[string]interface{})

			if r, ok := record.(ScheduleInterface); ok {
				r.EmbedSchedule().ActualStartAt = &now
				r.EmbedSchedule().ScheduledStartAt = nil
				updateMap["scheduled_start_at"] = r.EmbedSchedule().ScheduledStartAt
				updateMap["actual_start_at"] = r.EmbedSchedule().ActualStartAt
			}
			if _, ok := record.(ListInterface); ok {
				updateMap["list_updated"] = true
			}
			updateMap["status"] = StatusOnline
			updateMap["online_url"] = r.EmbedStatus().OnlineUrl
			if err = b.db.Model(record).Updates(updateMap).Error; err != nil {
				return
			}
		}

		// publish callback
		if r, ok := record.(AfterPublishInterface); ok {
			if err = r.AfterPublish(ctx, b.db, b.storage); err != nil {
				return
			}
		}
		return
	})
	return
}

func (b *Builder) WrapUnPublish(w func(in UnPublishFunc) UnPublishFunc) *Builder {
	b.unpublish = w(b.unpublish)
	return b
}

func (b *Builder) UnPublish(ctx context.Context, record any) (err error) {
	return b.unpublish(ctx, record)
}

// 幂等
func (b *Builder) defaultUnPublish(ctx context.Context, record any) (err error) {
	err = utils.Transact(b.db, func(tx *gorm.DB) (err error) {
		// unpublish content
		var objs []*PublishAction
		objs, err = b.getUnPublishActions(ctx, record)
		if err != nil {
			return
		}
		if err = UploadOrDelete(objs, b.storage); err != nil {
			return
		}
		// update status
		if _, ok := record.(StatusInterface); ok {
			updateMap := make(map[string]interface{})
			if r, ok := record.(ScheduleInterface); ok {
				now := b.db.NowFunc()
				r.EmbedSchedule().ActualEndAt = &now
				r.EmbedSchedule().ScheduledEndAt = nil
				updateMap["scheduled_end_at"] = r.EmbedSchedule().ScheduledEndAt
				updateMap["actual_end_at"] = r.EmbedSchedule().ActualEndAt
			}
			if _, ok := record.(ListInterface); ok {
				updateMap["list_deleted"] = true
			}
			updateMap["status"] = StatusOffline
			if err = b.db.Model(record).Updates(updateMap).Error; err != nil {
				return
			}
		}

		// unpublish callback
		if r, ok := record.(AfterUnPublishInterface); ok {
			if err = r.AfterUnPublish(ctx, b.db, b.storage); err != nil {
				return
			}
		}
		return
	})
	return
}

func UploadOrDelete(objs []*PublishAction, storage oss.StorageInterface) (err error) {
	for _, obj := range objs {
		if obj.IsDelete {
			fmt.Printf("deleting %s \n", obj.Url)
			err = storage.Delete(obj.Url)
		} else {
			fmt.Printf("uploading %s \n", obj.Url)
			_, err = storage.Put(obj.Url, strings.NewReader(obj.Content))
		}
		if err != nil {
			return
		}
	}
	return nil
}

func setPrimaryKeysConditionWithoutVersion(db *gorm.DB, record interface{}, s *schema.Schema) *gorm.DB {
	var querys []string
	var args []interface{}
	for _, p := range s.PrimaryFields {
		if p.Name == "Version" {
			continue
		}
		val, _ := p.ValueOf(db.Statement.Context, reflect.ValueOf(record))
		querys = append(querys, fmt.Sprintf("%s = ?", strcase.ToSnake(p.Name)))
		args = append(args, val)
	}
	return db.Where(strings.Join(querys, " AND "), args...)
}

func setPrimaryKeysConditionWithoutFields(db *gorm.DB, record interface{}, s *schema.Schema, ignoreFields ...string) *gorm.DB {
	var querys []string
	var args []interface{}
	for _, p := range s.PrimaryFields {
		if slices.Contains(ignoreFields, p.Name) {
			continue
		}
		val, _ := p.ValueOf(db.Statement.Context, reflect.ValueOf(record))
		querys = append(querys, fmt.Sprintf("%s = ?", strcase.ToSnake(p.Name)))
		args = append(args, val)
	}
	return db.Where(strings.Join(querys, " AND "), args...)
}

func (b *Builder) FullUrl(uri string) (s string) {
	s, _ = b.storage.GetURL(uri)
	return strings.TrimSuffix(b.storage.GetEndpoint(), "/") + "/" + strings.Trim(s, "/")
}
