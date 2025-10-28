package publish

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/oss"
	"gorm.io/gorm"

	vx "github.com/qor5/x/v3/ui/vuetifyx"
)

const (
	schedulePublishJobNamePrefix = "schedule-publisher"
	listPublishJobNamePrefix     = "list-publisher"
)

func RunPublisher(ctx context.Context, db *gorm.DB, storage oss.StorageInterface, publisher *Builder) {
	{ // schedule publisher
		scheduleP := NewSchedulePublishBuilder(publisher)

		for name, model := range publisher.nonVersionPublishModels {
			go RunJob(schedulePublishJobNamePrefix+"-"+name, time.Minute, time.Minute*5, func() {
				if err := scheduleP.Run(ctx, model); err != nil {
					log.Printf("schedule publisher error: %v\n", err)
				}
			})
		}

		for name, model := range publisher.versionPublishModels {
			go RunJob(schedulePublishJobNamePrefix+"-"+name, time.Minute, time.Minute*5, func() {
				if err := scheduleP.Run(ctx, model); err != nil {
					log.Printf("schedule publisher error: %v\n", err)
				}
			})
		}
	}

	{ // list publisher
		listP := NewListPublishBuilder(db, storage)
		for name, model := range publisher.listPublishModels {
			go RunJob(listPublishJobNamePrefix+"-"+name, time.Minute, time.Minute*5, func() {
				if err := listP.Run(ctx, model); err != nil {
					log.Printf("schedule publisher error: %v\n", err)
				}
			})
		}
	}
}

func RunJob(jobName string, interval, timeout time.Duration, f func()) {
	second := 1
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for now := range ticker.C {
		targetTime := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+1, second, 0, now.Location())
		time.Sleep(targetTime.Sub(now))

		start := time.Now()
		done := make(chan struct{})

		go func() {
			defer func() {
				stop := time.Now()
				log.Printf("job_name: %s, started_at: %s, stopped_at: %s, time_spent_ms: %d\n", jobName, start, stop, int64(stop.Sub(start)/time.Millisecond))
			}()
			f()
			done <- struct{}{}
		}()

		select {
		case <-done:
		case <-time.After(timeout):
			log.Printf("job_name: %s, started_at: %s, timeout: %s\n", jobName, start, time.Now())
			os.Exit(124)
		}
	}
}

const FilterKeyLive = "live"

func NewLiveFilterItem(ctx context.Context, columnPrefix string) (*vx.FilterItem, error) {
	evCtx := web.MustGetEventContext(ctx)
	msgr := i18n.MustGetModuleMessages(evCtx.R, I18nPublishKey, Messages_en_US).(*Messages)
	return &vx.FilterItem{
		Key:          FilterKeyLive,
		Label:        msgr.HeaderLive,
		ItemType:     vx.ItemTypeSelect,
		SQLCondition: columnPrefix + `status = ?`,
		Options: []*vx.SelectItem{
			{Text: msgr.StatusOnline, Value: StatusOnline},
			{Text: msgr.StatusOffline, Value: StatusOffline},
			{Text: msgr.StatusDraft, Value: StatusDraft},
		},
	}, nil
}
