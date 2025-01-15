package redirection

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/qor5/web/v3"

	"github.com/qor5/x/v3/oss"
	"github.com/qor5/x/v3/oss/s3"
	v "github.com/qor5/x/v3/ui/vuetify"

	"github.com/qor5/admin/v3/presets"
)

const (
	defaultConcurrency = 10
	defaultTimeout     = 5
)

type uploadFiles struct {
	NewFiles []*multipart.FileHeader
}

func (b *Builder) Get(ctx context.Context, path string) (*os.File, error) {
	return b.storage.Get(ctx, path)
}

func (b *Builder) GetStream(ctx context.Context, path string) (io.ReadCloser, error) {
	return b.storage.GetStream(ctx, path)
}

func (b *Builder) Put(ctx context.Context, path string, reader io.Reader) (obj *oss.Object, err error) {
	defer func() {
		if err == nil {
			b.createEmptyTargetRecord(path)
		}
	}()
	return b.storage.Put(ctx, path, reader)
}

func (b *Builder) Delete(ctx context.Context, path string) error {
	return b.storage.Delete(ctx, path)
}

func (b *Builder) List(ctx context.Context, path string) ([]*oss.Object, error) {
	return b.storage.List(ctx, path)
}

func (b *Builder) GetURL(ctx context.Context, path string) (string, error) {
	return b.storage.GetURL(ctx, path)
}

func (b *Builder) GetEndpoint(ctx context.Context) string {
	return b.storage.GetEndpoint(ctx)
}

func (b *Builder) createEmptyTargetRecord(path string) {
	var (
		tx  = b.db.Begin()
		err error
		m   ObjectRedirection
	)
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	tx.Where(`Source = ?`, path).Order("created_at desc").First(&m)
	if m.ID != 0 && m.Target == "" {
		return
	}
	err = tx.Create(&ObjectRedirection{
		Source: path,
		Target: "",
	}).Error
}

var (
	repeatedSourceErrors = errors.New("RepeatedSource")
	unreachableErrors    = errors.New("SomeURLAreUnreachable")
)

func (b *Builder) importRecords(ctx *web.EventContext, r *web.EventResponse) (err error) {
	var (
		uf             uploadFiles
		file           multipart.File
		body           []byte
		records        []ObjectRedirection
		existedSource  = make(map[string]bool)
		repeatedSource []string
		items          []ObjectRedirection
		urls           []string
	)
	ctx.MustUnmarshalForm(&uf)
	for _, fh := range uf.NewFiles {
		if file, err = fh.Open(); err != nil {
			return
		}
		if body, err = io.ReadAll(file); err != nil {
			return
		}
		if err = gocsv.UnmarshalBytes(body, &items); err != nil {
			return
		}
		for _, item := range items {
			if existedSource[item.Source] {
				if !slices.Contains(repeatedSource, item.Source) {
					repeatedSource = append(repeatedSource, item.Source)
				}
				continue
			}
			existedSource[item.Source] = true
			records = append(records, item)
		}
	}
	if len(repeatedSource) > 0 {
		presets.ShowMessage(r,
			fmt.Sprintf("%v:\n %v", repeatedSourceErrors.Error(),
				strings.Join(repeatedSource, "\n")), v.ColorError)
		return
	}
	if !checkURLsBatch(ctx.R, urls) {
		presets.ShowMessage(r, unreachableErrors.Error(), v.ColorError)
		return
	}
	tx := b.db.Begin()
	defer func() {
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()
	if err = tx.Save(&records).Error; err != nil {
		return
	}
	for _, item := range records {
		con := context.Background()
		context.WithValue(con, s3.WebsiteRedirectLocation{}, item.Source)
		if _, err = b.Put(con, item.Target, bytes.NewReader([]byte{})); err != nil {
			return
		}
	}
	return
}

func checkURL(url string) bool {
	client := http.Client{
		Timeout: time.Duration(defaultTimeout) * time.Second, // Set a timeout for the HTTP client
	}
	resp, err := client.Get(url) // Send a GET request to the URL
	if err != nil {
		return false // Return false if there is an error (e.g., unreachable URL)
	}
	defer resp.Body.Close() // Ensure the response body is closed after use
	return true             // Return true if the request succeeds
}

func constructFullURL(r *http.Request, relativePath string) string {
	// Determine the scheme
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	// Parse the base URL
	baseURL := &url.URL{
		Scheme: scheme,
		Host:   r.Host,
	}

	// Resolve the relative path
	relativeURL, err := url.Parse(relativePath)
	if err != nil {
		panic(err) // Handle error in production code
	}

	// Resolve relative URL against the base
	fullURL := baseURL.ResolveReference(relativeURL)

	return fullURL.String()
}

func prependBaseURL(r *http.Request, urls []string) {
	for i := 0; i < len(urls); i++ {
		if !strings.HasPrefix(urls[i], "http") {
			urls[i] = constructFullURL(r, urls[i])
		}
	}
}

func checkURLsBatch(r *http.Request, urls []string) bool {
	var wg sync.WaitGroup                          // WaitGroup to ensure all goroutines finish
	urlChan := make(chan string, len(urls))        // Channel for distributing URLs to workers
	resultChan := make(chan bool)                  // Channel for collecting successful results
	sem := make(chan struct{}, defaultConcurrency) // Semaphore to limit concurrency
	abort := make(chan struct{})                   // Channel to signal an early exit when a failure is detected
	prependBaseURL(r, urls)
	// Launch worker goroutines
	for i := 0; i < defaultConcurrency; i++ {
		wg.Add(1) // Increment WaitGroup counter for each goroutine
		go func() {
			defer wg.Done() // Decrement WaitGroup counter when the goroutine finishes
			for {
				select {
				case <-abort: // If an abort signal is received, exit the goroutine
					return
				case url, ok := <-urlChan: // Receive a URL from the channel
					if !ok { // If the channel is closed, exit the goroutine
						return
					}
					sem <- struct{}{}       // Acquire a semaphore slot
					result := checkURL(url) // Check the URL
					<-sem                   // Release the semaphore slot
					if !result {            // If the URL check fails
						select {
						case abort <- struct{}{}: // Send an abort signal
						default: // Prevent duplicate abort signals
						}
						return // Exit the goroutine immediately
					}
					resultChan <- true // Send success result to the result channel
				}
			}
		}()
	}

	// Send URLs to the channel
	go func() {
		for _, url := range urls {
			urlChan <- url // Send each URL to the channel
		}
		close(urlChan) // Close the URL channel after sending all URLs
	}()

	// Wait for all goroutines to finish and close result-related channels
	go func() {
		wg.Wait()         // Wait for all goroutines to complete
		close(resultChan) // Close the result channel
		close(abort)      // Close the abort channel
	}()

	// Collect results or handle aborts
	for range urls {
		select {
		case <-abort: // If an abort signal is received, return false
			return false
		case <-resultChan: // Consume successful results
		}
	}

	return true // If all URLs are reachable, return true
}
