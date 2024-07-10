package activity

import (
	"context"
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/stateful"
	v "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
)

func init() {
	stateful.RegisterActionableCompoType(&Timeline{})
}

// const NotifTImelineChanged = "NotifTImelineChanged"

type Timeline struct {
	ab *Builder `inject:""`

	ModelName       string `json:"model_name"`
	ModelKeys       string `json:"model_keys"`
	ShowAddNotesBox bool   `json:"show_add_notes_box"`
}

func (c *Timeline) CompoID() string {
	return fmt.Sprintf("Timeline:%s", c.ModelKeys)
}

func (c *Timeline) MarshalHTML(ctx context.Context) ([]byte, error) {
	children := []h.HTMLComponent{
		// TODO: i18n
		web.Scope().VSlot("{locals: timelineLocals,form}").Init("{showAddNoteBox:false}").FormInit("{note:''}").Children(
			v.VBtn("Add Note").Class("text-none mb-4").
				Attr("prepend-icon", "mdi-plus").Attr("variant", "tonal").Attr("color", "grey-darken-3").
				Attr("v-if", "!timelineLocals.showAddNoteBox").
				Attr("@click", "timelineLocals.showAddNoteBox = true"),
			h.Div().Class("d-flex flex-column").
				Attr("v-if", "!!timelineLocals.showAddNoteBox").Children(
				v.VTextarea().
					Attr("clearable", true).
					Attr("label", "Add Note"). // TODO: i18n
					Attr("variant", "outlined").
					Attr(web.VField("note", "")...),
				h.Div().Class("d-flex flex-row justify-end ga-2").Children(
					// TODO: i18n // TODO: ColorXXX
					v.VBtn("Cancel").Class("text-none").Attr("size", "small").Attr("color", "grey").
						Attr("@click", "timelineLocals.showAddNoteBox = false"),
					// TODO: i18n  // TODO: ColorXXX
					v.VBtn("Submit").Class("text-none").Attr("size", "small").Attr("color", "success").
						Attr("@click", stateful.PostAction(ctx, c,
							c.AddNote, AddNoteRequest{},
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
		children = append(children,
			h.Div().Class("d-flex flex-column ga-1").Children(
				h.Div().Class("d-flex flex-row align-center ga-2").Children(
					h.Div().Style("width: 8px; height: 8px; background-color: "+dotColor).Class("rounded-circle"),
					h.Div(h.Text(humanize.Time(log.CreatedAt))).Style("color: #757575"),
				),
				h.Div().Class("d-flex flex-row ga-2").Children(
					h.Div().Class("align-self-stretch").Style("background-color: "+dotColor+"; width: 1px; margin-top: -6px; margin-bottom: -2px; margin-left: 3.5px; margin-right: 3.5px;"),
					h.Div().Class("d-flex flex-column pb-3").Children(
						h.Div().Class("d-flex flex-row align-center ga-2").Children(
							v.VAvatar().Attr("style", "font-size: 12px; color: #3e63dd").Attr("color", "#E6EDFE").Attr("size", "x-small").Attr("density", "compact").Attr("rounded", true).Text(avatarText).Children(
								h.Iff(log.Creator.Avatar != "", func() h.HTMLComponent {
									return v.VImg().Attr("alt", creatorName).Attr("src", log.Creator.Avatar)
								}),
							),
							h.Div(h.Text(creatorName)).Style("font-weight: 500"),
						),
						h.Div().Class("d-flex flex-row align-center ga-2").Children(
							h.Div().Style("width: 16px"),
							h.Div(h.Text(humanContent(log))),
						),
					),
				),
			),
		)
	}
	return stateful.Actionable(ctx, c,
		// web.Listen(NotifyTodosChanged, stateful.ReloadAction(ctx, c, nil).Go()), // TODO:
		h.Div().Class("d-flex flex-column").Style("font-size: 14px").Children(
			children...,
		),
	).MarshalHTML(ctx)
}

type AddNoteRequest struct {
	Note string `json:"note"`
}

func (c *Timeline) AddNote(ctx context.Context, req AddNoteRequest) (r web.EventResponse, err error) {
	presets.ShowMessage(&r, req.Note, "") // TODO: i18n
	c.ShowAddNotesBox = false
	stateful.AppendReloadToResponse(&r, c)
	// r.Emit(NotifTImelineChanged)
	return
}
