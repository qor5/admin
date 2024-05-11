package testenv

import (
	"cmp"
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupDatabase(ctx context.Context, dbUser, dbPass, dbName string) (*gorm.DB, func() error, error) {
	container, err := testcontainers.GenericContainer(ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:        "postgres:16.3-alpine",
				ExposedPorts: []string{"5432/tcp"},
				Env: map[string]string{
					"POSTGRES_USER":     dbUser,
					"POSTGRES_PASSWORD": dbPass,
					"POSTGRES_DB":       dbName,
				},
				WaitingFor: wait.ForLog("database system is ready to accept connections"),
			},
			Started: true,
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("fail to start container: %w", err)
	}

	endpoint, err := container.Endpoint(ctx, "")
	if err != nil {
		return nil, nil, fmt.Errorf("fail to get endpoint: %w", err)
	}
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPass, endpoint, dbName)

	// WARN: required, dont know why
	time.Sleep(300 * time.Millisecond)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		defer container.Terminate(ctx)
		return nil, nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("no underlying sqlDB: %w", err)
	}

	return db, func() error {
		return cmp.Or(
			sqlDB.Close(),
			container.Terminate(context.Background()),
		)
	}, nil
}
