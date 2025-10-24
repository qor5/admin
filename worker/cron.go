package worker

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"reflect"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cast"
)

type cronJob struct {
	JobID   string
	Pid     int
	Command string
	Delete  bool `json:"-"`
}

func (job cronJob) ToString() string {
	marshal, _ := json.Marshal(job)
	return fmt.Sprintf("## BEGIN QOR JOB %v # %v\n%v\n## END QOR JOB\n", job.JobID, string(marshal), job.Command)
}

// Cron implemented a worker Queue based on cronjob
type cron struct {
	Jobs     []*cronJob
	CronJobs []string
	mutex    sync.Mutex `sql:"-"`
}

// NewCronQueue initialize a Cron queue
func NewCronQueue() Queue {
	return &cron{}
}

func (c *cron) parseJobs() []*cronJob {
	c.mutex.Lock()

	c.Jobs = []*cronJob{}
	c.CronJobs = []string{}
	if out, err := exec.Command("crontab", "-l").Output(); err == nil {
		var inQorJob bool
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			if strings.HasPrefix(line, "## BEGIN QOR JOB") {
				inQorJob = true
				if idx := strings.Index(line, "{"); idx > 1 {
					var job cronJob
					if json.Unmarshal([]byte(line[idx-1:]), &job) == nil {
						c.Jobs = append(c.Jobs, &job)
					}
				}
			}

			if !inQorJob {
				c.CronJobs = append(c.CronJobs, line)
			}

			if strings.HasPrefix(line, "## END QOR JOB") {
				inQorJob = false
			}
		}
	}
	return c.Jobs
}

func (c *cron) writeCronJob() error {
	defer c.mutex.Unlock()

	cmd := exec.Command("crontab", "-")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	stdin, _ := cmd.StdinPipe()
	for _, cronJob := range c.CronJobs {
		stdin.Write([]byte(cronJob + "\n"))
	}

	for _, job := range c.Jobs {
		if !job.Delete {
			stdin.Write([]byte(job.ToString() + "\n"))
		}
	}
	stdin.Close()
	return cmd.Run()
}

// Add a job to cron queue
func (c *cron) Add(ctx context.Context, job QueJobInterface) (err error) {
	c.parseJobs()
	defer c.writeCronJob()

	jobInfo, err := job.GetJobInfo()
	if err != nil {
		return err
	}

	var binaryFile string
	if binaryFile, err = filepath.Abs(os.Args[0]); err == nil {
		var jobs []*cronJob
		for _, cronJob := range c.Jobs {
			if cronJob.JobID != jobInfo.JobID {
				jobs = append(jobs, cronJob)
			}
		}

		if scheduler, ok := jobInfo.Argument.(Scheduler); ok && scheduler.GetScheduleTime() != nil {
			scheduleTime := scheduler.GetScheduleTime().In(time.Local)
			job.SetStatus(JobStatusScheduled)

			currentPath, _ := os.Getwd()
			jobs = append(jobs, &cronJob{
				JobID:   jobInfo.JobID,
				Command: fmt.Sprintf("%d %d %d %d * cd %v; %v --qor-job %v\n", scheduleTime.Minute(), scheduleTime.Hour(), scheduleTime.Day(), scheduleTime.Month(), currentPath, binaryFile, jobInfo.JobID),
			})
			c.Jobs = jobs
			return nil
		}
		cmd := exec.Command(binaryFile, "--qor-job", jobInfo.JobID)
		if err = cmd.Start(); err == nil {
			jobs = append(jobs, &cronJob{JobID: jobInfo.JobID, Pid: cmd.Process.Pid})
			cmd.Process.Release()
		}
		c.Jobs = jobs
	}

	return
}

// Run a job from cron queue
func (c *cron) run(ctx context.Context, qorJob QueJobInterface) (err error) {
	jobInfo, err := qorJob.GetJobInfo()
	if err != nil {
		return err
	}

	h := qorJob.GetHandler()
	if h == nil {
		panic(fmt.Sprintf("job %v no handler", jobInfo.JobName))
	}

	go func() {
		sigint := make(chan os.Signal, 1)

		// interrupt signal sent from terminal
		signal.Notify(sigint, syscall.SIGINT)
		// sigterm signal sent from kubernetes
		signal.Notify(sigint, syscall.SIGTERM)

		i := <-sigint

		qorJob.SetProgressText(fmt.Sprintf("Worker killed by signal %s", i.String()))
		qorJob.SetStatus(JobStatusKilled)

		qorJob.StopRefresh()
		os.Exit(int(reflect.ValueOf(i).Int()))
	}()

	qorJob.StartRefresh()
	defer qorJob.StopRefresh()

	err = h(ctx, qorJob)
	if err == nil {
		c.parseJobs()
		defer c.writeCronJob()
		for _, cronJob := range c.Jobs {
			if cronJob.JobID == jobInfo.JobID {
				cronJob.Delete = true
			}
		}
	}
	return err
}

// Kill a job from cron queue
func (c *cron) Kill(ctx context.Context, job QueJobInterface) (err error) {
	c.parseJobs()
	defer c.writeCronJob()

	jobInfo, err := job.GetJobInfo()
	if err != nil {
		return err
	}

	for _, cronJob := range c.Jobs {
		if cronJob.JobID == jobInfo.JobID {
			if process, err := os.FindProcess(cronJob.Pid); err == nil {
				if err = process.Kill(); err == nil {
					cronJob.Delete = true
					return job.SetStatus(JobStatusKilled)
				}
			}
			return err
		}
	}
	return errors.New("failed to find job")
}

// Remove a job from cron queue
func (c *cron) Remove(ctx context.Context, job QueJobInterface) error {
	c.parseJobs()
	defer c.writeCronJob()

	jobInfo, err := job.GetJobInfo()
	if err != nil {
		return err
	}

	for _, cronJob := range c.Jobs {
		if cronJob.JobID == jobInfo.JobID {
			if cronJob.Pid == 0 {
				cronJob.Delete = true
				return job.SetStatus(JobStatusKilled)
			}
			return errors.New("failed to remove current job as it is running")
		}
	}
	return errors.New("failed to find job")
}

func (c *cron) Listen(_ []*QorJobDefinition, getJob func(qorJobID uint) (QueJobInterface, error)) error {
	cmdLine := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	qorJobID := cmdLine.String("qor-job", "", "Qor Job ID")
	cmdLine.Parse(os.Args[1:])

	if *qorJobID != "" {
		id, err := cast.ToUintE(*qorJobID)
		if err != nil {
			fmt.Println(err)
			return err
		}
		job, err := getJob(id)
		if err != nil {
			fmt.Println(err)
			return err
		}
		err = c.doRunJob(context.Background(), job)
		if err == nil {
			return nil
		}
		fmt.Println(err)
		return err
	}

	return nil
}

func (c *cron) doRunJob(ctx context.Context, job QueJobInterface) error {
	defer func() {
		if r := recover(); r != nil {
			job.AddLog(string(debug.Stack()))
			job.SetProgressText(fmt.Sprint(r))
			job.SetStatus(JobStatusException)
		}
	}()

	if job.GetStatus() != JobStatusNew && job.GetStatus() != JobStatusScheduled {
		return errors.New("invalid job status, current status: " + job.GetStatus())
	}

	if err := job.SetStatus(JobStatusRunning); err == nil {
		runErr := c.run(ctx, job)
		if runErr == nil {
			return job.SetStatus(JobStatusDone)
		}

		job.SetProgressText(runErr.Error())
		job.SetStatus(JobStatusException)
	}

	return nil
}

func (*cron) Shutdown(ctx context.Context) error {
	return nil
}
