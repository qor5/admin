package bq

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// TestBigQueryExecution demonstrates executing SQL queries with BigQuery SDK.
// This test is skipped by default as it requires a real BigQuery environment.
// To run this test, set GOOGLE_APPLICATION_CREDENTIALS to a valid service account key file.
func TestBigQueryExecution(t *testing.T) {
	t.Skip("Skipping BigQuery execution test as it requires real credentials")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get the user's home directory to properly resolve credential path
	// Go doesn't expand ~ automatically unlike shell environments
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	credPath := filepath.Join(homeDir, ".config/gcloud/application_default_credentials.json")

	// Using WithCredentialsFile for demonstration purposes only
	// In production, credentials should be set via GOOGLE_APPLICATION_CREDENTIALS environment variable
	client, err := bigquery.NewClient(ctx, "product-data-sandbox",
		option.WithCredentialsFile(credPath))
	if err != nil {
		t.Fatalf("Failed to create BigQuery client: %v", err)
	}
	defer client.Close()

	// Define common result type used across subtests
	type Result struct {
		Value int64  `bigquery:"value"`
		Name  string `bigquery:"name"`
	}

	// 1. Basic query without parameters
	t.Run("BasicQuery", func(t *testing.T) {
		baseQuery := "SELECT 1 as value, 'test' as name"
		filterQuery := "SELECT * FROM (" + baseQuery + ") WHERE value = 1"

		q := client.Query(filterQuery)

		it, err := q.Read(ctx)
		if err != nil {
			t.Fatalf("Failed to execute query: %v", err)
		}

		var results []Result
		for {
			var row Result
			err := it.Next(&row)
			if err == iterator.Done {
				break
			}
			if err != nil {
				t.Fatalf("Error iterating over results: %v", err)
			}
			results = append(results, row)
		}

		assert.Len(t, results, 1)
		if len(results) > 0 {
			assert.Equal(t, int64(1), results[0].Value)
			assert.Equal(t, "test", results[0].Name)
		}
	})

	// 2. Demonstrating Named Parameters
	t.Run("NamedParameters", func(t *testing.T) {
		// Using @paramName syntax for named parameters
		namedParamQuery := "SELECT @value as value, @name as name"
		q := client.Query(namedParamQuery)

		// Setting parameter values
		q.Parameters = []bigquery.QueryParameter{
			{
				Name:  "value",
				Value: 42,
			},
			{
				Name:  "name",
				Value: "test-named-param",
			},
		}

		it, err := q.Read(ctx)
		if err != nil {
			t.Fatalf("Failed to execute named parameter query: %v", err)
		}

		var namedResults []Result
		for {
			var row Result
			err := it.Next(&row)
			if err == iterator.Done {
				break
			}
			if err != nil {
				t.Fatalf("Error iterating over named parameter results: %v", err)
			}
			namedResults = append(namedResults, row)
		}

		assert.Len(t, namedResults, 1)
		if len(namedResults) > 0 {
			assert.Equal(t, int64(42), namedResults[0].Value)
			assert.Equal(t, "test-named-param", namedResults[0].Name)
		}
	})

	// 3. Demonstrating Positional Parameters
	t.Run("PositionalParameters", func(t *testing.T) {
		// Using ? syntax for positional parameters
		posParamQuery := "SELECT ? as value, ? as name"
		q := client.Query(posParamQuery)

		// Setting parameter values in order matching the ? placeholders
		q.Parameters = []bigquery.QueryParameter{
			{
				Value: 99,
			},
			{
				Value: "test-positional-param",
			},
		}

		it, err := q.Read(ctx)
		if err != nil {
			t.Fatalf("Failed to execute positional parameter query: %v", err)
		}

		var posResults []Result
		for {
			var row Result
			err := it.Next(&row)
			if err == iterator.Done {
				break
			}
			if err != nil {
				t.Fatalf("Error iterating over positional parameter results: %v", err)
			}
			posResults = append(posResults, row)
		}

		assert.Len(t, posResults, 1)
		if len(posResults) > 0 {
			assert.Equal(t, int64(99), posResults[0].Value)
			assert.Equal(t, "test-positional-param", posResults[0].Name)
		}
	})

	// 4. Demonstrating Complex Parameters (arrays and structs)
	t.Run("ComplexParameters", func(t *testing.T) {
		// Using array parameters
		arrayParamQuery := "SELECT * FROM UNNEST(@values) as value"
		q := client.Query(arrayParamQuery)

		// Setting array parameter values
		q.Parameters = []bigquery.QueryParameter{
			{
				Name:  "values",
				Value: []int64{1, 2, 3, 4, 5},
			},
		}

		it, err := q.Read(ctx)
		if err != nil {
			t.Fatalf("Failed to execute array parameter query: %v", err)
		}

		var arrayResults []struct {
			Value int64 `bigquery:"value"`
		}

		for {
			var row struct {
				Value int64 `bigquery:"value"`
			}
			err := it.Next(&row)
			if err == iterator.Done {
				break
			}
			if err != nil {
				t.Fatalf("Error iterating over array parameter results: %v", err)
			}
			arrayResults = append(arrayResults, row)
		}

		assert.Len(t, arrayResults, 5)
	})
}
