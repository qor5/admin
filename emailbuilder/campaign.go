package emailbuilder

import (
	"fmt"

	"github.com/qor5/web/v3"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/presets"
)

type (
	MailCampaign struct {
		gorm.Model
		Segmentation string
		EmailDetail
	}
)

func (c *MailCampaign) PrimarySlug() string {
	return fmt.Sprintf("%d", c.ID)
}

func (c *MailCampaign) PrimaryColumnValuesBySlug(slug string) map[string]string {
	return map[string]string{
		"id": slug,
	}
}

func DefaultMailCampaign(pb *presets.Builder) *presets.ModelBuilder {
	mb := pb.Model(&MailCampaign{})
	mb.Listing("ID", "Subject")
	dp := mb.Detailing(EmailDetailField, "Segmentations")
	section := presets.NewSectionBuilder(mb, "Segmentations").Editing("Segmentation")
	section.EditingField("Segmentation").LazyWrapComponentFunc(func(in presets.FieldComponentFunc) presets.FieldComponentFunc {
		return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return presets.SelectField(obj, field, ctx).Items([]string{"segmentationA", "segmentationB", "segmentationC", "segmentationD"})
		}
	})
	mb.Editing("Subject")
	dp.Section(section)
	return mb
}
