package presets

const defaultBulkActionDialogWidth = "600"

type BulkActionBuilder struct {
	NameLabel

	buttonCompFunc                 ComponentFunc
	updateFunc                     BulkActionUpdateFunc
	compFunc                       BulkActionComponentFunc
	selectedIdsProcessorFunc       BulkActionSelectedIdsProcessorFunc
	selectedIdsProcessorNoticeFunc BulkActionSelectedIdsProcessorNoticeFunc

	dialogWidth string
	buttonColor string
}

func getBulkAction(actions []*BulkActionBuilder, name string) *BulkActionBuilder {
	for _, f := range actions {
		if f.name == name {
			return f
		}
	}
	return nil
}

func (b *BulkActionBuilder) Label(v string) (r *BulkActionBuilder) {
	b.label = v
	return b
}

func (b *BulkActionBuilder) ButtonCompFunc(v ComponentFunc) (r *BulkActionBuilder) {
	b.buttonCompFunc = v
	return b
}

func (b *BulkActionBuilder) UpdateFunc(v BulkActionUpdateFunc) (r *BulkActionBuilder) {
	b.updateFunc = v
	return b
}

func (b *BulkActionBuilder) ComponentFunc(v BulkActionComponentFunc) (r *BulkActionBuilder) {
	b.compFunc = v
	return b
}

func (b *BulkActionBuilder) SelectedIdsProcessorFunc(v BulkActionSelectedIdsProcessorFunc) (r *BulkActionBuilder) {
	b.selectedIdsProcessorFunc = v
	return b
}

func (b *BulkActionBuilder) SelectedIdsProcessorNoticeFunc(v BulkActionSelectedIdsProcessorNoticeFunc) (r *BulkActionBuilder) {
	b.selectedIdsProcessorNoticeFunc = v
	return b
}

func (b *BulkActionBuilder) DialogWidth(v string) (r *BulkActionBuilder) {
	b.dialogWidth = v
	return b
}

func (b *BulkActionBuilder) ButtonColor(v string) (r *BulkActionBuilder) {
	b.buttonColor = v
	return b
}

func (b *ListingBuilder) BulkAction(name string) (r *BulkActionBuilder) {
	builder := getBulkAction(b.bulkActions, name)
	if builder != nil {
		return builder
	}

	r = &BulkActionBuilder{}
	r.name = name
	r.buttonColor = "black"
	r.dialogWidth = defaultBulkActionDialogWidth
	b.bulkActions = append(b.bulkActions, r)
	return
}
