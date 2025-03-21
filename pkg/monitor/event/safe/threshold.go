package safe

import (
	"fmt"
	"strings"
	"time"
)

type Threshold struct {
	Timestamp         time.Time `json:"timestamp"`
	Monitor           string    `json:"monitor"`
	Group             string    `json:"name"`
	SafeAddress       string    `json:"address"`
	ActualThreshold   int       `json:"actual_threshold"`
	ExpectedThreshold int       `json:"expected_threshold"`
}

const (
	ThresholdType = "threshold"
)

func NewThreshold(timestamp time.Time, monitor, group, safeAddress string, actualThreshold, expectedThreshold int) *Threshold {
	return &Threshold{
		Timestamp:         timestamp,
		Monitor:           monitor,
		Group:             group,
		SafeAddress:       safeAddress,
		ActualThreshold:   actualThreshold,
		ExpectedThreshold: expectedThreshold,
	}
}

func (v *Threshold) GetType() string {
	return ThresholdType
}

func (v *Threshold) GetGroup() string {
	return "safe"
}

func (v *Threshold) GetMonitor() string {
	return v.Monitor
}

func (v *Threshold) GetTitle(includeMonitor, includeGroup bool) string {
	var sb strings.Builder

	if includeMonitor {
		sb.WriteString("[")
		sb.WriteString(v.Monitor)
		sb.WriteString("] ")
	}

	sb.WriteString("Safe account has unexpected threshold")

	return sb.String()
}

func (v *Threshold) GetDescriptionText(includeMonitor, includeGroup bool) string {
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

	sb.WriteString("\nActual Threshold: ")
	sb.WriteString(fmt.Sprintf("%d", v.ActualThreshold))

	sb.WriteString("\nExpected Threshold: ")
	sb.WriteString(fmt.Sprintf("%d", v.ExpectedThreshold))

	return sb.String()
}

func (v *Threshold) GetDescriptionMarkdown(includeMonitor, includeGroup bool) string {
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

	sb.WriteString("**Actual Threshold:** ")
	sb.WriteString(fmt.Sprintf("%d", v.ActualThreshold))
	sb.WriteString("\n")

	sb.WriteString("**Expected Threshold:** ")
	sb.WriteString(fmt.Sprintf("%d", v.ExpectedThreshold))

	return sb.String()
}

func (v *Threshold) GetDescriptionHTML(includeMonitor, includeGroup bool) string {
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

	sb.WriteString("<p><strong>Actual Threshold:</strong> ")
	sb.WriteString(fmt.Sprintf("%d", v.ActualThreshold))
	sb.WriteString("</p>")

	sb.WriteString("<p><strong>Expected Threshold:</strong> ")
	sb.WriteString(fmt.Sprintf("%d", v.ExpectedThreshold))
	sb.WriteString("</p>")

	return sb.String()
}
