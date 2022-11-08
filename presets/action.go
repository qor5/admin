package presets

type ActionBuilder struct {
	NameLabel

	buttonCompFunc                 ComponentFunc
	updateFunc                     ActionUpdateFunc
	compFunc                       ActionComponentFunc
	selectedIdsProcessorFunc       ActionSelectedIdsProcessorFunc
	selectedIdsProcessorNoticeFunc ActionSelectedIdsProcessorNoticeFunc

	dialogWidth string
	buttonColor string
}

const defaultBulkActionDialogWidth = "600"

func (b *ListingBuilder) BulkAction(name string) (r *ActionBuilder) {
	builder := getAction(b.bulkActions, name)
	if builder != nil {
		return builder
	}
	r = &ActionBuilder{}
	r.name = name
	r.dialogWidth = defaultBulkActionDialogWidth
	b.bulkActions = append(b.bulkActions, r)
	return
}

func (b *ListingBuilder) Action(name string) (r *ActionBuilder) {
	builder := getAction(b.actions, name)
	if builder != nil {
		return builder
	}

	r = &ActionBuilder{}
	r.name = name
	b.actions = append(b.actions, r)
	return
}

func getAction(actions []*ActionBuilder, name string) *ActionBuilder {
	for _, f := range actions {
		if f.name == name {
			return f
		}
	}
	return nil
}

func (b *ActionBuilder) UpdateFunc(v ActionUpdateFunc) (r *ActionBuilder) {
	b.updateFunc = v
	return b
}

func (b *ActionBuilder) Label(v string) (r *ActionBuilder) {
	b.label = v
	return b
}

func (b *ActionBuilder) ButtonColor(v string) (r *ActionBuilder) {
	b.buttonColor = v
	return b
}

func (b *ActionBuilder) DialogWidth(v string) (r *ActionBuilder) {
	b.dialogWidth = v
	return b
}

func (b *ActionBuilder) SelectedIdsProcessorFunc(v ActionSelectedIdsProcessorFunc) (r *ActionBuilder) {
	b.selectedIdsProcessorFunc = v
	return b
}

func (b *ActionBuilder) SelectedIdsProcessorNoticeFunc(v ActionSelectedIdsProcessorNoticeFunc) (r *ActionBuilder) {
	b.selectedIdsProcessorNoticeFunc = v
	return b
}

func (b *ActionBuilder) ButtonCompFunc(v ComponentFunc) (r *ActionBuilder) {
	b.buttonCompFunc = v
	return b
}

func (b *ActionBuilder) ComponentFunc(v ActionComponentFunc) (r *ActionBuilder) {
	b.compFunc = v
	return b
}

func (b *DetailingBuilder) Action(name string) (r *ActionBuilder) {
	builder := getAction(b.actions, name)
	if builder != nil {
		return builder
	}

	r = &ActionBuilder{}
	r.name = name
	b.actions = append(b.actions, r)
	return
}
