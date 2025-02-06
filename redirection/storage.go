package redirection

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/qor5/x/v3/oss"
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
		// m   Redirection
	)
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	err = tx.Create(&Redirection{
		Source: path,
		Target: "",
	}).Error
}

func (b *Builder) saver(ctx context.Context, record *Redirection) (err error) {
	tx := b.db.Begin()
	defer func() {
		if err == nil {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	if err = tx.Save(&record).Error; err != nil {
		return
	}
	err = b.redirection(ctx, record)
	return
}

func (b *Builder) redirection(ctx context.Context, record *Redirection) (err error) {
	var (
		client = b.s3Client
		bucket = client.Config.Bucket
		source = strings.TrimPrefix(record.Source, "/")
	)
	target := record.Target
	if !strings.HasPrefix(target, "http") {
		target = path.Join("/", record.Target)
	}
	if b.checkObjectExists(ctx, source) {
		_, err = client.S3.CopyObject(ctx, &s3.CopyObjectInput{
			Bucket:                  aws.String(bucket),
			CopySource:              aws.String(path.Join(bucket, record.Source)),
			Key:                     aws.String(strings.TrimPrefix(record.Source, "/")),
			WebsiteRedirectLocation: aws.String(target),
		})
	} else {
		params := &s3.PutObjectInput{
			Bucket:                  aws.String(client.Config.Bucket),
			Key:                     aws.String(client.ToS3Key(source)),
			ACL:                     types.ObjectCannedACL(client.Config.ACL),
			Body:                    bytes.NewReader([]byte{}),
			WebsiteRedirectLocation: aws.String(target),
		}
		if client.Config.CacheControl != "" {
			params.CacheControl = aws.String(client.Config.CacheControl)
		}
		_, err = client.S3.PutObject(ctx, params)
	}
	return
}

func (b *Builder) checkObjectExists(ctx context.Context, key string) bool {
	_, err := b.s3Client.S3.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(b.s3Client.Config.Bucket),
		Key:    aws.String(strings.TrimPrefix(key, "/")),
	})
	return err == nil
}

// checkURL checks if a single URL is reachable.
func checkURL(url string) bool {
	client := http.Client{
		Timeout: time.Duration(defaultTimeout) * time.Second, // Set a timeout for the HTTP client
	}
	resp, err := client.Get(url) // Send a GET request to the URL
	if err != nil {
		return false
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	return resp.StatusCode/100 == 2
}

// checkURLsBatch checks a batch of URLs and returns only the ones that failed.
func checkURLsBatch(urls map[string][]string) []string {
	var wg sync.WaitGroup                          // WaitGroup to ensure all goroutines finish
	urlChan := make(chan string, len(urls))        // Channel for distributing URLs to workers
	resultChan := make(chan string, len(urls))     // Channel for collecting failed results
	sem := make(chan struct{}, defaultConcurrency) // Semaphore to limit concurrency

	// Launch worker goroutines
	for i := 0; i < defaultConcurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done() // Decrement WaitGroup counter when the goroutine finishes
			for url := range urlChan {
				sem <- struct{}{}       // Acquire a semaphore slot
				status := checkURL(url) // Check the URL
				<-sem                   // Release the semaphore slot

				// If the URL check fails, send it to the result channel
				if !status {
					resultChan <- url
				}
			}
		}()
	}

	// Send URLs to the channel
	go func() {
		for u := range urls {
			urlChan <- u
		}
		close(urlChan) // Close the URL channel after sending all URLs
	}()

	// Wait for all goroutines to finish and close the result channel
	go func() {
		wg.Wait()
		close(resultChan) // Close the result channel
	}()

	// Collect failed results
	var failedURLs []string
	for failedURL := range resultChan {
		failedURLs = append(failedURLs, failedURL)
	}

	return failedURLs
}
