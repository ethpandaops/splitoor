package cmd

import (
	"context"
	"encoding/hex"

	"github.com/ethpandaops/splitoor/pkg/0xsplits/contract"
	"github.com/ethpandaops/splitoor/pkg/0xsplits/split"
	"github.com/ethpandaops/splitoor/pkg/ethereum/execution"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	statusElRPCURL        string
	statusContractAddress string
	statusSplitAddress    string
)

var statusSplitCmd = &cobra.Command{
	Use:   "status",
	Short: "Check split status",
	Long:  `Check the status of an existing split.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		initCommon()

		err := getSplitStatus(cmd.Context())
		if err != nil {
			log.Fatal(err)
		}

		return nil
	},
}

func init() {
	splitCmd.AddCommand(statusSplitCmd)

	statusSplitCmd.Flags().StringVar(&statusElRPCURL, "el-rpc-url", "", "Execution layer (EL) RPC URL")
	statusSplitCmd.Flags().StringVar(&statusContractAddress, "contract", "", "Contract address")
	statusSplitCmd.Flags().StringVar(&statusSplitAddress, "split", "", "Split address to check status")

	err := statusSplitCmd.MarkFlagRequired("el-rpc-url")
	if err != nil {
		log.WithError(err).Fatalf("Failed to mark flag %s as required", "el-rpc-url")
	}

	err = statusSplitCmd.MarkFlagRequired("split")
	if err != nil {
		log.WithError(err).Fatalf("Failed to mark flag %s as required", "split")
	}
}

func getSplitStatus(ctx context.Context) error {
	dpNodeConfig := &execution.Config{
		NodeAddress: statusElRPCURL,
	}

	dpNode := execution.NewNode(log, "execution", dpNodeConfig)
	if err := dpNode.Start(ctx); err != nil {
		return errors.Wrap(err, "failed to start execution node")
	}

	if statusContractAddress == "" {
		address, err := getSplitDefaultContractAddress(ctx, dpNode)
		if err != nil {
			return err
		}

		statusContractAddress = *address
	}

	config := &split.Config{
		ContractAddress: statusContractAddress,
		SplitAddress:    &statusSplitAddress,
	}

	client, err := split.NewClient(log, config)
	if err != nil {
		return err
	}

	contractABI, err := contract.GetSplitMainAbi()
	if err != nil {
		return err
	}

	controller, err := client.GetController(ctx, dpNode, contractABI)
	if err != nil {
		return err
	}

	hash, err := client.GetHash(ctx, dpNode, contractABI)
	if err != nil {
		return err
	}

	log.WithFields(logrus.Fields{
		"controller": *controller,
		"hash":       "0x" + hex.EncodeToString(hash[:]),
	}).Info("Split status")

	return nil
}
