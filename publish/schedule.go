package publish

import (
	"time"
)

type Schedule struct {
	ScheduledStartAt *time.Time `gorm:"index"`
	ScheduledEndAt   *time.Time `gorm:"index"`

	ActualStartAt *time.Time
	ActualEndAt   *time.Time
}

func (schedule Schedule) GetScheduledStartAt() *time.Time {
	return schedule.ScheduledStartAt
}

func (schedule Schedule) GetScheduledEndAt() *time.Time {
	return schedule.ScheduledEndAt
}

func (schedule *Schedule) SetScheduledStartAt(v *time.Time) {
	schedule.ScheduledStartAt = v
}

func (schedule *Schedule) SetScheduledEndAt(v *time.Time) {
	schedule.ScheduledEndAt = v
}

func (schedule Schedule) GetPublishedAt() *time.Time {
	return schedule.ActualStartAt
}

func (schedule Schedule) GetUnPublishedAt() *time.Time {
	return schedule.ActualEndAt
}

func (schedule *Schedule) SetPublishedAt(v *time.Time) {
	schedule.ActualStartAt = v
}

func (schedule *Schedule) SetUnPublishedAt(v *time.Time) {
	schedule.ActualEndAt = v
}
