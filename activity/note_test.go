package activity

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/qor5/admin/v3/presets"
	"github.com/stretchr/testify/require"
	"github.com/theplant/testingutils"
)

func TestGetNotesCounts(t *testing.T) {
	pb := presets.New()

	ab := New(db, testCurrentUser)
	pb.Use(ab)

	amb := ab.RegisterModel(Page{})
	amb.AddKeys("ID", "VersionName")
	resetDB()

	ctx := context.Background()
	page1 := Page{ID: 1, VersionName: "v1", Title: "test"}
	page2 := Page{ID: 2, VersionName: "v1", Title: "test"}
	ctxWithCurrentUser := context.WithValue(ctx, ctxKeyCurrentUser{}, currentUser)
	ctxWithAnotherUser := context.WithValue(ctx, ctxKeyCurrentUser{}, anotherUser)
	{ // currentUser create page1 => currentUser note => anotherUser note
		_, err := ab.OnCreate(ctxWithCurrentUser, page1)
		require.NoError(t, err)

		_, err = ab.Note(ctxWithCurrentUser, page1, &Note{
			Note:         "page1:foo",
			LastEditedAt: time.Now(),
		})
		require.NoError(t, err)

		_, err = ab.Note(ctxWithAnotherUser, page1, &Note{
			Note:         "page1:bar",
			LastEditedAt: time.Now(),
		})
		require.NoError(t, err)
	}

	{
		// anotherUser create page2 => currentUser note => anotherUser note
		_, err := ab.OnCreate(ctxWithAnotherUser, page2)
		require.NoError(t, err)

		_, err = ab.Note(ctxWithCurrentUser, page2, &Note{
			Note:         "page2:foo",
			LastEditedAt: time.Now(),
		})
		require.NoError(t, err)

		_, err = ab.Note(ctxWithAnotherUser, page2, &Note{
			Note:         "page2:bar",
			LastEditedAt: time.Now(),
		})
		require.NoError(t, err)
	}

	{
		// GetNotesCounts of page1 by currentUser
		noteCounts, err := ab.GetNotesCounts(
			ctxWithCurrentUser,
			ParseModelName(&Page{}),
			[]string{amb.ParseModelKeys(page1)},
		)
		require.NoError(t, err)
		require.Len(t, noteCounts, 1)
		require.Empty(t, testingutils.PrettyJsonDiff(`
	[
	  {
		"ModelName": "Page",
		"ModelKeys": "1:v1",
		"ModelLabel": "",
		"UnreadNotesCount": 1,
		"TotalNotesCount": 2
	  }
	]
		`, noteCounts))
	}

	{
		// GetAllNotesCounts by currentUser
		noteCounts, err := ab.GetNotesCounts(
			ctxWithCurrentUser,
			ParseModelName(&Page{}),
			nil,
		)
		require.NoError(t, err)
		require.Len(t, noteCounts, 2)
		require.Empty(t, testingutils.PrettyJsonDiff(`
	[
	  {
		"ModelName": "Page",
		"ModelKeys": "1:v1",
		"ModelLabel": "",
		"UnreadNotesCount": 1,
		"TotalNotesCount": 2
	  },
	  {
		"ModelName": "Page",
		"ModelKeys": "2:v1",
		"ModelLabel": "",
		"UnreadNotesCount": 1,
		"TotalNotesCount": 2
	  }
	]
		`, noteCounts))
	}

	{
		// GetAllNotesCounts which is owned by currentUser
		noteCounts, err := ab.GetNotesCounts(
			ctxWithCurrentUser,
			ParseModelName(&Page{}),
			nil,
			presets.SQLCondition{
				// only fetch notes owned by currentUser
				Query: "scope LIKE ?",
				Args:  []any{fmt.Sprintf("%%%s%%", ScopeWithOwner(currentUser.ID))},
				// Query: "scope = ?",
				// Args:  []any{ScopeWithOwnerID(currentUser.ID)},
			},
		)
		require.NoError(t, err)
		require.Len(t, noteCounts, 1)
		require.Empty(t, testingutils.PrettyJsonDiff(`
	[
	  {
		"ModelName": "Page",
		"ModelKeys": "1:v1",
		"ModelLabel": "",
		"UnreadNotesCount": 1,
		"TotalNotesCount": 2
	  }
	]
		`, noteCounts))
	}

	{
		// currentUser view page1 => getAllNotesCounts by currentUser to check unread notes count
		_, err := ab.Log(
			ctxWithCurrentUser,
			ActionLastView,
			page1,
			nil,
		)
		require.NoError(t, err)

		noteCounts, err := ab.GetNotesCounts(
			ctxWithCurrentUser,
			ParseModelName(&Page{}),
			nil,
		)
		require.NoError(t, err)
		require.Len(t, noteCounts, 2)
		require.Empty(t, testingutils.PrettyJsonDiff(`
	[
	  {
		"ModelName": "Page",
		"ModelKeys": "1:v1",
		"ModelLabel": "",
		"UnreadNotesCount": 0,
		"TotalNotesCount": 2
	  },
	  {
		"ModelName": "Page",
		"ModelKeys": "2:v1",
		"ModelLabel": "",
		"UnreadNotesCount": 1,
		"TotalNotesCount": 2
	  }
	]
		`, noteCounts))
	}
}
