package activity

import (
	"cmp"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/dustin/go-humanize"
	"github.com/pkg/errors"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	. "github.com/qor5/x/v3/ui/vuetify"
	v "github.com/qor5/x/v3/ui/vuetify"
	"github.com/qor5/x/v3/ui/vuetifyx"
	"github.com/samber/lo"
	"github.com/spf13/cast"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
)

const (
	I18nActivityKey    i18n.ModuleKey = "I18nActivityKey"
	paramHideDetailTop                = "hideDetailTop"
)

func (ab *Builder) Install(b *presets.Builder) error {
	if actual, loaded := ab.installedPresets.LoadOrStore(b, true); loaded && actual.(bool) {
		return errors.Errorf("activity: preset %q already installed", b.GetURIPrefix())
	}

	b.GetI18n().
		RegisterForModule(language.English, I18nActivityKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nActivityKey, Messages_zh_CN).
		RegisterForModule(language.Japanese, I18nActivityKey, Messages_ja_JP)

	if permB := b.GetPermission(); permB != nil {
		permB.CreatePolicies(ab.permPolicy)
	}

	mb := b.Model(&ActivityLog{}).MenuIcon("mdi-book-edit")
	return ab.logModelInstall(b, mb)
}

func (ab *Builder) IsPresetInstalled(pb *presets.Builder) bool {
	installed := false
	valInstalled, ok := ab.installedPresets.Load(pb)
	if ok {
		installed = valInstalled.(bool)
	}
	return installed
}

func (ab *Builder) defaultLogModelInstall(b *presets.Builder, mb *presets.ModelBuilder) error {
	var (
		lb = mb.Listing("CreatedAt", "User", "Action", "ModelKeys", "ModelLabel", "ModelName")
		dp = mb.Detailing("Detail").Drawer(true)
		eb = mb.Editing()
	)

	mb.LabelName(func(evCtx *web.EventContext, singular bool) string {
		msgr := i18n.MustGetModuleMessages(evCtx.R, I18nActivityKey, Messages_en_US).(*Messages)
		if singular {
			return msgr.ActivityLog
		}
		return msgr.ActivityLogs
	})

	// should use own DataOperator

	op := gorm2op.DataOperator(ab.db)
	dp.FetchFunc(func(obj any, id string, ctx *web.EventContext) (r any, err error) {
		r, err = op.Fetch(obj, id, ctx)
		if err != nil {
			return r, err
		}
		log := r.(*ActivityLog)
		if err := ab.supplyUsers(ctx.R.Context(), []*ActivityLog{log}); err != nil {
			return nil, err
		}
		return log, nil
	})

	eb.SaveFunc(func(obj any, id string, ctx *web.EventContext) error {
		return errors.New("should not be used")
	})
	eb.DeleteFunc(func(obj any, id string, ctx *web.EventContext) error {
		return errors.New("should not be used")
	})

	lb.SearchFunc(func(model any, params *presets.SearchParams, ctx *web.EventContext) (r any, totalCount int, err error) {
		params.SQLConditions = append(params.SQLConditions, &presets.SQLCondition{
			Query: "hidden = ?",
			Args:  []any{false},
		})
		r, totalCount, err = op.Search(model, params, ctx)
		if totalCount <= 0 {
			return
		}
		logs := r.([]*ActivityLog)
		if err := ab.supplyUsers(ctx.R.Context(), logs); err != nil {
			return nil, 0, err
		}
		return logs, totalCount, nil
	})

	// use mb.LabelName handle this now
	// lb.Title(func(evCtx *web.EventContext, style presets.ListingStyle, title string) (string, h.HTMLComponent, error) {
	// 	msgr := i18n.MustGetModuleMessages(evCtx.R, I18nActivityKey, Messages_en_US).(*Messages)
	// 	return msgr.ActivityLogs, nil, nil
	// })

	lb.WrapColumns(presets.CustomizeColumnLabel(func(evCtx *web.EventContext) (map[string]string, error) {
		msgr := i18n.MustGetModuleMessages(evCtx.R, I18nActivityKey, Messages_en_US).(*Messages)
		return map[string]string{
			"CreatedAt":  msgr.HeaderCreatedAt,
			"User":       msgr.HeaderUser,
			"Action":     msgr.HeaderAction,
			"ModelKeys":  msgr.HeaderModelKeys,
			"ModelLabel": msgr.HeaderModelLabel,
			"ModelName":  msgr.HeaderModelName,
		}, nil
	}))

	lb.NewButtonFunc(func(ctx *web.EventContext) h.HTMLComponent { return nil })
	lb.RowMenu().Empty()

	lb.Field("CreatedAt").Label(Messages_en_US.ModelCreatedAt).ComponentFunc(
		func(obj any, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return h.Td(h.Text(obj.(*ActivityLog).CreatedAt.Format(timeFormat)))
		},
	)
	lb.Field("User").Label(Messages_en_US.ModelUser).ComponentFunc(
		func(obj any, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return h.Td(h.Div().Attr("v-pre", true).Text(obj.(*ActivityLog).User.Name))
		},
	)
	lb.Field("ModelKeys").Label(Messages_en_US.ModelKeys)
	lb.Field("ModelName").Label(Messages_en_US.ModelName)
	lb.Field("ModelLabel").Label(Messages_en_US.ModelLabel).ComponentFunc(
		func(obj any, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			if obj.(*ActivityLog).ModelLabel == "" {
				return h.Td(h.Text("-"))
			}
			return h.Td(h.Div().Attr("v-pre", true).Text(obj.(*ActivityLog).ModelLabel))
		},
	)

	lb.FilterDataFunc(func(ctx *web.EventContext) vuetifyx.FilterData {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nActivityKey, Messages_en_US).(*Messages)
		actionLabels := defaultActionLabels(msgr)

		var actionOptions []*vuetifyx.SelectItem
		for _, action := range DefaultActions {
			actionOptions = append(actionOptions, &vuetifyx.SelectItem{
				Text:  cmp.Or(actionLabels[action], action),
				Value: action,
			})
		}
		actions := []string{}
		err := ab.db.Model(&ActivityLog{}).Select("DISTINCT action AS action").Pluck("action", &actions).Error
		if err != nil {
			panic(err)
		}

		for _, action := range actions {
			if action == ActionLastView {
				continue
			}
			label := i18n.PT(ctx.R, presets.ModelsI18nModuleKey, mb.Info().Label(), action)
			actionOptions = append(actionOptions, &vuetifyx.SelectItem{
				Text:  cmp.Or(label, actionLabels[action], action),
				Value: action,
			})
		}
		actionOptions = lo.UniqBy(actionOptions, func(item *vuetifyx.SelectItem) string { return item.Value })

		userIDs := []string{}
		err = ab.db.Model(&ActivityLog{}).Select("DISTINCT user_id AS id").Pluck("id", &userIDs).Error
		if err != nil {
			panic(err)
		}
		users, err := ab.findUsers(ctx.R.Context(), userIDs)
		if err != nil {
			panic(err)
		}
		var userOptions []*vuetifyx.SelectItem
		for _, user := range users {
			userOptions = append(userOptions, &vuetifyx.SelectItem{
				Text:  user.Name,
				Value: user.ID,
			})
		}

		var modelOptions []*vuetifyx.SelectItem
		for _, m := range ab.models {
			modelOptions = append(modelOptions, &vuetifyx.SelectItem{
				Text:  m.typ.Name(),
				Value: m.typ.Name(),
			})
		}

		filterData := []*vuetifyx.FilterItem{
			{
				Key:          "action",
				Label:        msgr.FilterAction,
				ItemType:     vuetifyx.ItemTypeSelect,
				SQLCondition: `action %s ?`,
				Options:      actionOptions,
			},
			{
				Key:          "created",
				Label:        msgr.FilterCreatedAt,
				ItemType:     vuetifyx.ItemTypeDatetimeRange,
				SQLCondition: `created_at %s ?`,
			},
		}
		if len(userOptions) > 0 {
			filterData = append(filterData, &vuetifyx.FilterItem{
				Key:          "user_id",
				Label:        msgr.FilterUser,
				ItemType:     vuetifyx.ItemTypeSelect,
				SQLCondition: `user_id %s ?`,
				Options:      userOptions,
			})
		}
		if len(modelOptions) > 0 {
			filterData = append(filterData, &vuetifyx.FilterItem{
				Key:          "model",
				Label:        msgr.FilterModel,
				ItemType:     vuetifyx.ItemTypeSelect,
				SQLCondition: `model_name %s ?`,
				Options:      modelOptions,
			})
		}
		return filterData
	})

	lb.FilterTabsFunc(func(ctx *web.EventContext) []*presets.FilterTab {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nActivityKey, Messages_en_US).(*Messages)
		filterTabs := []*presets.FilterTab{
			{
				Label: msgr.ActionAll,
				Query: url.Values{},
			},
		}
		actionLabels := defaultActionLabels(msgr)
		for _, action := range DefaultActions {
			filterTabs = append(filterTabs, &presets.FilterTab{
				Label: cmp.Or(actionLabels[action], action),
				Query: url.Values{"action": []string{action}},
			})
		}
		return filterTabs
	})

	// use mb.LabelName handle this now
	// dp.Title(func(evCtx *web.EventContext, obj any, style presets.DetailingStyle, title string) (string, h.HTMLComponent, error) {
	// 	msgr := i18n.MustGetModuleMessages(evCtx.R, I18nActivityKey, Messages_en_US).(*Messages)
	// 	return msgr.ActivityLog, nil, nil
	// })

	dp.Field("Detail").ComponentFunc(
		func(obj any, field *presets.FieldContext, ctx *web.EventContext) (r h.HTMLComponent) {
			var (
				log           = obj.(*ActivityLog)
				msgr          = i18n.MustGetModuleMessages(ctx.R, I18nActivityKey, Messages_en_US).(*Messages)
				hideDetailTop = cast.ToBool(ctx.R.Form.Get(paramHideDetailTop))
			)
			var children []h.HTMLComponent
			if !hideDetailTop {
				children = append(children, VCard().Elevation(0).Children(
					VCardTitle().Class("pa-0").Children(
						h.Text(" "+msgr.DiffDetail),
					),
					VCardText().Class("pa-0 pt-3").Children(
						VTable(
							h.Tbody(
								h.Tr(h.Td(h.Text(msgr.ModelUser)), h.Td().Attr("v-pre", true).Text(log.User.Name)),
								h.Tr(h.Td(h.Text(msgr.ModelUserID)), h.Td().Attr("v-pre", true).Text(log.UserID)),
								h.Tr(h.Td(h.Text(msgr.ModelAction)), h.Td().Attr("v-pre", true).Text(log.Action)),
								h.Tr(h.Td(h.Text(msgr.ModelName)), h.Td().Attr("v-pre", true).Text(log.ModelName)),
								h.Tr(h.Td(h.Text(msgr.ModelLabel)), h.Td().Attr("v-pre", true).Text(cmp.Or(log.ModelLabel, "-"))),
								h.Tr(h.Td(h.Text(msgr.ModelKeys)), h.Td().Attr("v-pre", true).Text(log.ModelKeys)),
								h.Iff(log.ModelLink != "", func() h.HTMLComponent {
									return h.Tr(h.Td(h.Text(msgr.ModelLink)), h.Td(
										v.VBtn(msgr.MoreInfo).Class("text-none text-overline d-flex align-center").
											Variant(v.VariantTonal).Color(v.ColorPrimary).Size(v.SizeXSmall).PrependIcon("mdi-open-in-new").
											Attr("@click", web.POST().
												EventFunc(actions.DetailingDrawer).
												Query(presets.ParamOverlay, actions.Dialog).
												URL(log.ModelLink).
												Go(),
											),
									))
								}),
								h.Tr(h.Td(h.Text(msgr.ModelCreatedAt)), h.Td(h.Text(log.CreatedAt.Format(timeFormat)))),
							),
						),
					),
				))
			}

			if d := field.Value(obj).(string); d != "" {
				switch log.Action {
				case ActionCreate, ActionView, ActionEdit, ActionDelete:
					children = append(children, DiffComponent(d, ctx.R))
				case ActionNote:
					note := &Note{}
					if err := json.Unmarshal([]byte(log.Detail), note); err != nil {
						panic(err)
					}
					children = append(children, VCard().Elevation(0).Children(
						VCardTitle().Class("pa-0").Children(
							h.Text(msgr.ActionNote),
						),
						VCardText().Class("mt-3 pa-3 border-thin rounded").Children(
							h.Div().Class("d-flex flex-column").Children(
								h.Pre(note.Note).Attr("v-pre", true).Style("white-space: pre-wrap"),
								h.Iff(!note.LastEditedAt.IsZero(), func() h.HTMLComponent {
									return h.Div().Class("text-caption font-italic").Style("color: #757575").Children(
										h.Text(msgr.LastEditedAt(humanize.Time(note.LastEditedAt))),
									)
								}),
							),
						),
					))
				default:
					children = append(children, h.Div().Attr("v-pre", true).Text(msgr.PerformAction(log.Action, log.Detail)))
				}
			}

			return h.Div().Class("d-flex flex-column ga-3").Children(children...)
		},
	)
	return nil
}

func DiffComponent(diffstr string, req *http.Request) h.HTMLComponent {
	var diffs []Diff
	err := json.Unmarshal([]byte(diffstr), &diffs)
	if err != nil {
		return nil
	}

	if len(diffs) == 0 {
		return nil
	}

	var (
		addediffs   []Diff
		changediffs []Diff
		deletediffs []Diff
	)

	for _, diff := range diffs {
		if diff.New == "" && diff.Old != "" {
			deletediffs = append(deletediffs, diff)
			continue
		}

		if diff.New != "" && diff.Old == "" {
			addediffs = append(addediffs, diff)
			continue
		}

		if diff.New != "" && diff.Old != "" {
			changediffs = append(changediffs, diff)
			continue
		}
	}
	var diffsElems []h.HTMLComponent
	msgr := i18n.MustGetModuleMessages(req, I18nActivityKey, Messages_en_US).(*Messages)

	if len(addediffs) > 0 {
		var elems []h.HTMLComponent
		for _, d := range addediffs {
			elems = append(elems, h.Tr(
				h.Td().Text(d.Field),
				h.Td().Attr("v-pre", true).Text(d.New),
			))
		}

		diffsElems = append(diffsElems,
			VCard().Elevation(0).Children(
				VCardTitle().Class("pa-0").Children(h.Text(msgr.DiffAdd)),
				VCardText().Class("pa-0 pt-3").Children(
					VTable(
						h.Thead(h.Tr(h.Th(msgr.DiffField), h.Th(msgr.DiffValue))),
						h.Tbody(elems...),
					),
				),
			))
	}

	if len(deletediffs) > 0 {
		var elems []h.HTMLComponent
		for _, d := range deletediffs {
			elems = append(elems, h.Tr(
				h.Td().Text(d.Field),
				h.Td().Attr("v-pre", true).Text(d.Old),
			))
		}

		diffsElems = append(diffsElems,
			VCard().Elevation(0).Children(
				VCardTitle().Class("pa-0").Children(h.Text(msgr.DiffDelete)),
				VCardText().Class("pa-0 pt-3").Children(
					VTable(
						h.Thead(h.Tr(h.Th(msgr.DiffField), h.Th(msgr.DiffValue))),
						h.Tbody(elems...),
					),
				),
			))
	}

	if len(changediffs) > 0 {
		var elems []h.HTMLComponent
		for _, d := range changediffs {
			elems = append(elems, h.Tr(
				h.Td().Text(d.Field),
				h.Td().Attr("v-pre", true).Text(d.Old),
				h.Td().Attr("v-pre", true).Text(d.New),
			))
		}

		diffsElems = append(diffsElems,
			VCard().Elevation(0).Children(
				VCardTitle().Class("pa-0").Children(h.Text(msgr.DiffChanges)),
				VCardText().Class("pa-0 pt-3").Children(
					VTable(
						h.Thead(h.Tr(h.Th(msgr.DiffField), h.Th(msgr.DiffOld), h.Th(msgr.DiffNew))),
						h.Tbody(elems...),
					),
				),
			))
	}
	return h.Components(diffsElems...)
}
