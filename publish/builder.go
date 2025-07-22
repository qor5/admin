package publish

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/iancoleman/strcase"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/oss"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/utils"
)

type (
	PublishFunc   func(ctx context.Context, record any) error
	UnPublishFunc func(ctx context.Context, record any) error

	Disablement struct {
		DisabledRename bool
		DisabledDelete bool
	}

	DisablementCheckFunc func(ctx *web.EventContext, obj any) *Disablement
)

type Builder struct {
	db                      *gorm.DB
	storage                 oss.StorageInterface
	ab                      *activity.Builder
	ctxValueProviders       []ContextValueFunc
	afterInstallFuncs       []func()
	autoSchedule            bool
	nonVersionPublishModels map[string]interface{}
	versionPublishModels    map[string]interface{}
	listPublishModels       map[string]interface{}

	publish              PublishFunc
	unpublish            UnPublishFunc
	disablementCheckFunc DisablementCheckFunc
}

type ContextValueFunc func(ctx context.Context) context.Context

func New(db *gorm.DB, storage oss.StorageInterface) *Builder {
	b := &Builder{
		db:                      db,
		storage:                 storage,
		nonVersionPublishModels: make(map[string]interface{}),
		versionPublishModels:    make(map[string]interface{}),
		listPublishModels:       make(map[string]interface{}),
	}
	b.publish = b.defaultPublish
	b.unpublish = b.defaultUnPublish
	b.disablementCheckFunc = b.defaultDisableByStatus
	return b
}

func (b *Builder) Activity(v *activity.Builder) (r *Builder) {
	b.ab = v
	return b
}

func (b *Builder) DisablementCheckFunc(v DisablementCheckFunc) (r *Builder) {
	b.disablementCheckFunc = v
	return b
}

func (b *Builder) WrapDisablementCheckFunc(w func(DisablementCheckFunc) DisablementCheckFunc) (r *Builder) {
	b.disablementCheckFunc = w(b.disablementCheckFunc)
	return b
}

func (b *Builder) WrapStorage(v func(oss.StorageInterface) oss.StorageInterface) (r *Builder) {
	b.storage = v(b.storage)
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
			b.versionPublishModels[m.Info().URIName()] = reflect.ValueOf(schedulePublishModel).Elem().Interface()
		}

		b.configVersionAndPublish(pb, m, db)
	} else {
		if schedulePublishModel, ok := obj.(ScheduleInterface); ok {
			b.nonVersionPublishModels[m.Info().URIName()] = reflect.ValueOf(schedulePublishModel).Elem().Interface()
		}
	}

	if model, ok := obj.(ListInterface); ok {
		if schedulePublishModel, ok := model.(ScheduleInterface); ok {
			b.listPublishModels[m.Info().URIName()] = reflect.ValueOf(schedulePublishModel).Elem().Interface()
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
	lb.WrapFilterDataFunc(func(in presets.FilterDataFunc) presets.FilterDataFunc {
		return func(ctx *web.EventContext) vx.FilterData {
			fd := in(ctx)
			for _, item := range fd {
				if item.Key == "f_"+activity.KeyHasUnreadNotes {
					item.SQLCondition = ListSubqueryConditionQueryPrefix + item.SQLCondition
				}
			}
			return fd
		}
	})
	lb.RowMenu().RowMenuItem("Delete").ComponentFunc(func(_ interface{}, _ string, _ *web.EventContext) h.HTMLComponent {
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
	lb.WrapColumns(presets.CustomizeColumnHeader(func(_ *web.EventContext, _ *presets.Column, th h.MutableAttrHTMLComponent) (h.MutableAttrHTMLComponent, error) {
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

const ListSubqueryConditionQueryPrefix = "__PUBLISH_LIST_SUBQUERY_CONDITION__: "

func makeSearchFunc(mb *presets.ModelBuilder, db *gorm.DB) func(searcher presets.SearchFunc) presets.SearchFunc {
	return func(searcher presets.SearchFunc) presets.SearchFunc {
		return func(ctx *web.EventContext, params *presets.SearchParams) (result *presets.SearchResult, err error) {
			stmt := &gorm.Statement{DB: db}
			stmt.Parse(params.Model)
			tn := stmt.Schema.Table

			var pks []string
			var wheres []string
			var args []any

			var newSQLConditions []*presets.SQLCondition
			for _, cond := range params.SQLConditions {
				if strings.HasPrefix(cond.Query, ListSubqueryConditionQueryPrefix) {
					wheres = append(wheres, fmt.Sprintf("(%s)", strings.TrimPrefix(cond.Query, ListSubqueryConditionQueryPrefix)))
					args = append(args, cond.Args...)
				} else {
					newSQLConditions = append(newSQLConditions, cond)
				}
			}
			params.SQLConditions = newSQLConditions

			for _, f := range stmt.Schema.Fields {
				if f.Name == "DeletedAt" {
					wheres = append(wheres, fmt.Sprintf("%s IS NULL", f.DBName))
				}
			}
			subqueryWhere := ""
			if len(wheres) > 0 {
				subqueryWhere = fmt.Sprintf("WHERE %s", strings.Join(wheres, " AND "))
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
			)`, pkc, pkc, pkc, pkc, StatusOnline, tn, subqueryWhere)

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
				)`, pkc, pkc, pkc, pkc, StatusOnline, tn, subqueryWhere)
			}

			con := presets.SQLCondition{
				Query: sql,
				Args:  args,
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
			RunPublisher(context.Background(), b.db, b.storage, b)
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

func (*Builder) getPublishContent(ctx context.Context, obj interface{}) (r string, err error) {
	var (
		mb PreviewBuilderInterface
		ok bool
	)
	builder := ctx.Value(utils.GetObjectName(obj))
	mb, ok = builder.(PreviewBuilderInterface)
	if !ok {
		return
	}
	r = mb.PreviewHTML(ctx, obj)
	return
}

func (*Builder) getObjectLiveUrl(ctx context.Context, db *gorm.DB, obj interface{}) (url string) {
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
		// update status
		if r, ok := record.(StatusInterface); ok {
			now := tx.NowFunc()
			if version, ok := record.(VersionInterface); ok {
				var modelSchema *schema.Schema
				modelSchema, err = schema.Parse(record, &sync.Map{}, tx.NamingStrategy)
				if err != nil {
					return
				}
				scope := setPrimaryKeysConditionWithoutVersion(tx.Model(reflect.New(modelSchema.ModelType).Interface()), record, modelSchema).Where("version <> ? AND status = ?", version.EmbedVersion().Version, StatusOnline)

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
			if r, ok := record.(ListInterface); ok {
				r.EmbedList().ListUpdated = true
				updateMap["list_updated"] = true
			}
			r.EmbedStatus().Status = StatusOnline
			updateMap["status"] = StatusOnline
			updateMap["online_url"] = r.EmbedStatus().OnlineUrl
			if err = tx.Model(record).Updates(updateMap).Error; err != nil {
				return
			}
		}

		if err = UploadOrDelete(ctx, objs, b.storage); err != nil {
			return
		}

		// publish callback
		if r, ok := record.(AfterPublishInterface); ok {
			if err = r.AfterPublish(ctx, tx, b.storage); err != nil {
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
		// update status
		if r, ok := record.(StatusInterface); ok {
			updateMap := make(map[string]interface{})
			if r, ok := record.(ScheduleInterface); ok {
				now := tx.NowFunc()
				r.EmbedSchedule().ActualEndAt = &now
				r.EmbedSchedule().ScheduledEndAt = nil
				updateMap["scheduled_end_at"] = r.EmbedSchedule().ScheduledEndAt
				updateMap["actual_end_at"] = r.EmbedSchedule().ActualEndAt
			}
			if r, ok := record.(ListInterface); ok {
				r.EmbedList().ListDeleted = true
				updateMap["list_deleted"] = true
			}
			r.EmbedStatus().Status = StatusOffline
			updateMap["status"] = StatusOffline
			if err = tx.Model(record).Updates(updateMap).Error; err != nil {
				return
			}
		}

		if err = UploadOrDelete(ctx, objs, b.storage); err != nil {
			return
		}

		// unpublish callback
		if r, ok := record.(AfterUnPublishInterface); ok {
			if err = r.AfterUnPublish(ctx, tx, b.storage); err != nil {
				return
			}
		}

		return
	})
	return
}

func UploadOrDelete(ctx context.Context, objs []*PublishAction, storage oss.StorageInterface) (err error) {
	for _, obj := range objs {
		if obj.IsDelete {
			fmt.Printf("deleting %s \n", obj.Url)
			err = storage.Delete(ctx, obj.Url)
		} else {
			fmt.Printf("uploading %s \n", obj.Url)
			_, err = storage.Put(ctx, obj.Url, strings.NewReader(obj.Content))
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

func (b *Builder) FullUrl(ctx context.Context, uri string) (string, error) {
	s, err := b.storage.GetURL(ctx, uri)
	if err != nil {
		return "", errors.Wrap(err, "get url")
	}
	return strings.TrimSuffix(b.storage.GetEndpoint(ctx), "/") + "/" + strings.Trim(s, "/"), nil
}

func (b *Builder) defaultDisableByStatus(_ *web.EventContext, obj any) *Disablement {
	status := obj.(StatusInterface).EmbedStatus().Status
	disabled := status == StatusOnline || status == StatusOffline
	return &Disablement{
		DisabledRename: disabled,
		DisabledDelete: disabled,
	}
}
