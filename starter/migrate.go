package starter

import (
	"context"

	"github.com/pkg/errors"
	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/login"
	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/role"
	"github.com/qor5/admin/v3/seo"
	"github.com/qor5/x/v3/perm"
	"gorm.io/gorm"
)

// AutoMigrate performs database migrations
func AutoMigrate(ctx context.Context, db *gorm.DB) error {
	db = db.WithContext(ctx)
	if err := db.AutoMigrate(
		&role.Role{},
		&User{},
		&perm.DefaultDBPolicy{},
	); err != nil {
		return errors.Wrap(err, "failed to auto migrate database")
	}

	if err := activity.AutoMigrate(db, ""); err != nil {
		return errors.Wrap(err, "failed to auto migrate activity")
	}

	if err := pagebuilder.AutoMigrate(db); err != nil {
		return errors.Wrap(err, "failed to auto migrate pagebuilder")
	}

	if err := seo.Migrate(db); err != nil {
		return errors.Wrap(err, "failed to auto migrate seo")
	}

	if err := createDefaultRolesIfEmpty(ctx, db); err != nil {
		return errors.Wrap(err, "failed to initialize default roles")
	}
	if err := login.AutoMigrateSession(db, ""); err != nil {
		return errors.Wrap(err, "failed to auto migrate login session")
	}
	return nil
}
