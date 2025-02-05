package split

import (
	"strings"
	"time"
)

type Controller struct {
	Timestamp          time.Time
	SplitAddress       string
	ExpectedController string
	ActualController   string
	Group              string
	Monitor            string
}

const (
	ControllerType = "split_controller"
)

func NewController(timestamp time.Time, monitor, group, splitAddress, expectedController, actualController string) *Controller {
	return &Controller{
		Timestamp:          timestamp,
		SplitAddress:       splitAddress,
		ExpectedController: expectedController,
		ActualController:   actualController,
		Group:              group,
		Monitor:            monitor,
	}
}

func (v *Controller) GetType() string {
	return ControllerType
}

func (v *Controller) GetGroup() string {
	return v.Group
}

func (v *Controller) GetMonitor() string {
	return v.Monitor
}

func (v *Controller) GetTitle(includeMonitor, includeGroup bool) string {
	var sb strings.Builder

	if includeMonitor {
		sb.WriteString("[")
		sb.WriteString(v.Monitor)
		sb.WriteString("] ")
	}

	sb.WriteString("Split controller has changed")

	return sb.String()
}

func (v *Controller) GetDescriptionText(includeMonitor, includeGroup bool) string {
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

	sb.WriteString("\nSplit Address: ")
	sb.WriteString(v.SplitAddress)
	sb.WriteString("\nExpected Controller address: ")
	sb.WriteString(v.ExpectedController)
	sb.WriteString("\nActual Controller address: ")
	sb.WriteString(v.ActualController)

	return sb.String()
}

func (v *Controller) GetDescriptionMarkdown(includeMonitor, includeGroup bool) string {
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

	sb.WriteString("**Split Address:** `")
	sb.WriteString(v.SplitAddress)
	sb.WriteString("`\n")

	sb.WriteString("**Expected Controller address:** `")
	sb.WriteString(v.ExpectedController)
	sb.WriteString("`\n")

	sb.WriteString("**Actual Controller address:** `")
	sb.WriteString(v.ActualController)
	sb.WriteString("`")

	return sb.String()
}

func (v *Controller) GetDescriptionHTML(includeMonitor, includeGroup bool) string {
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

	sb.WriteString("<p><strong>Split Address:</strong> ")
	sb.WriteString(v.SplitAddress)
	sb.WriteString("</p>")

	sb.WriteString("<p><strong>Expected Controller address:</strong> ")
	sb.WriteString(v.ExpectedController)
	sb.WriteString("</p>")

	sb.WriteString("<p><strong>Actual Controller address:</strong> ")
	sb.WriteString(v.ActualController)
	sb.WriteString("</p>")

	return sb.String()
}
