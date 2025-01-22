package contract

import (
	"context"
	"crypto/ecdsa"
	_ "embed"
	"time"

	"github.com/0xsequence/ethkit/go-ethereum/crypto"
	"github.com/ethpandaops/splitoor/pkg/ethereum/execution"
	"github.com/sirupsen/logrus"
)

type Deployer struct {
	log        logrus.FieldLogger
	config     *Config
	from       string
	privateKey *ecdsa.PrivateKey
	gasLimit   uint64
}

func NewDeployer(ctx context.Context, log logrus.FieldLogger, config *Config) (*Deployer, error) {
	privateKey, err := crypto.HexToECDSA(config.PrivateKey)
	if err != nil {
		return nil, err
	}

	return &Deployer{
		log:        log.WithField("module", "0xsplits/contract/deployer"),
		config:     config,
		from:       config.From,
		privateKey: privateKey,
		gasLimit:   config.GasLimit,
	}, nil
}

func (d *Deployer) Deploy(ctx context.Context, node *execution.Node) (*string, error) {
	txHash, err := node.DeployContract(ctx, &SplitMainBin, d.from, d.privateKey, d.gasLimit)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	d.log.WithField("tx", *txHash).Info("Waiting for transaction to be included in a block")

	for {
		var isPending bool

		_, isPending, err = node.TransactionByHash(ctx, *txHash)
		if err != nil {
			return nil, err
		}

		if !isPending {
			break
		}

		time.Sleep(time.Second)
	}

	receipt, err := node.TransactionReceipt(ctx, *txHash)
	if err != nil {
		return nil, err
	}

	contractAddress := receipt.ContractAddress.Hex()

	return &contractAddress, nil
}
