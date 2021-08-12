package media_library_view

import (
	"fmt"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	. "github.com/goplaid/x/vuetify"
	"github.com/jinzhu/gorm"
	"github.com/qor/media/media_library"
	h "github.com/theplant/htmlgo"
)

type MediaBoxConfigKey int

const MediaBoxConfig MediaBoxConfigKey = iota

func MediaBoxComponentFunc(db *gorm.DB) presets.FieldComponentFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		portalName := fmt.Sprintf("%s_Portal", field.Name)
		ctx.Hub.RegisterEventFunc(portalName, fileChooser(db, portalName))

		cfg := field.ContextValue(MediaBoxConfig).(*media_library.MediaBoxConfig)
		_ = cfg
		return h.Components(
			VCard(
				VCardActions(
					VBtn("Choose File").
						Depressed(true).
						Class("ml-2").
						OnClick(portalName),
				),
			).Class("mb-2"),
			web.Portal().Name(portalName),
		)
	}
}

func MediaBoxSetterFunc(db *gorm.DB) presets.FieldSetterFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
		return
	}
}

func fileChooser(db *gorm.DB, portalName string) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		//msgr := presets.MustGetMessages(ctx.R)
		uploadEventName := fmt.Sprintf("%s_upload", portalName)
		ctx.Hub.RegisterEventFunc(uploadEventName, uploadFile(db, portalName))

		files := []string{
			"https://cdn.vuetifyjs.com/images/cards/house.jpg",
			"https://cdn.vuetifyjs.com/images/cards/road.jpg",
			"https://cdn.vuetifyjs.com/images/cards/plane.jpg",
			"https://cdn.vuetifyjs.com/images/cards/house.jpg",
			"https://cdn.vuetifyjs.com/images/cards/road.jpg",
			"https://cdn.vuetifyjs.com/images/cards/plane.jpg",
			"https://cdn.vuetifyjs.com/images/cards/house.jpg",
			"https://cdn.vuetifyjs.com/images/cards/road.jpg",
			"https://cdn.vuetifyjs.com/images/cards/plane.jpg",
			"https://cdn.vuetifyjs.com/images/cards/house.jpg",
			"https://cdn.vuetifyjs.com/images/cards/road.jpg",
			"https://cdn.vuetifyjs.com/images/cards/plane.jpg",
		}

		ct := VContainer().Fluid(true)
		lastRow := VRow(
			VCol(
				VCard(
					VCardTitle(h.Text("Upload a file")),
					web.Bind(
						VFileInput().
							Class("justify-center").
							Label("New File").
							FieldName("NewFile").
							HideInput(true),
					).OnInput(uploadEventName),
				).Height(200),
			).Cols(3),
		)
		for _, f := range files {
			lastRow.AppendChildren(
				VCol(
					VCard(
						VImg().Src(f).Height(200),
					),
				).Cols(3),
			)
		}
		ct.AppendChildren(lastRow)

		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: portalName,
			Body: VDialog(
				VCard(
					VToolbar(
						VBtn("").
							Icon(true).
							Dark(true).
							Attr("@click", "vars.show = false").
							Children(
								VIcon("close"),
							),
						VToolbarTitle("Choose a File"),
					).Color("primary").
						MaxHeight(64).
						Flat(true).
						Dark(true),

					ct,
				).Tile(true),
			).
				Fullscreen(true).
				HideOverlay(true).
				Transition("dialog-bottom-transition").
				Scrollable(true).
				Attr("v-model", "vars.show").
				Attr(web.InitContextVars, `{show: false}`),
			AfterLoaded: "setTimeout(function(){ comp.vars.show = true }, 100)",
		})
		return
	}
}

func uploadFile(db *gorm.DB, portalName string) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		return
	}
}
