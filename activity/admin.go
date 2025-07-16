package activity

import (
	"cmp"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/perm"
	. "github.com/qor5/x/v3/ui/vuetify"
	"github.com/qor5/x/v3/ui/vuetifyx"
	"github.com/samber/lo"
	"github.com/spf13/cast"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
)

const (
	I18nActivityKey i18n.ModuleKey = "I18nActivityKey"

	paramHideDetailTop = "hideDetailTop"
)

func (ab *Builder) Install(b *presets.Builder) error {
	ab.mu.Lock()
	defer ab.mu.Unlock()
	lmb, ok := ab.logModelBuilders[b]
	if ok && lmb != nil {
		return errors.Errorf("activity: preset %q already installed", b.GetURIPrefix())
	}

	b.GetI18n().
		RegisterForModule(language.English, I18nActivityKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nActivityKey, Messages_zh_CN).
		RegisterForModule(language.Japanese, I18nActivityKey, Messages_ja_JP)

	if permB := b.GetPermission(); permB != nil {
		permB.CreatePolicies(ab.permPolicy)
	}

	lmb = b.Model(&ActivityLog{}).MenuIcon("mdi-book-edit")
	err := ab.logModelInstall(b, lmb)
	if err != nil {
		return err
	}

	ab.logModelBuilders[b] = lmb
	return err
}

func (ab *Builder) GetLogModelBuilder(pb *presets.Builder) *presets.ModelBuilder {
	ab.mu.RLock()
	defer ab.mu.RUnlock()
	val, ok := ab.logModelBuilders[pb]
	if !ok {
		return nil
	}
	return val
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
	setupDetailing(b, dp, op, ab)
	setupEditing(eb)
	setupListing(b, lb, op, ab)

	return nil
}

func setupEditing(eb *presets.EditingBuilder) {
	eb.SaveFunc(func(obj any, id string, ctx *web.EventContext) error {
		return errors.New("should not be used")
	})
	eb.DeleteFunc(func(obj any, id string, ctx *web.EventContext) error {
		return errors.New("should not be used")
	})
}

func setupListing(b *presets.Builder, lb *presets.ListingBuilder, op *gorm2op.DataOperatorBuilder, ab *Builder) {
	lb.RelayPagination(gorm2op.KeysetBasedPagination(true)).KeywordSearchOff(true)
	lb.SearchFunc(func(ctx *web.EventContext, params *presets.SearchParams) (result *presets.SearchResult, err error) {
		if !ab.skipResPermCheck {
			var modelLabels []string
			// err = ab.db.Model(&ActivityLog{}).Select("DISTINCT model_label AS model_label").Pluck("model_label", &modelLabels).Error
			// if err != nil {
			// 	return nil, err
			// }
			for _, m := range ab.models {
				if m.label != nil {
					modelLabels = append(modelLabels, m.label())
				}
			}
			signsNoPerm := []string{}
			modelLabels = lo.Uniq(modelLabels)
			for _, resourceSign := range modelLabels {
				if resourceSign == "" || resourceSign == NopModelLabel {
					continue
				}
				if b.GetVerifier().Spawn().SnakeOn(resourceSign).Do(presets.PermList).WithReq(ctx.R).IsAllowed() == nil {
					continue
				}
				signsNoPerm = append(signsNoPerm, resourceSign)
			}
			if len(signsNoPerm) > 0 {
				params.SQLConditions = append(params.SQLConditions, &presets.SQLCondition{
					Query: "model_label NOT IN ?",
					Args:  []any{signsNoPerm},
				})
			}
		}

		params.SQLConditions = append(params.SQLConditions, &presets.SQLCondition{
			Query: "hidden = ?",
			Args:  []any{false},
		})
		result, err = op.Search(ctx, params)
		if err != nil {
			return nil, err
		}
		logs := result.Nodes.([]*ActivityLog)
		if err := ab.supplyUsers(ctx.R.Context(), logs); err != nil {
			return nil, err
		}
		return result, nil
	})

	lb.WrapColumns(presets.CustomizeColumnLabel(func(evCtx *web.EventContext) (map[string]string, error) {
		msgr := i18n.MustGetModuleMessages(evCtx.R, I18nActivityKey, Messages_en_US).(*Messages)
		return map[string]string{
			"CreatedAt":  msgr.ModelCreatedAt,
			"User":       msgr.ModelUser,
			"Action":     msgr.ModelAction,
			"ModelKeys":  msgr.ModelKeys,
			"ModelLabel": msgr.ModelLabel,
			"ModelName":  msgr.ModelName,
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
	lb.Field("Action").Label(Messages_en_US.ModelAction).ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		action := obj.(*ActivityLog).Action
		label := getActionLabel(ctx, action)
		return h.Td(h.Div().Attr("v-pre", true).Text(label))
	})
	lb.Field("ModelKeys").Label(Messages_en_US.ModelKeys)
	lb.Field("ModelLabel").Label(Messages_en_US.ModelLabel).ComponentFunc(
		func(obj any, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			if obj.(*ActivityLog).ModelLabel == "" {
				return h.Td(h.Text(NopModelLabel))
			}
			return h.Td(h.Div().Attr("v-pre", true).Text(obj.(*ActivityLog).ModelLabel))
		},
	)
	lb.Field("ModelName").Label(Messages_en_US.ModelName).ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return h.Td(h.Div().Attr("v-pre", true).Text(i18n.T(ctx.R, presets.ModelsI18nModuleKey, obj.(*ActivityLog).ModelName)))
	})

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
		var actions []string
		err := ab.db.Model(&ActivityLog{}).Select("DISTINCT action AS action").Pluck("action", &actions).Error
		if err != nil {
			panic(err)
		}

		for _, action := range actions {
			if action == ActionLastView {
				continue
			}
			label := actionLabels[action]
			if label == "" {
				label = i18n.PT(ctx.R, presets.ModelsI18nModuleKey, I18nActionLabelPrefix, action)
			}
			actionOptions = append(actionOptions, &vuetifyx.SelectItem{
				Text:  label,
				Value: action,
			})
		}
		actionOptions = lo.UniqBy(actionOptions, func(item *vuetifyx.SelectItem) string { return item.Value })

		var userIDs []string
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

		var modelNames []string
		err = ab.db.Model(&ActivityLog{}).Select("DISTINCT model_name AS model_name").Pluck("model_name", &modelNames).Error
		if err != nil {
			panic(err)
		}
		var modelNameOptions []*vuetifyx.SelectItem
		for _, modelName := range modelNames {
			modelNameOptions = append(modelNameOptions, &vuetifyx.SelectItem{
				Text:  i18n.T(ctx.R, presets.ModelsI18nModuleKey, modelName),
				Value: modelName,
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
				ItemType:     vuetifyx.ItemTypeDatetimeRangePicker,
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
		if len(modelNameOptions) > 0 {
			filterData = append(
				filterData,
				&vuetifyx.FilterItem{
					Key:          "model_name",
					Label:        msgr.FilterModel,
					ItemType:     vuetifyx.ItemTypeSelect,
					SQLCondition: `model_name %s ?`,
					Options:      modelNameOptions,
				},
				&vuetifyx.FilterItem{
					Key:          "model_keys",
					Label:        msgr.FilterModelKeys,
					ItemType:     vuetifyx.ItemTypeString,
					SQLCondition: `model_keys %s ?`,
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
}

func setupDetailing(b *presets.Builder, dp *presets.DetailingBuilder, op *gorm2op.DataOperatorBuilder, ab *Builder) {
	dp.FetchFunc(func(obj any, id string, ctx *web.EventContext) (r any, err error) {
		r, err = op.Fetch(obj, id, ctx)
		if err != nil {
			return r, err
		}
		log := r.(*ActivityLog)

		if !ab.skipResPermCheck {
			if log.ModelLabel != "" && log.ModelLabel != NopModelLabel {
				if b.GetVerifier().Spawn().SnakeOn(log.ModelLabel).Do(presets.PermGet).WithReq(ctx.R).IsAllowed() != nil {
					return nil, perm.PermissionDenied
				}
			}
		}

		if err := ab.supplyUsers(ctx.R.Context(), []*ActivityLog{log}); err != nil {
			return nil, err
		}
		return log, nil
	})

	dp.Field("Detail").ComponentFunc(
		func(obj any, field *presets.FieldContext, ctx *web.EventContext) (r h.HTMLComponent) {
			var (
				log           = obj.(*ActivityLog)
				msgr          = i18n.MustGetModuleMessages(ctx.R, I18nActivityKey, Messages_en_US).(*Messages)
				pmsgr         = presets.MustGetMessages(ctx.R)
				hideDetailTop = cast.ToBool(ctx.R.Form.Get(paramHideDetailTop))
				actionLabel   = getActionLabel(ctx, log.Action)
				hasPermGET    = true
			)
			if log.ModelLabel != "" && log.ModelLabel != NopModelLabel {
				if b.GetVerifier().Spawn().SnakeOn(log.ModelLabel).Do(presets.PermGet).WithReq(ctx.R).IsAllowed() != nil {
					hasPermGET = false
				}
			}
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
								h.Tr(h.Td(h.Text(msgr.ModelAction)), h.Td().Attr("v-pre", true).Text(actionLabel)),
								h.Tr(h.Td(h.Text(msgr.ModelName)), h.Td().Attr("v-pre", true).Text(
									i18n.T(ctx.R, presets.ModelsI18nModuleKey, obj.(*ActivityLog).ModelName),
								)),
								h.Tr(h.Td(h.Text(msgr.ModelLabel)), h.Td().Attr("v-pre", true).Text(cmp.Or(log.ModelLabel, NopModelLabel))),
								h.Tr(h.Td(h.Text(msgr.ModelKeys)), h.Td().Attr("v-pre", true).Text(log.ModelKeys)),
								h.Iff(hasPermGET && log.ModelLink != "", func() h.HTMLComponent {
									onClick := web.Plaid().PushStateURL(log.ModelLink).Go()
									href := log.ModelLink
									return h.Tr(h.Td(h.Text(msgr.ModelLink)), h.Td(
										h.A(
											VBtn(msgr.MoreInfo).Class("text-none text-overline d-flex align-center").
												Variant(VariantTonal).Color(ColorPrimary).Size(SizeXSmall).PrependIcon("mdi-open-in-new"),
										).Href(href).Attr("@click", fmt.Sprintf(`(e) => {
											if (e.metaKey || e.ctrlKey) { return; }
											e.stopPropagation();
											e.preventDefault();
											%s;
										}`, onClick)),
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
								h.Div().Class("text-body-2").Style("white-space: pre-wrap").Text(fmt.Sprintf(`{{%q}}`, note.Note)),
								h.Iff(!note.LastEditedAt.IsZero(), func() h.HTMLComponent {
									return h.Div().Class("text-caption font-italic").Style("color: #757575").Children(
										h.Text(msgr.LastEditedAt(pmsgr.HumanizeTime(note.LastEditedAt))),
									)
								}),
							),
						),
					))
				default:
					children = append(children, h.Div().Attr("v-pre", true).Text(msgr.PerformAction(actionLabel, log.Detail)))
				}
			}

			return h.Div().Class("d-flex flex-column ga-3").Children(children...)
		},
	)
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
		tableHeaders := []map[string]any{
			{"title": msgr.DiffField, "key": "Field", "sortable": false, "width": "40%"},
			{"title": msgr.DiffValue, "key": "New", "sortable": false, "width": "60%"},
		}

		diffsElems = append(diffsElems,
			VCard().Elevation(0).Children(
				VCardTitle().Class("pa-0").Children(h.Text(msgr.DiffAdd)),
				VCardText().Class("pa-0 pt-3").Children(
					VDataTable().ItemsPerPage(-1).HideDefaultFooter(true).Headers(tableHeaders).Items(addediffs),
				),
			))
	}

	if len(deletediffs) > 0 {
		tableHeaders := []map[string]any{
			{"title": msgr.DiffField, "key": "Field", "sortable": false, "width": "40%"},
			{"title": msgr.DiffValue, "key": "Old", "sortable": false, "width": "60%"},
		}
		diffsElems = append(diffsElems,
			VCard().Elevation(0).Children(
				VCardTitle().Class("pa-0").Children(h.Text(msgr.DiffDelete)),
				VCardText().Class("pa-0 pt-3").Children(
					VDataTable().ItemsPerPage(-1).HideDefaultFooter(true).Headers(tableHeaders).Items(deletediffs),
				),
			))
	}

	if len(changediffs) > 0 {
		tableHeaders := []map[string]any{
			{"title": msgr.DiffField, "key": "Field", "sortable": false, "width": "20%"},
			{"title": msgr.DiffOld, "key": "Old", "sortable": false, "width": "40%"},
			{"title": msgr.DiffNew, "key": "New", "sortable": false, "width": "40%"},
		}

		diffsElems = append(diffsElems,
			VCard().Elevation(0).Children(
				VCardTitle().Class("pa-0").Children(h.Text(msgr.DiffChanges)),
				VCardText().Class("pa-0 pt-3").Children(
					VDataTable().ItemsPerPage(-1).HideDefaultFooter(true).Headers(tableHeaders).Items(changediffs),
				),
			))
	}
	return h.Components(diffsElems...)
}
