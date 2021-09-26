package publish

type Builder struct {
}

func New() *Builder {
	return &Builder{}
}

func (b *Builder) Sync(models ...interface{}) error {
	return nil
}

// 幂等
func (b *Builder) Publish(record interface{}) error {
	return nil
}

func (b *Builder) UnPublish(record interface{}) error {
	return nil
}
