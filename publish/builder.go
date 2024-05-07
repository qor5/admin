package publish

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/iancoleman/strcase"
	"github.com/qor/oss"
	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Builder struct {
	db      *gorm.DB
	storage oss.StorageInterface
	context context.Context
	models  []*presets.ModelBuilder
	ab      *activity.Builder
}

func New(db *gorm.DB, storage oss.StorageInterface) *Builder {
	return &Builder{
		db:      db,
		storage: storage,
		context: context.Background(),
	}
}

func (b *Builder) Models(vs ...*presets.ModelBuilder) (r *Builder) {
	b.models = append(b.models, vs...)
	return b
}

func (b *Builder) Activity(v *activity.Builder) (r *Builder) {
	b.ab = v
	return b
}

func (b *Builder) Install(pb *presets.Builder) {
	configure(pb, b)
	return
}

func (b *Builder) WithValue(key, val interface{}) *Builder {
	b.context = context.WithValue(b.context, key, val)
	return b
}

const (
	PublishContextKeyPageBuilder  = "pagebuilder"
	PublishContextKeyL10nBuilder  = "l10nbuilder"
	PublishContextKeyEventContext = "eventcontext"
)

func (b *Builder) WithPageBuilder(val interface{}) *Builder {
	b.context = context.WithValue(b.context, PublishContextKeyPageBuilder, val)
	return b
}

func (b *Builder) WithL10nBuilder(val interface{}) *Builder {
	b.context = context.WithValue(b.context, PublishContextKeyL10nBuilder, val)
	return b
}

func (b *Builder) WithEventContext(val interface{}) *Builder {
	b.context = context.WithValue(b.context, PublishContextKeyEventContext, val)
	return b
}

func (b *Builder) Context() context.Context {
	return b.context
}

// 幂等
func (b *Builder) Publish(record interface{}) (err error) {
	err = utils.Transact(b.db, func(tx *gorm.DB) (err error) {
		// publish content
		if r, ok := record.(PublishInterface); ok {
			var objs []*PublishAction
			objs, err = r.GetPublishActions(b.db, b.context, b.storage)
			if err != nil {
				return
			}
			if err = UploadOrDelete(objs, b.storage); err != nil {
				return
			}
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
				scope := SetPrimaryKeysConditionWithoutVersion(b.db.Model(reflect.New(modelSchema.ModelType).Interface()), record, modelSchema).Where("version <> ? AND status = ?", version.GetVersion(), StatusOnline)
				var count int64
				if err = scope.Count(&count).Error; err != nil {
					return
				}

				// update old version
				if count > 0 {
					var oldVersionUpdateMap = make(map[string]interface{})
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
			}
			var updateMap = make(map[string]interface{})

			if r, ok := record.(ScheduleInterface); ok {
				r.SetPublishedAt(&now)
				r.SetScheduledStartAt(nil)
				updateMap["scheduled_start_at"] = r.GetScheduledStartAt()
				updateMap["actual_start_at"] = r.GetPublishedAt()
			}
			if _, ok := record.(ListInterface); ok {
				updateMap["list_updated"] = true
			}
			updateMap["status"] = StatusOnline
			updateMap["online_url"] = r.GetOnlineUrl()
			if err = b.db.Model(record).Updates(updateMap).Error; err != nil {
				return
			}
		}

		// publish callback
		if r, ok := record.(AfterPublishInterface); ok {
			if err = r.AfterPublish(b.db, b.storage, b.context); err != nil {
				return
			}
		}
		return
	})
	return
}

func (b *Builder) UnPublish(record interface{}) (err error) {
	err = utils.Transact(b.db, func(tx *gorm.DB) (err error) {
		// unpublish content
		if r, ok := record.(UnPublishInterface); ok {
			var objs []*PublishAction
			objs, err = r.GetUnPublishActions(b.db, b.context, b.storage)
			if err != nil {
				return
			}
			if err = UploadOrDelete(objs, b.storage); err != nil {
				return
			}
		}

		// update status
		if _, ok := record.(StatusInterface); ok {
			var updateMap = make(map[string]interface{})
			if r, ok := record.(ScheduleInterface); ok {
				now := b.db.NowFunc()
				r.SetUnPublishedAt(&now)
				r.SetScheduledEndAt(nil)
				updateMap["scheduled_end_at"] = r.GetScheduledEndAt()
				updateMap["actual_end_at"] = r.GetUnPublishedAt()
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
			if err = r.AfterUnPublish(b.db, b.storage, b.context); err != nil {
				return
			}
		}
		return
	})
	return
}

func (b *Builder) Sync(models ...interface{}) error {
	return nil
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

func SetPrimaryKeysConditionWithoutVersion(db *gorm.DB, record interface{}, s *schema.Schema) *gorm.DB {
	querys := []string{}
	args := []interface{}{}
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
