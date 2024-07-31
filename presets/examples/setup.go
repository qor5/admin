package examples

import (
	"fmt"
	"net/url"
	"reflect"
	"time"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3"
	. "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

type Thumb struct {
	Name string
}

type Customer struct {
	ID              int
	Name            string
	Email           string
	Description     string
	Thumb1          *Thumb `gorm:"-"`
	CompanyID       int
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ApprovedAt      *time.Time
	TermAgreedAt    *time.Time
	ApprovalComment string
	LanguageCode    string
	Events          []*Event `gorm:"-"`
}

func (c *Customer) PageTitle() string {
	return c.Name
}

type Note struct {
	ID         int
	SourceType string
	SourceID   int
	Content    string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type CreditCard struct {
	ID              int
	CustomerID      int
	Number          string
	ExpireYearMonth string
	Name            string
	Type            string
	Phone           string
	Email           string
}

type Payment struct {
	ID                   int
	CustomerID           int
	CurrencyCode         string
	Amount               int
	PaymentMethodID      int
	StatementDescription string
	Description          string
	AuthorizeOnly        bool
	CreatedAt            time.Time
}

type Event struct {
	ID          int
	SourceType  string // Payment, Customer
	SourceID    int
	CreatedAt   time.Time
	Type        string
	Description string
}

type Language struct {
	Code string `gorm:"unique;not null"`
	Name string
}

func (l *Language) PrimarySlug() string {
	return l.Code
}

func (l *Language) PrimaryColumnValuesBySlug(slug string) map[string]string {
	return map[string]string{
		"code": slug,
	}
}

type Company struct {
	ID   int
	Name string
}

type Product struct {
	ID        int
	Name      string
	OwnerName string
}

func (*Product) TableName() string {
	return "preset_products"
}

func addListener(v any) h.HTMLComponent {
	simpleReload := web.Plaid().MergeQuery(true).Go()
	return web.Listen(
		presets.NotifModelsCreated(v), simpleReload,
		presets.NotifModelsUpdated(v), simpleReload,
		presets.NotifModelsDeleted(v), simpleReload,
	)
}

func Preset1(db *gorm.DB) (r *presets.Builder) {
	err := db.AutoMigrate(
		&Customer{},
		&Note{},
		&CreditCard{},
		&Payment{},
		&Event{},
		&Company{},
		&Product{},
		&Language{},
	)
	if err != nil {
		panic(err)
	}

	p := presets.New().URIPrefix("/admin")

	p.BrandFunc(func(ctx *web.EventContext) h.HTMLComponent {
		return h.Components(
			VIcon("mdi-directions-boat").Class("pr-2"),
			VToolbarTitle("My Admin"),
		)
	})
	// .BrandTitle("My Admin")

	writeFieldDefaults := p.FieldDefaults(presets.WRITE)
	writeFieldDefaults.FieldType(&Thumb{}).ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		i, err := reflectutils.Get(obj, field.Name)
		if err != nil {
			panic(err)
		}
		return h.Text(i.(*Thumb).Name)
	})

	p.FieldDefaults(presets.LIST).FieldType(&Thumb{}).ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		i, err := reflectutils.Get(obj, field.Name)
		if err != nil {
			panic(err)
		}
		return h.Text(i.(*Thumb).Name)
	})

	p.FieldDefaults(presets.DETAIL).FieldType([]*Event{}).ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		events := reflectutils.MustGet(obj, field.Name).([]*Event)
		typeName := reflect.ValueOf(obj).Elem().Type().Name()
		objId := fmt.Sprint(reflectutils.MustGet(obj, "ID"))

		dt := vx.DataTable(events).WithoutHeader(true).LoadMoreAt(20, "Show More").LoadMoreURL(fmt.Sprintf("/admin/events?sourceType=%s&sourceId=%s", typeName, objId))

		dt.Column("Type")
		dt.Column("Description")

		dt.RowMenuItemFuncs(presets.EditDeleteRowMenuItemFuncs(field.ModelInfo, "/admin/events",
			url.Values{"model": []string{typeName}, "model_id": []string{objId}})...)

		return vx.Card(
			dt,
		).HeaderTitle(field.Label).
			Actions(
				addListener(&Event{}),
				VBtn("Add Event").
					Variant(VariantFlat).Attr("@click",
					web.Plaid().EventFunc(actions.New).
						Query("model", typeName).
						Query("model_id", objId).
						URL("/admin/events").
						Go(),
				),
			).Class("mb-4")
	})

	p.DataOperator(gorm2op.DataOperator(db))

	p.MenuGroup("Customer Management").Icon("group").SubItems("my_customers", "company")
	mp := p.Model(&Product{}).MenuIcon("laptop")
	mp.Listing().PerPage(3)

	m := p.Model(&Customer{}).URIName("my_customers")
	p.Model(&Company{})
	m.Labels(
		"Name", "名字",
		"Bool1", "性别",
		"Float1", "体重",
		"CompanyID", "公司",
	).Placeholders(
		"Name", "请输入你的名字",
	)

	l := m.Listing("Name", "CompanyID", "ApprovalComment").SearchColumns("name", "email", "description").PerPage(5).SelectableColumns(true)
	l.Field("Name").Label("列表的名字")
	l.Field("CompanyID").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		u := obj.(*Customer)
		var comp Company
		err := db.Find(&comp, u.CompanyID).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			panic(err)
		}
		return h.Td(
			h.A().Text(comp.Name).
				Attr("@click.stop",
					web.Plaid().URL("/admin/companies").
						EventFunc(actions.Edit).
						Query(presets.ParamID, fmt.Sprint(comp.ID)).
						Go()),
		)
	})

	l.BulkAction("Approve").Label("Approve").UpdateFunc(func(selectedIds []string, ctx *web.EventContext, r *web.EventResponse) (err error) {
		comment := ctx.R.FormValue("ApprovalComment")
		if len(comment) < 10 {
			ctx.Flash = "comment should larger than 10"
			return
		}
		err = db.Model(&Customer{}).
			Where("id IN (?)", selectedIds).
			Updates(map[string]interface{}{"approved_at": time.Now(), "approval_comment": comment}).Error
		if err != nil {
			ctx.Flash = err.Error()
		} else {
			r.Emit(
				presets.NotifModelsUpdated(&Customer{}),
				presets.PayloadModelsUpdated{Ids: selectedIds},
			)
		}
		return
	}).ComponentFunc(func(selectedIds []string, ctx *web.EventContext) h.HTMLComponent {
		comment := ctx.R.FormValue("ApprovalComment")
		errorMessage := ""
		if ctx.Flash != nil {
			errorMessage = ctx.Flash.(string)
		}
		return VTextField().
			Attr(web.VField("ApprovalComment", comment)...).
			Label("Content").
			ErrorMessages(errorMessage)
	})

	l.BulkAction("Delete").Label("Delete").UpdateFunc(func(selectedIds []string, ctx *web.EventContext, r *web.EventResponse) (err error) {
		err = db.Where("id IN (?)", selectedIds).Delete(&Customer{}).Error
		if err == nil {
			r.Emit(
				presets.NotifModelsDeleted(&Customer{}),
				presets.PayloadModelsDeleted{Ids: selectedIds},
			)
		}
		return
	}).ComponentFunc(func(selectedIds []string, ctx *web.EventContext) h.HTMLComponent {
		return h.Div().Text(fmt.Sprintf("Are you sure you want to delete %s ?", selectedIds)).Class("title deep-orange--text")
	})

	l.FilterDataFunc(func(ctx *web.EventContext) vx.FilterData {
		var companyOptions []*vx.SelectItem
		err := db.Model(&Company{}).Select("name as text, id as value").Scan(&companyOptions).Error
		if err != nil {
			panic(err)
		}

		return []*vx.FilterItem{
			{
				Key:          "created",
				Label:        "Created",
				Folded:       true,
				ItemType:     vx.ItemTypeDatetimeRange,
				SQLCondition: `extract(epoch from created_at) %s ?`,
			},
			{
				Key:          "approved",
				Label:        "Approved",
				ItemType:     vx.ItemTypeDatetimeRange,
				SQLCondition: `extract(epoch from approved_at) %s ?`,
			},
			{
				Key:          "name",
				Label:        "Name",
				Folded:       true,
				ItemType:     vx.ItemTypeString,
				SQLCondition: `name %s ?`,
			},
			{
				Key:          "company",
				Label:        "Company",
				ItemType:     vx.ItemTypeSelect,
				SQLCondition: `company_id %s ?`,
				Options:      companyOptions,
			},
		}
	})

	l.FilterTabsFunc(func(ctx *web.EventContext) []*presets.FilterTab {
		var c Company
		db.First(&c)
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
				Label: "The Plant",
				Query: url.Values{"company": []string{fmt.Sprint(c.ID)}},
			},
			{
				Label: "Approved",
				Query: url.Values{"approved.gt": []string{fmt.Sprint(1)}},
			},
		}
	})

	ef := m.Editing("Name", "CompanyID", "LanguageCode").
		ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
			cu := obj.(*Customer)
			if len(cu.Name) < 5 {
				err.FieldError("Name", "input more than 5 chars")
				err.GlobalError("there are some errors")
			}
			return
		})
	ef.Field("LanguageCode").Label("语言").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		u := obj.(*Customer)
		var langs []Language
		err := db.Find(&langs).Error
		if err != nil {
			panic(err)
		}
		return VAutocomplete().
			Attr(web.VField(field.Name, u.LanguageCode)...).
			Label(field.Label).
			Items(langs).
			ItemTitle("Name").
			ItemValue("Code").
			Multiple(false)
	})

	ef.Field("CompanyID").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		u := obj.(*Customer)
		var companies []*Company
		err := db.Find(&companies).Error
		if err != nil {
			panic(err)
		}
		return VSelect().
			Attr(web.VField("CompanyID", u.CompanyID)...).
			Label(field.Label).
			Items(companies).
			ItemTitle("Name").
			ItemValue("ID").
			Multiple(false)
	})

	dp := m.Detailing("MainInfo", "Details", "Cards", "Events")

	dp.FetchFunc(func(obj interface{}, id string, ctx *web.EventContext) (r interface{}, err error) {
		cus := &Customer{}
		err = db.Find(cus, id).Error
		if err != nil {
			return
		}

		var events []*Event
		err = db.Where("source_type = ? AND source_id = ?", "Customer", id).Find(&events).Error
		if err != nil {
			return
		}
		cus.Events = events
		r = cus
		return
	})

	dp.Field("MainInfo").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		cu := obj.(*Customer)

		title := cu.Name
		if len(title) == 0 {
			title = cu.Description
		}

		var notes []*Note
		err := db.Where("source_type = 'Customer' AND source_id = ?", cu.ID).
			Order("id DESC").
			Find(&notes).Error
		if err != nil {
			panic(err)
		}

		dt := vx.DataTable(notes).WithoutHeader(true).LoadMoreAt(2, "Show More")

		dt.Column("Content").CellComponentFunc(func(obj interface{}, fieldName string, ctx *web.EventContext) h.HTMLComponent {
			n := obj.(*Note)
			return h.Td(h.Div(
				h.Div(
					VIcon("comment").Color("blue").Size(SizeSmall).Class("pr-2"),
					h.Text(n.Content),
				).Class("body-1"),
				h.Div(
					h.Text(n.CreatedAt.Format("Jan 02,15:04 PM")),
					h.Text(" by Felix Sun"),
				).Class("grey--text pl-7 body-2"),
			).Class("my-3"))
		})

		cusID := fmt.Sprint(cu.ID)
		dt.RowMenuItemFuncs(presets.EditDeleteRowMenuItemFuncs(field.ModelInfo, "/admin/notes",
			url.Values{"model": []string{"Customer"}, "model_id": []string{cusID}})...)

		return vx.Card(
			dt,
		).HeaderTitle(title).
			Actions(
				addListener(&Note{}),
				VBtn("Add Note").
					Variant(VariantFlat).
					Attr("@click",
						web.Plaid().EventFunc(actions.New).
							Query("model", "Customer").
							Query("model_id", cusID).
							URL("/admin/notes").
							Go(),
					),
			).Class("mb-4")
	})

	dp.Field("Details").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		cu := obj.(*Customer)
		cusID := fmt.Sprint(cu.ID)

		var lang Language
		db.Where("code = ?", cu.LanguageCode).First(&lang)

		var termAgreed string
		if cu.TermAgreedAt != nil {
			termAgreed = cu.TermAgreedAt.Format("Jan 02,15:04 PM")
		}

		detail := vx.DetailInfo(
			vx.DetailColumn(
				vx.DetailField(vx.OptionalText(cu.Name).ZeroLabel("No Name")).Label("Name"),
				vx.DetailField(vx.OptionalText(cu.Email).ZeroLabel("No Email")).Label("Email"),
				vx.DetailField(vx.OptionalText(cu.Description).ZeroLabel("No Description")).Label("Description"),
				vx.DetailField(vx.OptionalText(cusID).ZeroLabel("No ID")).Label("ID"),
				vx.DetailField(vx.OptionalText(cu.CreatedAt.Format("Jan 02,15:04 PM")).ZeroLabel("")).Label("Created"),
				vx.DetailField(vx.OptionalText(termAgreed).ZeroLabel("Not Agreed Yet")).Label("Terms Agreed"),
				vx.DetailField(vx.OptionalText(lang.Name).ZeroLabel("No Language")).Label("Language"),
			).Header("ACCOUNT INFORMATION"),
			vx.DetailColumn().Header("BILLING INFORMATION"),
		)

		return vx.Card(detail).HeaderTitle("Details").
			Actions(
				web.Listen(
					m.NotifModelsUpdated(), web.Plaid().MergeQuery(true).Go(),
				),
				VBtn("Agree Terms").
					Variant(VariantFlat).Class("mr-2").
					Attr("@click", web.Plaid().
						EventFunc(actions.Action).
						Query(presets.ParamAction, "AgreeTerms").
						Query("customerID", cusID).
						Go()),

				VBtn("Update details").
					Variant(VariantFlat).
					Attr("@click", web.Plaid().
						EventFunc(actions.Edit).
						Query("customerID", cusID).
						URL("/admin/customers").Go()),
			).Class("mb-4")
	})

	dp.Field("Cards").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		cu := obj.(*Customer)
		cusID := fmt.Sprint(cu.ID)

		var cards []*CreditCard
		err := db.Where("customer_id = ?", cu.ID).Order("id ASC").Find(&cards).Error
		if err != nil {
			panic(err)
		}

		dt := vx.DataTable(cards).
			WithoutHeader(true).
			RowExpandFunc(func(obj interface{}, ctx *web.EventContext) h.HTMLComponent {
				card := obj.(*CreditCard)
				return vx.DetailInfo(
					vx.DetailColumn(
						vx.DetailField(vx.OptionalText(card.Name).ZeroLabel("No Name")).Label("Name"),
						vx.DetailField(vx.OptionalText(card.Number).ZeroLabel("No Number")).Label("Number"),
						vx.DetailField(vx.OptionalText(card.ExpireYearMonth).ZeroLabel("No Expires")).Label("Expires"),
						vx.DetailField(vx.OptionalText(card.Type).ZeroLabel("No Type")).Label("Type"),
						vx.DetailField(vx.OptionalText(card.Phone).ZeroLabel("No phone provided")).Label("Phone"),
						vx.DetailField(vx.OptionalText(card.Email).ZeroLabel("No email provided")).Label("Email"),
					),
				)
			}).RowMenuItemFuncs(
			presets.EditDeleteRowMenuItemFuncs(
				field.ModelInfo, "/admin/credit-cards",
				url.Values{"customerID": []string{cusID}},
			)...)

		dt.Column("Type")
		dt.Column("Number")
		dt.Column("ExpireYearMonth")

		return vx.Card(dt).HeaderTitle("Cards").
			Actions(
				addListener(&CreditCard{}),
				VBtn("Add Card").
					Variant(VariantFlat).
					Attr("@click",
						web.Plaid().EventFunc(
							actions.New).Query("customerID", cusID).
							URL("/admin/credit-cards").
							Go()),
			).Class("mb-4")
	})

	dp.Action("AgreeTerms").UpdateFunc(func(id string, ctx *web.EventContext, r *web.EventResponse) (err error) {
		if ctx.R.FormValue("Agree") != "true" {
			ve := &web.ValidationErrors{}
			ve.GlobalError("You must agree the terms")
			err = ve
			return
		}

		err = db.Model(&Customer{}).Where("id = ?", id).
			Updates(map[string]interface{}{"term_agreed_at": time.Now()}).Error
		if err == nil {
			r.Emit(
				presets.NotifModelsUpdated(&Customer{}),
				presets.PayloadModelsUpdated{Ids: []string{id}},
			)
		}
		return
	}).ComponentFunc(func(id string, ctx *web.EventContext) h.HTMLComponent {
		var alert h.HTMLComponent

		if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
			alert = VAlert(h.Text(ve.GetGlobalError())).Border("left").
				Type("error").
				Elevation(2)
		}

		return h.Components(
			alert,
			VCheckbox().
				Attr(web.VField("Agree", ctx.R.FormValue("Agree"))...).
				Label("Agree the terms"),
		)
	})

	p.Model(&Note{}).
		InMenu(false).
		Editing("Content").
		SetterFunc(func(obj interface{}, ctx *web.EventContext) {
			note := obj.(*Note)
			note.SourceID = ctx.ParamAsInt("model_id")
			note.SourceType = ctx.R.FormValue("model")
		})

	p.Model(&Event{}).
		InMenu(false).
		Editing("Type", "Description").
		SetterFunc(func(obj interface{}, ctx *web.EventContext) {
			note := obj.(*Event)
			note.SourceID = ctx.ParamAsInt("model_id")
			note.SourceType = ctx.R.FormValue("model")
		})

	cc := p.Model(&CreditCard{}).
		InMenu(false)

	ccedit := cc.Editing("ExpireYearMonth", "Phone", "Email").
		SetterFunc(func(obj interface{}, ctx *web.EventContext) {
			card := obj.(*CreditCard)
			card.CustomerID = ctx.ParamAsInt("customerID")
		})

	ccedit.Creating("Number")

	p.Model(&Language{}).PrimaryField("Code")

	return p
}
