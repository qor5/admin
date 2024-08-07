package login

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/stateful"
	"github.com/qor5/x/v3/i18n"
	v "github.com/qor5/x/v3/ui/vuetify"
	"github.com/samber/lo"
	h "github.com/theplant/htmlgo"
	"golang.org/x/exp/maps"
	"golang.org/x/text/language"
)

func (b *ProfileBuilder) Install(pb *presets.Builder) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.logoutURL == "" && b.lsb == nil {
		return errors.Errorf("profile: missing logout URL")
	}
	if b.pb != nil {
		return errors.Errorf("profile: already installed")
	}
	b.pb = pb
	pb.GetI18n().
		RegisterForModule(language.English, I18nAdminLoginKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nAdminLoginKey, Messages_zh_CN).
		RegisterForModule(language.Japanese, I18nAdminLoginKey, Messages_ja_JP)
	pb.ProfileFunc(func(evCtx *web.EventContext) h.HTMLComponent {
		return b.NewCompo(evCtx, "")
	})

	dc := pb.GetDependencyCenter()
	injectorName := b.injectorName()
	dc.RegisterInjector(injectorName)
	dc.MustProvide(injectorName, func() *ProfileBuilder {
		return b
	})
	return nil
}

type ProfileField struct {
	Name  string
	Value string
}

type Profile struct {
	ID          string
	Name        string
	Avatar      string
	Roles       []string
	Status      string
	Fields      []*ProfileField
	NotifCounts []*activity.NoteCount
}

func (u *Profile) GetFirstRole() string {
	role := "-"
	if len(u.Roles) > 0 {
		role = u.Roles[0]
	}
	return role
}

type ProfileBuilder struct {
	mu sync.RWMutex
	pb *presets.Builder

	lsb                 *SessionBuilder
	logoutURL           string
	disableNotification bool
	currentProfileFunc  func(ctx context.Context) (*Profile, error)
	renameCallback      func(ctx context.Context, newName string) error
}

func NewProfileBuilder(
	currentProfileFunc func(ctx context.Context) (*Profile, error),
	renameCallback func(ctx context.Context, newName string) error,
) *ProfileBuilder {
	return &ProfileBuilder{
		currentProfileFunc: currentProfileFunc,
		renameCallback:     renameCallback,
	}
}

func (b *ProfileBuilder) SessionBuilder(lsb *SessionBuilder) *ProfileBuilder {
	b.lsb = lsb
	return b
}

func (b *ProfileBuilder) LogoutURL(s string) *ProfileBuilder {
	b.logoutURL = s
	return b
}

func (b *ProfileBuilder) DisableNotification(v bool) *ProfileBuilder {
	b.disableNotification = v
	return b
}

func (b *ProfileBuilder) injectorName() string {
	return "__profile__"
}

func (b *ProfileBuilder) NewCompo(evCtx *web.EventContext, idSuffix string) h.HTMLComponent {
	b.mu.RLock()
	pb := b.pb
	b.mu.RUnlock()
	if pb == nil {
		panic("profile: not installed")
	}
	return h.Div().Class("w-100").Children(
		b.pb.GetDependencyCenter().MustInject(b.injectorName(), &ProfileCompo{
			ID: b.pb.GetURIPrefix() + idSuffix,
		}),
	)
}

func init() {
	stateful.RegisterActionableCompoType(&ProfileCompo{})
}

type ProfileCompo struct {
	b *ProfileBuilder `inject:""`

	ID string `json:"id"`
}

func (c *ProfileCompo) CompoID() string {
	return fmt.Sprintf("ProfileCompo:%s", c.ID)
}

func (c *ProfileCompo) MustGetEventContext(ctx context.Context) (*web.EventContext, *Messages) {
	evCtx := web.MustGetEventContext(ctx)
	return evCtx, i18n.MustGetModuleMessages(evCtx.R, I18nAdminLoginKey, Messages_en_US).(*Messages)
}

func (c *ProfileCompo) MarshalHTML(ctx context.Context) ([]byte, error) {
	user, err := c.b.currentProfileFunc(ctx)
	if err != nil {
		return nil, err
	}

	showBellCompo := !c.b.disableNotification && len(user.NotifCounts) > 0
	return stateful.Actionable(ctx, c, web.Scope().VSlot("{ locals: xlocals }").Init("{ userCardVisible: false }").Children(
		h.Div().Class("d-flex align-center ga-2 pa-3").Children(
			v.VAvatar().Class("text-body-1 font-weight-medium text-primary bg-primary-lighten-2").Size(v.SizeLarge).Density(v.DensityCompact).Rounded(true).
				Text(activity.FirstUpperWord(user.Name)).Children(
				h.Iff(user.Avatar != "", func() h.HTMLComponent {
					return v.VImg().Attr("alt", user.Name).Attr("src", user.Avatar)
				}),
			),
			h.Div().Class("d-flex flex-column flex-1-1").StyleIf("max-width: 119px", showBellCompo).Children(
				h.Div().Class("d-flex align-center ga-2 pt-1").Children(
					h.Div().Attr("v-pre", true).Text(user.Name).Class("flex-1-1 text-subtitle-2 text-secondary text-truncate"),
					c.userCardCompo(ctx, user, "xlocals.userCardVisible"),
				),
				h.Div().Class("text-overline text-grey-darken-1").Text(strings.ToUpper(user.GetFirstRole())),
			),
			h.Iff(showBellCompo, func() h.HTMLComponent {
				return h.Div().Class("d-flex align-center px-4 me-n3 border-s-sm h-50").Children(
					c.bellCompo(ctx, user.NotifCounts),
				)
			}),
		).Attr("@click", "xlocals.userCardVisible = !xlocals.userCardVisible"),
	)).MarshalHTML(ctx)
}

func (c *ProfileCompo) bellCompo(ctx context.Context, notifCounts []*activity.NoteCount) h.HTMLComponent {
	evCtx, msgr := c.MustGetEventContext(ctx)

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
		title := i18n.T(evCtx.R, presets.ModelsI18nModuleKey, modelName)

		listItem := v.VListItem().Children(
			v.VListItemTitle(h.Text(title)),
			v.VListItemSubtitle(h.Text(msgr.UnreadMessages(lo.SumBy(counts, unreadBy)))),
		)

		var href string
		hasModelLabel, ok := lo.Find(counts, func(item *activity.NoteCount) bool {
			return item.ModelLabel != "" && item.ModelLabel != activity.NopModelLabel
		})
		if ok {
			href = activity.GetHasUnreadNotesHref(hasModelLabel.ModelLabel)
		}
		if href != "" {
			listItem.Href(href)
		}

		listItems = append(listItems, listItem)
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
			h.Components(
				lo.Map(modelNames, func(modelName string, _ int) h.HTMLComponent {
					return web.Listen(
						activity.NotifiLastViewedAtUpdated(modelName),
						stateful.ReloadAction(ctx, c, nil).Go(),
					)
				})...,
			),
		),
		v.VCard(v.VList(listItems...)),
	)
}

func (c *ProfileCompo) userCardCompo(ctx context.Context, user *Profile, vmodel string) h.HTMLComponent {
	_, msgr := c.MustGetEventContext(ctx)

	children := []h.HTMLComponent{}
	for _, field := range user.Fields {
		children = append(children, h.Div().Class("d-flex flex-column ga-2").Children(
			h.Div().Attr("v-pre", true).Class("text-body-2 text-grey-darken-2").Text(field.Name),
			h.Div().Attr("v-pre", true).Class("text-subtitle-2 font-weight-medium text-grey-darken-4").Text(field.Value),
		))
	}
	var clickLogout string
	if c.b.lsb != nil {
		clickLogout = web.Plaid().URL(c.b.lsb.GetLoginBuilder().LogoutURL).Go()
	} else {
		clickLogout = web.Plaid().URL(c.b.logoutURL).Go()
	}
	children = append(children, h.Div().Class("d-flex flex-column ga-2").Children(
		h.Iff(c.b.lsb != nil, func() h.HTMLComponent {
			return v.VBtn(msgr.ViewLoginSessions).Variant(v.VariantTonal).Color(v.ColorSecondary).Attr("@click", c.b.lsb.OpenSessionsDialog())
		}),
		v.VBtn(msgr.Logout).Variant(v.VariantTonal).Color(v.ColorError).Attr("@click", clickLogout),
	))

	renameAction := stateful.PostAction(ctx, c,
		c.Rename, RenameRequest{},
		stateful.WithAppendFix(`v.request.name = xlocals.name`),
	).Go()
	compo := v.VMenu().CloseOnContentClick(false).Children(
		web.Slot().Name("activator").Scope(`{props}`).Children(
			v.VBtn("").Attr("v-bind", "props").Size(16).Icon(true).Density(v.DensityCompact).Variant(v.VariantText).Children(
				v.VIcon("mdi-chevron-right").Size(16),
			),
		),
		v.VCard().Width(300).Children(
			v.VCardText().Class("pa-0").Children(
				h.Div().Class("d-flex align-center ga-6 pa-6 bg-grey-lighten-4").Children(
					v.VAvatar().Class("text-h3 font-weight-medium text-primary bg-primary-lighten-2 rounded-lg").Size(80).Density(v.DensityCompact).
						Text(activity.FirstUpperWord(user.Name)).Children(
						h.Iff(user.Avatar != "", func() h.HTMLComponent {
							return v.VImg().Attr("alt", user.Name).Attr("src", user.Avatar)
						}),
					),
					h.Div().Class("flex-1-1 d-flex flex-column ga-4").Style("max-width:148px").Children(
						h.Div().Class("d-flex flex-column").Children(
							web.Scope().VSlot(`{ locals: xlocals }`).Init(fmt.Sprintf(`{editShow:false, name: %q}`, user.Name)).Children(
								h.Div().Attr("v-if", "!xlocals.editShow").Class("d-flex align-center ga-2").Children(
									h.Div().Attr("v-pre", true).Text(user.Name).Class("text-subtitle-1 font-weight-medium text-truncate"),
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
						}).Else(func() h.HTMLComponent {
							return h.Div().Style("height:20px")
						}),
					),
				),
				h.Div().Class("d-flex flex-column ga-6 pa-6").Children(children...),
			),
		),
	)
	if vmodel != "" {
		compo.Attr("v-model", vmodel)
	}
	return compo
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
	web.AppendRunScripts(&r, web.Plaid().MergeQuery(true).Go())
	return r, nil
}
