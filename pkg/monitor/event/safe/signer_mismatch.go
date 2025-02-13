package safe

import (
	"strings"
	"time"
)

type SignerMismatch struct {
	Timestamp   time.Time `json:"timestamp"`
	Monitor     string    `json:"monitor"`
	Group       string    `json:"name"`
	SafeAddress string    `json:"address"`
}

const (
	SignerMismatchType = "signer_mismatch"
)

func NewSignerMismatch(timestamp time.Time, monitor, group, safeAddress string) *SignerMismatch {
	return &SignerMismatch{
		Timestamp:   timestamp,
		Monitor:     monitor,
		Group:       group,
		SafeAddress: safeAddress,
	}
}

func (v *SignerMismatch) GetType() string {
	return SignerMismatchType
}

func (v *SignerMismatch) GetGroup() string {
	return "safe"
}

func (v *SignerMismatch) GetMonitor() string {
	return v.Monitor
}

func (v *SignerMismatch) GetTitle(includeMonitor, includeGroup bool) string {
	var sb strings.Builder

	if includeMonitor {
		sb.WriteString("[")
		sb.WriteString(v.Monitor)
		sb.WriteString("] ")
	}

	sb.WriteString("Safe account has unexpected owners")

	return sb.String()
}

func (v *SignerMismatch) GetDescriptionText(includeMonitor, includeGroup bool) string {
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

	sb.WriteString("\nSafe Account: ")
	sb.WriteString(v.SafeAddress)

	return sb.String()
}

func (v *SignerMismatch) GetDescriptionMarkdown(includeMonitor, includeGroup bool) string {
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

	sb.WriteString("**Safe Account:** `")
	sb.WriteString(v.SafeAddress)
	sb.WriteString("`")

	return sb.String()
}

func (v *SignerMismatch) GetDescriptionHTML(includeMonitor, includeGroup bool) string {
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

	sb.WriteString("<p><strong>Safe Account:</strong> ")
	sb.WriteString(v.SafeAddress)
	sb.WriteString("</p>")

	return sb.String()
}
