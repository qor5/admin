package examples_vuetify

// @snippet_begin(VuetifyListSample)
import (
	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/examples"
	. "github.com/qor5/x/v3/ui/vuetify"
	. "github.com/theplant/htmlgo"
)

type ListItem struct {
	Type          string `json:"type,omitempty"`
	Title         string `json:"title,omitempty"`
	PrependAvatar string `json:"prependAvatar,omitempty"`
	Subtitle      string `json:"subtitle,omitempty"`
	Inset         bool   `json:"inset,omitempty"`
}

func HelloVuetifyList(ctx *web.EventContext) (pr web.PageResponse, err error) {
	wrapper := func(children ...HTMLComponent) HTMLComponent {
		return VContainer(
			VCard(children...),
		).GridList(Md).TextAlign(Xs, Center)
	}

	items := []ListItem{
		{
			Type:  "subheader",
			Title: "Today",
		},
		{
			Title:         "Brunch this weekend?",
			PrependAvatar: "https://cdn.vuetifyjs.com/images/lists/1.jpg",
			Subtitle:      `<span class="text-primary">Ali Connors</span> &mdash; I'll be in your neighborhood doing errands this weekend. Do you want to hang out?`,
		},
		{
			Type:  "divider",
			Inset: true,
		},
		{
			Title:         "Summer BBQ",
			PrependAvatar: "https://cdn.vuetifyjs.com/images/lists/2.jpg",
			Subtitle:      `<span class="text-primary">to Alex, Scott, Jennifer</span> &mdash; Wish I could come, but I'm out of town this weekend.`,
		},
		{
			Type:  "divider",
			Inset: true,
		},
		{
			Title:         "Oui oui",
			PrependAvatar: "https://cdn.vuetifyjs.com/images/lists/3.jpg",
			Subtitle:      `<span class="text-primary">Sandra Adams</span> &mdash; Do you have Paris recommendations? Have you ever been?`,
		},
		{
			Type:  "divider",
			Inset: true,
		},
		{
			Title:         "Birthday gift",
			PrependAvatar: "https://cdn.vuetifyjs.com/images/lists/4.jpg",
			Subtitle:      `<span class="text-primary">Trevor Hansen</span> &mdash; Have any ideas about what we should get Heidi for her birthday?`,
		},
		{
			Type:  "divider",
			Inset: true,
		},
		{
			Title:         "Recipe to try",
			PrependAvatar: "https://cdn.vuetifyjs.com/images/lists/5.jpg",
			Subtitle:      `<span class="text-primary">Britta Holt</span> &mdash; We should eat this: Grate, Squash, Corn, and tomatillo Tacos.`,
		},
	}

	pr.Body = wrapper(
		VToolbar(
			// VToolbarSideIcon(),
			VToolbarTitle("Inbox").Class("text-white"),
			VSpacer(),
			VBtn("").Icon(true).Children(
				VIcon("search"),
			),
		).Color("cyan"),
		VList(
			Template(
				Div().Attr("v-html", "subtitle"),
			).Attr("v-slot:subtitle", "{ subtitle }"),
		).Lines("three").ItemProps(true).Items(items),
	)

	return
}

var HelloVuetifyListPB = web.Page(HelloVuetifyList)

var HelloVuetifyListPath = examples.URLPathByFunc(HelloVuetifyList)

// @snippet_end
