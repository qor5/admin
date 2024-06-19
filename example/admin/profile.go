package admin

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/qor5/admin/v3/example/models"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/login"
	"github.com/qor5/x/v3/perm"
	. "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

const (
	signOutAllSessionEvent  = "signOutAllSessionEvent"
	loginSessionDialogEvent = "loginSessionDialogEvent"
	accountRenameEvent      = "accountRenameEvent"

	paramName = "name"
)

func profile(db *gorm.DB) presets.ComponentFunc {
	return func(ctx *web.EventContext) h.HTMLComponent {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nExampleKey, Messages_en_US).(*Messages)

		u := getCurrentUser(ctx.R)
		if u == nil {
			return VBtn("Login").Variant(VariantText).Href("/auth/login")
		}

		var roles []string
		for _, role := range u.Roles {
			roles = append(roles, role.Name)
		}

		total := notifierCount(db)(ctx)
		content := notifierComponent(db)(ctx)
		icon := VIcon("mdi-bell-outline").Size(20).Color("grey-darken-1")
		notification := VMenu().Children(
			h.Template().Attr("v-slot:activator", "{ props }").Children(
				VBtn("").Icon(true).Children(
					h.If(total > 0,
						VBadge(
							icon,
						).Content(total).Dot(true).Color("error"),
					).Else(icon),
				).Attr("v-bind", "props").
					Density(DensityCompact).
					Variant(VariantText),
			),
			VCard(content),
		)

		prependProfile := VCard(
			web.Slot(
				VAvatar().Text(getAvatarShortName(u)).Color(ColorPrimaryLighten2).Size(SizeLarge).Class(fmt.Sprintf("rounded-lg text-%s", ColorPrimary)),
			).Name(VSlotPrepend),
			web.Slot(
				h.Div(
					h.Div(h.Text(u.Name)).Class(fmt.Sprintf(`text-subtitle-2 text-%s`, ColorSecondary)),
					VBtn("").
						Icon(true).Density(DensityCompact).Variant(VariantText).Children(
						VIcon("mdi-chevron-right").Size(SizeSmall),
					),
				).Class("d-flex justify-space-between align-center"),
			).Name(VSlotTitle),
			web.Slot(
				h.Div(h.Text(roles[0])),
			).Name(VSlotSubtitle),
		).Class(W100).Flat(true)

		profileMenuCard := VMenu(
			web.Slot().Name("activator").Scope(`{props}`).Children(
				prependProfile.Attr("v-bind", "props"),
			),
			VCard(
				VCardText(
					web.Scope(
						VCard(
							web.Slot(
								VAvatar().Text(getAvatarShortName(u)).Color(ColorPrimaryLighten2).
									Size(SizeXLarge).Class(fmt.Sprintf("rounded-lg text-%s", ColorPrimary), "mr-6"),
							).Name(VSlotPrepend),
							web.Slot(
								VTextField(
									web.Slot(
										VIcon("mdi-pencil-outline").Attr("@click", "locals.editShow=true"),
									).Name(VSlotAppend),
								).HideDetails(true).
									Autofocus(true).
									Color(ColorPrimary).
									Attr(":variant", fmt.Sprintf(`locals.editShow?"%s":"%s"`, VariantOutlined, VariantPlain)).
									Attr(":readonly", `!locals.editShow`).
									Attr(web.VField(paramName, u.Name)...).
									Attr("@blur", "locals.editShow=false").
									Attr("@keyup.enter", web.Plaid().EventFunc(accountRenameEvent).
										URL("/profile").Query(presets.ParamID, u.ID).Go()),
							).Name(VSlotTitle),
							web.Slot(
								h.Div(
									h.Text(roles[0]),
								),
								h.Div(
									VChip(h.Text(u.Status)).Label(true).Density(DensityCompact).Color(ColorSuccess).Class("px-1"),
								).Class("mt-2"),
							).Name(VSlotSubtitle),
						).Variant(VariantTonal).Rounded(false),
					).VSlot(`{ locals }`).Init(`{editShow:false}`),
				).Class("pa-0"),
				VCardText(
					VRow(
						VCol(
							vx.VXReadonlyField().Label(msgr.Email).Value(u.Account),
						),
					).Class("my-n6"),
					VRow(
						VCol(
							vx.VXReadonlyField().Label(msgr.Company).Value(u.Company),
						),
					).Class("my-n6"),
					VRow(
						VBtn("View login sessions").
							Attr("@click", web.Plaid().URL("/profile").EventFunc(loginSessionDialogEvent).Query(presets.ParamID, u.ID).Go()).
							Variant(VariantTonal).Color(ColorSecondary).Class(W100, "mt-6"),
					),
					VRow(
						VBtn("Logout").Attr("@click", web.Plaid().URL(logoutURL).Go()).Variant(VariantTonal).Color(ColorError).Class(W100, "mt-2"),
					),
				).Class("my-6 mx-2"),
			)).Location(LocationEnd).CloseOnContentClick(false).Width(300)

		profileNewLook := VCard(
			VCardTitle(
				profileMenuCard,
				VCardText(
					h.Div(
						notification,
					).Class("border-s-md", "pl-4", H75)),
			).Class("d-inline-flex align-center justify-space-between  justify-center pa-0", W100),
		).Class("pa-0").Class(W100)
		return profileNewLook
	}
}

type Profile struct{}

func configProfile(b *presets.Builder, db *gorm.DB) {
	m := b.Model(&Profile{}).URIName("profile").
		MenuIcon("mdi-account").Label("Profile").InMenu(false)
	m.RegisterEventFunc(signOutAllSessionEvent, func(ctx *web.EventContext) (r web.EventResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nExampleKey, Messages_en_US).(*Messages)

		u := getCurrentUser(ctx.R)

		if u.GetAccountName() == loginInitialUserEmail {
			return r, perm.PermissionDenied
		}

		if err = expireOtherSessionLogs(db, ctx.R, u.ID); err != nil {
			return r, err
		}

		presets.ShowMessage(&r, msgr.SignOutAllSuccessfullyTips, "")
		r.Reload = true
		return
	})
	m.RegisterEventFunc(loginSessionDialogEvent, loginSession(db))
	m.RegisterEventFunc(accountRenameEvent, accountRename(m, db))
}

func getAvatarShortName(u *models.User) string {
	name := u.Name
	if name == "" {
		name = u.Account
	}
	if rs := []rune(name); len(rs) > 1 {
		name = string(rs[:1])
	}

	return strings.ToUpper(name)
}

func loginSession(db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nExampleKey, Messages_en_US).(*Messages)
		presetsMsgr := presets.MustGetMessages(ctx.R)
		uid := ctx.Param(presets.ParamID)
		u := &models.User{}
		if err = db.First(&u, uid).Error; err != nil {
			return
		}
		var items []*models.LoginSession
		if err = db.Where("user_id = ?", u.ID).Find(&items).Error; err != nil {
			return
		}

		isPublicUser := false
		if u.GetAccountName() == loginInitialUserEmail {
			isPublicUser = true
		}

		currentTokenHash := getStringHash(login.GetSessionToken(loginBuilder, ctx.R), LoginTokenHashLen)

		var (
			expired        = msgr.Expired
			active         = msgr.Active
			currentSession = msgr.CurrentSession
		)

		activeDevices := make(map[string]struct{})
		for _, item := range items {
			if isPublicUser {
				item.IP = msgr.HideIPTips
			}

			if isTokenValid(*item) {
				item.Status = expired
			} else {
				item.Status = active
				activeDevices[fmt.Sprintf("%s#%s", item.Device, item.IP)] = struct{}{}
			}
			if item.TokenHash == currentTokenHash {
				item.Status = currentSession
			}

			item.Time = humanize.Time(item.CreatedAt)
		}

		{
			newItems := make([]*models.LoginSession, 0, len(items))
			for _, item := range items {
				if item.Status == expired {
					_, ok := activeDevices[fmt.Sprintf("%s#%s", item.Device, item.IP)]
					if ok {
						continue
					}
				}
				newItems = append(newItems, item)
			}
			items = newItems
		}

		if isPublicUser {
			if len(items) > 10 {
				items = items[:10]
			}
		}

		sort.Slice(items, func(i, j int) bool {
			if items[j].Status == currentSession {
				return false
			}
			if items[i].Status == expired &&
				items[j].Status == active {
				return false
			}
			if items[i].CreatedAt.Sub(items[j].CreatedAt) < 0 {
				return false
			}
			return true
		})

		sessionTableHeaders := []DataTableHeader{
			{msgr.Time, "Time", "25%", false},
			{msgr.Device, "Device", "25%", false},
			{msgr.IPAddress, "IP", "25%", false},
			{"", "Status", "25%", true},
		}

		body := web.Scope().VSlot("{locals}").Init("{dialog:true}").Children(
			VDialog(
				VCard(
					VCardTitle(
						h.Text(msgr.LoginSessions),
						VBtn("").Icon("mdi-close").Variant(VariantText).Attr("@click", "locals.dialog=false"),
					).Class("d-flex justify-space-between align-center", W100),
					VRow(
						VCol(VCardSubtitle(h.Text(msgr.LoginSessionsTips))),
						VCol(
							h.If(!isPublicUser,
								VBtn("").Attr("@click", web.Plaid().EventFunc(signOutAllSessionEvent).Go()).
									Variant(VariantOutlined).Color("primary").
									Children(VIcon("warning").Size(SizeSmall), h.Text(msgr.SignOutAllOtherSessions))),
						).Class("text-right mt-6 mr-4"),
					),
					h.Div(
						VDataTable().Headers(sessionTableHeaders).
							Items(items).
							ItemsPerPage(-1).HideDefaultFooter(true),
						VCardActions(VSpacer(), VBtn(presetsMsgr.Cancel).Variant(VariantOutlined).Attr("@click", "locals.dialog=false")).Class("pa-0"),
					).Class("pa-6"),
				).Class("mx-2 mt-12 mb-4"),
			).Attr("v-model", "locals.dialog").Width(780),
		)

		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{Name: presets.DialogPortalName, Body: body})

		return
	}
}

func accountRename(mb *presets.ModelBuilder, db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		var (
			uid  = ctx.Param(presets.ParamID)
			name = ctx.Param(paramName)
			u    = &models.User{}
		)
		if err = db.First(u, uid).Error; err != nil {
			return
		}
		if mb.Info().Verifier().Do(presets.PermUpdate).ObjectOn(u).WithReq(ctx.R).IsAllowed() != nil {
			presets.ShowMessage(&r, perm.PermissionDenied.Error(), ColorError)
			return
		}
		u.Name = name
		if err = db.Save(u).Error; err != nil {
			return
		}
		r.PushState = web.Location(url.Values{})
		return
	}
}
