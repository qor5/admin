package publish

import (
	"context"
	"log"
	"reflect"

	"github.com/hashicorp/go-multierror"
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

// model is a empty struct
// example: Product{}
func (b *SchedulePublishBuilder) Run(model interface{}) (err error) {
	var scope *gorm.DB
	if m, ok := model.(SchedulePublisher); ok {
		scope = m.SchedulePublisherDBScope(b.publisher.db)
	} else {
		scope = b.publisher.db
	}
	reqCtx := b.publisher.WithContextValues(context.Background())

	// If model is Product{}
	// Generate a records: []*Product{}
	records := reflect.MakeSlice(reflect.SliceOf(reflect.New(reflect.TypeOf(model)).Type()), 0, 0).Interface()
	flagTime := scope.NowFunc()
	var unpublishAfterPublishRecords []interface{}

	{
		tempRecords := records
		err = scope.Where("scheduled_end_at <= ?", flagTime).Order("scheduled_end_at").Find(&tempRecords).Error
		if err != nil {
			return
		}
		needUnpublishReflectValues := reflect.ValueOf(tempRecords)
		for i := 0; i < needUnpublishReflectValues.Len(); i++ {
			{
				record := needUnpublishReflectValues.Index(i).Interface().(ScheduleInterface)
				if record.EmbedSchedule().ScheduledStartAt != nil && record.EmbedSchedule().ScheduledStartAt.Sub(*record.EmbedSchedule().ScheduledEndAt) < 0 {
					unpublishAfterPublishRecords = append(unpublishAfterPublishRecords, record)
					continue
				}
			}
			record := needUnpublishReflectValues.Index(i).Interface()
			if err2 := b.publisher.UnPublish(reqCtx, record); err2 != nil {
				log.Printf("error: %s\n", err2)
				err = multierror.Append(err, err2).ErrorOrNil()
			}
		}
	}

	{
		tempRecords := records
		err = scope.Where("scheduled_start_at <= ?", flagTime).Order("scheduled_start_at").Find(&tempRecords).Error
		if err != nil {
			return
		}
		needPublishReflectValues := reflect.ValueOf(tempRecords)
		for i := 0; i < needPublishReflectValues.Len(); i++ {
			record := needPublishReflectValues.Index(i).Interface()
			if err2 := b.publisher.Publish(reqCtx, record); err2 != nil {
				log.Printf("error: %s\n", err2)
				err = multierror.Append(err, err2).ErrorOrNil()
			}
		}
	}

	for _, record := range unpublishAfterPublishRecords {
		if err2 := b.publisher.UnPublish(reqCtx, record); err2 != nil {
			log.Printf("error: %s\n", err2)
			err = multierror.Append(err, err2).ErrorOrNil()
		}
	}
	return
}
