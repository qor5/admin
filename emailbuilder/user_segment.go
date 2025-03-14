package emailbuilder

import (
	_ "embed"
	"fmt"
	"time"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	v "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
)

type (
	UserSegment struct {
		gorm.Model
		Name       string `gorm:"uniqueIndex"`
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
				TotalUsers: 3000,
				Change:     0.51,
			},
			{
				Name:       "Purchases down",
				TotalUsers: 200,
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
	var err error
	if err = db.AutoMigrate(&UserSegment{}); err != nil {
		panic(err)
	}
	// TODO:demo data
	initRecords(db)

	mb := pb.Model(&UserSegment{}).Label("User Segments").RightDrawerWidth("840")

	cb := mb.Editing().Creating("Name", "Conditions")
	cb.Field("Name").Label("Title")
	cb.Field("Conditions").Label("Conditions").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return h.Div().Class("d-flex flex-column ga-2").Children(
			h.Text("Conditions"),
			vx.VXSegmentForm("Conditions").Options(dummyTags(ctx.R.Context())),
		)
	})

	listing := mb.Listing("ID", "Name", "TotalUsers", "Change", "CreatedAt", "UpdatedAt")
	listing.NewButtonFunc(func(ctx *web.EventContext) h.HTMLComponent {
		msgr := i18n.MustGetModuleMessages(ctx.R, presets.CoreI18nModuleKey, presets.Messages_en_US).(*presets.Messages)
		return h.Components(
			v.VBtn(msgr.New).
				Color(v.ColorPrimary).
				Variant(v.VariantElevated).
				Theme("light").Class("ml-2").
				Attr("@click", web.Plaid().URL(mb.Info().ListingHref()).EventFunc(actions.New).Query(presets.ParamOverlay, actions.Dialog).Go()),
		)
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
	yesterday := now.AddDate(0, 0, -1) // 昨天
	var (
		activeUsersDateArr []string
		activeUsersDataArr []interface{}

		// 添加14天数据的数组
		activeUsers14DaysDateArr []string
		activeUsers14DaysDataArr []interface{}
	)
	baseValue := 200
	// 定义稳定的浮动模式
	fluctuations := []int{20, 35, -15, -30, 40, 10, -25}

	// 从昨天往前推7天，并且按照从最早到最近的顺序排列
	for i := 6; i >= 0; i-- {
		datePoint := yesterday.AddDate(0, 0, -i) // 从昨天往前推i天
		activeUsersDateArr = append(activeUsersDateArr, datePoint.Format("01-02"))
		value := baseValue + fluctuations[6-i] // 反向使用波动数组以保持相同的变化模式
		activeUsersDataArr = append(activeUsersDataArr, value)
	}

	// 从昨天往前推14天，并且按照从最早到最近的顺序排列
	baseValue14Days := 180 // 稍微低一点的基准值，以区分两条线
	// 更长的波动数组，14天
	fluctuations14Days := []int{15, 25, -10, -20, 30, 10, -15, 5, 40, -25, -5, 20, 30, -10}
	for i := 13; i >= 0; i-- {
		datePoint := yesterday.AddDate(0, 0, -i) // 从昨天往前推i天
		activeUsers14DaysDateArr = append(activeUsers14DaysDateArr, datePoint.Format("01-02"))
		value := baseValue14Days + fluctuations14Days[13-i] // 反向使用波动数组以保持相同的变化模式
		activeUsers14DaysDataArr = append(activeUsers14DaysDataArr, value)
	}
	detailing := mb.Detailing("Charts")
	detailing.Title(func(evCtx *web.EventContext, obj any, style presets.DetailingStyle, defaultTitle string) (title string, titleCompo h.HTMLComponent, err error) {
		title = obj.(*UserSegment).Name
		return
	})
	se := presets.NewSectionBuilder(mb, "Charts").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return web.Scope(
			h.Div(
				vx.VXTabs(
					v.VTab(h.Text("Demographics Info")).Value(0),
					v.VTab(h.Text("User Activity")).Value(1),
				).Attr("v-model", "xLocals.tab"),

				v.VTabsWindow(
					v.VTabsWindowItem(
						h.Div(
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
									).Class("border border-gray-500 rounded-lg"),
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
									).Class("border border-gray-500 rounded-lg"),
								).Cols(6),
							).Class("mt-4"),
							v.VRow(
								v.VCol(
									h.Div(
										vx.VXChart().Children(
											web.Slot(
												h.Div(
													h.Button("Past 7 Days").
														Class("text-body-2 rounded text-no-wrap border-0 flex-grow-1 d-flex align-center justify-center rounded px-2").
														Attr(":style", `
														 currentIndex === 0
														? 'background-color: #fff; color: #4a4a4a;'
														: 'background-color: transparent; color: rgb(117, 117, 117);'`).Attr("@click", "toggle(0)"),
													h.Button("Past 14 Days").Attr(":style", `
														 currentIndex === 1
														? 'background-color: #fff; color: #4a4a4a;'
														: 'background-color: transparent; color: rgb(117, 117, 117);'`).
														Class("text-body-2 rounded text-no-wrap border-0 flex-grow-1 d-flex align-center justify-center rounded px-2").
														Attr("@click", "toggle(1)"),
												).Class("d-flex align-center bg-grey-lighten-3 rounded pa-1 mr-4 mt-4").Style("height: 32px;"),
											).Name("action").Scope("{ list, currentIndex, toggle }"),
										).Presets("barChart").Options([]vx.VXChartOption{
											{
												Title: &vx.VXChartOptionTitle{Text: "Daily Active Users (7 Days)"},
												XAxis: &vx.VXChartOptionXAxis{
													Data: activeUsersDateArr,
												},
												Series: &[]vx.VXChartOptionSeries{
													{
														Name: "7 days",
														Data: activeUsersDataArr,
													},
												},
											},
											{
												Title: &vx.VXChartOptionTitle{Text: "Daily Active Users (14 Days)"},
												XAxis: &vx.VXChartOptionXAxis{
													Data: activeUsers14DaysDateArr,
												},
												Series: &[]vx.VXChartOptionSeries{
													{
														Name: "14 days",
														Data: activeUsers14DaysDataArr,
													},
												},
											},
										}),
									).Class("border border-gray-500 rounded-lg"),
								).Cols(12),
							).Class("mt-4"),
						),
					).Value(0),
					v.VTabsWindowItem(

						v.VRow(
							v.VCol(
								h.Div(
									vx.VXChart().Presets("funnelChart").Children(
										web.Slot(
											h.Div(
												h.Button("Past 7 Days").
													Class("text-body-2 rounded text-no-wrap border-0 flex-grow-1 d-flex align-center justify-center rounded px-2").
													Attr(":style", `
														 currentIndex === 0
														? 'background-color: #fff; color: #4a4a4a;'
														: 'background-color: transparent; color: rgb(117, 117, 117);'`).Attr("@click", "toggle(0)"),
												h.Button("Past 14 Days").Attr(":style", `
														 currentIndex === 1
														? 'background-color: #fff; color: #4a4a4a;'
														: 'background-color: transparent; color: rgb(117, 117, 117);'`).
													Class("text-body-2 rounded text-no-wrap border-0 flex-grow-1 d-flex align-center justify-center rounded px-2").
													Attr("@click", "toggle(1)"),
												h.Button("Past 30 Days").Attr(":style", `
														 currentIndex === 2
														? 'background-color: #fff; color: #4a4a4a;'
														: 'background-color: transparent; color: rgb(117, 117, 117);'`).
													Class("text-body-2 rounded text-no-wrap border-0 flex-grow-1 d-flex align-center justify-center rounded px-2").
													Attr("@click", "toggle(2)"),
											).Class("d-flex align-center bg-gre"+
												"y-lighten-3 rounded pa-1 mr-4 mt-4").Style("height: 32px;"),
										).Name("action").Scope("{ list, currentIndex, toggle }"),
									).Options([]vx.VXChartOption{
										{
											Title: &vx.VXChartOptionTitle{Text: "User Activity (7 Days)"},
											Series: &[]vx.VXChartOptionSeries{
												{
													Name: "7 days",
													Data: []interface{}{
														map[string]interface{}{"value": 1840863, "name": "View Products"},
														map[string]interface{}{"value": 588604, "name": "Add Products To Cart"},
														map[string]interface{}{"value": 202022, "name": "Purchase Products"},
													},
												},
											},
										},
										{
											Title: &vx.VXChartOptionTitle{Text: "User Activity (14 Days)"},
											Series: &[]vx.VXChartOptionSeries{
												{
													Name: "14 days",
													Data: []interface{}{
														map[string]interface{}{"value": 1840863 * 1.2, "name": "View Products"},
														map[string]interface{}{"value": 588604 * 1.2, "name": "Add Products To Cart"},
														map[string]interface{}{"value": 202022 * 1.2, "name": "Purchase Products"},
													},
												},
											},
										},
										{
											Title: &vx.VXChartOptionTitle{Text: "User Activity (30 Days)"},
											Series: &[]vx.VXChartOptionSeries{
												{
													Name: "30 days",
													Data: []interface{}{
														map[string]interface{}{"value": 1840863 * 1.4, "name": "View Products"},
														map[string]interface{}{"value": 588604 * 1.4, "name": "Add Products To Cart"},
														map[string]interface{}{"value": 202022 * 1.4, "name": "Purchase Products"},
													},
												},
											},
										},
									}),
								).Class("border border-gray-500 rounded-lg").Style("height: 600px;"),
							).Cols(12),
						).Class("mt-4"),
					).Value(1),
				).Attr("v-model", "xLocals.tab"),
			).Class("px-2"),
		).VSlot("{locals:xLocals}").Init("{tab:0}")
	})

	detailing.Section(se)
	return mb
}
