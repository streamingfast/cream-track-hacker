package main

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

type address []byte

func (b address) Pretty() string {
	return "0x" + hex.EncodeToString(b)
}

func newAddress(input string) (address, error) {
	out, err := hex.DecodeString(sanitizeHex(input))
	if err != nil {
		return nil, fmt.Errorf("invalid address %q: %w", input, err)
	}

	byteCount := len(out)
	if byteCount > 20 {
		out = out[byteCount-20:]
	}

	return address(out), nil
}

func sanitizeHex(input string) string {
	if has0xPrefix(input) {
		input = input[2:]
	}

	if len(input)%2 != 0 {
		input = "0" + input
	}

	return strings.ToLower(input)
}

func has0xPrefix(input string) bool {
	return len(input) >= 2 && input[0] == '0' && (input[1] == 'x' || input[1] == 'X')
}

type hash []byte

func (b hash) Pretty() string {
	return "0x" + hex.EncodeToString(b)
}

type addressSet []string

func (s addressSet) contains(address string) bool {
	for _, candidate := range s {
		if candidate == address {
			return true
		}
	}

	return false
}

func formatTokenAmount(in *big.Int, decimals, truncateDecimalCount uint) string {
	if in == nil {
		return ""
	}

	if decimals == 0 {
		return in.String()
	}

	var isNegative bool
	if in.Sign() < 0 {
		isNegative = true
		in = new(big.Int).Abs(in)
	}

	bigDecimals := decimalsInBigInt(uint32(decimals))
	whole := new(big.Int).Div(in, bigDecimals)

	reminder := new(big.Int).Rem(in, bigDecimals).String()
	missingLeadingZeros := int(decimals) - len(reminder)
	fractional := strings.Repeat("0", missingLeadingZeros) + reminder
	if truncateDecimalCount != 0 && len(fractional) > int(truncateDecimalCount) {
		fractional = fractional[0:truncateDecimalCount]
	}

	if isNegative {
		return fmt.Sprintf("-%s.%s", whole, fractional)
	}

	return fmt.Sprintf("%s.%s", whole, fractional)
}

var _10b = big.NewInt(10)
var _1e18b = new(big.Int).Exp(_10b, big.NewInt(18), nil)

func decimalsInBigInt(decimal uint32) *big.Int {
	if decimal == 18 {
		return _1e18b
	}

	return new(big.Int).Exp(_10b, big.NewInt(int64(decimal)), nil)
}
