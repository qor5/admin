package presets

type ActionBuilder struct {
	NameLabel

	buttonCompFunc ComponentFunc
	updateFunc     ActionUpdateFunc
	compFunc       ActionComponentFunc

	dialogWidth string
	buttonColor string
}

func getAction(actions []*ActionBuilder, name string) *ActionBuilder {
	for _, f := range actions {
		if f.name == name {
			return f
		}
	}
	return nil
}

func (b *ActionBuilder) Label(v string) (r *ActionBuilder) {
	b.label = v
	return b
}

// ButtonCompFunc defines the components of the button area.
func (b *ActionBuilder) ButtonCompFunc(v ComponentFunc) (r *ActionBuilder) {
	b.buttonCompFunc = v
	return b
}

// UpdateFunc defines event when the button is clicked.
func (b *ActionBuilder) UpdateFunc(v ActionUpdateFunc) (r *ActionBuilder) {
	b.updateFunc = v
	return b
}

// ComponentFunc defines the components in dialog of button click.
func (b *ActionBuilder) ComponentFunc(v ActionComponentFunc) (r *ActionBuilder) {
	b.compFunc = v
	return b
}

func (b *ActionBuilder) DialogWidth(v string) (r *ActionBuilder) {
	b.dialogWidth = v
	return b
}

// ButtonColor defines the color of default button if buttonCompFunc is not defined.
func (b *ActionBuilder) ButtonColor(v string) (r *ActionBuilder) {
	b.buttonColor = v
	return b
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
