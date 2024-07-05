package examples_vuetify

// @snippet_begin(VuetifyComponentsKitchen)

import (
	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/examples"
	. "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
)

var globalCities = []string{"Tokyo", "Hangzhou", "Shanghai"}

type formVals struct {
	Cities1 []string
	Cities2 []string

	MyItem string
}

var fv = formVals{
	Cities1: []string{
		"TK",
		"LD",
	},

	Cities2: []string{
		"Hangzhou",
		"Shanghai",
	},

	MyItem: "VItem2",
}

func VuetifyComponentsKitchen(ctx *web.EventContext) (pr web.PageResponse, err error) {
	var chips h.HTMLComponents
	for _, city := range globalCities {
		chips = append(chips,
			VChip(h.Text(city)).
				Closable(true).
				Attr("@click:close", web.POST().EventFunc("removeCity").Query("city", city).Go()),
		)
	}

	pr.Body = VContainer(
		examples.PrettyFormAsJSON(ctx),
		web.Scope(
			h.H1("Chips delete"),
			chips,

			h.H1("Chips group"),
			h.H2("Cities1"),
			VChipGroup(
				VChip(
					h.Text("Hangzhou"),
					VIcon("mdi-star").End(true),
				).Value("HZ"),
				VChip(h.Text("Shanghai")).Value("SH").Filter(true).Label(true),
				VChip(h.Text("Tokyo")).Value("TK").Filter(true),
				VChip(h.Text("New York")).Value("NY"),
				VChip(h.Text("London")).Value("LD"),
			).SelectedClass("bg-indigo").
				Attr("v-model", "form.Cities1").
				Multiple(true),
			h.H2("Cities2"),
			VAutocomplete().
				Items(globalCities).
				Chips(true).
				ClosableChips(true).
				Multiple(true).
				Attr("v-model", "form.Cities2"),

			h.H1("Items Group"),

			VItemGroup(
				VContainer(
					VRow(
						VCol(
							VItem(
								VCard(
									VCardTitle(h.Text("Item1")),
								).
									Height(200).
									Attr(":class", "['d-flex align-center', selectedClass]").
									Attr("@click", "toggle"),
							).Value("VItem1").Attr("v-slot", "{isSelected, selectedClass, toggle}"),
						),

						VCol(
							VItem(
								VCard(
									VCardTitle(h.Text("Item2")),
								).
									Height(200).
									Attr(":class", "['d-flex align-center', selectedClass]").
									Attr("@click", "toggle"),
							).Value("VItem2").Attr("v-slot", "{isSelected, selectedClass, toggle}"),
						),
					),
				),
			).
				SelectedClass("bg-primary").
				Attr("v-model", "form.MyItem"),

			VBtn("Submit").
				OnClick("submit"),
		).VSlot("{ locals, form }").FormInit(h.JSONString(fv)),
	)
	return
}

func submit2(ctx *web.EventContext) (r web.EventResponse, err error) {
	fv = formVals{}
	ctx.MustUnmarshalForm(&fv)

	r.Reload = true
	return
}

func removeCity(ctx *web.EventContext) (r web.EventResponse, err error) {
	city := ctx.R.FormValue("city")
	newCities := make([]string, 0)
	for _, c := range globalCities {
		if c != city {
			newCities = append(newCities, c)
		}
	}
	globalCities = newCities
	r.Reload = true
	return
}

var VuetifyComponentsKitchenPB = web.Page(VuetifyComponentsKitchen).
	EventFunc("removeCity", removeCity).
	EventFunc("submit", submit2)

var VuetifyComponentsKitchenPath = examples.URLPathByFunc(VuetifyComponentsKitchen)

// @snippet_end
