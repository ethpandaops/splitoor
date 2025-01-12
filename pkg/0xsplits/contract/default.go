package contract

import (
	_ "embed"

	"github.com/0xsequence/ethkit/ethcoder"
)

//go:embed SplitMain.bin
var SplitMainBin []byte

//go:embed SplitMain.json
var SplitMainAbi []byte

var defaultContracts = map[string]string{
	"1":        "0x2ed6c4B5dA6378c7897AC67Ba9e43102Feb694EE", // mainnet
	"11155111": "0x54E4a6014D36c381fC43b7E24A1492F556139a6F", // sepolia
	"17000":    "0xfC8a305728051367797DADE6Aa0344E0987f5286", // holesky
}

func GetDefaultContractAddress(chainID string) *string {
	if addr, ok := defaultContracts[chainID]; ok {
		return &addr
	}

	return nil
}

func GetSplitMainAbi() (*ethcoder.ABI, error) {
	wrappedABI := ethcoder.NewABI()

	err := wrappedABI.AddABIFromJSON(string(SplitMainAbi))
	if err != nil {
		return nil, err
	}

	return &wrappedABI, nil
}
