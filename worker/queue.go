package worker

type Queue interface {
	Add(QorJobInterface) error
	Run(QorJobInterface) error
	Kill(QorJobInterface) error
	Remove(QorJobInterface) error
}
