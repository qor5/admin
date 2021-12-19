package publish

import (
	"context"
	"log"
	"reflect"
	"time"

	"gorm.io/gorm"
)

type SchedulePublishBuilder struct {
	publisher *Builder
	context   context.Context
}

func NewSchedulePublishBuilder(publisher *Builder) *SchedulePublishBuilder {
	return &SchedulePublishBuilder{
		publisher: publisher,
		context:   context.Background(),
	}
}

func (b *SchedulePublishBuilder) WithValue(key, val interface{}) *SchedulePublishBuilder {
	b.context = context.WithValue(b.context, key, val)
	return b
}

type SchedulePublisher interface {
	SchedulePublisherDBScope(db *gorm.DB) *gorm.DB
}

func (b *SchedulePublishBuilder) Run(model interface{}) (err error) {
	db := b.publisher.db
	var scope *gorm.DB
	if m, ok := model.(SchedulePublisher); ok {
		scope = m.SchedulePublisherDBScope(b.publisher.db)
	} else {
		scope = db
	}

	records := reflect.MakeSlice(reflect.SliceOf(reflect.New(reflect.TypeOf(model)).Type()), 0, 0).Interface()

	{
		tempRecords := records
		err = scope.Where("status = ? AND scheduled_end_at <= ?", StatusOnline, db.NowFunc().Add(time.Minute)).Find(&tempRecords).Error
		if err != nil {
			return
		}
		publishedReflectValues := reflect.ValueOf(tempRecords)
		for i := 0; i < publishedReflectValues.Len(); i++ {
			if record, ok := publishedReflectValues.Index(i).Interface().(UnPublishInterface); ok {
				if err2 := b.publisher.UnPublish(record); err2 != nil {
					log.Printf("error: %s\n", err2)
					err = err2
				}
			}
		}
	}

	{
		tempRecords := records
		err = scope.Where("status = ? AND scheduled_start_at <= ?", StatusDraft, db.NowFunc().Add(time.Minute)).Find(&tempRecords).Error
		if err != nil {
			return
		}
		approvedReflectValues := reflect.ValueOf(tempRecords)
		for i := 0; i < approvedReflectValues.Len(); i++ {
			if record, ok := approvedReflectValues.Index(i).Interface().(PublishInterface); ok {
				if err2 := b.publisher.Publish(record); err2 != nil {
					log.Printf("error: %s\n", err2)
					err = err2
				}
			}
		}
	}
	return
}
