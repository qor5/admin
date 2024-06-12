package presets

func (mb *ModelBuilder) Use(vs ...ModelPlugin) (r *ModelBuilder) {
	mb.plugins = append(mb.plugins, vs...) // Only for debug for now
	for _, mp := range vs {
		err := mp.ModelInstall(mb.p, mb)
		if err != nil {
			panic(err)
		}
	}

	return mb
}

func (b *Builder) Use(vs ...Plugin) (r *Builder) {
	b.plugins = append(b.plugins, vs...) // Only for debug for now
	for _, p := range vs {
		err := p.Install(b)
		if err != nil {
			panic(err)
		}
	}
	return b
}
