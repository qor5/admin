package exchange

type ImporterExecOption interface {
	iePrivate()
}

type ExporterExecOption interface {
	eePrivate()
}

type ExecOption interface {
	ImporterExecOption
	ExporterExecOption
}

func MaxParamsPerSQL(v int) ExecOption {
	return &maxParamsPerSQLOption{v}
}

type maxParamsPerSQLOption struct {
	v int
}

var _ ImporterExecOption = (*maxParamsPerSQLOption)(nil)
var _ ExporterExecOption = (*maxParamsPerSQLOption)(nil)

func (o *maxParamsPerSQLOption) iePrivate() {}
func (o *maxParamsPerSQLOption) eePrivate() {}
