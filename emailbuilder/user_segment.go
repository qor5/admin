package emailbuilder

import (
	"fmt"
	"path"
	"time"

	"github.com/qor5/web/v3"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/presets"
	v "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
)

type (
	UserSegment struct {
		gorm.Model
		Name       string
		TotalUsers int
		Change     float64
	}
)

const (
	upIcon = `<svg width="24" height="17" viewBox="0 0 16 17" fill="none" xmlns="http://www.w3.org/2000/svg">
<path d="M14.9722 7.86587L14.5275 4.14502L11.0828 5.62032L12.4833 6.42888C12.4501 6.46044 12.4196 6.49593 12.3926 6.53521L9.25769 11.098L6.42316 8.00556C6.00623 7.55069 5.28089 7.57824 4.89967 8.06342L1.4755 12.4214C1.24802 12.7109 1.2983 13.13 1.58782 13.3575C1.87733 13.5849 2.29643 13.5346 2.52391 13.2451L5.70552 9.1959L8.55794 12.3079C8.9925 12.782 9.75514 12.7285 10.1193 12.1985L13.4916 7.29026C13.5393 7.2208 13.5721 7.1455 13.5906 7.06817L14.9722 7.86587Z" fill="#30A46C"/>
</svg>`
	downIcon = `<svg width="24" height="25" viewBox="0 0 24 25" fill="none" xmlns="http://www.w3.org/2000/svg">
<path d="M18.968 14.0539L18.5235 17.7731L15.0803 16.2985L16.4803 15.4902C16.4471 15.4587 16.4167 15.4232 16.3897 15.384L13.2564 10.8235L10.4235 13.9142C10.0066 14.3691 9.28122 14.3415 8.89999 13.8563L5.47745 9.50046C5.24997 9.21094 5.30026 8.79184 5.58977 8.56436C5.87928 8.33688 6.29838 8.38717 6.52586 8.67668L9.70585 12.7238L12.5566 9.61364C12.9912 9.13953 13.7538 9.19298 14.118 9.72305L17.4887 14.6289C17.5364 14.6983 17.5691 14.7736 17.5876 14.8509L18.968 14.0539Z" fill="#E5484D"/>
</svg>`
)

func initRecords(db *gorm.DB) {
	var (
		count int64
		ms    = []UserSegment{
			{
				Name:       "Purchases Up",
				TotalUsers: 200,
				Change:     0.51,
			},
			{
				Name:       "Purchases down",
				TotalUsers: 120,
				Change:     -0.21,
			},
		}
	)
	db.Model(&UserSegment{}).Count(&count)
	if count == 0 {
		db.Save(&ms)
	}
}

func ConfigUserSegment(pb *presets.Builder, db *gorm.DB) *presets.ModelBuilder {
	var (
		err error
	)
	if err = db.AutoMigrate(&UserSegment{}); err != nil {
		panic(err)
	}
	//TODO:demo data
	initRecords(db)

	mb := pb.Model(&UserSegment{})
	mb.Editing().Creating("Name")
	listing := mb.Listing("ID", "Name", "TotalUsers", "Change", "CreatedAt", "UpdatedAt").WrapCell(func(in presets.CellProcessor) presets.CellProcessor {
		return func(evCtx *web.EventContext, cell h.MutableAttrHTMLComponent, id string, obj any) (h.MutableAttrHTMLComponent, error) {
			cell.SetAttr("@click", web.Plaid().PushState(true).URL(path.Join(pb.GetURIPrefix(), "chart")).Go())
			return cell, nil
		}
	})
	listing.Field("Change").Label("% Change").LazyWrapComponentFunc(func(in presets.FieldComponentFunc) presets.FieldComponentFunc {
		return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			var (
				val   = field.Value(obj).(float64)
				color = v.ColorSuccess
				icon  = upIcon
			)
			val *= 100
			if val <= 0 {
				val *= -1
				color = v.ColorError
				icon = downIcon
			}
			return h.Td(h.RawHTML(icon), h.Span(fmt.Sprintf("%v%%", val))).Class("text-"+color, "align-center", "d-flex")
		}
	})

	now := time.Now()
	now = now.AddDate(0, 0, -now.Day()+1)
	var (
		activeUsersDateArr []string
		activeUsersDataArr []interface{}
	)
	baseValue := 200
	// 定义稳定的浮动模式
	fluctuations := []int{20, 35, -15, -30, 40, 10, -25}
	for i := 0; i < 7; i++ {
		activeUsersDateArr = append(activeUsersDateArr, now.Format("01-02"))
		value := baseValue + fluctuations[i]
		activeUsersDataArr = append(activeUsersDataArr, value)
		now = now.AddDate(0, 0, 1) // Move to next day
	}

	cp := presets.NewCustomPage(pb).Body(func(ctx *web.EventContext) h.HTMLComponent {
		return web.Scope(
			vx.VXTabs().Attr("v-model", "xLocals.tab").Children(
				v.VTab().Value(0).Text("Demographics Info"),
				v.VTab().Value(1).Text("User Active"),
			),
			v.VTabsWindow().Attr("v-model", "xLocals.tab").Children(
				v.VTabsWindowItem().Value(0).Children(
					v.VRow(
						v.VCol(
							h.Div(
								vx.VXChart().Presets("barChart").Options(vx.VXChartOption{
									Title: &vx.VXChartOptionTitle{Text: "Age"},
									XAxis: &vx.VXChartOptionXAxis{
										Data: []string{"0-18", "18-25", "25-65", "65+"},
									},
									Series: &[]vx.VXChartOptionSeries{
										{
											Name: "Age",
											Data: []interface{}{100, 300, 500, 200},
										},
									},
								}),
							).Class("chart-container border border-gray-500 rounded-lg"),
						).Cols(6),
						v.VCol(
							h.Div(
								vx.VXChart().Presets("pieChart").Options(vx.VXChartOption{
									Title: &vx.VXChartOptionTitle{Text: "Gender"},
									Series: &[]vx.VXChartOptionSeries{
										{
											Name: "Gender",
											Data: []interface{}{
												map[string]interface{}{
													"name":  "Male",
													"value": 45,
												},
												map[string]interface{}{
													"name":  "Female",
													"value": 55,
												},
											},
										},
									},
								}),
							).Class("chart-container border border-gray-500 rounded-lg"),
						).Cols(6),
					).Class("mt-4"),
					v.VRow(
						v.VCol(
							h.Div(
								vx.VXChart().Presets("barChart").Options(vx.VXChartOption{
									Title: &vx.VXChartOptionTitle{Text: "Daily Active Users"},
									XAxis: &vx.VXChartOptionXAxis{
										Data: activeUsersDateArr,
									},
									Series: &[]vx.VXChartOptionSeries{
										{
											Name: "7 days",
											Data: activeUsersDataArr,
										},
									},
								}),
							).Class("chart-container border border-gray-500 rounded-lg")),
					).Class("mt-4"),
				),
			),
			v.VTabsWindowItem().Value(1).Children(
				v.VCard().Elevation(0).Children(
					v.VCardText(
						h.Div().Class("border border-dashed text-primary font-weight-bold border-primary text-center border-opacity-100 pa-4").Text("Detailed user information will go here"),
					),
				),
			),
		).VSlot("{locals:xLocals}").Init(`{tab:0}`)
	})
	pb.HandleCustomPage("chart", cp)
	return mb
}
