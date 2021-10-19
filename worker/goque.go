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
	"gorm.io/gorm"
)

type goque struct {
	db *gorm.DB
	q  que.Queue
}

func NewGoQueQueue(db *gorm.DB) Queue {
	if db == nil {
		panic("db can not be nil")
	}

	var q que.Queue
	{
		rdb, err := db.DB()
		if err != nil {
			panic(err)
		}
		q, err = pg.New(rdb)
		if err != nil {
			panic(err)
		}
	}

	return &goque{
		db: db,
		q:  q,
	}
}

func (q *goque) Add(job QorJobInterface) error {
	args, err := job.GetArgument()
	if err != nil {
		return err
	}
	runAt := time.Now()
	if scheduler, ok := args.(Scheduler); ok && scheduler.GetScheduleTime() != nil {
		runAt = scheduler.GetScheduleTime().In(time.Local)
		job.SetStatus(JobStatusScheduled)
	}

	ids, err := q.q.Enqueue(context.Background(), nil, que.Plan{
		Queue: "worker_" + job.GetJobName(),
		Args:  que.Args(job.GetJobID(), args),
		RunAt: runAt,
	})
	if err != nil {
		return err
	}
	return job.SetQueJobID(ids[0])
}

func (q *goque) run(ctx context.Context, job QorJobInterface) error {
	job.StartRefresh()
	defer job.StopRefresh()

	return job.GetHandler()(ctx, job)
}

func (q *goque) expireJob(id int64) error {
	return q.db.Exec("update goque_jobs set expired_at = now() where id = ?", id).Error
}

func (q *goque) isJobExpired(id int64) (bool, error) {
	db, err := q.db.DB()
	if err != nil {
		return false, err
	}

	var expiredAt time.Time
	err = db.QueryRow("select expired_at from goque_jobs where id = $1", id).Scan(&expiredAt)
	if err != nil {
		return false, err
	}
	if time.Now().Sub(expiredAt) > 0 {
		return true, nil
	}

	return false, nil
}

func (q *goque) Kill(job QorJobInterface) error {
	err := job.SetStatus(JobStatusKilled)
	if err != nil {
		return err
	}

	return q.expireJob(job.GetQueJobID())
}

func (q *goque) Remove(job QorJobInterface) error {
	err := job.SetStatus(JobStatusKilled)
	if err != nil {
		return err
	}

	return q.expireJob(job.GetQueJobID())
}

func (q *goque) Listen(jobDefs []*QorJobDefinition, getJob func(qorJobID uint) (QorJobInterface, error)) error {
	for i, _ := range jobDefs {
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
				var job QorJobInterface
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

				if job.GetStatus() != JobStatusNew && job.GetStatus() != JobStatusScheduled {
					job.SetStatus(JobStatusKilled)
					return errors.New("invalid job status, current status: " + job.GetStatus())
				}

				err = job.SetStatus(JobStatusRunning)
				if err != nil {
					return err
				}

				hctx, cf := context.WithCancel(context.Background())
				queJobID := job.GetQueJobID()
				hDoneC := make(chan struct{})
				isExpired := false
				go func() {
					timer := time.NewTicker(time.Second)
					for {
						select {
						case <-hDoneC:
							return
						case <-timer.C:
							isExpired, _ = q.isJobExpired(queJobID)
							if isExpired {
								job.SetStatus(JobStatusKilled)
								cf()
								return
							}
						}
					}
				}()
				err = q.run(hctx, job)
				if !isExpired {
					hDoneC <- struct{}{}
				}
				if err != nil {
					job.SetProgressText(err.Error())
					job.SetStatus(JobStatusException)
					return err
				}

				if !isExpired {
					err = job.SetStatus(JobStatusDone)
					if err != nil {
						return err
					}
				}
				return qj.Done(ctx)
			},
		})
		if err != nil {
			panic(err)
		}

		go func() {
			err := worker.Run()
			fmt.Println("worker Run() error:", err)
		}()
	}

	return nil
}

func (q *goque) parseArgs(data []byte, args ...interface{}) error {
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
