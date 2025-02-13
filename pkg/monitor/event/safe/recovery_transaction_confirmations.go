package safe

import (
	"fmt"
	"strings"
	"time"
)

type RecoveryTransactionConfirmations struct {
	Timestamp             time.Time
	SafeAddress           string
	Group                 string
	Monitor               string
	RecoveryTransactionID string
	NumConfirmations      int
	ExpectedConfirmations int
}

const (
	RecoveryTransactionConfirmationsType = "safe_recovery_transaction_confirmations"
)

func NewRecoveryTransactionConfirmations(timestamp time.Time, monitor, group, safeAddress, recoveryTxID string, numConfirmations, expectedConfirmations int) *RecoveryTransactionConfirmations {
	return &RecoveryTransactionConfirmations{
		Timestamp:             timestamp,
		SafeAddress:           safeAddress,
		Group:                 group,
		Monitor:               monitor,
		RecoveryTransactionID: recoveryTxID,
		NumConfirmations:      numConfirmations,
		ExpectedConfirmations: expectedConfirmations,
	}
}

func (v *RecoveryTransactionConfirmations) GetType() string {
	return RecoveryTransactionConfirmationsType
}

func (v *RecoveryTransactionConfirmations) GetGroup() string {
	return v.Group
}

func (v *RecoveryTransactionConfirmations) GetMonitor() string {
	return v.Monitor
}

func (v *RecoveryTransactionConfirmations) GetTitle(includeMonitor, includeGroup bool) string {
	var sb strings.Builder

	if includeMonitor {
		sb.WriteString("[")
		sb.WriteString(v.Monitor)
		sb.WriteString("] ")
	}

	sb.WriteString("Safe account has a recovery transaction with incorrect number of confirmations")

	return sb.String()
}

func (v *RecoveryTransactionConfirmations) GetDescriptionText(includeMonitor, includeGroup bool) string {
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

	sb.WriteString("\nCurrent Confirmations: ")
	sb.WriteString(fmt.Sprintf("%d", v.NumConfirmations))

	sb.WriteString("\nExpected Confirmations: ")
	sb.WriteString(fmt.Sprintf("%d", v.ExpectedConfirmations))

	return sb.String()
}

func (v *RecoveryTransactionConfirmations) GetDescriptionMarkdown(includeMonitor, includeGroup bool) string {
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
	sb.WriteString("`\n")

	sb.WriteString("**Current Confirmations:** ")
	sb.WriteString(fmt.Sprintf("%d", v.NumConfirmations))
	sb.WriteString("\n")

	sb.WriteString("**Expected Confirmations:** ")
	sb.WriteString(fmt.Sprintf("%d", v.ExpectedConfirmations))

	return sb.String()
}

func (v *RecoveryTransactionConfirmations) GetDescriptionHTML(includeMonitor, includeGroup bool) string {
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

	sb.WriteString("<p><strong>Current Confirmations:</strong> ")
	sb.WriteString(fmt.Sprintf("%d", v.NumConfirmations))
	sb.WriteString("</p>")

	sb.WriteString("<p><strong>Expected Confirmations:</strong> ")
	sb.WriteString(fmt.Sprintf("%d", v.ExpectedConfirmations))
	sb.WriteString("</p>")

	return sb.String()
}
