package activity

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
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

func (c *TimelineCompo) MustGetEventContext(ctx context.Context) (*web.EventContext, *Messages) {
	evCtx := web.MustGetEventContext(ctx)
	return evCtx, i18n.MustGetModuleMessages(evCtx.R, I18nActivityKey, Messages_en_US).(*Messages)
}

func (c *TimelineCompo) humanContent(ctx context.Context, log *ActivityLog, forceTextColor string) h.HTMLComponent {
	_, msgr := c.MustGetEventContext(ctx)
	switch log.Action {
	case ActionNote:
		note := &Note{}
		if err := json.Unmarshal([]byte(log.Detail), note); err != nil {
			return h.Text(fmt.Sprintf("Failed to unmarshal detail: %v", err))
		}
		return h.Components(
			h.Div().Attr("v-if", "!xlocals.showEditBox").Class("d-flex flex-column").Children(
				h.Div(h.Text(msgr.AddedANote)).ClassIf(forceTextColor, forceTextColor != ""),
				h.Pre(note.Note).Attr("v-pre", true).Style("white-space: pre-wrap").ClassIf(forceTextColor, forceTextColor != ""),
				h.Iff(!note.LastEditedAt.IsZero(), func() h.HTMLComponent {
					return h.Div().Class("text-caption font-italic").Class(lo.If(forceTextColor != "", forceTextColor).Else("text-grey-darken-1")).Children(
						h.Text(msgr.LastEditedAt(humanize.Time(note.LastEditedAt))),
					)
				}),
			),
			h.Div().Attr("v-if", "!!xlocals.showEditBox").Class("flex-grow-1 d-flex flex-column mt-4").Style("position: relative").Children(
				v.VTextarea().Rows(3).Attr("row-height", "12").Clearable(false).AutoGrow(true).Label("").Variant(v.VariantOutlined).
					Attr(web.VField("note", note.Note)...),
				h.Div().Class("d-flex flex-row ga-1").Style("position: absolute; top: 6px; right: 6px").Children(
					v.VBtn("").Variant(v.VariantText).Color("grey-darken-3").Size(v.SizeXSmall).Icon("mdi-close").
						Attr("@click", "xlocals.showEditBox = false"),
					v.VBtn("").Variant(v.VariantText).Color("grey-darken-3").Size(v.SizeXSmall).Icon("mdi-check").
						Attr("@click", stateful.PostAction(ctx, c,
							c.UpdateNote, UpdateNoteRequest{
								LogID: log.ID,
							},
							stateful.WithAppendFix(`v.request.note = form["note"];`),
						).Go()),
				),
			),
		)
	case ActionView:
		return h.Div(h.Text(msgr.Viewed)).ClassIf(forceTextColor, forceTextColor != "")
	case ActionCreate:
		return h.Div(h.Text(msgr.Created)).ClassIf(forceTextColor, forceTextColor != "")
	case ActionEdit:
		diffs := []Diff{}
		if err := json.Unmarshal([]byte(log.Detail), &diffs); err != nil {
			return h.Text(fmt.Sprintf("Failed to unmarshal detail: %v", err))
		}
		presetInstalled := c.ab.IsPresetInstalled(c.mb.GetPresetsBuilder())
		return h.Div().Class("d-flex flex-row align-center ga-2").Children(
			h.Div(h.Text(msgr.EditedNFields(len(diffs)))).ClassIf(forceTextColor, forceTextColor != ""),
			h.Iff(presetInstalled, func() h.HTMLComponent {
				return v.VBtn(msgr.MoreInfo).Class("text-none text-overline d-flex align-center").
					Variant(v.VariantTonal).Color(v.ColorPrimary).Size(v.SizeXSmall).PrependIcon("mdi-open-in-new").
					Attr("@click", web.POST().
						EventFunc(actions.DetailingDrawer).
						Query(presets.ParamOverlay, actions.Dialog).
						URL(fmt.Sprintf("%s/activity-logs/%d", c.mb.GetPresetsBuilder().GetURIPrefix(), log.ID)).
						Query(paramHideDetailTop, true).
						Go(),
					)
			}),
		)
	case ActionDelete:
		return h.Div(h.Text(msgr.Deleted)).ClassIf(forceTextColor, forceTextColor != "")
	default:
		return h.Div().Attr("v-pre", true).Text(msgr.PerformAction(log.Action, log.Detail)).ClassIf(forceTextColor, forceTextColor != "")
	}
}

func (c *TimelineCompo) MarshalHTML(ctx context.Context) ([]byte, error) {
	user, err := c.ab.currentUserFunc(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current user")
	}

	evCtx, msgr := c.MustGetEventContext(ctx)
	if c.mb.Info().Verifier().Do(presets.PermGet).WithReq(evCtx.R).IsAllowed() != nil {
		return h.Div().Attr("v-pre", true).Text(perm.PermissionDenied.Error()).MarshalHTML(ctx)
	}

	children := []h.HTMLComponent{
		h.Div().Class("text-h6 mb-8").Text(msgr.Activities),
		web.Scope().VSlot("{locals: xlocals,form}").Init("{showEditBox:false}").Children(
			v.VBtn(msgr.AddNote).Attr("v-if", "!xlocals.showEditBox").
				Class("text-none mb-4").Variant(v.VariantTonal).Color("grey-darken-3").Size(v.SizeDefault).PrependIcon("mdi-plus").
				Attr("@click", "xlocals.showEditBox = true"),
			h.Div().Attr("v-if", "!!xlocals.showEditBox").Class("d-flex flex-column").Style("position: relative").Children(
				v.VTextarea().Rows(3).Attr("row-height", "12").Clearable(false).AutoGrow(true).Label(msgr.AddNote).Variant(v.VariantOutlined).
					Attr(web.VField("note", "")...),
				h.Div().Class("d-flex flex-row ga-1").Style("position: absolute; top: 6px; right: 6px").Children(
					v.VBtn("").Variant(v.VariantText).Color("grey-darken-3").Size(v.SizeXSmall).Icon("mdi-close").
						Attr("@click", "xlocals.showEditBox = false"),
					v.VBtn("").Variant(v.VariantText).Color("grey-darken-3").Size(v.SizeXSmall).Icon("mdi-check").
						Attr("@click", stateful.PostAction(ctx, c,
							c.CreateNote, CreateNoteRequest{},
							stateful.WithAppendFix(`v.request.note = form["note"];`),
						).Go()),
				),
			),
		),
	}

	logs, err := c.ab.getActivityLogs(ctx, c.ModelName, c.ModelKeys)
	if err != nil {
		return nil, err
	}

	for i, log := range logs {
		creatorName := log.Creator.Name
		if creatorName == "" {
			creatorName = msgr.UnknownCreator
		}
		avatarText := ""
		if log.Creator.Avatar == "" {
			avatarText = strings.ToUpper(string([]rune(creatorName)[0:1]))
		}
		dotColor := v.ColorSuccess
		if i != 0 {
			dotColor = "grey-lighten-2"
		}
		var child h.HTMLComponent = h.Div().Class("d-flex flex-column ga-1").Children(
			h.Div().Class("d-flex flex-row align-center ga-2").Children(
				h.Div().Class("bg-"+dotColor).Style("width: 8px; height: 8px;").Class("rounded-circle"),
				h.Div(h.Text(humanize.Time(log.CreatedAt))).Class(lo.If(i != 0, "text-grey").Else("text-grey-darken-1")),
			),
			h.Div().Class("d-flex flex-row ga-2").Children(
				h.Div().Class("bg-"+dotColor).Class("align-self-stretch").Style("width: 1px; margin: -6px 3.5px -2px 3.5px;"),
				h.Div().Class("flex-grow-1 d-flex flex-column pb-3").Children(
					h.Div().Class("d-flex flex-row align-center ga-2").Children(
						v.VAvatar().Class("text-overline font-weight-medium text-primary bg-primary-lighten-2").Size(v.SizeXSmall).Density(v.DensityCompact).Rounded(true).Text(avatarText).Children(
							h.Iff(log.Creator.Avatar != "", func() h.HTMLComponent {
								return v.VImg().Attr("alt", creatorName).Attr("src", log.Creator.Avatar)
							}),
						),
						h.Div().Attr("v-pre", true).Text(creatorName).Class("font-weight-medium").ClassIf("text-grey", i != 0),
					),
					h.Div().Class("d-flex flex-row align-center ga-2").Children(
						h.Div().Style("width: 16px"),
						c.humanContent(ctx, log, lo.If(i != 0, "text-grey").Else("")),
					),
				),
			),
		)
		if log.Action == ActionNote {
			child = web.Scope().VSlot("{locals: xlocals, form}").Init("{showEditBox:false}").Children(
				v.VHover().Disabled(log.CreatorID != user.ID).Children(
					web.Slot().Name("default").Scope("{ isHovering, props }").Children(
						h.Div().Class("d-flex flex-column").Style("position: relative").Attr("v-bind", "props").Children(
							h.Div().Attr("v-if", "isHovering && !xlocals.showEditBox").Class("d-flex flex-row ga-1").
								Style("position: absolute; top: 21px; right: 16px").Children(
								v.VBtn("").Variant(v.VariantText).Color("grey-darken-3").Size(v.SizeXSmall).Icon("mdi-square-edit-outline").
									Attr("@click", "xlocals.showEditBox = true"),
								v.VBtn("").Variant(v.VariantText).Color("grey-darken-3").Size(v.SizeXSmall).Icon("mdi-delete").
									Attr("@click", fmt.Sprintf(`toplocals.deletingLogID = %d`, log.ID)),
							),
							child,
						),
					),
				),
			)
		}
		children = append(children, child)
	}

	if len(logs) == 0 {
		children = append(children, h.Div().Class("text-body-2 text-grey align-self-center mb-4").Text(msgr.NoActivitiesYet))
	}

	reloadAction := fmt.Sprintf(`
	if (!!payload.models && payload.models.length > 0 && payload.models.every(obj => obj.Hidden === true)) {
		return
	}
	%s
	`, stateful.ReloadAction(ctx, c, nil).Go())
	return stateful.Actionable(ctx, c,
		web.Listen(
			presets.NotifModelsCreated(&ActivityLog{}), reloadAction,
			presets.NotifModelsUpdated(&ActivityLog{}), reloadAction,
			presets.NotifModelsDeleted(&ActivityLog{}), reloadAction,
		),
		web.Scope().VSlot("{locals: toplocals, form}").Init(`{ deletingLogID: -1 }`).Children(
			v.VDialog().MaxWidth("520px").
				Attr(":model-value", `toplocals.deletingLogID !== -1`).
				Attr("@update:model-value", `(value) => { toplocals.deletingLogID = value ? toplocals.deletingLogID : -1; }`).Children(
				v.VCard(
					v.VCardTitle(h.Text(msgr.DeleteNoteDialogTitle)),
					v.VCardText(h.Text(msgr.DeleteNoteDialogText)),
					v.VCardActions(
						v.VSpacer(),
						v.VBtn(msgr.Cancel).Variant(v.VariantFlat).Size(v.SizeSmall).Class("ml-2").
							Attr("@click", `toplocals.deletingLogID = -1`),
						v.VBtn(msgr.Delete).Color(v.ColorError).Variant(v.VariantTonal).Size(v.SizeSmall).
							Attr("@click", stateful.PostAction(ctx, c,
								c.DeleteNote, DeleteNoteRequest{},
								stateful.WithAppendFix(`v.request.log_id = toplocals.deletingLogID`),
							).Go()),
					),
				),
			),
			h.Div().Class("d-flex flex-column mb-8").Style("text-body-2").Children(
				children...,
			),
		),
	).MarshalHTML(ctx)
}

type CreateNoteRequest struct {
	Note string `json:"note"`
}

func (c *TimelineCompo) CreateNote(ctx context.Context, req CreateNoteRequest) (r web.EventResponse, _ error) {
	if c.ModelName == "" || c.ModelKeys == "" {
		return r, perm.PermissionDenied
	}

	evCtx, msgr := c.MustGetEventContext(ctx)
	if c.mb.Info().Verifier().Do(presets.PermGet).WithReq(evCtx.R).IsAllowed() != nil {
		return r, perm.PermissionDenied
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
		return r, perm.PermissionDenied
	}

	evCtx, msgr := c.MustGetEventContext(ctx)
	if c.mb.Info().Verifier().Do(presets.PermGet).WithReq(evCtx.R).IsAllowed() != nil {
		return r, perm.PermissionDenied
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
	if log.CreatorID != user.ID {
		presets.ShowMessage(&r, msgr.YouAreNotTheNoteCreator, v.ColorError)
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
	r.Emit(presets.NotifModelsUpdated(&ActivityLog{}), presets.PayloadModelsUpdated{
		Ids:    []string{fmt.Sprint(log.ID)},
		Models: []any{log},
	})
	return
}

type DeleteNoteRequest struct {
	LogID uint `json:"log_id"`
}

func (c *TimelineCompo) DeleteNote(ctx context.Context, req DeleteNoteRequest) (r web.EventResponse, _ error) {
	if c.ModelName == "" || c.ModelKeys == "" {
		return r, perm.PermissionDenied
	}

	evCtx, msgr := c.MustGetEventContext(ctx)
	if c.mb.Info().Verifier().Do(presets.PermGet).WithReq(evCtx.R).IsAllowed() != nil {
		return r, perm.PermissionDenied
	}

	user, err := c.ab.currentUserFunc(ctx)
	if err != nil {
		presets.ShowMessage(&r, msgr.FailedToGetCurrentUser, v.ColorError)
		return
	}

	result := c.ab.db.Where("id = ? AND creator_id = ?", req.LogID, user.ID).Delete(&ActivityLog{})
	if err := result.Error; err != nil {
		presets.ShowMessage(&r, msgr.FailedToDeleteNote, v.ColorError)
		return
	}
	if result.RowsAffected == 0 {
		presets.ShowMessage(&r, msgr.YouAreNotTheNoteCreator, v.ColorError)
		return
	}
	presets.ShowMessage(&r, msgr.SuccessfullyDeletedNote, v.ColorSuccess)
	r.Emit(presets.NotifModelsDeleted(&ActivityLog{}), presets.PayloadModelsDeleted{
		Ids: []string{fmt.Sprint(req.LogID)},
	})
	return
}
