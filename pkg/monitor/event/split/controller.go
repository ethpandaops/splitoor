package split

import (
	"fmt"
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

func (v *Controller) GetTitle() string {
	return fmt.Sprintf("[%s] %s split controller has changed", v.Monitor, v.Group)
}

func (v *Controller) GetDescription() string {
	return fmt.Sprintf(`
Timestamp: %s
Monitor: %s
Group: %s
Split Address: %s
Expected Controller address: %s
Actual Controller address: %s`, v.Timestamp.UTC().Format("2006-01-02 15:04:05 UTC"), v.Monitor, v.Group, v.SplitAddress, v.ExpectedController, v.ActualController)
}
