package presets

import (
	"testing"
	"time"

	"github.com/dustin/go-humanize"
)

func TestMessages_en_US_HumanizeTime(t *testing.T) {
	msgr := Messages_en_US

	now := time.Now()
	cases := []struct {
		name, got, exp string
	}{
		{"now", msgr.HumanizeTime(now), "now"},
		{"1 second ago", msgr.HumanizeTime(now.Add(-1 * time.Second)), "1 second ago"},
		{"12 seconds ago", msgr.HumanizeTime(now.Add(-12 * time.Second)), "12 seconds ago"},
		{"30 seconds ago", msgr.HumanizeTime(now.Add(-30 * time.Second)), "30 seconds ago"},
		{"45 seconds ago", msgr.HumanizeTime(now.Add(-45 * time.Second)), "45 seconds ago"},
		{"1 minute ago", msgr.HumanizeTime(now.Add(-63 * time.Second)), "1 minute ago"},
		{"15 minutes ago", msgr.HumanizeTime(now.Add(-15 * time.Minute)), "15 minutes ago"},
		{"1 hour ago", msgr.HumanizeTime(now.Add(-63 * time.Minute)), "1 hour ago"},
		{"2 hours ago", msgr.HumanizeTime(now.Add(-2 * time.Hour)), "2 hours ago"},
		{"21 hours ago", msgr.HumanizeTime(now.Add(-21 * time.Hour)), "21 hours ago"},
		{"1 day ago", msgr.HumanizeTime(now.Add(-26 * time.Hour)), "1 day ago"},
		{"2 days ago", msgr.HumanizeTime(now.Add(-49 * time.Hour)), "2 days ago"},
		{"3 days ago", msgr.HumanizeTime(now.Add(-3 * humanize.Day)), "3 days ago"},
		{"1 week ago (1)", msgr.HumanizeTime(now.Add(-7 * humanize.Day)), "1 week ago"},
		{"1 week ago (2)", msgr.HumanizeTime(now.Add(-12 * humanize.Day)), "1 week ago"},
		{"2 weeks ago", msgr.HumanizeTime(now.Add(-15 * humanize.Day)), "2 weeks ago"},
		{"1 month ago", msgr.HumanizeTime(now.Add(-39 * humanize.Day)), "1 month ago"},
		{"3 months ago", msgr.HumanizeTime(now.Add(-99 * humanize.Day)), "3 months ago"},
		{"1 year ago (1)", msgr.HumanizeTime(now.Add(-365 * humanize.Day)), "1 year ago"},
		{"1 year ago (2)", msgr.HumanizeTime(now.Add(-400 * humanize.Day)), "1 year ago"},
		{"2 years ago (1)", msgr.HumanizeTime(now.Add(-548 * humanize.Day)), "2 years ago"},
		{"2 years ago (2)", msgr.HumanizeTime(now.Add(-725 * humanize.Day)), "2 years ago"},
		{"2 years ago (3)", msgr.HumanizeTime(now.Add(-800 * humanize.Day)), "2 years ago"},
		{"3 years ago", msgr.HumanizeTime(now.Add(-3 * humanize.Year)), "3 years ago"},
		{"long ago", msgr.HumanizeTime(now.Add(-humanize.LongTime)), "a long while ago"},
	}
	for _, test := range cases {
		if test.got != test.exp {
			t.Errorf("On %v, expected '%v', but got '%v'",
				test.name, test.exp, test.got)
		}
	}
}
