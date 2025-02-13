package validator

import (
	"fmt"
	"strings"
	"time"
)

type WithdrawalCredentials struct {
	Timestamp time.Time
	Pubkey    string
	Code      int64
	Group     string
	Monitor   string
}

const (
	WithdrawalCredentialsType = "validator_withdrawal_credentials"
)

func NewWithdrawalCredentials(timestamp time.Time, code int64, pubkey, group, monitor string) *WithdrawalCredentials {
	return &WithdrawalCredentials{
		Timestamp: timestamp,
		Pubkey:    pubkey,
		Code:      code,
		Group:     group,
		Monitor:   monitor,
	}
}

func (v *WithdrawalCredentials) GetType() string {
	return WithdrawalCredentialsType
}

func (v *WithdrawalCredentials) GetGroup() string {
	return v.Group
}

func (v *WithdrawalCredentials) GetMonitor() string {
	return v.Monitor
}

func (v *WithdrawalCredentials) GetTitle(includeMonitor, includeGroup bool) string {
	var sb strings.Builder

	if includeMonitor {
		sb.WriteString("[")
		sb.WriteString(v.Monitor)
		sb.WriteString("] ")
	}

	sb.WriteString("Validator has unexpected withdrawal credentials type")

	return sb.String()
}

func (v *WithdrawalCredentials) GetDescriptionText(includeMonitor, includeGroup bool) string {
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
	sb.WriteString("\nCode: 0x")
	sb.WriteString(fmt.Sprintf("%02x", v.Code))

	return sb.String()
}

func (v *WithdrawalCredentials) GetDescriptionMarkdown(includeMonitor, includeGroup bool) string {
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

	sb.WriteString("**Code:** `0x")
	sb.WriteString(fmt.Sprintf("%02x", v.Code))
	sb.WriteString("`")

	return sb.String()
}

func (v *WithdrawalCredentials) GetDescriptionHTML(includeMonitor, includeGroup bool) string {
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

	sb.WriteString("<p><strong>Code:</strong> 0x")
	sb.WriteString(fmt.Sprintf("%02x", v.Code))
	sb.WriteString("</p>")

	return sb.String()
}
