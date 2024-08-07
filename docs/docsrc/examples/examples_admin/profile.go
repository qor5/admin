package examples_admin

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	plogin "github.com/qor5/admin/v3/login"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	v "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

type ProfileUser struct {
	gorm.Model
	Email   string
	Name    string
	Avatar  string
	Role    string
	Status  string
	Company string
}

func profileExample(b *presets.Builder, db *gorm.DB, currentUser *ProfileUser, customize func(profileB *plogin.ProfileBuilder)) http.Handler {
	b.GetI18n().SupportLanguages(language.English, language.SimplifiedChinese, language.Japanese)
	b.DataOperator(gorm2op.DataOperator(db))

	getCurrentUser := func(_ context.Context) (*ProfileUser, error) {
		return currentUser, nil
	}

	profielB := plogin.NewProfileBuilder(
		func(ctx context.Context) (*plogin.Profile, error) {
			u, err := getCurrentUser(ctx)
			if err != nil {
				return nil, err
			}
			return &plogin.Profile{
				ID:     fmt.Sprint(u.ID),
				Name:   u.Name,
				Avatar: u.Avatar,
				Roles:  []string{u.Role},
				Status: u.Status,
				Fields: []*plogin.ProfileField{
					{Name: "Email", Value: u.Email},
					{Name: "Company", Value: u.Company},
				},
				NotifCounts: nil,
			}, nil
		},
		func(ctx context.Context, newName string) error {
			u, err := getCurrentUser(ctx)
			if err != nil {
				return err
			}
			u.Name = newName
			// if err := db.Save(u).Error; err != nil {
			// 	return errors.Wrap(err, "failed to update user name")
			// }
			return nil
		},
	)
	customize(profielB)

	b.Use(profielB)

	return b
}

var currentProfileUser = &ProfileUser{
	Model:   gorm.Model{ID: 1},
	Email:   "admin@theplant.jp",
	Name:    "admin",
	Avatar:  "https://i.pravatar.cc/300",
	Role:    "Admin",
	Status:  "Active",
	Company: "The Plant",
}

func ProfileExample(b *presets.Builder, db *gorm.DB) http.Handler {
	return profileExample(b, db, currentProfileUser, func(pb *plogin.ProfileBuilder) {
		pb.DisableNotification(true).LogoutURL("auth/logout").CustomizeButtons(func(ctx context.Context, buttons ...h.HTMLComponent) ([]h.HTMLComponent, error) {
			return slices.Concat([]h.HTMLComponent{
				v.VBtn("Change Password").Variant(v.VariantTonal).Color(v.ColorSecondary).Attr("@click", "console.log('clicked change password')"),
			}, buttons), nil
		})
	})
}
