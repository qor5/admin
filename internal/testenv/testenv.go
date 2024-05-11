package testenv

import (
	"cmp"
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"gorm.io/gorm"
)

type TestEnvBuilder struct {
	ctx                    context.Context
	dbUser, dbPass, dbName string
}

func New() *TestEnvBuilder {
	return &TestEnvBuilder{}
}

func (b *TestEnvBuilder) DBUser(v string) *TestEnvBuilder {
	b.dbUser = v
	return b
}

func (b *TestEnvBuilder) DBPass(v string) *TestEnvBuilder {
	b.dbPass = v
	return b
}

func (b *TestEnvBuilder) DBName(v string) *TestEnvBuilder {
	b.dbName = v
	return b
}

func (b *TestEnvBuilder) Context(ctx context.Context) *TestEnvBuilder {
	b.ctx = ctx
	return b
}

type TestEnv struct {
	DB       *gorm.DB
	tearDown func() error
	tornDown atomic.Bool
}

func (env *TestEnv) TearDown() error {
	if !env.tornDown.CompareAndSwap(false, true) {
		return errors.New("torn down")
	}
	return env.tearDown()
}

func (b *TestEnvBuilder) SetUp() (env *TestEnv, xerr error) {
	ctx := cmp.Or(b.ctx, context.Background())
	dbUser := cmp.Or(b.dbUser, "test_user")
	dbPass := cmp.Or(b.dbPass, "test_pass")
	dbName := cmp.Or(b.dbName, "test_db")

	cctx, ccancel := context.WithTimeout(ctx, 10*time.Second)
	defer ccancel()
	db, dbCloser, err := setupDatabase(cctx, dbUser, dbPass, dbName)
	if err != nil {
		return nil, err
	}
	return &TestEnv{
		DB: db,
		tearDown: func() error {
			return dbCloser()
		},
	}, nil
}

func (b *TestEnvBuilder) SetUpWithT(t *testing.T) (*TestEnv, error) {
	env, err := b.SetUp()
	if err != nil {
		return nil, err
	}
	t.Cleanup(func() {
		if err := env.TearDown(); err != nil {
			t.Logf("fail to tear down: %v", err)
		}
	})
	return env, nil
}
