package split

import (
	"context"
	"fmt"

	"github.com/0xsequence/ethkit/ethcoder"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/ethpandaops/splitoor/pkg/ethereum/execution"
)

func (c *Client) GetController(ctx context.Context, node *execution.Node, contractABI *ethcoder.ABI) (*string, error) {
	if c.splitAddress == nil {
		return nil, fmt.Errorf("split address is required")
	}

	calldata, err := contractABI.EncodeMethodCalldataFromStringValues("getController", []string{*c.splitAddress})
	if err != nil {
		return nil, err
	}

	status, err := node.ReadContract(ctx, c.contractAddress, calldata, nil)
	if err != nil {
		return nil, err
	}

	values, err := contractABI.RawABI().Methods["getController"].Outputs.UnpackValues(status)
	if err != nil {
		return nil, err
	}

	controllerAddress, ok := values[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid controller address")
	}

	controller := controllerAddress.Hex()

	return &controller, nil
}

func (c *Client) GetHash(ctx context.Context, node *execution.Node, contractABI *ethcoder.ABI) (*[32]uint8, error) {
	if c.splitAddress == nil {
		return nil, fmt.Errorf("split address is required")
	}

	calldata, err := contractABI.EncodeMethodCalldataFromStringValues("getHash", []string{*c.splitAddress})
	if err != nil {
		return nil, err
	}

	hash, err := node.ReadContract(ctx, c.contractAddress, calldata, nil)
	if err != nil {
		return nil, err
	}

	values, err := contractABI.RawABI().Methods["getHash"].Outputs.UnpackValues(hash)
	if err != nil {
		return nil, err
	}

	rsp, ok := values[0].([32]uint8)
	if !ok {
		return nil, fmt.Errorf("invalid hash")
	}

	return &rsp, nil
}
