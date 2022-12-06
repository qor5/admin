package presets

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
	"github.com/qor5/admin/presets/actions"
	"github.com/qor5/web"
	"github.com/qor5/x/perm"
)

type ModelBuilder struct {
	p                   *Builder
	model               interface{}
	primaryField        string
	modelType           reflect.Type
	menuGroupName       string
	notInMenu           bool
	menuIcon            string
	uriName             string
	defaultURLQueryFunc func(*http.Request) url.Values
	label               string
	fieldLabels         []string
	placeholders        []string
	listing             *ListingBuilder
	detailing           *DetailingBuilder
	editing             *EditingBuilder
	creating            *EditingBuilder
	writeFields         *FieldsBuilder
	hasDetailing        bool
	rightDrawerWidth    string
	link                string
	layoutConfig        *LayoutConfig
	modelInfo           *ModelInfo
	singleton           bool
	web.EventsHub
}

func NewModelBuilder(p *Builder, model interface{}) (mb *ModelBuilder) {
	mb = &ModelBuilder{p: p, model: model, primaryField: "ID"}
	mb.modelType = reflect.TypeOf(model)
	if mb.modelType.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("model %#+v must be pointer", model))
	}
	modelstr := mb.modelType.String()
	modelName := modelstr[strings.LastIndex(modelstr, ".")+1:]
	mb.label = strcase.ToCamel(inflection.Plural(modelName))
	mb.uriName = inflection.Plural(strcase.ToKebab(modelName))
	mb.modelInfo = &ModelInfo{mb: mb}
	// Be aware the uriName here is still the original struct
	mb.newListing()
	mb.newDetailing()
	mb.newEditing()

	return
}

func (mb *ModelBuilder) RightDrawerWidth(v string) *ModelBuilder {
	mb.rightDrawerWidth = v
	return mb
}

func (mb *ModelBuilder) Link(v string) *ModelBuilder {
	mb.link = v
	return mb
}

func (mb *ModelBuilder) registerDefaultEventFuncs() {
	mb.RegisterEventFunc(actions.New, mb.editing.formNew)
	mb.RegisterEventFunc(actions.Edit, mb.editing.formEdit)
	mb.RegisterEventFunc(actions.DeleteConfirmation, mb.listing.deleteConfirmation)
	mb.RegisterEventFunc(actions.Update, mb.editing.defaultUpdate)
	mb.RegisterEventFunc(actions.DoDelete, mb.editing.doDelete)
	mb.RegisterEventFunc(actions.DoBulkAction, mb.listing.doBulkAction)
	mb.RegisterEventFunc(actions.DoListingAction, mb.listing.doListingAction)
	mb.RegisterEventFunc(actions.OpenBulkActionDialog, mb.listing.openBulkActionDialog)
	mb.RegisterEventFunc(actions.OpenActionDialog, mb.listing.openActionDialog)
	mb.RegisterEventFunc(actions.Action, mb.detailing.formDrawerAction)
	mb.RegisterEventFunc(actions.DoAction, mb.detailing.doAction)
	mb.RegisterEventFunc(actions.DetailingDrawer, mb.detailing.showInDrawer)
	mb.RegisterEventFunc(actions.ReloadList, mb.listing.reloadList)
	mb.RegisterEventFunc(actions.OpenListingDialog, mb.listing.openListingDialog)
	mb.RegisterEventFunc(actions.UpdateListingDialog, mb.listing.updateListingDialog)

	// list editor
	mb.RegisterEventFunc(actions.AddRowEvent, addListItemRow(mb))
	mb.RegisterEventFunc(actions.RemoveRowEvent, removeListItemRow(mb))
	mb.RegisterEventFunc(actions.SortEvent, sortListItems(mb))
}

func (mb *ModelBuilder) NewModel() (r interface{}) {
	return reflect.New(mb.modelType.Elem()).Interface()
}

func (mb *ModelBuilder) NewModelSlice() (r interface{}) {
	return reflect.New(reflect.SliceOf(mb.modelType)).Interface()
}

func (mb *ModelBuilder) newListing() (lb *ListingBuilder) {
	mb.listing = &ListingBuilder{
		mb:            mb,
		FieldsBuilder: *mb.p.listFieldDefaults.InspectFields(mb.model),
		dialogWidth:   "1200px",
	}
	if mb.p.dataOperator != nil {
		mb.listing.SearchFunc(mb.p.dataOperator.Search)
	}

	rmb := mb.listing.RowMenu()
	rmb.RowMenuItem("Edit").ComponentFunc(editRowMenuItemFunc(mb.Info(), "", url.Values{}))
	rmb.RowMenuItem("Delete").ComponentFunc(deleteRowMenuItemFunc(mb.Info(), "", url.Values{}))
	return
}

func (mb *ModelBuilder) newEditing() (r *EditingBuilder) {
	mb.writeFields, mb.listing.searchColumns = mb.p.writeFieldDefaults.inspectFieldsAndCollectName(mb.model, reflect.TypeOf(""))
	mb.editing = &EditingBuilder{mb: mb, FieldsBuilder: *mb.writeFields}
	if mb.p.dataOperator != nil {
		mb.editing.FetchFunc(mb.p.dataOperator.Fetch)
		mb.editing.SaveFunc(mb.p.dataOperator.Save)
		mb.editing.DeleteFunc(mb.p.dataOperator.Delete)
	}
	return
}

func (mb *ModelBuilder) newDetailing() (r *DetailingBuilder) {
	mb.detailing = &DetailingBuilder{mb: mb, FieldsBuilder: *mb.p.detailFieldDefaults.InspectFields(mb.model)}
	if mb.p.dataOperator != nil {
		mb.detailing.FetchFunc(mb.p.dataOperator.Fetch)
	}
	return
}

func (mb *ModelBuilder) Info() (r *ModelInfo) {
	return mb.modelInfo
}

type ModelInfo struct {
	mb *ModelBuilder
}

func (b ModelInfo) ListingHref() string {
	return fmt.Sprintf("%s/%s", b.mb.p.prefix, b.mb.uriName)
}

func (b ModelInfo) EditingHref(id string) string {
	return fmt.Sprintf("%s/%s/%s/edit", b.mb.p.prefix, b.mb.uriName, id)
}

func (b ModelInfo) DetailingHref(id string) string {
	return fmt.Sprintf("%s/%s/%s", b.mb.p.prefix, b.mb.uriName, id)
}

func (b ModelInfo) HasDetailing() bool {
	return b.mb.hasDetailing
}

func (b ModelInfo) DetailingInDrawer() bool {
	return b.mb.detailing.drawer
}

func (b ModelInfo) PresetsPrefix() string {
	return b.mb.p.prefix
}

func (b ModelInfo) URIName() string {
	return b.mb.uriName
}

func (b ModelInfo) Label() string {
	return b.mb.label
}

func (b ModelInfo) Verifier() *perm.Verifier {
	return b.mb.p.verifier.Spawn().
		SnakeOn(b.mb.menuGroupName).
		SnakeOn(b.mb.uriName)
}

func (mb *ModelBuilder) URIName(v string) (r *ModelBuilder) {
	mb.uriName = v
	return mb
}

func (mb *ModelBuilder) DefaultURLQueryFunc(v func(*http.Request) url.Values) (r *ModelBuilder) {
	mb.defaultURLQueryFunc = v
	return mb
}

func (mb *ModelBuilder) PrimaryField(v string) (r *ModelBuilder) {
	mb.primaryField = v
	return mb
}

func (mb *ModelBuilder) InMenu(v bool) (r *ModelBuilder) {
	mb.notInMenu = !v
	return mb
}

func (mb *ModelBuilder) MenuIcon(v string) (r *ModelBuilder) {
	mb.menuIcon = v
	return mb
}

func (mb *ModelBuilder) Label(v string) (r *ModelBuilder) {
	mb.label = v
	return mb
}

func (mb *ModelBuilder) Labels(vs ...string) (r *ModelBuilder) {
	mb.fieldLabels = append(mb.fieldLabels, vs...)
	return mb
}

func (mb *ModelBuilder) LayoutConfig(v *LayoutConfig) (r *ModelBuilder) {
	mb.layoutConfig = v
	return mb
}

func (mb *ModelBuilder) Placeholders(vs ...string) (r *ModelBuilder) {
	mb.placeholders = append(mb.placeholders, vs...)
	return mb
}

func (mb *ModelBuilder) Singleton(v bool) (r *ModelBuilder) {
	mb.singleton = v
	return mb
}

func (mb *ModelBuilder) getComponentFuncField(field *FieldBuilder) (r *FieldContext) {
	r = &FieldContext{
		ModelInfo: mb.Info(),
		Name:      field.name,
		Label:     mb.getLabel(field.NameLabel),
	}
	return
}

func (mb *ModelBuilder) getLabel(field NameLabel) (r string) {
	if len(field.label) > 0 {
		return field.label
	}

	for i := 0; i < len(mb.fieldLabels)-1; i = i + 2 {
		if mb.fieldLabels[i] == field.name {
			return mb.fieldLabels[i+1]
		}
	}

	return humanizeString(field.name)
}
