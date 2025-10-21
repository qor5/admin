package starter_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/qor5/admin/v3/starter"
	"github.com/qor5/x/v3/gormx"
	"github.com/stretchr/testify/require"
	"github.com/theplant/inject"
	"github.com/theplant/inject/lifecycle"
	"gorm.io/gorm"

	_ "embed"
)

func TestMain(m *testing.M) {
	m.Run()
}

func setupDummyUserOptions() *starter.UpsertUserOptions {
	return &starter.UpsertUserOptions{
		Email:    "test@example.com",
		Password: "test123456789",
		Role:     []string{starter.RoleAdmin},
	}
}

func setupTestConfig(ctx context.Context) (*starter.Config, error) {
	loader, err := starter.InitializeConfig()
	if err != nil {
		return nil, err
	}
	conf, err := loader(ctx, "testdata/config.yaml")
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func setupTestHandlerFactory(ctors ...any) []any {
	ctors = append(ctors,
		setupDummyUserOptions,
	)
	return []any{
		starter.SetupTestHandlerFactory(ctors...),
	}
}

type env struct {
	lc      *lifecycle.Lifecycle
	handler *starter.Handler
}

func newTestEnv(t *testing.T, ctors ...any) *env {
	lc, err := lifecycle.Start(context.Background(),
		lifecycle.SetupSignal,
		gormx.SetupTestSuiteFactory(),
		func(testSuite *gormx.TestSuite) *gorm.DB {
			return testSuite.DB()
		},
		setupTestConfig,
		http.NewServeMux,
		setupTestHandlerFactory(ctors...),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := lc.Stop(context.Background())
		require.NoError(t, err)
	})

	handler := inject.MustResolve[*starter.Handler](lc)

	return &env{
		lc:      lc,
		handler: handler,
	}
}
