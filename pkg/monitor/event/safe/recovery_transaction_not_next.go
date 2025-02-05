package safe

import (
	"strings"
	"time"
)

type RecoveryTransactionNotNext struct {
	Timestamp             time.Time
	SafeAddress           string
	Group                 string
	Monitor               string
	RecoveryTransactionID string
}

const (
	RecoveryTransactionNotNextType = "safe_recovery_transaction_not_next"
)

func NewRecoveryTransactionNotNext(timestamp time.Time, monitor, group, safeAddress, recoveryTxID string) *RecoveryTransactionNotNext {
	return &RecoveryTransactionNotNext{
		Timestamp:             timestamp,
		SafeAddress:           safeAddress,
		Group:                 group,
		Monitor:               monitor,
		RecoveryTransactionID: recoveryTxID,
	}
}

func (v *RecoveryTransactionNotNext) GetType() string {
	return RecoveryTransactionNotNextType
}

func (v *RecoveryTransactionNotNext) GetGroup() string {
	return v.Group
}

func (v *RecoveryTransactionNotNext) GetMonitor() string {
	return v.Monitor
}

func (v *RecoveryTransactionNotNext) GetTitle(includeMonitor, includeGroup bool) string {
	var sb strings.Builder

	if includeMonitor {
		sb.WriteString("[")
		sb.WriteString(v.Monitor)
		sb.WriteString("] ")
	}

	sb.WriteString("Safe account has a recovery transaction that is not next in queue")

	return sb.String()
}

func (v *RecoveryTransactionNotNext) GetDescriptionText(includeMonitor, includeGroup bool) string {
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

	sb.WriteString("\nRecovery Transaction: ")
	sb.WriteString(v.RecoveryTransactionID)

	return sb.String()
}

func (v *RecoveryTransactionNotNext) GetDescriptionMarkdown(includeMonitor, includeGroup bool) string {
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
	sb.WriteString("`\n")

	sb.WriteString("**Recovery Transaction:** `")
	sb.WriteString(v.RecoveryTransactionID)
	sb.WriteString("`")

	return sb.String()
}

func (v *RecoveryTransactionNotNext) GetDescriptionHTML(includeMonitor, includeGroup bool) string {
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

	sb.WriteString("<p><strong>Recovery Transaction:</strong> ")
	sb.WriteString(v.RecoveryTransactionID)
	sb.WriteString("</p>")

	return sb.String()
}
