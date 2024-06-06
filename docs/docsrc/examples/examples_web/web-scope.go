package examples_web

import (
	"github.com/qor5/docs/v3/docsrc/examples"
	"github.com/qor5/docs/v3/docsrc/utils"
	"github.com/qor5/web/v3"
	. "github.com/qor5/x/v3/ui/vuetify"
	. "github.com/theplant/htmlgo"
)

// @snippet_begin(WebScopeUseLocalsSample1)
func WebScopeUseLocals(ctx *web.EventContext) (pr web.PageResponse, err error) {
	pr.Body = VContainer(
		VIcon("home"),
		VBtn("Test Can Not Change Other Scope").
			PrependIcon("mdi-home").
			Attr("@click", `locals.btnLabel = "YES"`),
		web.Scope(
			VCard(
				VList(
					VListSubheader(
						Text("REPORTS"),
					),
					VListItem(
						web.Slot(
							VIcon("").Attr(":icon", "item.icon"),
						).Name("prepend"),
						VListItemTitle().Attr("v-text", "item.text"),
					).Attr("v-for", "(item, i) in locals.items").
						Attr(":key", "i").
						Attr(":value", "i").
						Attr("color", "red"),
				).Attr("dense", true).Attr("v-model:selected", "locals.selectedItem"),

				VCardActions(
					VBtn("").
						Variant("text").
						Attr("v-text", "locals.btnLabel + locals.selectedItem").
						Attr("@click", `
if (locals.btnLabel == "Add") {
	locals.items.push({text: "B", icon: "mdi-check"});
	locals.btnLabel = "Remove";
} else {
	locals.items.splice(locals.selectedItem, 1)
	locals.btnLabel = "Add";
}`),
				),
			).Class("mx-auto").
				Attr("max-width", "300").
				Attr("tile", ""),
		).Init(`{ selectedItem: 0, btnLabel:"Add", items: [{text: "A", icon: "mdi-clock"}]}`).
			VSlot("{ locals }"),
	)
	return
}

var UseLocalsPB = web.Page(WebScopeUseLocals)

// @snippet_end

// @snippet_begin(WebScopeUsePlaidFormSample1)
var materialID, materialName, rawMaterialID, rawMaterialName, countryID, countryName, productName string

func WebScopeUseForm(ctx *web.EventContext) (pr web.PageResponse, err error) {
	pr.Body = Div(
		H3("Form Content"),
		utils.PrettyFormAsJSON(ctx),
		web.Scope(

			Div(
				Div(
					Fieldset(
						Legend("Product Form"),
						Div(
							Label("Product Name"),
							Input("").
								Type("text").
								Attr("v-model", "form.ProductName"),
						),
						Div(
							Label("Material ID"),
							Input("").
								Type("text").Disabled(true).
								Attr("v-model", "form.MaterialID"),
						),

						web.Scope(
							Fieldset(
								Legend("Material Form"),

								Div(
									Label("Material Name"),
									Input("").
										Type("text").
										Attr("v-model", "form.MaterialName"),
								),
								Div(
									Label("Raw Material ID"),
									Input("").
										Type("text").Disabled(true).
										Attr("v-model", "form.RawMaterialID"),
								),
								web.Scope(
									Fieldset(
										Legend("Raw Material Form"),

										Div(
											Label("Raw Material Name"),
											Input("").
												Type("text").
												Attr("v-model", "form.RawMaterialName"),
										),

										Button("Send").Style(`background: orange;`).Attr("@click", web.POST().EventFunc("updateValue").Go()),
									).Style(`background: orange;`),
								).VSlot("{ form, locals }").FormInit(struct{ RawMaterialName string }{RawMaterialName: rawMaterialName}),

								Button("Send").Style(`background: brown;`).Attr("@click", web.POST().EventFunc("updateValue").Go()),
							).Style(`background: brown;`),
						).VSlot("{ form, locals }"),

						Div(
							Label("Country ID"),
							Input("").
								Type("text").Disabled(true).
								Attr("v-model", "form.CountryID"),
						),

						web.Scope(
							Fieldset(
								Legend("Country Of Origin Form"),

								Div(
									Label("Country Name"),
									Input("").
										Type("text").
										Attr("v-model", "form.CountryName"),
								),

								Button("Send").Style(`background: red;`).Attr("@click", web.POST().EventFunc("updateValue").Go()),
							).Style(`background: red;`),
						).VSlot("{ form, locals }").FormInit(struct{ CountryName string }{CountryName: countryName}),

						Div(
							Button("Send").Style(`background: grey;`).Attr("@click", web.POST().EventFunc("updateValue").Go())),
					).Style(`background: grey;`)),
			).Style(`width:600px;`),
		).VSlot("{ locals, form }").FormInit("{ProductName: 'Product1', MaterialID: '55', MaterialName: 'Material1', RawMaterialID: '77', RawMaterialName: 'RawMaterial1', CountryID: '88', CountryName: 'Country1'}"),
	)

	return
}

func updateValue(ctx *web.EventContext) (er web.EventResponse, err error) {
	ctx.R.ParseForm()
	if v := ctx.R.Form.Get("ProductName"); v != "" {
		productName = v
	}
	if v := ctx.R.Form.Get("MaterialName"); v != "" {
		materialName = v
		materialID = "66"
	}
	if v := ctx.R.Form.Get("RawMaterialName"); v != "" {
		rawMaterialName = v
		rawMaterialID = "88"
	}
	if v := ctx.R.Form.Get("CountryName"); v != "" {
		countryName = v
		countryID = "99"
	}
	er.Reload = true
	return
}

var UsePlaidFormPB = web.Page(WebScopeUseForm).
	EventFunc("updateValue", updateValue)

// @snippet_end

var (
	WebScopeUseLocalsPath = examples.URLPathByFunc(WebScopeUseLocals)
	WebScopeUseFormPath   = examples.URLPathByFunc(WebScopeUseForm)
)
