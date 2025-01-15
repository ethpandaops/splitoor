package split

import (
	"bytes"
	"encoding/hex"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/crypto"
)

type HashParams struct {
	Accounts              []string
	PercentageAllocations []uint32
	DistributorFee        uint32

	mu sync.Mutex
}

func (p *HashParams) order() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	accounts, allocations, err := ParseRecipients(p.Accounts, p.PercentageAllocations)
	if err != nil {
		return err
	}

	p.Accounts = accounts
	p.PercentageAllocations = allocations

	return nil
}

// Calculate the hash of the split
// requires abi.encodePacked non standard packed mode
// https://docs.soliditylang.org/en/latest/abi-spec.html#non-standard-packed-mode
func (c *Client) CalculateHash(params *HashParams) ([]byte, error) {
	if err := params.order(); err != nil {
		return nil, err
	}

	data := encodePacked(
		encodeAddressArrayPadded(params.Accounts),
		encodeUint32ArrayPadded(params.PercentageAllocations),
		encodeUint32NoPad(0),
	)

	return crypto.Keccak256(data), nil
}

func encodePacked(parts ...[]byte) []byte {
	return bytes.Join(parts, nil)
}

func encodeAddressArrayPadded(addrs []string) []byte {
	out := make([][]byte, 0, len(addrs))

	for _, addr := range addrs {
		out = append(out, encodeAddressPadded(addr))
	}

	return bytes.Join(out, nil)
}

func encodeAddressPadded(addr string) []byte {
	raw := decode20Bytes(addr)
	padded := make([]byte, 32)
	copy(padded[12:], raw)

	return padded
}

func decode20Bytes(addr string) []byte {
	s := strings.TrimPrefix(addr, "0x")

	decoded, err := hex.DecodeString(s)
	if err != nil {
		return nil
	}

	if len(decoded) != 20 {
		return nil
	}

	return decoded
}

func encodeUint32ArrayPadded(arr []uint32) []byte {
	out := make([][]byte, 0, len(arr))

	for _, v := range arr {
		out = append(out, encodeUint32Padded(v))
	}

	return bytes.Join(out, nil)
}

func encodeUint32Padded(value uint32) []byte {
	b := make([]byte, 32)
	b[28] = byte(value >> 24)
	b[29] = byte(value >> 16)
	b[30] = byte(value >> 8)
	b[31] = byte(value)

	return b
}

func encodeUint32NoPad(value uint32) []byte {
	b := make([]byte, 4)
	b[0] = byte(value >> 24)
	b[1] = byte(value >> 16)
	b[2] = byte(value >> 8)
	b[3] = byte(value)

	return b
}
