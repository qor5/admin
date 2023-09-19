package worker

import "context"

//go:generate moq -pkg mock -out mock/queue.go . Queue

type QorJobDefinition struct {
	Name    string
	Handler JobHandler
}

type Queue interface {
	Add(ctx context.Context, job QueJobInterface) error
	Kill(ctx context.Context, job QueJobInterface) error
	Remove(ctx context.Context, job QueJobInterface) error
	Listen(jobDefs []*QorJobDefinition, getJob func(qorJobID uint) (QueJobInterface, error)) error
	Shutdown(ctx context.Context) error
}
