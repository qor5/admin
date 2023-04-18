package publish

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/qor/oss"
	"gorm.io/gorm"
)

func RunPublisher(db *gorm.DB, storage oss.StorageInterface, publisher *Builder) {
	{ // schedule publisher
		scheduleP := NewSchedulePublishBuilder(publisher)

		for name, model := range NonVersionPublishModels {
			name := name
			model := model
			go RunJob("schedule-publisher"+"-"+name, time.Minute, time.Minute*5, func() {
				if err := scheduleP.Run(model); err != nil {
					panic(err)
				}
			})
		}

		for name, model := range VersionPublishModels {
			name := name
			model := model
			go RunJob("schedule-publisher"+"-"+name, time.Minute, time.Minute*5, func() {
				if err := scheduleP.Run(model); err != nil {
					panic(err)
				}
			})
		}
	}

	{ // list publisher
		listP := NewListPublishBuilder(db, storage)
		for name, model := range ListPublishModels {
			name := name
			model := model
			go RunJob("list-publisher"+"-"+name, time.Minute, time.Minute*5, func() {
				if err := listP.Run(model); err != nil {
					panic(err)
				}
			})
		}
	}
}

func RunJob(jobName string, interval time.Duration, timeout time.Duration, f func()) {
	t := time.Tick(interval)
	for range t {
		start := time.Now()
		s := make(chan bool, 1)
		go func() {
			defer func() {
				stop := time.Now()
				log.Printf("job_name: %s, started_at: %s, stopped_at: %s, time_spent_ms: %s\n", jobName, start, stop, fmt.Sprintf("%f", float64(stop.Sub(start))/float64(time.Millisecond)))
			}()
			f()
			s <- true
		}()

		select {
		case <-s:
		case <-time.After(timeout):
			log.Printf("job_name: %s, started_at: %s, timeout: %s\n", jobName, start, time.Now())
			os.Exit(124)
		}
	}
}
