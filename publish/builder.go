package publish

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/iancoleman/strcase"
	"github.com/qor/oss"
	"github.com/qor/qor5/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Builder struct {
	db      *gorm.DB
	storage oss.StorageInterface
}

func New(db *gorm.DB, storage oss.StorageInterface) *Builder {
	return &Builder{
		db:      db,
		storage: storage,
	}
}

// 幂等
func (b *Builder) Publish(record interface{}) (err error) {
	err = utils.Transact(b.db, func(tx *gorm.DB) (err error) {
		// publish content
		if r, ok := record.(PublishInterface); ok {
			var objs []*PublishAction
			objs = r.GetPublishActions(b.db)
			if err = b.UploadOrDelete(objs); err != nil {
				return
			}
		}

		// update status
		if r, ok := record.(StatusInterface); ok {
			if _, ok := record.(VersionInterface); ok {
				var modelSchema *schema.Schema
				modelSchema, err = schema.Parse(record, &sync.Map{}, b.db.NamingStrategy)
				if err != nil {
					return
				}
				if err = SetPrimaryKeysConditionWithoutVersion(b.db.Model(reflect.New(modelSchema.ModelType).Interface()), record, modelSchema).Where("status = ?", StatusOnline).Updates(map[string]interface{}{"status": StatusOffline}).Error; err != nil {
					return
				}
			}
			if err = b.db.Model(record).Updates(map[string]interface{}{"status": StatusOnline, "online_id": r.GetOnlineID()}).Error; err != nil {
				return
			}
		}

		// TODO update schedule

		// publish callback
		if r, ok := record.(AfterPublishInterface); ok {
			if err = r.AfterPublish(b.db, b.storage); err != nil {
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
			objs = r.GetUnPublishActions(b.db)
			if err = b.UploadOrDelete(objs); err != nil {
				return
			}
		}

		// update status
		if _, ok := record.(StatusInterface); ok {
			if err = b.db.Model(record).Updates(map[string]interface{}{"status": StatusOffline}).Error; err != nil {
				return
			}
		}

		// unpublish callback
		if r, ok := record.(AfterUnPublishInterface); ok {
			if err = r.AfterUnPublish(b.db, b.storage); err != nil {
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

func (b *Builder) UploadOrDelete(objs []*PublishAction) (err error) {
	for _, obj := range objs {
		if obj.IsDelete {
			err = b.storage.Delete(obj.Url)
		} else {
			_, err = b.storage.Put(obj.Url, strings.NewReader(obj.Content))
		}
		if err != nil {
			return
		}
	}
	return nil
}

func SetPrimaryKeysConditionWithoutVersion(db *gorm.DB, record interface{}, s *schema.Schema) *gorm.DB {
	conds := []string{}
	for _, p := range s.PrimaryFields {
		if p.Name == "VersionName" {
			continue
		}
		val, _ := p.ValueOf(reflect.ValueOf(record))
		conds = append(conds, fmt.Sprintf("%s = %v", strcase.ToSnake(p.Name), val))
	}

	return db.Where(strings.Join(conds, " AND "))
}
