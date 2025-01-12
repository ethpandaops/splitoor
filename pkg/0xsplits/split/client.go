package split

import (
	"github.com/sirupsen/logrus"
)

type Client struct {
	log             logrus.FieldLogger
	contractAddress string
	splitAddress    *string
}

func NewClient(log logrus.FieldLogger, config *Config) (*Client, error) {
	return &Client{
		log:             log.WithField("module", "0xsplits/split/client"),
		contractAddress: config.ContractAddress,
		splitAddress:    config.SplitAddress,
	}, nil
}
