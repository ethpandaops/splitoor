package event

type Event interface {
	GetType() string
	GetText() string
	GetMarkdown() string
	GetGroup() string
}
