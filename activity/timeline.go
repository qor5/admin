package activity

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/stateful"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/perm"
	v "github.com/qor5/x/v3/ui/vuetify"
	"github.com/samber/lo"
	h "github.com/theplant/htmlgo"
)

type Note struct {
	Note         string    `json:"note"`
	LastEditedAt time.Time `json:"last_edited_at"`
}

func init() {
	stateful.RegisterActionableCompoType(&TimelineCompo{})
}

type TimelineCompo struct {
	ab *Builder              `inject:""`
	mb *presets.ModelBuilder `inject:""`

	ID        string `json:"id"`
	ModelName string `json:"model_name"`
	ModelKeys string `json:"model_keys"`
	ModelLink string `json:"model_link"`
}

func (c *TimelineCompo) CompoID() string {
	return fmt.Sprintf("TimelineCompo:%s", c.ID)
}

func (c *TimelineCompo) VarCurrentActive() string {
	return fmt.Sprintf("__current_active_of_%s__", stateful.MurmurHash3(c.CompoID()))
}

func (*TimelineCompo) MustGetEventContext(ctx context.Context) (*web.EventContext, *Messages) {
	evCtx := web.MustGetEventContext(ctx)
	return evCtx, i18n.MustGetModuleMessages(evCtx.R, I18nActivityKey, Messages_en_US).(*Messages)
}

func (c *TimelineCompo) humanContent(ctx context.Context, log *ActivityLog) h.HTMLComponent {
	evCtx, msgr := c.MustGetEventContext(ctx)
	pmsgr := presets.MustGetMessages(evCtx.R)
	switch log.Action {
	case ActionNote:
		note := &Note{}
		if err := json.Unmarshal([]byte(log.Detail), note); err != nil {
			return h.Text(fmt.Sprintf("Failed to unmarshal detail: %v", err))
		}
		return h.Components(
			h.Div().Attr("v-if", "!xlocals.showEditBox").Class("d-flex flex-column").Children(
				h.Div(h.Text(msgr.AddedANote)),
				h.Div().Class("text-body-2").Style("white-space: pre-wrap").Text(fmt.Sprintf(`{{%q}}`, note.Note)),
				h.Iff(!note.LastEditedAt.IsZero(), func() h.HTMLComponent {
					return h.Div().Class("text-caption font-italic").Class("text-grey-darken-1").Children(
						h.Text(msgr.LastEditedAt(pmsgr.HumanizeTime(note.LastEditedAt))),
					)
				}),
			),
			h.Div().Attr("v-if", "!!xlocals.showEditBox").Class("flex-grow-1 d-flex flex-column mt-4").Style("position: relative").Children(
				v.VTextarea().Rows(2).Attr(":row-height", "12").Clearable(false).AutoGrow(true).Label("").Variant(v.VariantOutlined).
					Color(v.ColorPrimary).Class("text-grey-darken-3 textarea-with-bottom-btns").
					Attr(web.VField("note", note.Note)...),
				h.Div().Class("d-flex flex-row ga-2").Style("position: absolute; bottom: 32px; right: 12px").Children(
					v.VBtn("").Variant(v.VariantText).Color("grey-darken-3").Size(16).
						Attr("@click", "xlocals.showEditBox = false; toplocals.editing = false ").Children(
						v.VIcon("mdi-close").Size(16),
					),
					v.VBtn("").Variant(v.VariantText).Color(v.ColorPrimary).Size(16).
						Attr("@click", stateful.PostAction(ctx, c,
							c.UpdateNote, UpdateNoteRequest{
								LogID: log.ID,
							},
							stateful.WithAppendFix(`v.request.note = form["note"];`),
						).Go()).Children(
						v.VIcon("mdi-check").Size(16),
					),
				).Attr("v-on-mounted", `({watch}) => {
					watch(form, (val) => {
						toplocals.edited = true;
					})
				}`),
			),
		)
	case ActionView:
		return h.Div(h.Text(msgr.Viewed))
	case ActionCreate:
		return h.Div(h.Text(msgr.Created))
	case ActionEdit:
		diffs := []Diff{}
		if err := json.Unmarshal([]byte(log.Detail), &diffs); err != nil {
			return h.Text(fmt.Sprintf("Failed to unmarshal detail: %v", err))
		}

		// logModelBuilder := c.ab.GetLogModelBuilder(c.mb.GetPresetsBuilder())
		return h.Div().Class("d-flex flex-row align-center ga-2").Children(
			h.Div(h.Text(msgr.EditedNFields(len(diffs)))),
			// h.Iff(logModelBuilder != nil, func() h.HTMLComponent {
			// 	return v.VBtn(msgr.MoreInfo).Class("text-none text-overline d-flex align-center").
			// 		Variant(v.VariantTonal).Color(v.ColorPrimary).Size(v.SizeXSmall).PrependIcon("mdi-open-in-new").
			// 		Attr("@click", web.POST().
			// 			EventFunc(actions.DetailingDrawer).
			// 			Query(presets.ParamOverlay, actions.Dialog).
			// 			URL(logModelBuilder.Info().ListingHref()).
			// 			Query(presets.ParamID, fmt.Sprint(log.ID)).
			// 			Query(paramHideDetailTop, true).
			// 			Query(presets.ParamVarCurrentActive, c.VarCurrentActive()).
			// 			Go(),
			// 		)
			// }),
		)
	case ActionDelete:
		return h.Div(h.Text(msgr.Deleted))
	default:
		return h.Div().Attr("v-pre", true).Text(msgr.PerformAction(getActionLabel(evCtx, log.Action), log.Detail))
	}
}

func (c *TimelineCompo) MarshalHTML(ctx context.Context) ([]byte, error) {
	evCtx, msgr := c.MustGetEventContext(ctx)
	if c.mb.Info().Verifier().Do(presets.PermGet).WithReq(evCtx.R).IsAllowed() != nil {
		return h.Div().Attr("v-pre", true).Text(perm.PermissionDenied.Error()).MarshalHTML(ctx)
	}

	canAddNote := c.mb.Info().Verifier().Do(PermAddNote).WithReq(evCtx.R).IsAllowed() == nil
	canEditNote := c.mb.Info().Verifier().Do(PermEditNote).WithReq(evCtx.R).IsAllowed() == nil
	canDeleteNote := c.mb.Info().Verifier().Do(PermDeleteNote).WithReq(evCtx.R).IsAllowed() == nil

	children := []h.HTMLComponent{
		web.Scope().VSlot("{locals: xlocals, form}").Init("{showEditBox:false}").Children(
			h.Div().Class("d-flex flex-column ga-4 mb-8").Children(
				h.Div().Class("d-flex align-center ga-2").Children(
					h.Div().Class("text-h6").Text(msgr.Activities),
					v.VSpacer(),
					h.Iff(canAddNote, func() h.HTMLComponent {
						return v.VBtn(msgr.AddNote).Attr(":disabled", "xlocals.showEditBox || toplocals.editing").
							Class("text-caption").Variant(v.VariantTonal).Color("grey-darken-3").Size(v.SizeSmall).PrependIcon("mdi-plus").
							Attr("@click", "xlocals.showEditBox = true; toplocals.editing = true")
					}),
				),
				v.VDivider(),
				h.Div().Attr("v-if", "!!xlocals.showEditBox").Class("d-flex flex-column mb-n6").Style("position: relative").Children(
					v.VTextarea().Rows(2).Attr(":row-height", "12").Clearable(false).AutoGrow(true).Label("").Placeholder(msgr.AddNote).Variant(v.VariantOutlined).
						Color(v.ColorPrimary).Class("text-grey-darken-3 textarea-with-bottom-btns").
						Attr(web.VField("note", "")...),
					h.Div().Class("d-flex flex-row ga-2").Style("position: absolute; bottom: 32px; right: 12px").Children(
						v.VBtn("").Variant(v.VariantText).Color("grey-darken-3").Size(16).
							Attr("@click", "xlocals.showEditBox = false; toplocals.editing = false").Children(
							v.VIcon("mdi-close").Size(16),
						),
						v.VBtn("").Variant(v.VariantText).Color("grey-darken-3").Size(16).
							Attr("@click", stateful.PostAction(ctx, c,
								c.CreateNote, CreateNoteRequest{},
								stateful.WithAppendFix(`v.request.note = form["note"];`),
							).Go()).Children(
							v.VIcon("mdi-check").Size(16),
						),
					).Attr("v-on-mounted", `({watch}) => {
						watch(form, (val) => {
							toplocals.edited = true;
						})
					}`),
				),
			),
		),
	}

	logs, hasMore, err := c.ab.findLogsForTimeline(ctx, c.ModelName, c.ModelKeys)
	if err != nil {
		return nil, err
	}

	pmsgr := presets.MustGetMessages(evCtx.R)

	user, err := c.ab.currentUserFunc(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current user")
	}

	logModelBuilder := c.ab.GetLogModelBuilder(c.mb.GetPresetsBuilder())
	varCurrentActive := c.VarCurrentActive()
	for i, log := range logs {
		userName := log.User.Name
		if userName == "" {
			userName = msgr.UnknownUser
		}
		avatarText := ""
		if log.User.Avatar == "" {
			avatarText = strings.ToUpper(string([]rune(userName)[0:1]))
		}
		isAccent := true
		dotColor := v.ColorSuccess
		if i != 0 {
			isAccent = false
			dotColor = "grey-lighten-2"
		}
		idStr := fmt.Sprint(log.ID)

		hotspot := h.Div().Attr(":class", fmt.Sprintf(`{ "bg-grey-lighten-4": isHovering || vars.%s == %q }`, varCurrentActive, idStr)).
			Class("flex-grow-1 d-flex flex-column pe-1 pb-3 rounded").Style("padding-left: 2px;").Children(
			h.Div().Class("d-flex flex-row align-center ga-2").Children(
				v.VAvatar().Class("text-overline font-weight-medium text-primary bg-primary-lighten-2").Size(v.SizeXSmall).Density(v.DensityCompact).Rounded(true).Text(avatarText).Children(
					h.Iff(log.User.Avatar != "", func() h.HTMLComponent {
						return v.VImg().Attr("alt", userName).Attr("src", log.User.Avatar)
					}),
				),
				h.Div().Attr(":class", fmt.Sprintf(`{ "text-grey": !xlocals.isAccent && !isHovering && vars.%s != %q }`, varCurrentActive, idStr)).
					Class("font-weight-medium flex-grow-1").Children(
					h.Div().Attr("v-pre", true).Text(userName),
				),
				h.Iff(log.Action == ActionEdit && logModelBuilder != nil, func() h.HTMLComponent {
					return v.VIcon("mdi-chevron-right").
						Attr("v-if", fmt.Sprintf(`isHovering || vars.%s == %q`, varCurrentActive, idStr)).
						Size(v.SizeSmall).Class("text-grey-darken-4")
				}),
			),
			h.Div().Class("d-flex flex-row align-center ga-2").Children(
				h.Div().Style("width: 16px; flex-shrink:0"),
				h.Div().Attr(":class", fmt.Sprintf(`{ "text-grey": !xlocals.isAccent && !isHovering && vars.%s != %q }`, varCurrentActive, idStr)).
					Class("flex-grow-1").Children(
					c.humanContent(ctx, log),
				),
			),
		)
		if log.Action == ActionEdit && logModelBuilder != nil {
			hotspot.Attr("@click", web.POST().
				EventFunc(actions.DetailingDrawer).
				Query(presets.ParamOverlay, actions.Dialog).
				URL(logModelBuilder.Info().ListingHref()).
				Query(presets.ParamID, idStr).
				Query(paramHideDetailTop, true).
				Query(presets.ParamVarCurrentActive, c.VarCurrentActive()).
				Go())
		}

		var child h.HTMLComponent = h.Div().Class("d-flex flex-column ga-1").Children(
			h.Div().Class("d-flex flex-row align-center ga-2").Children(
				h.Div().Class("bg-"+dotColor).Style("width: 8px; height: 8px;").Class("rounded-circle"),
				h.Div(h.Text(pmsgr.HumanizeTime(log.CreatedAt))).Class(lo.If(isAccent, "text-grey").Else("text-grey-darken-1")),
			),
			h.Div().Class("d-flex flex-row ga-2").Children(
				h.Div().Class("bg-"+dotColor).Class("align-self-stretch").Style("width: 1px; margin: -6px 3.5px -2px 3.5px;"),
				hotspot,
			),
		)

		if log.Action == ActionNote && log.UserID == user.ID {
			child = h.Div().Class("d-flex flex-column").Style("position: relative").Children(
				h.Div().Attr("v-if", "isHovering && !xlocals.showEditBox && !toplocals.editing").Class("d-flex flex-row ga-1").
					Style("position: absolute; top: 21px; right: 16px").Children(
					h.Iff(canEditNote, func() h.HTMLComponent {
						return v.VBtn("").Variant(v.VariantText).Color("grey-darken-3").Size(v.SizeXSmall).Icon("mdi-square-edit-outline").
							Attr("@click", "xlocals.showEditBox = true; toplocals.editing = true")
					}),
					h.Iff(canDeleteNote, func() h.HTMLComponent {
						return v.VBtn("").Variant(v.VariantText).Color("grey-darken-3").Size(v.SizeXSmall).Icon("mdi-delete").
							Attr("@click", fmt.Sprintf(`toplocals.deletingLogID = %q`, idStr))
					}),
				),
				child,
			)
		}

		hoverable := (log.Action == ActionNote && log.UserID == user.ID) ||
			(log.Action == ActionEdit && logModelBuilder != nil)
		children = append(children, v.VHover().Disabled(!hoverable).Children(
			web.Slot().Name("default").Scope("{ isHovering, props }").Children(
				h.Div().Class("d-flex flex-column").Attr("v-bind", "props").Children(
					web.Scope().VSlot("{locals: xlocals, form}").Init(fmt.Sprintf(`{ showEditBox: false, isAccent: %t }`, isAccent)).Children(
						child,
					),
				),
			),
		))
	}

	if hasMore {
		children = append(children,
			h.Div().Class("d-flex flex-row ga-2").Children(
				h.Div().Class("bg-grey-lighten-2").Class("align-self-stretch").Style("width: 1px; margin: -6px 3.5px -2px 3.5px;"),
				h.Iff(logModelBuilder != nil, func() h.HTMLComponent {
					return v.VBtn(msgr.ViewAll).Variant(v.VariantText).Color(v.ColorPrimary).Size(v.SizeSmall).
						AppendIcon("mdi-chevron-right").Class("text-caption ms-n2").
						Attr("@click", web.Plaid().
							URL(logModelBuilder.Info().ListingHref()).
							Query("f_model_name", c.ModelName).
							Query("f_model_keys", c.ModelKeys).
							PushState(true).
							Go(),
						)
				}).Else(func() h.HTMLComponent {
					return h.Div().Class("text-caption text-grey").Text(msgr.CannotShowMore)
				}),
			),
		)
	}

	if len(logs) == 0 {
		children = append(children, h.Div().Class("text-body-2 text-grey align-self-center mb-4").Text(msgr.NoActivitiesYet))
	}

	varEditing := fmt.Sprintf(`__activity_editing_of_%s__`, stateful.MurmurHash3(c.CompoID()))
	return stateful.Actionable(ctx, c,
		web.Listen(
			presets.NotifModelsCreated(&ActivityLog{}), fmt.Sprintf(`
			if (!!payload.models && payload.models.length > 0 && payload.models.every(obj => obj.Hidden === true)) {
				return
			}
			%s
			`, stateful.ReloadAction(ctx, c, nil).Go()),
			presets.NotifModelsUpdated(&ActivityLog{}), fmt.Sprintf(`
			if (!!payload.models && Object.keys(payload.models).length > 0 && Object.values(payload.models).every(obj => obj.Hidden === true)) {
				return
			}
			%s
			`, stateful.ReloadAction(ctx, c, nil).Go()),
			presets.NotifModelsDeleted(&ActivityLog{}), stateful.ReloadAction(ctx, c, nil).Go(),
		),
		web.Scope().VSlot("{locals: toplocals}").Init(`{ deletingLogID: "", editing: false, edited: false }`).Children(
			h.Div().Class("activity-timeline-wrap").
				Attr("v-on-mounted", fmt.Sprintf(`({watch, watchEffect}) => {
					watch(() => toplocals.editing, (val) => {
						if (!val) {
							toplocals.edited = false
						}
					}, { immediate: true })
					watchEffect(() => {
						vars.%s.%s = toplocals.editing && toplocals.edited
						vars.%s = toplocals.deletingLogID
					})
				}`, presets.VarsPresetsDataChanged, varEditing, varCurrentActive)).
				Attr("v-on-unmounted", fmt.Sprintf(`() => { 
					delete(vars.%s.%s)
					delete(vars.%s)
				}`, presets.VarsPresetsDataChanged, varEditing, varCurrentActive)).
				Children(
					children...,
				),
			v.VDialog().MaxWidth("520px").
				Attr(":model-value", `toplocals.deletingLogID !== ""`).
				Attr("@update:model-value", `(value) => { toplocals.deletingLogID = value ? toplocals.deletingLogID : ""; }`).Children(
				v.VCard(
					v.VCardTitle(h.Text(msgr.DeleteNoteDialogTitle)),
					v.VCardText(h.Text(msgr.DeleteNoteDialogText)),
					v.VCardActions(
						v.VSpacer(),
						v.VBtn(msgr.Cancel).Variant(v.VariantFlat).Size(v.SizeSmall).Class("ml-2").
							Attr("@click", `toplocals.deletingLogID = ""`),
						v.VBtn(msgr.Delete).Color(v.ColorError).Variant(v.VariantTonal).Size(v.SizeSmall).
							Attr("@click", stateful.PostAction(ctx, c,
								c.DeleteNote, DeleteNoteRequest{},
								stateful.WithAppendFix(`v.request.log_id = parseInt(toplocals.deletingLogID, 10)`),
							).Go()),
					),
				),
			),
		),
	).MarshalHTML(ctx)
}

type CreateNoteRequest struct {
	Note string `json:"note"`
}

func (c *TimelineCompo) CreateNote(ctx context.Context, req CreateNoteRequest) (r web.EventResponse, _ error) {
	if c.ModelName == "" || c.ModelKeys == "" {
		presets.ShowMessage(&r, perm.PermissionDenied.Error(), v.ColorError)
		return
	}

	evCtx, msgr := c.MustGetEventContext(ctx)
	if c.mb.Info().Verifier().Do(PermAddNote).WithReq(evCtx.R).IsAllowed() != nil {
		presets.ShowMessage(&r, perm.PermissionDenied.Error(), v.ColorError)
		return
	}

	req.Note = strings.TrimSpace(req.Note)
	if req.Note == "" {
		presets.ShowMessage(&r, msgr.NoteCannotBeEmpty, v.ColorError)
		return
	}

	log, err := c.ab.MustGetModelBuilder(c.mb).create(ctx, ActionNote, c.ModelName, c.ModelKeys, c.ModelLink, &Note{
		Note: req.Note,
	})
	if err != nil {
		presets.ShowMessage(&r, msgr.FailedToCreateNote, v.ColorError)
		return
	}

	presets.ShowMessage(&r, msgr.SuccessfullyCreatedNote, v.ColorSuccess)
	r.Emit(presets.NotifModelsCreated(&ActivityLog{}), presets.PayloadModelsCreated{
		Models: []any{log},
	})
	return
}

type UpdateNoteRequest struct {
	LogID uint   `json:"log_id"`
	Note  string `json:"note"`
}

func (c *TimelineCompo) UpdateNote(ctx context.Context, req UpdateNoteRequest) (r web.EventResponse, _ error) {
	if c.ModelName == "" || c.ModelKeys == "" {
		presets.ShowMessage(&r, perm.PermissionDenied.Error(), v.ColorError)
		return
	}

	evCtx, msgr := c.MustGetEventContext(ctx)
	if c.mb.Info().Verifier().Do(PermEditNote).WithReq(evCtx.R).IsAllowed() != nil {
		presets.ShowMessage(&r, perm.PermissionDenied.Error(), v.ColorError)
		return
	}

	req.Note = strings.TrimSpace(req.Note)
	if req.Note == "" {
		presets.ShowMessage(&r, msgr.NoteCannotBeEmpty, v.ColorError)
		return
	}

	user, err := c.ab.currentUserFunc(ctx)
	if err != nil {
		presets.ShowMessage(&r, msgr.FailedToGetCurrentUser, v.ColorError)
		return
	}

	log := &ActivityLog{}
	if err := c.ab.db.Where("id = ?", req.LogID).First(log).Error; err != nil {
		presets.ShowMessage(&r, msgr.FailedToGetNote, v.ColorError)
		return
	}
	if log.UserID != user.ID {
		presets.ShowMessage(&r, msgr.YouAreNotTheNoteUser, v.ColorError)
		return
	}

	note := &Note{}
	if err := json.Unmarshal([]byte(log.Detail), note); err != nil {
		return r, errors.Wrap(err, "failed to unmarshal note")
	}
	if note.Note == req.Note {
		stateful.AppendReloadToResponse(&r, c)
		return
	}

	log.Detail = h.JSONString(&Note{
		Note:         req.Note,
		LastEditedAt: c.ab.db.NowFunc(),
	})
	if err := c.ab.db.Save(log).Error; err != nil {
		presets.ShowMessage(&r, msgr.FailedToUpdateNote, v.ColorError)
		return
	}

	presets.ShowMessage(&r, msgr.SuccessfullyUpdatedNote, v.ColorSuccess)

	id := fmt.Sprint(log.ID)
	r.Emit(presets.NotifModelsUpdated(&ActivityLog{}), presets.PayloadModelsUpdated{
		Ids:    []string{id},
		Models: map[string]any{id: log},
	})
	return
}

type DeleteNoteRequest struct {
	LogID uint `json:"log_id"`
}

func (c *TimelineCompo) DeleteNote(ctx context.Context, req DeleteNoteRequest) (r web.EventResponse, _ error) {
	if c.ModelName == "" || c.ModelKeys == "" {
		presets.ShowMessage(&r, perm.PermissionDenied.Error(), v.ColorError)
		return
	}

	evCtx, msgr := c.MustGetEventContext(ctx)
	if c.mb.Info().Verifier().Do(PermDeleteNote).WithReq(evCtx.R).IsAllowed() != nil {
		presets.ShowMessage(&r, perm.PermissionDenied.Error(), v.ColorError)
		return
	}

	user, err := c.ab.currentUserFunc(ctx)
	if err != nil {
		presets.ShowMessage(&r, msgr.FailedToGetCurrentUser, v.ColorError)
		return
	}

	result := c.ab.db.Where("id = ? AND user_id = ?", req.LogID, user.ID).Delete(&ActivityLog{})
	if err := result.Error; err != nil {
		presets.ShowMessage(&r, msgr.FailedToDeleteNote, v.ColorError)
		return
	}
	if result.RowsAffected == 0 {
		presets.ShowMessage(&r, msgr.YouAreNotTheNoteUser, v.ColorError)
		return
	}
	presets.ShowMessage(&r, msgr.SuccessfullyDeletedNote, v.ColorSuccess)
	r.Emit(presets.NotifModelsDeleted(&ActivityLog{}), presets.PayloadModelsDeleted{
		Ids: []string{fmt.Sprint(req.LogID)},
	})
	return
}
