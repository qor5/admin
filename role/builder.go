package role

import (
	"net/http"
	"time"

	"github.com/ory/ladon"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/perm"
	. "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

type Builder struct {
	db        *gorm.DB
	actions   []*DefaultOptionItem
	resources []*DefaultOptionItem
	// editorSubject is the subject that has permission to edit roles
	// empty value means anyone can edit roles
	editorSubject    string
	roleMb           *presets.ModelBuilder
	AfterInstallFunc presets.ModelInstallFunc
}

func New(db *gorm.DB) *Builder {
	return &Builder{
		db: db,
		actions: []*DefaultOptionItem{
			{Text: "All", Value: "*"},
			{Text: "List", Value: presets.PermList},
			{Text: "Get", Value: presets.PermGet},
			{Text: "Create", Value: presets.PermCreate},
			{Text: "Update", Value: presets.PermUpdate},
			{Text: "Delete", Value: presets.PermDelete},
		},
	}
}

func (b *Builder) Actions(vs []*DefaultOptionItem) *Builder {
	b.actions = vs
	return b
}

func (b *Builder) AfterInstall(v presets.ModelInstallFunc) *Builder {
	b.AfterInstallFunc = v
	return b
}

func (b *Builder) Resources(vs []*DefaultOptionItem) *Builder {
	b.resources = vs
	return b
}

func (b *Builder) EditorSubject(v string) *Builder {
	b.editorSubject = v
	return b
}

func (b *Builder) Install(pb *presets.Builder) error {
	if b.editorSubject != "" {
		permB := pb.GetPermission()
		if permB == nil {
			panic("pb does not have a permission builder")
		}
		ctxf := permB.GetContextFunc()
		ssf := permB.GetSubjectsFunc()
		permB.ContextFunc(func(r *http.Request, objs []interface{}) perm.Context {
			c := make(perm.Context)
			if ctxf != nil {
				c = ctxf(r, objs)
			}
			ss := ssf(r)
			hasRoleEditorSubject := false
			for _, s := range ss {
				if s == b.editorSubject {
					hasRoleEditorSubject = true
					break
				}
			}
			c["has_role_editor_subject"] = hasRoleEditorSubject
			return c
		})
		permB.CreatePolicies(
			perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).ToDo(perm.Anything).On("*:roles:*").Given(perm.Conditions{
				"has_role_editor_subject": &ladon.BooleanCondition{
					BooleanValue: false,
				},
			}),
			perm.PolicyFor(b.editorSubject).WhoAre(perm.Allowed).ToDo(perm.Anything).On("*:roles:*"),
		)
	}

	b.roleMb = pb.Model(&Role{})

	ed := b.roleMb.Editing(
		"Name",
		"Permissions",
	)

	permFb := pb.NewFieldsBuilder(presets.WRITE).Model(&perm.DefaultDBPolicy{}).Only("Effect", "Actions", "Resources")
	ed.Field("Permissions").Nested(permFb)

	permFb.Field("Effect").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return VSelect().
			Variant(FieldVariantUnderlined).
			Items([]string{perm.Allowed, perm.Denied}).
			Value(field.StringValue(obj)).
			Label(field.Label).
			Attr(web.VField(field.FormKey, field.StringValue(obj))...)
	}).SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
		p := obj.(*perm.DefaultDBPolicy)
		p.Effect = ctx.R.FormValue(field.FormKey)
		return
	})
	permFb.Field("Actions").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		policy := obj.(*perm.DefaultDBPolicy)
		return VAutocomplete().
			Variant(FieldVariantUnderlined).
			Label(field.Label).
			Attr(web.VField(field.FormKey, policy.Actions)...).
			Multiple(true).
			Chips(true).
			ClosableChips(true).
			Items(b.actions).ItemTitle("text").ItemValue("value")
	})

	permFb.Field("Resources").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		policy := obj.(*perm.DefaultDBPolicy)
		return VAutocomplete().
			Variant(FieldVariantUnderlined).
			Attr(web.VField(field.FormKey, policy.Resources)...).
			Label(field.Label).
			Multiple(true).
			Chips(true).
			ClosableChips(true).
			Items(b.resources).ItemTitle("text").ItemValue("value")
	}).SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
		p := obj.(*perm.DefaultDBPolicy)
		p.Resources = ctx.R.Form[field.FormKey]
		return
	})

	ed.FetchFunc(func(obj interface{}, id string, ctx *web.EventContext) (r interface{}, err error) {
		return gorm2op.DataOperator(b.db.Preload("Permissions")).Fetch(obj, id, ctx)
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
			if err = b.db.Delete(&perm.DefaultDBPolicy{}, "refer_id = ?", r.ID).Error; err != nil {
				return
			}
		}
		if err = gorm2op.DataOperator(b.db.Session(&gorm.Session{FullSaveAssociations: true})).Save(obj, id, ctx); err != nil {
			return
		}
		startFrom := time.Now().Add(-1 * time.Second)
		pb.GetPermission().LoadDBPoliciesToMemory(b.db, &startFrom)
		return
	})

	ed.DeleteFunc(func(obj interface{}, id string, ctx *web.EventContext) (err error) {
		return b.db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Delete(&perm.DefaultDBPolicy{}, "refer_id = ?", id).Error; err != nil {
				return err
			}
			if err := tx.Delete(&Role{}, "id = ?", id).Error; err != nil {
				return err
			}

			return nil
		})
	})

	if b.AfterInstallFunc != nil {
		return b.AfterInstallFunc(pb, b.roleMb)
	}

	return nil
}
