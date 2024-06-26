package examples_presets

import (
	"fmt"
	"net/url"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

func PresetsHelloWorldX(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilderX,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	err := db.AutoMigrate(
		&Customer{},
		&Company{},
		&Address{},
	)
	if err != nil {
		panic(err)
	}

	b.DataOperator(gorm2op.DataOperator(db))
	mb = b.Model(&Customer{})
	ce = mb.Editing()

	{
		cl := mb.Listing()
		cl.OrderableFields([]*presets.OrderableField{
			{
				FieldName: "ID",
				DBColumn:  "id",
			},
			{
				FieldName: "Name",
				DBColumn:  "name",
			},
		})
		cl.SelectableColumns(true)
		cl.BulkAction("Delete").Label("Delete").
			UpdateFunc(func(selectedIds []string, ctx *web.EventContext) (err error) {
				err = db.Where("id IN (?)", selectedIds).Delete(&Customer{}).Error
				return
			}).
			ComponentFunc(func(selectedIds []string, ctx *web.EventContext) h.HTMLComponent {
				return h.Div().Text(fmt.Sprintf("Are you sure you want to delete %s ?", selectedIds)).Class("title deep-orange--text")
			})
		cl.FilterTabsFunc(func(ctx *web.EventContext) []*presets.FilterTab {
			return []*presets.FilterTab{
				{
					Label: "All",
					Query: url.Values{"all": []string{"1"}},
				},
				{
					Label: "Felix",
					Query: url.Values{"name.ilike": []string{"felix"}},
				},
				{
					Label: "Approved",
					Query: url.Values{"approved.gt": []string{fmt.Sprint(1)}},
				},
			}
		})
		cl.FilterDataFunc(func(ctx *web.EventContext) vuetifyx.FilterData {
			msgr := i18n.MustGetModuleMessages(ctx.R, presets.ModelsI18nModuleKey, Messages_en_US).(*Messages)
			var companyOptions []*vuetifyx.SelectItem
			err := db.Model(&Company{}).Select("name as text, id as value").Scan(&companyOptions).Error
			if err != nil {
				panic(err)
			}

			return []*vuetifyx.FilterItem{
				{
					Key:      "created",
					Label:    msgr.CustomersFilterCreated,
					ItemType: vuetifyx.ItemTypeDatetimeRange,
					// SQLCondition: `cast(strftime('%%s', created_at) as INTEGER) %s ?`,
					SQLCondition: `created_at %s ?`,
				},
				{
					Key:      "approved",
					Label:    msgr.CustomersFilterApproved,
					ItemType: vuetifyx.ItemTypeDatetimeRange,
					// SQLCondition: `cast(strftime('%%s', created_at) as INTEGER) %s ?`,
					SQLCondition: `created_at %s ?`,
				},
				{
					Key:          "name",
					Label:        msgr.CustomersFilterName,
					ItemType:     vuetifyx.ItemTypeString,
					SQLCondition: `name %s ?`,
				},
				{
					Key:          "company",
					Label:        msgr.CustomersFilterCompany,
					ItemType:     vuetifyx.ItemTypeSelect,
					SQLCondition: `company_id %s ?`,
					Options:      companyOptions,
				},
			}
		})
	}
	{
		cl := mb.ListingX()
		cl.OrderableFields([]*presets.OrderableField{
			{
				FieldName: "ID",
				DBColumn:  "id",
			},
			{
				FieldName: "Name",
				DBColumn:  "name",
			},
		})
		cl.SelectableColumns(true)
		cl.BulkAction("Delete").Label("Delete").
			UpdateFunc(func(selectedIds []string, ctx *web.EventContext) (err error) {
				err = db.Where("id IN (?)", selectedIds).Delete(&Customer{}).Error
				return
			}).
			ComponentFunc(func(selectedIds []string, ctx *web.EventContext) h.HTMLComponent {
				return h.Div().Text(fmt.Sprintf("Are you sure you want to delete %s ?", selectedIds)).Class("title deep-orange--text")
			})
		cl.FilterTabsFunc(func(ctx *web.EventContext) []*presets.FilterTab {
			return []*presets.FilterTab{
				{
					Label: "All",
					Query: url.Values{"all": []string{"1"}},
				},
				{
					Label: "Felix",
					Query: url.Values{"name.ilike": []string{"felix"}},
				},
				{
					Label: "Approved",
					Query: url.Values{"approved.gt": []string{fmt.Sprint(1)}},
				},
			}
		})
		cl.FilterDataFunc(func(ctx *web.EventContext) vuetifyx.FilterData {
			msgr := i18n.MustGetModuleMessages(ctx.R, presets.ModelsI18nModuleKey, Messages_en_US).(*Messages)
			var companyOptions []*vuetifyx.SelectItem
			err := db.Model(&Company{}).Select("name as text, id as value").Scan(&companyOptions).Error
			if err != nil {
				panic(err)
			}

			return []*vuetifyx.FilterItem{
				{
					Key:      "created",
					Label:    msgr.CustomersFilterCreated,
					ItemType: vuetifyx.ItemTypeDatetimeRange,
					// SQLCondition: `cast(strftime('%%s', created_at) as INTEGER) %s ?`,
					SQLCondition: `created_at %s ?`,
				},
				{
					Key:      "approved",
					Label:    msgr.CustomersFilterApproved,
					ItemType: vuetifyx.ItemTypeDatetimeRange,
					// SQLCondition: `cast(strftime('%%s', created_at) as INTEGER) %s ?`,
					SQLCondition: `created_at %s ?`,
				},
				{
					Key:          "name",
					Label:        msgr.CustomersFilterName,
					ItemType:     vuetifyx.ItemTypeString,
					SQLCondition: `name %s ?`,
				},
				{
					Key:          "company",
					Label:        msgr.CustomersFilterCompany,
					ItemType:     vuetifyx.ItemTypeSelect,
					SQLCondition: `company_id %s ?`,
					Options:      companyOptions,
				},
			}
		})
	}
	return
}
