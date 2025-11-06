package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/tnclong/go-que"
	"github.com/tnclong/go-que/pg"
	"go.uber.org/multierr"
	"gorm.io/gorm"
)

type goque struct {
	q   que.Queue
	db  *gorm.DB
	wks []*que.Worker
}

// NewGoQueQueue creates a new go-que based Queue (default queue implementation).
// Does not run migrations - call worker.AutoMigrate() first.
func NewGoQueQueue(db *gorm.DB) Queue {
	return newGoQueQueue(db)
}

// newGoQueQueue creates a Queue without migrations.
func newGoQueQueue(db *gorm.DB) Queue {
	if db == nil {
		panic("db can not be nil")
	}

	rdb, err := db.DB()
	if err != nil {
		panic(err)
	}

	// Always disable auto-migration in queue creation
	// Migration is handled by AutoMigrate() function
	q, err := pg.NewWithOptions(pg.Options{
		DB:        rdb,
		DBMigrate: false,
	})
	if err != nil {
		panic(err)
	}

	return &goque{
		q:  q,
		db: db,
	}
}

func (q *goque) Add(ctx context.Context, job QueJobInterface) error {
	jobInfo, err := job.GetJobInfo()
	if err != nil {
		return err
	}
	runAt := time.Now()
	if scheduler, ok := jobInfo.Argument.(Scheduler); ok && scheduler.GetScheduleTime() != nil {
		runAt = scheduler.GetScheduleTime().In(time.Local)
		job.SetStatus(JobStatusScheduled)
	}

	_, err = q.q.Enqueue(ctx, nil, que.Plan{
		Queue: "worker_" + jobInfo.JobName,
		Args:  que.Args(jobInfo.JobID, jobInfo.Argument),
		RunAt: runAt,
	})
	if err != nil {
		return err
	}

	return nil
}

func (*goque) run(ctx context.Context, job QueJobInterface) error {
	job.StartRefresh()
	defer job.StopRefresh()

	return job.GetHandler()(ctx, job)
}

func (*goque) Kill(ctx context.Context, job QueJobInterface) error {
	return job.SetStatus(JobStatusKilled)
}

func (*goque) Remove(ctx context.Context, job QueJobInterface) error {
	return job.SetStatus(JobStatusCancelled)
}

func (q *goque) Listen(jobDefs []*QorJobDefinition, getJob func(qorJobID uint) (QueJobInterface, error)) error {
	for i := range jobDefs {
		jd := jobDefs[i]
		if jd.Handler == nil {
			panic(fmt.Sprintf("job %s handler is nil", jd.Name))
		}
		worker, err := que.NewWorker(que.WorkerOptions{
			Queue:                     "worker_" + jd.Name,
			Mutex:                     q.q.Mutex(),
			MaxLockPerSecond:          10,
			MaxBufferJobsCount:        0,
			MaxPerformPerSecond:       2,
			MaxConcurrentPerformCount: 1,
			Perform: func(ctx context.Context, qj que.Job) (err error) {
				var job QueJobInterface
				{
					var sid string
					err = q.parseArgs(qj.Plan().Args, &sid)
					if err != nil {
						return err
					}
					id, err := strconv.Atoi(sid)
					if err != nil {
						return err
					}
					job, err = getJob(uint(id))
					if err != nil {
						return err
					}
				}

				defer func() {
					if r := recover(); r != nil {
						job.AddLog(string(debug.Stack()))
						job.SetProgressText(fmt.Sprint(r))
						job.SetStatus(JobStatusException)
						panic(r)
					}
				}()

				if job.GetStatus() == JobStatusCancelled {
					return qj.Expire(ctx, errors.New("job is cancelled"))
				}
				if job.GetStatus() != JobStatusNew && job.GetStatus() != JobStatusScheduled {
					job.SetStatus(JobStatusKilled)
					return errors.New("invalid job status, current status: " + job.GetStatus())
				}

				err = job.SetStatus(JobStatusRunning)
				if err != nil {
					return err
				}

				hctx, cf := context.WithCancel(ctx)
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
							if status == JobStatusKilled {
								isAborted = true
								cf()
								return
							}
						}
					}
				}()
				err = q.run(hctx, job)
				if !isAborted {
					hDoneC <- struct{}{}
				}
				if err != nil {
					job.SetProgressText(err.Error())
					job.SetStatus(JobStatusException)
					return err
				}
				if isAborted {
					return qj.Expire(ctx, errors.New("manually aborted"))
				}

				err = job.SetStatus(JobStatusDone)
				if err != nil {
					return err
				}
				return qj.Done(ctx)
			},
		})
		if err != nil {
			panic(err)
		}
		q.wks = append(q.wks, worker)
		go func() {
			if err := worker.Run(); err != nil {
				q.db.Create(&GoQueError{
					Error: fmt.Sprintf("worker Run() error: %s", err.Error()),
				})
			}
		}()
	}

	return nil
}

func (q *goque) Shutdown(ctx context.Context) error {
	var errs error
	for _, wk := range q.wks {
		if err := wk.Stop(ctx); err != nil {
			errs = multierr.Append(errs, err)
		}
	}
	return errs
}

func (*goque) parseArgs(data []byte, args ...interface{}) error {
	d := json.NewDecoder(bytes.NewReader(data))
	if _, err := d.Token(); err != nil {
		return err
	}
	for _, arg := range args {
		if err := d.Decode(arg); err != nil {
			return err
		}
	}
	return nil
}
