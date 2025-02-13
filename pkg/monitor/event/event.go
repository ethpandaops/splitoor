package event

type Event interface {
	GetMonitor() string
	GetType() string
	GetTitle(includeMonitor, includeGroup bool) string
	GetDescriptionText(includeMonitor, includeGroup bool) string
	GetDescriptionMarkdown(includeMonitor, includeGroup bool) string
	GetDescriptionHTML(includeMonitor, includeGroup bool) string
	GetGroup() string
}
