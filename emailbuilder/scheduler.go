package emailbuilder

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/pkg/errors"
	"github.com/tnclong/go-que"
	"gorm.io/gorm"
)

// CampaignScheduler manages email campaign scheduling using go-que
type CampaignScheduler struct {
	Queue   string
	Enqueue func(context.Context, []que.Plan) ([]int64, error)
}

// CampaignJobArgs contains the arguments for a campaign job
type CampaignJobArgs struct {
	CampaignID uint `json:"campaignID"`
}

// ScheduleCampaign schedules a campaign based on its configuration
func (cs *CampaignScheduler) ScheduleCampaign(ctx context.Context, campaign *EmailCampaign) (int64, error) {
	// Convert campaign parameters to cron expression
	cronExpr := buildCronExpression(campaign)
	if cronExpr == "" {
		return 0, errors.Errorf("invalid schedule configuration")
	}

	// Create job arguments
	args := CampaignJobArgs{
		CampaignID: campaign.ID,
	}
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal job arguments: %w", err)
	}

	// Create retry policy - convert int to int32 for MaxRetryCount
	retryPolicy := que.RetryPolicy{
		MaxRetryCount:          int32(campaign.RetryCount),
		InitialInterval:        30 * time.Second,
		MaxInterval:            1 * time.Hour,
		NextIntervalMultiplier: 2.0,
		IntervalRandomPercent:  20,
	}

	// Create a unique ID for the job to prevent duplicates
	uniqueID := fmt.Sprintf("campaign:%d", campaign.ID)

	// Create a plan for the scheduler
	plan := que.Plan{
		Queue:           cs.Queue,
		Args:            argsJSON,
		RunAt:           campaign.StartTime,
		RetryPolicy:     retryPolicy,
		UniqueID:        &uniqueID,
		UniqueLifecycle: que.Lockable,
	}

	// Schedule the job
	ids, err := cs.Enqueue(ctx, []que.Plan{plan})
	if err != nil {
		return 0, fmt.Errorf("failed to schedule campaign: %w", err)
	}

	if len(ids) > 0 {
		return ids[0], nil
	}

	return 0, nil
}

// CancelSchedule cancels a scheduled campaign
func (cs *CampaignScheduler) CancelSchedule(ctx context.Context, campaign *EmailCampaign) error {
	// Implementation would depend on go-que's API for cancelling jobs
	// This might involve updating the job status in the database
	return nil
}

// ProcessEmailCampaignJob is the worker function to process a scheduled email campaign
func ProcessEmailCampaignJob(ctx context.Context, job que.Job, db *gorm.DB) error {
	// Extract job arguments
	var args CampaignJobArgs
	if err := json.Unmarshal(job.Plan().Args, &args); err != nil {
		return fmt.Errorf("failed to unmarshal job arguments: %w", err)
	}

	// Log job execution
	log.Printf("Processing email campaign job for campaign ID: %d", args.CampaignID)

	// Fetch the campaign
	var campaign EmailCampaign
	if err := db.First(&campaign, args.CampaignID).Error; err != nil {
		return fmt.Errorf("failed to fetch campaign: %w", err)
	}

	// Process the campaign (send emails, update status, etc.)
	// Implementation depends on your email sending logic

	// Update campaign status to sent
	campaign.Status = StatusSent
	if err := db.Save(&campaign).Error; err != nil {
		return fmt.Errorf("failed to update campaign status: %w", err)
	}

	return job.Done(ctx)
}

// RegisterEmailCampaignWorker registers the email campaign worker with go-que
func RegisterEmailCampaignWorker(db *gorm.DB, queue string, mutex que.Mutex) (*que.Worker, error) {
	worker, err := que.NewWorker(que.WorkerOptions{
		Queue: queue,
		Mutex: mutex,
		Perform: func(ctx context.Context, job que.Job) error {
			return ProcessEmailCampaignJob(ctx, job, db)
		},
		MaxLockPerSecond:          5,
		MaxPerformPerSecond:       10,
		MaxConcurrentPerformCount: 5,
		MaxBufferJobsCount:        10,
	})

	return worker, err
}
