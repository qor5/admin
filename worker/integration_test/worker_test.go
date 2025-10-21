package integration_test

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/qor5/admin/v3/worker"
	integration "github.com/qor5/admin/v3/worker/integration_test"
)

func TestJobSelectList(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/workers?__execute_event__=presets_New", http.NoBody)
	w := httptest.NewRecorder()
	pb.ServeHTTP(w, r)
	body := w.Body.String()
	expectItems := []string{"noArgJob", "progressTextJob", "argJob", "longRunningJob", "scheduleJob", "errorJob", "panicJob"}
	for _, ei := range expectItems {
		if ok := strings.Contains(body, ei); !ok {
			t.Fatalf("want item %q, but not found\n", ei)
		}
	}
}

func TestJobForm(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/workers?__execute_event__=worker_selectJob&jobName=argJob", http.NoBody)
	w := httptest.NewRecorder()
	pb.ServeHTTP(w, r)
	body := w.Body.String()
	expectItems := []string{"F1", "F2", "F3"}
	for _, ei := range expectItems {
		if ok := strings.Contains(body, ei); !ok {
			t.Fatalf("want item %q, but not found\n", ei)
		}
	}
}

func TestJobLog(t *testing.T) {
	cleanData()
	mustCreateJob(map[string]string{
		"Job": "noArgJob",
	})
	integration.ConsumeQueItem()
	j := mustGetFirstJob()
	r := httptest.NewRequest(http.MethodPost, fmt.Sprintf(`/workers?__execute_event__=worker_updateJobProgressing&job=noArgJob&jobID=%d`, j.ID), http.NoBody)
	w := httptest.NewRecorder()
	pb.ServeHTTP(w, r)
	body := w.Body.String()
	expectItems := []string{"hoho1", "hoho2", "hoho3"}
	for _, ei := range expectItems {
		if ok := strings.Contains(body, ei); !ok {
			t.Fatalf("want item %q, but not found\n", ei)
		}
	}

	cleanData()
	mustCreateJob(map[string]string{
		"Job": "argJob",
		"F1":  "f1val",
		"F2":  "2",
		"F3":  "true",
	})
	integration.ConsumeQueItem()
	j = mustGetFirstJob()
	r = httptest.NewRequest(http.MethodPost, fmt.Sprintf(`/workers?__execute_event__=worker_updateJobProgressing&job=argJob&jobID=%d`, j.ID), http.NoBody)
	w = httptest.NewRecorder()
	pb.ServeHTTP(w, r)
	body = w.Body.String()
	expectItems = []string{"F1", "f1val", "F2:2", "F3:true"}
	for _, ei := range expectItems {
		if ok := strings.Contains(body, ei); !ok {
			t.Fatalf("want item %q, but not found\n", ei)
		}
	}
}

func TestJobActions(t *testing.T) {
	// abort
	{
		cleanData()
		mustCreateJob(map[string]string{
			"Job": "longRunningJob",
		})
		j := mustGetFirstJob()
		go integration.ConsumeQueItem()
		time.Sleep(time.Second)
		r := httptest.NewRequest(http.MethodPost, fmt.Sprintf(`/workers/%d?__execute_event__=worker_abortJob&job=longRunningJob&jobID=%d`, j.ID, j.ID), http.NoBody)
		w := httptest.NewRecorder()
		pb.ServeHTTP(w, r)
		time.Sleep(2 * time.Second)
		r = httptest.NewRequest(http.MethodPost, fmt.Sprintf(`/workers?__execute_event__=worker_updateJobProgressing&job=longRunningJob&jobID=%d`, j.ID), http.NoBody)
		w = httptest.NewRecorder()
		pb.ServeHTTP(w, r)
		body := w.Body.String()
		expectItems := []string{"Killed", "job aborted"}
		for _, ei := range expectItems {
			if ok := strings.Contains(body, ei); !ok {
				t.Fatalf("want item %q, but not found\n", ei)
			}
		}
	}

	// rerun
	{
		cleanData()
		mustCreateJob(map[string]string{
			"Job": "noArgJob",
		})
		integration.ConsumeQueItem()
		j := mustGetFirstJob()
		if j.Status != worker.JobStatusDone {
			t.Fatalf("want status %q, got %q", worker.JobStatusDone, j.Status)
		}
		r := httptest.NewRequest(http.MethodPost, fmt.Sprintf(`/workers/%d?__execute_event__=worker_rerunJob&job=noArgJob&jobID=%d`, j.ID, j.ID), http.NoBody)
		w := httptest.NewRecorder()
		pb.ServeHTTP(w, r)
		j = mustGetFirstJob()
		if j.Status != worker.JobStatusNew {
			t.Fatalf("want status %q, got %q", worker.JobStatusNew, j.Status)
		}
	}

	// update
	{
		cleanData()
		mustCreateJob(map[string]string{
			"Job":          "scheduleJob",
			"ScheduleTime": time.Now().Add(time.Hour).Local().Format(`2006-01-02 15:04`),
		})
		j := mustGetFirstJob()
		rBody := bytes.NewBuffer(nil)
		mw := multipart.NewWriter(rBody)
		st := time.Now().Add(2 * time.Hour).Local().Format(`2006-01-02 15:04`)
		mw.WriteField("ScheduleTime", st)
		mw.Close()
		r := httptest.NewRequest(http.MethodPost, fmt.Sprintf(`/workers/%d?__execute_event__=worker_updateJob&job=scheduleJob&jobID=%d`, j.ID, j.ID), rBody)
		r.Header.Add("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", mw.Boundary()))
		w := httptest.NewRecorder()
		pb.ServeHTTP(w, r)
		r = httptest.NewRequest(http.MethodGet, fmt.Sprintf(`/workers/%d`, j.ID), http.NoBody)
		w = httptest.NewRecorder()
		pb.ServeHTTP(w, r)
		if !strings.Contains(w.Body.String(), st) {
			t.Fatalf("want updated schedule time %q", st)
		}
	}
}
