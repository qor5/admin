package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/qor/oss/s3"
	"github.com/qor/qor5/example/admin"
	"github.com/qor/qor5/publish"
)

func main() {
	db := admin.ConnectDB()
	storage := s3.New(&s3.Config{
		Bucket:  os.Getenv("S3_Bucket"),
		Region:  os.Getenv("S3_Region"),
		Session: session.Must(session.NewSession()),
	})
	admin.NewConfig()

	{ // schedule publisher
		publisher := publish.New(db, storage)
		scheduleP := publish.NewSchedulePublishBuilder(publisher)

		for name, model := range publish.NonVersionPublishModels {
			name := name
			model := model
			go RunJob("schedule-publisher"+"-"+name, time.Minute, time.Minute*5, func() {
				if err := scheduleP.Run(model); err != nil {
					panic(err)
				}
			})
		}

		for name, model := range publish.VersionPublishModels {
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
		listP := publish.NewListPublishBuilder(db, storage)
		for name, model := range publish.ListPublishModels {
			name := name
			model := model
			go RunJob("list-publisher"+"-"+name, time.Minute, time.Minute*5, func() {
				if err := listP.Run(model); err != nil {
					panic(err)
				}
			})
		}
	}

	select {}
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
