package publish

import (
	"log"
	"os"
	"time"

	"github.com/qor/oss"
	"gorm.io/gorm"
)

const (
	schedulePublishJobNamePrefix = "schedule-publisher"
	listPublishJobNamePrefix     = "list-publisher"
)

func RunPublisher(db *gorm.DB, storage oss.StorageInterface, publisher *Builder) {
	{ // schedule publisher
		scheduleP := NewSchedulePublishBuilder(publisher)

		for name, model := range NonVersionPublishModels {
			go RunJob(schedulePublishJobNamePrefix+"-"+name, time.Minute, time.Minute*5, func() {
				if err := scheduleP.Run(model); err != nil {
					log.Printf("schedule publisher error: %v\n", err)
				}
			})
		}

		for name, model := range VersionPublishModels {
			go RunJob(schedulePublishJobNamePrefix+"-"+name, time.Minute, time.Minute*5, func() {
				if err := scheduleP.Run(model); err != nil {
					log.Printf("schedule publisher error: %v\n", err)
				}
			})
		}
	}

	{ // list publisher
		listP := NewListPublishBuilder(db, storage)
		for name, model := range ListPublishModels {
			go RunJob(listPublishJobNamePrefix+"-"+name, time.Minute, time.Minute*5, func() {
				if err := listP.Run(model); err != nil {
					log.Printf("schedule publisher error: %v\n", err)
				}
			})
		}
	}
}

func RunJob(jobName string, interval time.Duration, timeout time.Duration, f func()) {
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
