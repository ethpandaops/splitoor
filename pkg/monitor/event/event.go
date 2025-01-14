package event

type Event interface {
	GetMonitor() string
	GetType() string
	GetTitle() string
	GetDescription() string
	GetGroup() string
}
