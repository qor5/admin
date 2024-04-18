package presets

type FooterActionBuilder struct {
	NameLabel
	buttonCompFunc ComponentFunc
}

func getFooterAction(actions []*FooterActionBuilder, name string) *FooterActionBuilder {
	for _, f := range actions {
		if f.name == name {
			return f
		}
	}
	return nil
}

func (b *FooterActionBuilder) ButtonCompFunc(v ComponentFunc) (r *FooterActionBuilder) {
	b.buttonCompFunc = v
	return b
}

func (b *ListingBuilder) FooterAction(name string) (r *FooterActionBuilder) {
	builder := getFooterAction(b.footerActions, name)
	if builder != nil {
		return builder
	}
	r = &FooterActionBuilder{}
	r.name = name
	b.footerActions = append(b.footerActions, r)
	return
}
