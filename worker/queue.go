package worker

//go:generate moq -pkg mock -out mock/queue.go . Queue

type QorJobDefinition struct {
	Name    string
	Handler JobHandler
}

type Queue interface {
	Add(QueJobInterface) error
	Kill(QueJobInterface) error
	Remove(QueJobInterface) error
	Listen(jobDefs []*QorJobDefinition, getJob func(qorJobID uint) (QueJobInterface, error)) error
}
