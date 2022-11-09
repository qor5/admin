package role

import (
	"time"

	. "github.com/qor5/ui/vuetify"
	"github.com/qor5/web"
	"github.com/qor5/x/perm"
	"github.com/qor5/admin/presets"
	"github.com/qor5/admin/presets/gorm2op"
	"github.com/qor5/admin/listeditor"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

var DefaultActions = []DefaultOptionItem{
	{Text: "All", Value: "*"},
	{Text: "List", Value: presets.PermList},
	{Text: "Get", Value: presets.PermGet},
	{Text: "Create", Value: presets.PermCreate},
	{Text: "Update", Value: presets.PermUpdate},
	{Text: "Delete", Value: presets.PermDelete},
}

func Configure(b *presets.Builder, db *gorm.DB, actions []DefaultOptionItem, resources []DefaultOptionItem) {
	role := b.Model(&Role{})
	listeditor.Configure(role)

	ed := role.Editing(
		"Name",
		"Permissions",
	)

	permFb := b.NewFieldsBuilder(presets.WRITE).Model(&perm.DefaultDBPolicy{}).Only("Effect", "Actions", "Resources")
	ed.ListField("Permissions", permFb).ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return listeditor.New(field).Value(field.Value(obj))
	})

	permFb.Field("Effect").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return VSelect().
			Items([]string{perm.Allowed, perm.Denied}).
			Value(field.StringValue(obj)).
			Label(field.Label).
			FieldName(field.FormKey)
	}).SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
		p := obj.(*perm.DefaultDBPolicy)
		p.Effect = ctx.R.FormValue(field.FormKey)
		return
	})
	permFb.Field("Actions").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return VAutocomplete().
			Value(field.Value(obj)).
			Label(field.Label).
			FieldName(field.FormKey).
			Multiple(true).Chips(true).DeletableChips(true).
			Items(actions)
	})

	permFb.Field("Resources").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return VAutocomplete().
			Value(field.Value(obj)).
			Label(field.Label).
			FieldName(field.FormKey).
			Multiple(true).Chips(true).DeletableChips(true).
			Items(resources)
	}).SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
		p := obj.(*perm.DefaultDBPolicy)
		p.Resources = ctx.R.Form[field.FormKey]
		return
	})

	ed.FetchFunc(func(obj interface{}, id string, ctx *web.EventContext) (r interface{}, err error) {
		return gorm2op.DataOperator(db.Preload("Permissions")).Fetch(obj, id, ctx)
	})

	ed.ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		u := obj.(*Role)
		if u.Name == "" {
			err.FieldError("Name", "Name is required")
			return
		}
		for _, p := range u.Permissions {
			p.Subject = u.Name
		}
		return
	})

	ed.SaveFunc(func(obj interface{}, id string, ctx *web.EventContext) (err error) {
		r := obj.(*Role)
		if r.ID != 0 {
			if err = db.Delete(&perm.DefaultDBPolicy{}, "refer_id = ?", r.ID).Error; err != nil {
				return
			}
		}
		if err = gorm2op.DataOperator(db.Session(&gorm.Session{FullSaveAssociations: true})).Save(obj, id, ctx); err != nil {
			return
		}
		startFrom := time.Now().Add(-1 * time.Second)
		b.GetPermission().LoadDBPoliciesToMemory(db, &startFrom)
		return
	})

	ed.DeleteFunc(func(obj interface{}, id string, ctx *web.EventContext) (err error) {
		r := obj.(*Role)
		err = db.Select("Permissions").Delete(&r).Error
		return
	})

}
