package validator

import (
	"strings"
	"time"
)

type Status struct {
	Timestamp time.Time
	Pubkey    string
	Status    string
	Group     string
	Monitor   string
}

const (
	StatusType = "validator_status"
)

func NewStatus(timestamp time.Time, status, pubkey, group, monitor string) *Status {
	return &Status{
		Timestamp: timestamp,
		Pubkey:    pubkey,
		Status:    status,
		Group:     group,
		Monitor:   monitor,
	}
}

func (v *Status) GetType() string {
	return StatusType
}

func (v *Status) GetGroup() string {
	return v.Group
}

func (v *Status) GetMonitor() string {
	return v.Monitor
}

func (v *Status) GetTitle(includeMonitor, includeGroup bool) string {
	var sb strings.Builder

	if includeMonitor {
		sb.WriteString("[")
		sb.WriteString(v.Monitor)
		sb.WriteString("] ")
	}

	sb.WriteString("Validator has unexpectedly status")

	return sb.String()
}

func (v *Status) GetDescriptionText(includeMonitor, includeGroup bool) string {
	var sb strings.Builder

	sb.WriteString("\nTimestamp: ")
	sb.WriteString(v.Timestamp.UTC().Format("2006-01-02 15:04:05 UTC"))

	if includeMonitor {
		sb.WriteString("\nMonitor: ")
		sb.WriteString(v.Monitor)
	}

	if includeGroup {
		sb.WriteString("\nGroup: ")
		sb.WriteString(v.Group)
	}

	sb.WriteString("\nPubkey: ")
	sb.WriteString(v.Pubkey)
	sb.WriteString("\nStatus: ")
	sb.WriteString(v.Status)

	return sb.String()
}

func (v *Status) GetDescriptionMarkdown(includeMonitor, includeGroup bool) string {
	var sb strings.Builder

	sb.WriteString("**Timestamp:** ")
	sb.WriteString(v.Timestamp.UTC().Format("2006-01-02 15:04:05 UTC"))
	sb.WriteString("\n")

	if includeMonitor {
		sb.WriteString("**Monitor:** ")
		sb.WriteString(v.Monitor)
		sb.WriteString("\n")
	}

	if includeGroup {
		sb.WriteString("**Group:** ")
		sb.WriteString(v.Group)
		sb.WriteString("\n")
	}

	sb.WriteString("**Pubkey:** `")
	sb.WriteString(v.Pubkey)
	sb.WriteString("`\n")

	sb.WriteString("**Status:** ")
	sb.WriteString(v.Status)

	return sb.String()
}

func (v *Status) GetDescriptionHTML(includeMonitor, includeGroup bool) string {
	var sb strings.Builder

	sb.WriteString("<p><strong>Timestamp:</strong> ")
	sb.WriteString(v.Timestamp.UTC().Format("2006-01-02 15:04:05 UTC"))
	sb.WriteString("</p>")

	if includeMonitor {
		sb.WriteString("<p><strong>Monitor:</strong> ")
		sb.WriteString(v.Monitor)
		sb.WriteString("</p>")
	}

	if includeGroup {
		sb.WriteString("<p><strong>Group:</strong> ")
		sb.WriteString(v.Group)
		sb.WriteString("</p>")
	}

	sb.WriteString("<p><strong>Pubkey:</strong> ")
	sb.WriteString(v.Pubkey)
	sb.WriteString("</p>")

	sb.WriteString("<p><strong>Status:</strong> ")
	sb.WriteString(v.Status)
	sb.WriteString("</p>")

	return sb.String()
}
