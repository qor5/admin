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
}

func NewSchedulePublishBuilder(publisher *Builder) *SchedulePublishBuilder {
	return &SchedulePublishBuilder{
		publisher: publisher,
	}
}

type SchedulePublisher interface {
	SchedulePublishDBScope(db *gorm.DB) *gorm.DB
	ScheduleUnPublishDBScope(db *gorm.DB) *gorm.DB
}

type ctxKeyScheduleRecordsFinder struct{}

type ScheduleOperation string

const (
	ScheduleOperationPublish   ScheduleOperation = "publish"
	ScheduleOperationUnPublish ScheduleOperation = "unpublish"
)

type ScheduleRecordsFinderFunc func(ctx context.Context, operation ScheduleOperation, b *Builder, db *gorm.DB, records any) error

func WithScheduleRecordsFinder(ctx context.Context, f ScheduleRecordsFinderFunc) context.Context {
	return context.WithValue(ctx, ctxKeyScheduleRecordsFinder{}, f)
}

// model is a empty struct
// example: Product{}
func (b *SchedulePublishBuilder) Run(ctx context.Context, model interface{}) (err error) {
	reqCtx := b.publisher.WithContextValues(ctx)

	// If model is Product{}
	// Generate a records: []*Product{}
	records := reflect.MakeSlice(reflect.SliceOf(reflect.New(reflect.TypeOf(model)).Type()), 0, 0).Interface()
	flagTime := b.publisher.db.NowFunc()
	var unpublishAfterPublishRecords []interface{}

	{
		tempRecords := records
		scope := b.publisher.db

		fn, ok := ctx.Value(ctxKeyScheduleRecordsFinder{}).(ScheduleRecordsFinderFunc)
		if ok && fn != nil {
			err = fn(reqCtx, ScheduleOperationUnPublish, b.publisher, scope, &tempRecords)
			if err != nil {
				return
			}
		} else {
			if m, ok := model.(SchedulePublisher); ok {
				scope = m.ScheduleUnPublishDBScope(scope)
			}
			err = scope.Where("scheduled_end_at <= ?", flagTime).Order("scheduled_end_at").Find(&tempRecords).Error
			if err != nil {
				return
			}
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
		scope := b.publisher.db

		fn, ok := ctx.Value(ctxKeyScheduleRecordsFinder{}).(ScheduleRecordsFinderFunc)
		if ok && fn != nil {
			err = fn(reqCtx, ScheduleOperationPublish, b.publisher, scope, &tempRecords)
			if err != nil {
				return
			}
		} else {
			if m, ok := model.(SchedulePublisher); ok {
				scope = m.SchedulePublishDBScope(scope)
			}
			err = scope.Where("scheduled_start_at <= ?", flagTime).Order("scheduled_start_at").Find(&tempRecords).Error
			if err != nil {
				return
			}
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
