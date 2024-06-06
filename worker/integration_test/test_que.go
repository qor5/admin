package integration

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/qor5/admin/v3/worker"
	"github.com/qor5/admin/v3/worker/mock"
)

var items []worker.QueJobInterface

var Que = &mock.QueueMock{
	AddFunc: func(ctx context.Context, job worker.QueJobInterface) error {
		jobInfo, err := job.GetJobInfo()
		if err != nil {
			return err
		}
		if scheduler, ok := jobInfo.Argument.(worker.Scheduler); ok && scheduler.GetScheduleTime() != nil {
			job.SetStatus(worker.JobStatusScheduled)
		}
		items = append(items, job)
		return nil
	},
	KillFunc: func(ctx context.Context, job worker.QueJobInterface) error {
		return job.SetStatus(worker.JobStatusKilled)
	},
	ListenFunc: func(jobDefs []*worker.QorJobDefinition, getJob func(qorJobID uint) (worker.QueJobInterface, error)) error {
		return nil
	},
	RemoveFunc: func(ctx context.Context, job worker.QueJobInterface) error {
		return job.SetStatus(worker.JobStatusCancelled)
	},
	ShutdownFunc: func(ctx context.Context) error {
		return nil
	},
}

func ConsumeQueItem() (err error) {
	if len(items) == 0 {
		return
	}

	job := items[0]
	items = items[1:]
	defer func() {
		if r := recover(); r != nil {
			job.AddLog(string(debug.Stack()))
			job.SetProgressText(fmt.Sprint(r))
			job.SetStatus(worker.JobStatusException)
			panic(r)
		}
	}()

	if job.GetStatus() == worker.JobStatusCancelled {
		return
	}
	if job.GetStatus() != worker.JobStatusNew && job.GetStatus() != worker.JobStatusScheduled {
		job.SetStatus(worker.JobStatusKilled)
		return errors.New("invalid job status, current status: " + job.GetStatus())
	}

	err = job.SetStatus(worker.JobStatusRunning)
	if err != nil {
		return err
	}

	hctx, cf := context.WithCancel(context.Background())
	hDoneC := make(chan struct{})
	isAborted := false
	go func() {
		timer := time.NewTicker(time.Second)
		for {
			select {
			case <-hDoneC:
				return
			case <-timer.C:
				status, _ := job.FetchAndSetStatus()
				if status == worker.JobStatusKilled {
					isAborted = true
					cf()
					return
				}
			}
		}
	}()
	job.StartRefresh()
	err = job.GetHandler()(hctx, job)
	job.StopRefresh()
	if !isAborted {
		hDoneC <- struct{}{}
	}
	if err != nil {
		job.SetProgressText(err.Error())
		job.SetStatus(worker.JobStatusException)
		return err
	}
	if isAborted {
		return
	}

	err = job.SetStatus(worker.JobStatusDone)
	if err != nil {
		return err
	}
	return
}
