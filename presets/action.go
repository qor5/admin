package presets

type ActionBuilder struct {
	NameLabel

	// ButtonCompFunc will define the components of the button area.
	buttonCompFunc ComponentFunc
	// UpdateFunc will be executed when the button is clicked.
	updateFunc ActionUpdateFunc
	// CompFunc will define the components in dialog of button click.
	compFunc ActionComponentFunc

	dialogWidth string
	// If buttonCompFunc is not defined, use the default button style with buttonColor.
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

func (b *ActionBuilder) ButtonCompFunc(v ComponentFunc) (r *ActionBuilder) {
	b.buttonCompFunc = v
	return b
}

func (b *ActionBuilder) UpdateFunc(v ActionUpdateFunc) (r *ActionBuilder) {
	b.updateFunc = v
	return b
}

func (b *ActionBuilder) ComponentFunc(v ActionComponentFunc) (r *ActionBuilder) {
	b.compFunc = v
	return b
}

func (b *ActionBuilder) DialogWidth(v string) (r *ActionBuilder) {
	b.dialogWidth = v
	return b
}

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
