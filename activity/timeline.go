package activity

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/stateful"
	"github.com/qor5/x/v3/perm"
	v "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
)

type Note struct {
	Note         string    `json:"note"`
	LastEditedAt time.Time `json:"last_edited_at"`
}

func init() {
	stateful.RegisterActionableCompoType(&Timeline{})
}

type Timeline struct {
	ab *Builder              `inject:""`
	mb *presets.ModelBuilder `inject:""`

	ID        string `json:"id"`
	ModelName string `json:"model_name"`
	ModelKeys string `json:"model_keys"`
	ModelLink string `json:"model_link"`
}

func (c *Timeline) CompoID() string {
	return fmt.Sprintf("Timeline:%s", c.ID)
}

func (c *Timeline) HumanContent(ctx context.Context, log *ActivityLog) h.HTMLComponent {
	// TODO: i18n
	switch log.Action {
	case ActionNote:
		note := &Note{}
		if err := json.Unmarshal([]byte(log.Detail), note); err != nil {
			return h.Text(fmt.Sprintf("Failed to unmarshal detail: %v", err))
		}
		return h.Components(
			h.Div().Attr("v-if", "!xlocals.showEditBox").Class("d-flex flex-column").Children(
				h.Text("Added a note :"),
				h.Pre(note.Note).Style("white-space: pre-wrap"),
				h.Iff(!note.LastEditedAt.IsZero(), func() h.HTMLComponent {
					return h.Div().Class("text-caption font-italic").Style("color: #757575").Children(
						h.Text(fmt.Sprintf("(edited at %s)", humanize.Time(note.LastEditedAt))),
					)
				}),
			),
			h.Div().Attr("v-if", "!!xlocals.showEditBox").Class("flex-grow-1 d-flex flex-column mt-4").Style("position: relative").Children(
				v.VTextarea().Rows(3).Attr("row-height", "12").Clearable(false).AutoGrow(true).Label("").Variant(v.VariantOutlined).
					Attr(web.VField("note", note.Note)...),
				h.Div().Class("d-flex flex-row ga-1").Style("position: absolute; top: 6px; right: 6px").Children(
					// TODO: i18n
					v.VBtn("").Variant(v.VariantText).Color("grey-darken-3").Size(v.SizeXSmall).Icon("mdi-close").
						Attr("@click", "xlocals.showEditBox = false"),
					// TODO: i18n
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
		return h.Text("Viewed")
	case ActionCreate:
		return h.Text("Created")
	case ActionEdit:
		diffs := []Diff{}
		if err := json.Unmarshal([]byte(log.Detail), &diffs); err != nil {
			return h.Text(fmt.Sprintf("Failed to unmarshal detail: %v", err))
		}
		return h.Div().Class("d-flex flex-row align-center ga-2").Children(
			h.Text(fmt.Sprintf("Edited %d fields ", len(diffs))),
			v.VBtn("More Info").Class("text-none text-overline d-flex align-center").
				Variant(v.VariantTonal).Color(v.ColorPrimary).Size(v.SizeXSmall).PrependIcon("mdi-open-in-new").
				Attr("@click", web.POST().
					EventFunc(actions.DetailingDrawer).
					Query(presets.ParamOverlay, actions.Dialog).
					URL(fmt.Sprintf("%s/activity-logs/%d", c.mb.GetPresetsBuilder().GetURIPrefix(), log.ID)).
					Query(paramHideModelLink, true).
					Go(),
				),
		)
	case ActionDelete:
		return h.Text("Deleted")
	default:
		return h.Text(fmt.Sprintf("Performed action %q with detail %v ", log.Action, log.Detail))
	}
}

func (c *Timeline) MarshalHTML(ctx context.Context) ([]byte, error) {
	// TODO:
	// evCtx := web.MustGetEventContext(ctx)
	// utilMsgr := i18n.MustGetModuleMessages(evCtx.R, utils.I18nUtilsKey, Messages_en_US).(*utils.Messages)

	children := []h.HTMLComponent{
		// TODO: i18n
		h.Div().Class("text-h6 mb-8").Text("Activity"),
		web.Scope().VSlot("{locals: xlocals,form}").Init("{showEditBox:false}").Children(
			v.VBtn("Add Note").Attr("v-if", "!xlocals.showEditBox").
				Class("text-none mb-4").Variant(v.VariantTonal).Color("grey-darken-3").Size(v.SizeDefault).PrependIcon("mdi-plus").
				Attr("@click", "xlocals.showEditBox = true"),
			h.Div().Attr("v-if", "!!xlocals.showEditBox").Class("d-flex flex-column").Style("position: relative").Children(
				// TODO: i18n
				v.VTextarea().Rows(3).Attr("row-height", "12").Clearable(false).AutoGrow(true).Label("Add Note").Variant(v.VariantOutlined).
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
			creatorName = "Unknown" // TODO: i18n
		}
		avatarText := ""
		if log.Creator.Avatar == "" {
			avatarText = strings.ToUpper(creatorName[0:1])
		}
		// TODO: v.ColorXXX ?
		dotColor := "#30a46c"
		if i != 0 {
			dotColor = "#e0e0e0"
		}
		var child h.HTMLComponent = h.Div().Class("d-flex flex-column ga-1").Children(
			h.Div().Class("d-flex flex-row align-center ga-2").Children(
				h.Div().Style("width: 8px; height: 8px; background-color: "+dotColor).Class("rounded-circle"),
				h.Div(h.Text(humanize.Time(log.CreatedAt))).Style("color: #757575"),
			),
			h.Div().Class("d-flex flex-row ga-2").Children(
				h.Div().Class("align-self-stretch").Style("background-color: "+dotColor+"; width: 1px; margin: -6px 3.5px -2px 3.5px;"),
				h.Div().Class("flex-grow-1 d-flex flex-column pb-3").Children(
					h.Div().Class("d-flex flex-row align-center ga-2").Children(
						v.VAvatar().Class("text-overline").Attr("style", "color: #3e63dd").Attr("color", "#E6EDFE").Attr("size", "x-small").Attr("density", "compact").Attr("rounded", true).Text(avatarText).Children(
							h.Iff(log.Creator.Avatar != "", func() h.HTMLComponent {
								return v.VImg().Attr("alt", creatorName).Attr("src", log.Creator.Avatar)
							}),
						),
						h.Div(h.Text(creatorName)).Class("font-weight-medium"),
					),
					h.Div().Class("d-flex flex-row align-center ga-2").Children(
						h.Div().Style("width: 16px"),
						c.HumanContent(ctx, log),
					),
				),
			),
		)
		if log.Action == ActionNote {
			child = web.Scope().VSlot("{locals: xlocals, form}").Init("{showEditBox:false}").Children(
				v.VHover().Disabled(log.CreatorID != c.ab.currentUserFunc(ctx).ID).Children(
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
	return stateful.Actionable(ctx, c,
		web.Listen(
			presets.NotifModelsCreated(&ActivityLog{}), stateful.ReloadAction(ctx, c, nil).Go(),
			presets.NotifModelsUpdated(&ActivityLog{}), stateful.ReloadAction(ctx, c, nil).Go(),
			presets.NotifModelsDeleted(&ActivityLog{}), stateful.ReloadAction(ctx, c, nil).Go(),
		),
		web.Scope().VSlot("{locals: toplocals, form}").Init(`{ deletingLogID: -1 }`).Children(
			v.VDialog().MaxWidth("520px").
				Attr(":model-value", `toplocals.deletingLogID !== -1`).
				Attr("@update:model-value", `(value) => { toplocals.deletingLogID = value ? toplocals.deletingLogID : -1; }`).Children(
				v.VCard(
					v.VCardTitle(h.Text("Delete note")),                               // TODO: i18n
					v.VCardText(h.Text("Are you sure you want to delete this note?")), // TODO: i18n
					v.VCardActions(
						v.VSpacer(),
						v.VBtn("Cancel").Variant(v.VariantFlat).Size(v.SizeSmall).Class("ml-2").
							Attr("@click", `toplocals.deletingLogID = -1`),
						v.VBtn("Delete").Color(v.ColorError).Variant(v.VariantTonal).Size(v.SizeSmall).
							Attr("@click", stateful.PostAction(ctx, c,
								c.DeleteNote, DeleteNoteRequest{},
								stateful.WithAppendFix(`v.request.log_id = toplocals.deletingLogID`),
							).Go()),
					),
				),
			),
			h.Div().Class("d-flex flex-column").Style("text-body-2").Children(
				children...,
			),
		),
	).MarshalHTML(ctx)
}

type CreateNoteRequest struct {
	Note string `json:"note"`
}

func (c *Timeline) CreateNote(ctx context.Context, req CreateNoteRequest) (r web.EventResponse, _ error) {
	if c.ModelName == "" || c.ModelKeys == "" {
		return r, perm.PermissionDenied
	}

	req.Note = strings.TrimSpace(req.Note)
	if req.Note == "" {
		// TODO: field error ?
		presets.ShowMessage(&r, "Note cannot be blank", "error") // TODO: i18n
		return
	}

	// TODO: 需要单独封装方法供外界显式调用
	log, err := c.ab.MustGetModelBuilder(c.mb).create(ctx, ActionNote, c.ModelName, c.ModelKeys, c.ModelLink, &Note{
		Note: req.Note,
	})
	if err != nil {
		presets.ShowMessage(&r, "Failed to add note", "error") // TODO: i18n
		return
	}

	presets.ShowMessage(&r, "Successfully added note", "") // TODO: i18n
	r.Emit(presets.NotifModelsCreated(&ActivityLog{}), presets.PayloadModelsCreated{
		Models: []any{log},
	})
	return
}

type UpdateNoteRequest struct {
	LogID uint   `json:"log_id"`
	Note  string `json:"note"`
}

func (c *Timeline) UpdateNote(ctx context.Context, req UpdateNoteRequest) (r web.EventResponse, _ error) {
	if c.ModelName == "" || c.ModelKeys == "" {
		return r, perm.PermissionDenied
	}

	req.Note = strings.TrimSpace(req.Note)
	if req.Note == "" {
		// TODO: field error ?
		presets.ShowMessage(&r, "Note cannot be blank", "error") // TODO: i18n
		return
	}

	creator := c.ab.currentUserFunc(ctx)
	if creator == nil {
		presets.ShowMessage(&r, "Failed to get current user", "error") // TODO: i18n
		return
	}

	// TODO: 需要单独封装方法供外界显式调用
	log := &ActivityLog{}
	if err := c.ab.db.Where("id = ?", req.LogID).First(log).Error; err != nil {
		presets.ShowMessage(&r, "Failed to get note", "error") // TODO: i18n
		return
	}
	if log.CreatorID != creator.ID {
		presets.ShowMessage(&r, "You are not the creator of this note", "error") // TODO: i18n
		return
	}

	note := &Note{}
	if err := json.Unmarshal([]byte(log.Detail), note); err != nil {
		presets.ShowMessage(&r, "Failed to unmarshal note", "error") // TODO: i18n
		return
	}
	if note.Note == req.Note {
		stateful.AppendReloadToResponse(&r, c)
		return
	}

	log.Detail = h.JSONString(&Note{
		Note:         req.Note,
		LastEditedAt: time.Now(),
	})
	if err := c.ab.db.Save(log).Error; err != nil {
		presets.ShowMessage(&r, "Failed to update note", "error") // TODO: i18n
		return
	}

	presets.ShowMessage(&r, "Successfully updated note", "") // TODO: i18n
	r.Emit(presets.NotifModelsUpdated(&ActivityLog{}), presets.PayloadModelsUpdated{
		Ids:    []string{fmt.Sprint(log.ID)},
		Models: []any{log},
	})
	return
}

type DeleteNoteRequest struct {
	LogID uint `json:"log_id"`
}

func (c *Timeline) DeleteNote(ctx context.Context, req DeleteNoteRequest) (r web.EventResponse, _ error) {
	if c.ModelName == "" || c.ModelKeys == "" {
		return r, perm.PermissionDenied
	}

	creator := c.ab.currentUserFunc(ctx)
	if creator == nil {
		presets.ShowMessage(&r, "Failed to get current user", "error") // TODO: i18n
		return
	}

	// TODO: 需要一个弹窗确认？
	// TODO: 需要单独封装方法供外界显式调用
	result := c.ab.db.Where("id = ? AND creator_id = ?", req.LogID, creator.ID).Delete(&ActivityLog{})
	if err := result.Error; err != nil {
		presets.ShowMessage(&r, "Failed to delete note", "error") // TODO: i18n
		return
	}
	if result.RowsAffected == 0 {
		presets.ShowMessage(&r, "You are not the creator of this note", "error") // TODO: i18n
		return
	}
	presets.ShowMessage(&r, "Successfully deleted note", "") // TODO: i18n
	// TODO: PayloadModelsDeleted 还是应该存在删除前的内容？否则一些地方无法直接细粒度更新
	r.Emit(presets.NotifModelsDeleted(&ActivityLog{}), presets.PayloadModelsDeleted{
		Ids: []string{fmt.Sprint(req.LogID)},
	})
	return
}
