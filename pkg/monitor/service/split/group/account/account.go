package account

import (
	"context"

	"github.com/sirupsen/logrus"
)

type Account struct {
	log                 logrus.FieldLogger
	name                string
	monitor             string
	address             string
	allocation          uint32
	shouldGatherMetrics bool
}

func NewAccount(log logrus.FieldLogger, monitor, name, address string, allocation uint32, shouldGatherMetrics bool) *Account {
	return &Account{
		log:                 log,
		monitor:             monitor,
		name:                name,
		address:             address,
		allocation:          allocation,
		shouldGatherMetrics: shouldGatherMetrics,
	}
}

func (a *Account) Start(ctx context.Context) error {
	return nil
}

func (a *Account) Stop(ctx context.Context) error {
	return nil
}

func (a *Account) Address() string {
	return a.address
}

func (a *Account) Allocation() uint32 {
	return a.allocation
}
