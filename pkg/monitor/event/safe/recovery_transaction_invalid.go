package safe

import (
	"strings"
	"time"
)

type RecoveryTransactionInvalid struct {
	Timestamp             time.Time
	SafeAddress           string
	Group                 string
	Monitor               string
	RecoveryTransactionID string
	Reason                string
}

const (
	RecoveryTransactionInvalidType = "safe_recovery_transaction_invalid"
)

func NewRecoveryTransactionInvalid(timestamp time.Time, monitor, group, safeAddress, txID, reason string) *RecoveryTransactionInvalid {
	return &RecoveryTransactionInvalid{
		Timestamp:             timestamp,
		SafeAddress:           safeAddress,
		Group:                 group,
		Monitor:               monitor,
		RecoveryTransactionID: txID,
		Reason:                reason,
	}
}

func (v *RecoveryTransactionInvalid) GetType() string {
	return RecoveryTransactionInvalidType
}

func (v *RecoveryTransactionInvalid) GetGroup() string {
	return v.Group
}

func (v *RecoveryTransactionInvalid) GetMonitor() string {
	return v.Monitor
}

func (v *RecoveryTransactionInvalid) GetTitle(includeMonitor, includeGroup bool) string {
	var sb strings.Builder

	if includeMonitor {
		sb.WriteString("[")
		sb.WriteString(v.Monitor)
		sb.WriteString("] ")
	}

	sb.WriteString("Safe account has invalid recovery transaction")

	return sb.String()
}

func (v *RecoveryTransactionInvalid) GetDescriptionText(includeMonitor, includeGroup bool) string {
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

	sb.WriteString("\nTransaction ID: ")
	sb.WriteString(v.RecoveryTransactionID)

	sb.WriteString("\nReason: ")
	sb.WriteString(v.Reason)

	return sb.String()
}

func (v *RecoveryTransactionInvalid) GetDescriptionMarkdown(includeMonitor, includeGroup bool) string {
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

	sb.WriteString("**Transaction ID:** `")
	sb.WriteString(v.RecoveryTransactionID)
	sb.WriteString("`\n")

	sb.WriteString("**Reason:** ")
	sb.WriteString(v.Reason)

	return sb.String()
}

func (v *RecoveryTransactionInvalid) GetDescriptionHTML(includeMonitor, includeGroup bool) string {
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

	sb.WriteString("<p><strong>Transaction ID:</strong> ")
	sb.WriteString(v.RecoveryTransactionID)
	sb.WriteString("</p>")

	sb.WriteString("<p><strong>Reason:</strong> ")
	sb.WriteString(v.Reason)
	sb.WriteString("</p>")

	return sb.String()
}
