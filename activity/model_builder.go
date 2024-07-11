package activity

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/qor5/admin/v3/presets"
	"github.com/samber/lo"
)

// @snippet_begin(ActivityModelBuilder)
// a unique model builder is consist of typ and presetModel
type ModelBuilder struct {
	typ           reflect.Type                 // model type
	ab            *Builder                     // activity builder
	presetModel   *presets.ModelBuilder        // preset model builder
	skip          uint8                        // skip the refined data operator of the presetModel
	keys          []string                     // primary keys
	ignoredFields []string                     // ignored fields
	typeHandlers  map[reflect.Type]TypeHandler // type handlers
	link          func(any) string             // display the model link on the admin detail page
}

// @snippet_end

// AddKeys add keys to the model builder
func (mb *ModelBuilder) AddKeys(keys ...string) *ModelBuilder {
	mb.keys = lo.Uniq(append(mb.keys, keys...))
	return mb
}

// Keys set keys for the model builder
func (mb *ModelBuilder) Keys(keys ...string) *ModelBuilder {
	mb.keys = keys
	return mb
}

// LinkFunc set the link that linked to the modified record
func (mb *ModelBuilder) LinkFunc(f func(any) string) *ModelBuilder {
	mb.link = f
	return mb
}

// SkipCreate skip the created action for preset.ModelBuilder
func (mb *ModelBuilder) SkipCreate() *ModelBuilder {
	if mb.presetModel == nil {
		panic("SkipCreate method only supports presets.ModelBuilder")
	}

	if mb.skip&Create == 0 {
		mb.skip |= Create
	}
	return mb
}

// SkipUpdate skip the update action for preset.ModelBuilder
func (mb *ModelBuilder) SkipUpdate() *ModelBuilder {
	if mb.presetModel == nil {
		panic("SkipUpdate method only supports presets.ModelBuilder")
	}

	if mb.skip&Update == 0 {
		mb.skip |= Update
	}
	return mb
}

// SkipDelete skip the delete action for preset.ModelBuilder
func (mb *ModelBuilder) SkipDelete() *ModelBuilder {
	if mb.presetModel == nil {
		panic("SkipDelete method only supports presets.ModelBuilder")
	}

	if mb.skip&Delete == 0 {
		mb.skip |= Delete
	}
	return mb
}

// AddIgnoredFields append ignored fields to the default ignored fields, this would not overwrite the default ignored fields
func (mb *ModelBuilder) AddIgnoredFields(fields ...string) *ModelBuilder {
	mb.ignoredFields = lo.Uniq(append(mb.ignoredFields, fields...))
	return mb
}

// IgnoredFields set ignored fields to replace the default ignored fields with the new set.
func (mb *ModelBuilder) IgnoredFields(fields ...string) *ModelBuilder {
	mb.ignoredFields = fields
	return mb
}

// AddTypeHanders add type handers for the model builder
func (mb *ModelBuilder) AddTypeHanders(v any, f TypeHandler) *ModelBuilder {
	if mb.typeHandlers == nil {
		mb.typeHandlers = map[reflect.Type]TypeHandler{}
	}
	mb.typeHandlers[reflect.Indirect(reflect.ValueOf(v)).Type()] = f
	return mb
}

// KeysValue get model keys value
func (mb *ModelBuilder) KeysValue(v any) string {
	var (
		stringBuilder = strings.Builder{}
		reflectValue  = reflect.Indirect(reflect.ValueOf(v))
		reflectType   = reflectValue.Type()
	)

	for _, key := range mb.keys {
		if fields, ok := reflectType.FieldByName(key); ok {
			if reflectValue.FieldByName(key).IsZero() {
				continue
			}
			if fields.Anonymous { // TODO: 为什么匿名的需要多一个 FiledByName 呢？
				stringBuilder.WriteString(fmt.Sprintf("%v:", reflectValue.FieldByName(key).FieldByName(key).Interface()))
			} else {
				stringBuilder.WriteString(fmt.Sprintf("%v:", reflectValue.FieldByName(key).Interface()))
			}
		}
	}

	return strings.TrimRight(stringBuilder.String(), ":")
}

// AddRecords add records log
func (mb *ModelBuilder) AddRecords(ctx context.Context, action string, vs ...any) error {
	if len(vs) == 0 {
		return errors.New("data are empty")
	}

	creator := mb.ab.currentUserFunc(ctx)

	switch action {
	case ActionView:
		for _, v := range vs {
			err := mb.AddViewRecord(creator, v)
			if err != nil {
				return err
			}
		}

	case ActionDelete:
		for _, v := range vs {
			err := mb.AddDeleteRecord(creator, v)
			if err != nil {
				return err
			}
		}
	case ActionCreate:
		for _, v := range vs {
			err := mb.AddCreateRecord(creator, v)
			if err != nil {
				return err
			}
		}
	case ActionEdit:
		for _, v := range vs {
			err := mb.AddEditRecord(creator, v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// AddCustomizedRecord add customized record
func (mb *ModelBuilder) AddCustomizedRecord(ctx context.Context, action string, diff bool, obj any) error {
	creator := mb.ab.currentUserFunc(ctx)
	if !diff {
		return mb.save(creator, action, obj, "")
	}

	old, ok := fetchOld(obj, mb.ab.db)
	if !ok {
		return fmt.Errorf("can't find old data for %+v ", obj)
	}
	return mb.addDiff(action, creator, old, obj)
}

// AddViewRecord add view record
func (mb *ModelBuilder) AddViewRecord(creator *User, v any) error {
	return mb.save(creator, ActionView, v, "")
}

// AddDeleteRecord	add delete record
func (mb *ModelBuilder) AddDeleteRecord(creator *User, v any) error {
	return mb.save(creator, ActionDelete, v, "")
}

// AddSaverRecord will save a create log or a edit log
func (mb *ModelBuilder) AddSaveRecord(creator *User, new any) error {
	old, ok := fetchOld(new, mb.ab.db)
	if !ok {
		return mb.AddCreateRecord(creator, new)
	}
	return mb.AddEditRecordWithOld(creator, old, new)
}

// AddCreateRecord add create record
func (mb *ModelBuilder) AddCreateRecord(creator *User, v any) error {
	return mb.save(creator, ActionCreate, v, "")
}

// AddEditRecord add edit record
func (mb *ModelBuilder) AddEditRecord(creator *User, new any) error {
	old, ok := fetchOld(new, mb.ab.db)
	if !ok {
		return fmt.Errorf("can't find old data for %+v ", new)
	}
	return mb.AddEditRecordWithOld(creator, old, new)
}

// AddEditRecord add edit record
func (mb *ModelBuilder) AddEditRecordWithOld(creator *User, old, new any) error {
	return mb.addDiff(ActionEdit, creator, old, new)
}

func (mb *ModelBuilder) addDiff(action string, creator *User, old, new any) error {
	diffs, err := mb.Diff(old, new)
	if err != nil {
		return err
	}

	if len(diffs) == 0 {
		return nil
	}

	b, err := json.Marshal(diffs)
	if err != nil {
		return err
	}

	return mb.save(creator, action, new, string(b))
}

// Diff get diffs between old and new value
func (mb *ModelBuilder) Diff(old, new any) ([]Diff, error) {
	return NewDiffBuilder(mb).Diff(old, new)
}

func (mb *ModelBuilder) save(creator *User, action string, v any, diffs string) error {
	if mb.ab.findUsersFunc == nil { // TODO: 先简单处理吧
		user := &ActivityUser{
			Name:   creator.Name,
			Avatar: creator.Avatar,
		}
		user.ID = creator.ID
		// TODO: 这样 CreatedAt 会是空值
		if err := mb.ab.db.Save(user).Error; err != nil {
			return err
		}
	}

	log := &ActivityLog{}

	log.CreatedAt = time.Now()
	log.CreatorID = creator.ID
	log.Action = action
	log.ModelName = modelName(v)
	log.ModelKeys = mb.KeysValue(v)

	if mb.presetModel != nil && mb.presetModel.Info().URIName() != "" {
		log.ModelLabel = mb.presetModel.Info().URIName()
	} else {
		log.ModelLabel = "-"
	}

	if f := mb.link; f != nil {
		log.ModelLink = f(v)
	}

	if diffs == "" && action == ActionEdit {
		return nil
	}

	if action == ActionEdit {
		log.Detail = diffs
	}

	err := mb.ab.db.Save(log).Error
	if err != nil {
		return err
	}
	return nil
}

func (mb *ModelBuilder) createEditLog(ctx context.Context, old any, new any) (*ActivityLog, error) {
	diffs, err := mb.Diff(old, new)
	if err != nil {
		return nil, err
	}

	if len(diffs) == 0 {
		return nil, nil
	}

	return mb.createWithObj(ctx, ActionEdit, new, diffs)
}

func (mb *ModelBuilder) modelLink(v any) string {
	if f := mb.link; f != nil {
		return f(v)
	}
	return ""
}

func (mb *ModelBuilder) createWithObj(ctx context.Context, action string, obj any, detail any) (*ActivityLog, error) {
	return mb.create(ctx, action, modelName(obj), mb.KeysValue(obj), mb.modelLink(obj), detail)
}

func (mb *ModelBuilder) create(
	ctx context.Context,
	action string,
	modelName, modelKeys, modelLink string,
	detail any,
) (*ActivityLog, error) {
	creator := mb.ab.currentUserFunc(ctx)
	if creator == nil {
		return nil, errors.New("current user is nil")
	}

	if mb.ab.findUsersFunc == nil { // TODO: 先简单处理吧
		user := &ActivityUser{
			Name:   creator.Name,
			Avatar: creator.Avatar,
		}
		user.ID = creator.ID
		// TODO: 这样 CreatedAt 会是空值，回头再优化
		if err := mb.ab.db.Save(user).Error; err != nil {
			return nil, err
		}
	}

	log := &ActivityLog{
		CreatorID:  creator.ID,
		Action:     action,
		ModelName:  modelName,
		ModelKeys:  modelKeys,
		ModelLabel: "-",
		ModelLink:  modelLink,
	}
	if mb.presetModel != nil {
		log.ModelLabel = cmp.Or(mb.presetModel.Info().URIName(), log.ModelLabel)
	}
	log.CreatedAt = time.Now()

	detailJson, err := json.Marshal(detail)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal detail")
	}
	log.Detail = string(detailJson)

	if err := mb.ab.db.Create(log).Error; err != nil {
		return nil, err
	}
	return log, nil
}
