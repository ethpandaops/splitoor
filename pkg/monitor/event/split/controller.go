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
	Source       string
	Group        string
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

func (v *Controller) GetText() string {
	return fmt.Sprintf("Split %s (%s) controller has changed to %s", v.SplitName, v.SplitAddress, v.Controller)
}

func (v *Controller) GetMarkdown() string {
	return fmt.Sprintf("Split %s (%s) controller has changed to %s", v.SplitName, v.SplitAddress, v.Controller)
}
