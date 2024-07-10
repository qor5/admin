package activity

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type MultiPrimaryKey struct {
	ID          uint   `gorm:"primary_key"`
	Version     string `gorm:"primary_key"`
	Title       string
	Description string
}

func TestFindOld(t *testing.T) {
	require.NoError(t, db.AutoMigrate(&MultiPrimaryKey{}))
	// old := MultiPrimaryKey{}
	// require.NoError(t, db.Where(&MultiPrimaryKey{
	// 	ID:          1,
	// 	Version:     "v1",
	// 	Title:       "title",
	// 	Description: "description",
	// }).First(&old).Error)

	db.First(&MultiPrimaryKey{
		ID:          1,
		Version:     "v1",
		Title:       "title",
		Description: "description",
	})
}
