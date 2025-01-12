package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/ethpandaops/splitoor/pkg/0xsplits/contract"
	"github.com/ethpandaops/splitoor/pkg/ethereum/execution"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var splitCmd = &cobra.Command{
	Use:   "split",
	Short: "Manage splits",
	Long:  `Create, update, or check status of splits.`,
}

func init() {
	rootCmd.AddCommand(splitCmd)
}

func getSplitDefaultContractAddress(ctx context.Context, dpNode *execution.Node) (*string, error) {
	var address *string

	chainID, err := dpNode.ChainID(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get chain id")
	}

	address = contract.GetDefaultContractAddress(chainID.String())
	if address != nil {
		return address, nil
	}

	return nil, fmt.Errorf("contract address is required (chain id: %s)", chainID.String())
}
func parseRecipients(accounts, percentageAllocations string) (recipients []string, allocations []uint32, err error) {
	acc := strings.Split(accounts, ",")
	alloc := strings.Split(percentageAllocations, ",")

	if len(acc) < 2 {
		return nil, nil, fmt.Errorf("must specify at least 2 recipients")
	}

	if len(acc) != len(alloc) {
		return nil, nil, fmt.Errorf("number of accounts and percentage allocations must match")
	}

	allocations = make([]uint32, 0)

	var total uint32

	for _, v := range alloc {
		val, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid percentage allocation %s: %v", v, err)
		}

		allocation := uint32(val)
		total += allocation
		allocations = append(allocations, allocation)
	}

	if total != 1000000 {
		return nil, nil, fmt.Errorf("percentage allocations must sum to 1000000, got %d", total)
	}

	return acc, allocations, nil
}
