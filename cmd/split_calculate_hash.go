package cmd

import (
	"encoding/hex"

	"github.com/ethpandaops/splitoor/pkg/0xsplits/split"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	calculateHashRecipients     string
	calculateHashPercentages    string
	calculateHashDistributorFee uint32
)

var calculateHashSplitCmd = &cobra.Command{
	Use:   "calculate-hash",
	Short: "Calculate split hash",
	Long:  `Calculate the keccak256 hash of split recipients and percentages.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		initCommon()

		err := calculateSplitHash()
		if err != nil {
			log.Fatal(err)
		}

		return nil
	},
}

func init() {
	splitCmd.AddCommand(calculateHashSplitCmd)

	calculateHashSplitCmd.Flags().StringVar(&calculateHashRecipients, "recipients", "", "Comma-separated list of recipient addresses")
	calculateHashSplitCmd.Flags().StringVar(&calculateHashPercentages, "percentages", "", "Comma-separated list of percentages as an integer where 999999 = 99.9999%. Must sum to 1000000")
	calculateHashSplitCmd.Flags().Uint32Var(&calculateHashDistributorFee, "distributor-fee", 0, "Distributor fee percentage")

	err := calculateHashSplitCmd.MarkFlagRequired("recipients")
	if err != nil {
		log.WithError(err).Fatalf("Failed to mark flag %s as required", "recipients")
	}

	err = calculateHashSplitCmd.MarkFlagRequired("percentages")
	if err != nil {
		log.WithError(err).Fatalf("Failed to mark flag %s as required", "percentages")
	}
}

func calculateSplitHash() error {
	accounts, allocations, err := parseRecipients(calculateHashRecipients, calculateHashPercentages)
	if err != nil {
		return err
	}

	hashParams := &split.HashParams{
		Accounts:              accounts,
		PercentageAllocations: allocations,
		DistributorFee:        calculateHashDistributorFee,
	}

	hash, err := split.CalculateHash(hashParams)
	if err != nil {
		return err
	}

	log.WithFields(logrus.Fields{
		"hash": "0x" + hex.EncodeToString(hash),
	}).Info("Split hash")

	return nil
}
