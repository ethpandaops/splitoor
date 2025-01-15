package validator

import (
	"fmt"
	"time"
)

type Controller struct {
	Timestamp    time.Time
	SplitAddress string
	SplitName    string
	Controller   string
	Group        string
	Monitor      string
}

const (
	ControllerType = "split_controller"
)

func (v *Controller) GetType() string {
	return ControllerType
}

func (v *Controller) GetGroup() string {
	return v.Group
}

func (v *Controller) GetMonitor() string {
	return v.Monitor
}

func (v *Controller) GetTitle() string {
	return fmt.Sprintf("[%s] %s split controller has changed", v.Monitor, v.Group)
}

func (v *Controller) GetDescription() string {
	return fmt.Sprintf(`
Timestamp: %s
Monitor: %s
Group: %s
Split Name: %s
Split Address: %s
Controller: %s`, v.Timestamp.UTC().Format("2006-01-02 15:04:05 UTC"), v.Monitor, v.Group, v.SplitName, v.SplitAddress, v.Controller)
}
