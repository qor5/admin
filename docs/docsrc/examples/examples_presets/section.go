package examples_presets

import (
	"fmt"
	"strings"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

// SectionDemo is a model for testing section scenarios
type SectionDemo struct {
	gorm.Model
	Name        string
	Email       string
	Age         int
	Description string
	Items       []*SectionDemoItem `gorm:"foreignKey:SectionDemoID"`
}

// SectionDemoItem is a list item for testing IsList section
type SectionDemoItem struct {
	gorm.Model
	SectionDemoID uint
	Title         string
	Content       string
}

// Scenario 1: Singleton with Editing using Section
func PresetsSectionSingleton(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	err := db.AutoMigrate(&SectionDemo{})
	if err != nil {
		panic(err)
	}
	b.DataOperator(gorm2op.DataOperator(db))

	mb = b.Model(&SectionDemo{}).Singleton(true)

	// Create section for basic info
	basicSection := presets.NewSectionBuilder(mb, "BasicInfo").
		Editing("Name", "Email").
		WrapValidator(func(in presets.ValidateFunc) presets.ValidateFunc {
			return func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
				m := obj.(*SectionDemo)
				if m.Name == "" {
					err.FieldError("BasicInfo.Name", "Name is required")
				}
				if m.Email == "" {
					err.FieldError("BasicInfo.Email", "Email is required")
				}
				return
			}
		})

	// Create section for additional info
	additionalSection := presets.NewSectionBuilder(mb, "AdditionalInfo").
		Editing("Age", "Description").
		WrapValidator(func(in presets.ValidateFunc) presets.ValidateFunc {
			return func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
				m := obj.(*SectionDemo)
				if m.Age < 0 {
					err.FieldError("AdditionalInfo.Age", "Age must be non-negative")
				}
				if m.Description == "" {
					err.FieldError("AdditionalInfo.Description", "Description is required")
				}
				return
			}
		})

	ce = mb.Editing("BasicInfo", "AdditionalInfo").
		Section(basicSection, additionalSection)

	// Editing level validator for cross-field validation
	ce.ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		m := obj.(*SectionDemo)
		// Cross-field validation: if age > 100, description must mention "senior"
		if m.Age > 100 && m.Description != "" {
			if !strings.Contains(strings.ToLower(m.Description), "senior") {
				err.GlobalError("For age > 100, description should mention 'senior'")
			}
		}
		return
	})

	return
}

// Scenario 2: Normal Editing + Detailing with Section
func PresetsSectionDetailingNormal(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	err := db.AutoMigrate(&SectionDemo{})
	if err != nil {
		panic(err)
	}
	b.DataOperator(gorm2op.DataOperator(db))

	mb = b.Model(&SectionDemo{})

	// Create section for detailing page
	detailSection := presets.NewSectionBuilder(mb, "Details").
		Editing("Name", "Email", "Age", "Description").
		WrapValidator(func(in presets.ValidateFunc) presets.ValidateFunc {
			return func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
				m := obj.(*SectionDemo)
				if m.Name == "" {
					err.FieldError("Details.Name", "Name is required")
				}
				if m.Email == "" {
					err.FieldError("Details.Email", "Email is required")
				}
				if m.Age < 0 {
					err.FieldError("Details.Age", "Age must be non-negative")
				}
				return
			}
		})

	dp = mb.Detailing("Details").Drawer(true)
	dp.Section(detailSection)

	// Normal editing without section
	ce = mb.Editing("Name", "Email", "Age", "Description")
	ce.ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		m := obj.(*SectionDemo)
		if m.Name == "" {
			err.FieldError("Name", "Name is required")
		}
		if m.Email == "" {
			err.FieldError("Email", "Email is required")
		}
		if m.Age < 0 {
			err.FieldError("Age", "Age must be non-negative")
		}
		// Cross-field validation
		if m.Name != "" && m.Email != "" && !strings.Contains(strings.ToLower(m.Email), strings.ToLower(m.Name)) {
			err.GlobalError("Email should contain the name")
		}
		return
	})

	return
}

// Scenario 3: Editing with Clone Section
func PresetsSectionEditingClone(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	err := db.AutoMigrate(&SectionDemo{})
	if err != nil {
		panic(err)
	}
	b.DataOperator(gorm2op.DataOperator(db))

	mb = b.Model(&SectionDemo{})

	// Create section that will be used in both detailing and editing
	sharedSection := presets.NewSectionBuilder(mb, "SharedInfo").
		Editing("Name", "Email").
		WrapValidator(func(in presets.ValidateFunc) presets.ValidateFunc {
			return func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
				m := obj.(*SectionDemo)
				if m.Name == "" {
					err.FieldError("SharedInfo.Name", "Name is required")
				}
				if len(m.Name) > 50 {
					err.FieldError("SharedInfo.Name", "Name must be less than 50 characters")
				}
				if m.Email == "" {
					err.FieldError("SharedInfo.Email", "Email is required")
				}
				return
			}
		})

	additionalSection := presets.NewSectionBuilder(mb, "AdditionalInfo").
		Editing("Age", "Description").
		WrapValidator(func(in presets.ValidateFunc) presets.ValidateFunc {
			return func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
				m := obj.(*SectionDemo)
				if m.Age < 0 || m.Age > 150 {
					err.FieldError("AdditionalInfo.Age", "Age must be between 0 and 150")
				}
				return
			}
		})

	// Detailing uses original section
	dp = mb.Detailing("SharedInfo", "AdditionalInfo").Drawer(true)
	dp.Section(sharedSection, additionalSection)

	// Editing uses cloned section
	ce = mb.Editing("SharedInfo", "AdditionalInfo").
		Section(sharedSection.Clone(), additionalSection.Clone())

	// Editing level validator for cross-field validation
	ce.ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		m := obj.(*SectionDemo)
		// Cross-field validation
		if m.Age > 0 && m.Description == "" {
			err.GlobalError("Description is required when age is specified")
		}
		return
	})

	return
}

// Scenario 4: Detailing with IsList Section
func PresetsSectionIsList(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	err := db.AutoMigrate(&SectionDemo{}, &SectionDemoItem{})
	if err != nil {
		panic(err)
	}
	b.DataOperator(gorm2op.DataOperator(db))

	mb = b.Model(&SectionDemo{})

	// Wrap fetch to load items
	mb.Detailing().WrapFetchFunc(func(in presets.FetchFunc) presets.FetchFunc {
		return func(obj interface{}, id string, ctx *web.EventContext) (r interface{}, err error) {
			r, err = in(obj, id, ctx)
			if err != nil {
				return
			}
			m := r.(*SectionDemo)
			if err = db.Where("section_demo_id = ?", m.ID).Find(&m.Items).Error; err != nil {
				return
			}
			return m, nil
		}
	})

	// Basic info section
	basicSection := presets.NewSectionBuilder(mb, "BasicInfo").
		Editing("Name", "Email").
		WrapValidator(func(in presets.ValidateFunc) presets.ValidateFunc {
			return func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
				m := obj.(*SectionDemo)
				if m.Name == "" {
					err.FieldError("BasicInfo.Name", "Name is required")
				}
				return
			}
		})

	// List section for items
	itemsSection := presets.NewSectionBuilder(mb, "Items").
		IsList(&SectionDemoItem{}).
		Editing("Title", "Content").
		ElementEditComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			item := obj.(*SectionDemoItem)
			return h.Div(
				vx.VXField().Label("Title").
					Attr(web.VField(fmt.Sprintf("%s.Title", field.FormKey), item.Title)...),
				vx.VXField().Label("Content").
					Attr(web.VField(fmt.Sprintf("%s.Content", field.FormKey), item.Content)...),
			).Class("d-flex flex-column ga-2")
		}).
		ElementShowComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			item := obj.(*SectionDemoItem)
			return h.Div(
				h.Strong("Title: "), h.Text(item.Title),
				h.Br(),
				h.Strong("Content: "), h.Text(item.Content),
			)
		}).
		WrapValidator(func(in presets.ValidateFunc) presets.ValidateFunc {
			return func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
				m := obj.(*SectionDemo)
				for i, item := range m.Items {
					if item.Title == "" {
						err.FieldError("Items", fmt.Sprintf("Item %d: Title is required", i+1))
					}
					if item.Content == "" {
						err.FieldError("Items", fmt.Sprintf("Item %d: Content is required", i+1))
					}
				}
				return
			}
		})

	dp = mb.Detailing("BasicInfo", "Items").Drawer(true)
	dp.Section(basicSection, itemsSection)

	// Normal editing (IsList section cannot be used in editing)
	ce = mb.Editing("Name", "Email")
	ce.ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		m := obj.(*SectionDemo)
		if m.Name == "" {
			err.FieldError("Name", "Name is required")
		}
		if m.Email == "" {
			err.FieldError("Email", "Email is required")
		}
		return
	})

	return
}
