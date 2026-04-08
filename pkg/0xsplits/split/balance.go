package split

import (
	"context"
	"errors"
	"math/big"

	"github.com/0xsequence/ethkit/ethcoder"
	"github.com/ethpandaops/splitoor/pkg/ethereum/execution"
)

func (c *Client) GetETHBalance(ctx context.Context, node *execution.Node, contractABI *ethcoder.ABI, address string) (*big.Int, error) {
	calldata, err := contractABI.EncodeMethodCalldataFromStringValues("getETHBalance", []string{address})
	if err != nil {
		return nil, err
	}

	balance, err := node.ReadContract(ctx, c.contractAddress, calldata, nil)
	if err != nil {
		return nil, err
	}

	values, err := contractABI.RawABI().Methods["getETHBalance"].Outputs.UnpackValues(balance)
	if err != nil {
		return nil, err
	}

	bigBalance, ok := values[0].(*big.Int)
	if !ok {
		return nil, errors.New("invalid balance")
	}

	return bigBalance, nil
}
