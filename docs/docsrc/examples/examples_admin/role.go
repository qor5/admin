package examples_admin

import (
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/role"
	"github.com/qor5/x/v3/perm"
	"github.com/qor5/x/v3/ui/vuetify"
	"gorm.io/gorm"
)

func rolePieces() {
	var db *gorm.DB
	// @snippet_begin(RolePermEnableDBPolicy)
	perm.New().
		Policies(
		// static policies
		).
		DBPolicy(perm.NewDBPolicy(db))
	// @snippet_end

	// @snippet_begin(RoleSetResources)
	rb := role.New(db).
		Resources([]*vuetify.DefaultOptionItem{
			{Text: "All", Value: "*"},
			{Text: "Posts", Value: "*:posts:*"},
			{Text: "Customers", Value: "*:customers:*"},
			{Text: "Products", Value: "*:products:*"},
		})
	// @snippet_end

	// @snippet_begin(RoleSetActions)
	// default value
	rb.Actions([]*vuetify.DefaultOptionItem{
		{Text: "All", Value: "*"},
		{Text: "List", Value: presets.PermList},
		{Text: "Get", Value: presets.PermGet},
		{Text: "Create", Value: presets.PermCreate},
		{Text: "Update", Value: presets.PermUpdate},
		{Text: "Delete", Value: presets.PermDelete},
	})
	// @snippet_end

	// @snippet_begin(RoleSetEditorSubject)
	rb.EditorSubject("RoleEditor")
	// @snippet_end

	var presetsBuilder *presets.Builder
	// @snippet_begin(RoleAttachToPresetsBuilder)
	rb.Install(presetsBuilder)
	// @snippet_end
}
