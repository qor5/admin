package worker

type QorJobDefinition struct {
	Name    string
	Handler JobHandler
}

type Queue interface {
	Add(QorJobInterface) error
	Kill(QorJobInterface) error
	Remove(QorJobInterface) error
	Listen(jobDefs []*QorJobDefinition, getJob func(qorJobID uint) (QorJobInterface, error)) error
}
