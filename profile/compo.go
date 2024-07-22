package profile

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/jinzhu/inflection"
	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/stateful"
	"github.com/qor5/x/v3/i18n"
	v "github.com/qor5/x/v3/ui/vuetify"
	"github.com/samber/lo"
	h "github.com/theplant/htmlgo"
	"golang.org/x/exp/maps"
)

const logoutURL = "/auth/logout"

func init() {
	stateful.RegisterActionableCompoType(&ProfileCompo{})
}

type ProfileCompo struct {
	b *Builder `inject:""`

	ID string `json:"id"`
}

func (c *ProfileCompo) CompoID() string {
	return fmt.Sprintf("ProfileCompo:%s", c.ID)
}

func (c *ProfileCompo) MustGetEventContext(ctx context.Context) (*web.EventContext, *Messages) {
	evCtx := web.MustGetEventContext(ctx)
	return evCtx, i18n.MustGetModuleMessages(evCtx.R, I18nProfileKey, Messages_en_US).(*Messages)
}

func (c *ProfileCompo) MarshalHTML(ctx context.Context) ([]byte, error) {
	user, err := c.b.currentUserFunc(ctx)
	if err != nil {
		return nil, err
	}

	prepend := v.VCard().Flat(true).Children(
		web.Slot().Name(v.VSlotPrepend).Children(
			v.VAvatar().Class("text-body-1 font-weight-medium text-primary bg-primary-lighten-2").Size(v.SizeLarge).Density(v.DensityCompact).Rounded(true).
				Text(ShortName(user.Name)).Children(
				h.Iff(user.Avatar != "", func() h.HTMLComponent {
					return v.VImg().Attr("alt", user.Name).Attr("src", user.Avatar)
				}),
			),
		),
		web.Slot().Name(v.VSlotTitle).Children(
			h.Div().Class("d-flex align-center ga-2 pt-1").Children(
				h.Div().Attr("v-pre", true).Text(user.Name).Class("text-subtitle-2 text-secondary"),
				c.userCompo(ctx, user),
			),
		),
		web.Slot().Name(v.VSlotSubtitle).Children(
			h.Div().Class("text-overline").Text(strings.ToUpper(user.GetFirstRole())),
		),
	)
	return stateful.Actionable(ctx, c, h.Div().Class("d-flex align-center ga-0").Children(
		prepend,
		v.VSpacer(),
		h.Iff(len(user.NotifCounts) > 0, func() h.HTMLComponent {
			return h.Div().Class("d-flex align-center px-4 border-s-sm h-50").Children(
				c.bellCompo(ctx, user.NotifCounts),
			)
		}),
	)).MarshalHTML(ctx)
}

func (c *ProfileCompo) bellCompo(ctx context.Context, notifCounts []*activity.NoteCount) h.HTMLComponent {
	_, msgr := c.MustGetEventContext(ctx)

	unreadBy := func(item *activity.NoteCount) int { return int(item.UnreadNotesCount) }
	unreadCount := lo.SumBy(notifCounts, unreadBy)

	listItems := []h.HTMLComponent{}
	groups := lo.GroupBy(notifCounts, func(item *activity.NoteCount) string {
		return item.ModelName
	})
	modelNames := maps.Keys(groups)
	sort.Strings(modelNames)
	for _, modelName := range modelNames {
		counts := groups[modelName]
		title := inflection.Plural(modelName)
		// TODO: i18n
		// TODO: href?
		href := fmt.Sprintf(
			"/%s?active_filter_tab=hasUnreadNotes&f_hasUnreadNotes=1",
			strings.ToLower(title),
		)
		listItems = append(listItems, v.VListItem().Href(href).Children(
			v.VListItemTitle(h.Text(title)),
			v.VListItemSubtitle(h.Text(msgr.UnreadMessages(lo.SumBy(counts, unreadBy)))),
		))
	}

	icon := v.VIcon("mdi-bell-outline").Size(20).Color("grey-darken-1")
	return v.VMenu().Children(
		web.Slot().Name("activator").Scope(`{props}`).Children(
			v.VBtn("").Attr("v-bind", "props").Size(36).Icon(true).Density(v.DensityCompact).Variant(v.VariantText).Children(
				h.Iff(unreadCount > 0, func() h.HTMLComponent {
					return v.VBadge(icon).Content(unreadCount).Dot(true).Color(v.ColorError)
				}).Else(func() h.HTMLComponent {
					return icon
				}),
			),
		),
		v.VCard(v.VList(listItems...)),
	)
}

func (c *ProfileCompo) userCompo(ctx context.Context, user *User) h.HTMLComponent {
	_, msgr := c.MustGetEventContext(ctx)

	children := []h.HTMLComponent{}
	for _, field := range user.Fields {
		children = append(children, h.Div().Class("d-flex flex-column ga-2").Children(
			h.Div().Attr("v-pre", true).Class("text-body-2 text-grey-darken-2").Text(field.Name),
			h.Div().Attr("v-pre", true).Class("text-subtitle-2 font-weight-medium text-grey-darken-4").Text(field.Value),
		))
	}
	children = append(children, h.Div().Class("d-flex flex-column ga-2").Children(
		v.VBtn(msgr.ViewLoginSessions).Variant(v.VariantTonal).Color(v.ColorSecondary),
		v.VBtn(msgr.Logout).Variant(v.VariantTonal).Color(v.ColorError).Attr("@click", web.Plaid().URL(logoutURL).Go()),
	))

	renameAction := stateful.PostAction(ctx, c,
		c.Rename, RenameRequest{},
		stateful.WithAppendFix(`v.request.name = xlocals.name`),
	).Go()
	return v.VMenu().CloseOnContentClick(false).Children(
		web.Slot().Name("activator").Scope(`{props}`).Children(
			v.VBtn("").Attr("v-bind", "props").Size(16).Icon(true).Density(v.DensityCompact).Variant(v.VariantText).Children(
				v.VIcon("mdi-chevron-right").Size(16),
			),
		),
		v.VCard().Width(300).Children(
			v.VCardText().Class("pa-0").Children(
				h.Div().Class("d-flex align-center ga-6 pa-6 bg-grey-lighten-4").Children(
					v.VAvatar().Class("text-h3 font-weight-medium text-primary bg-primary-lighten-2 rounded-lg").Size(80).Density(v.DensityCompact).
						Text(ShortName(user.Name)).Children(
						h.Iff(user.Avatar != "", func() h.HTMLComponent {
							return v.VImg().Attr("alt", user.Name).Attr("src", user.Avatar)
						}),
					),
					h.Div().Class("flex-grow-1 d-flex flex-column ga-4").Children(
						h.Div().Class("d-flex flex-column").Children(
							web.Scope().VSlot(`{ locals: xlocals }`).Init(fmt.Sprintf(`{editShow:false, name: %q}`, user.Name)).Children(
								h.Div().Attr("v-if", "!xlocals.editShow").Class("d-flex align-center ga-2").Children(
									h.Div().Attr("v-pre", true).Text(user.Name).Class("text-subtitle-1 font-weight-medium"),
									v.VBtn("").Size(20).Variant(v.VariantText).Color(v.ColorGreyDarken1).
										Attr("@click", "xlocals.editShow = true").Children(
										v.VIcon("mdi-pencil-outline"),
									),
								),
								h.Div().Attr("v-else", true).Style("height:24px").Class("d-flex align-center ga-2").Children(
									v.VTextField().Class("text-subtitle-1 font-weight-medium mt-n2").
										HideDetails(true).Density(v.DensityCompact).Autofocus(true).
										Color(v.ColorPrimary).
										Variant(v.VariantPlain).
										Attr("v-model", "xlocals.name").
										Attr("@keyup.enter", renameAction),
									h.Div().Class("d-flex align-center ga-1").Children(
										v.VBtn("").Size(20).Variant(v.VariantText).Color(v.ColorGreyDarken1).
											Attr("@click", "xlocals.editShow = false").Children(v.VIcon("mdi-close")),
										v.VBtn("").Size(20).Variant(v.VariantText).Color(v.ColorGreyDarken1).
											Attr("@click", renameAction).Children(v.VIcon("mdi-check")),
									),
								),
							),
							h.Div().Class("text-subtitle-2 font-weight-medium text-grey-darken-1").Text(user.GetFirstRole()),
						),
						h.Iff(user.Status != "", func() h.HTMLComponent {
							return v.VChip().Text(user.Status).
								Attr("style", "height:20px").Class("align-self-start px-1 text-caption").
								Label(true).Density(v.DensityCompact).Color(v.ColorSuccess)
						}),
					),
				),
				h.Div().Class("d-flex flex-column ga-6 pa-6").Children(children...),
			),
		),
	)
}

type RenameRequest struct {
	Name string `json:"name"`
}

func (c *ProfileCompo) Rename(ctx context.Context, req RenameRequest) (r web.EventResponse, _ error) {
	err := c.b.renameCallback(ctx, req.Name)
	if err != nil {
		presets.ShowMessage(&r, err.Error(), v.ColorError)
		return
	}
	_, msgr := c.MustGetEventContext(ctx)
	presets.ShowMessage(&r, msgr.SuccessfullyRename, v.ColorSuccess)
	r.Reload = true
	return r, nil
}
