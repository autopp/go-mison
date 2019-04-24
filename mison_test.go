package mison

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func bitsToUint32(bits string) uint32 {
	var ret uint32
	n := len(bits)
	for i := 0; i < n; i++ {
		ret <<= 1
		if bits[i] == '1' {
			ret |= 1
		}
	}

	return ret
}

func Uint32ToBits(x uint32) string {
	bits := make([]byte, 32)
	for i := 31; i >= 0; i-- {
		if x&1 == 1 {
			bits[i] = '1'
		} else {
			bits[i] = '0'
		}
		x >>= 1
	}

	return string(bits)
}

func Uint32SliceToBits(slice []uint32) []string {
	ret := make([]string, len(slice))
	for i, x := range slice {
		ret[i] = Uint32ToBits(x)
	}

	return ret
}

func TestRemoveRightmost1(t *testing.T) {
	cases := []struct {
		bits     string
		expected string
	}{
		{bits: "11101000", expected: "11100000"},
		{bits: "01111111", expected: "01111110"},
		{bits: "00000000", expected: "00000000"},
	}

	for _, tt := range cases {
		title := fmt.Sprintf("input: %s, expected: %s", tt.bits, tt.expected)
		t.Run(title, func(t *testing.T) {
			assert.Equal(t, bitsToUint32(tt.expected), removeRightmost1(bitsToUint32(tt.bits)))
		})
	}
}

func TestExtractRightmost1(t *testing.T) {
	cases := []struct {
		bits     string
		expected string
	}{
		{bits: "11101000", expected: "00001000"},
		{bits: "01111111", expected: "00000001"},
		{bits: "00000000", expected: "00000000"},
	}

	for _, tt := range cases {
		title := fmt.Sprintf("input: %s, expected: %s", tt.bits, tt.expected)
		t.Run(title, func(t *testing.T) {
			assert.Equal(t, bitsToUint32(tt.expected), extractRightmost1(bitsToUint32(tt.bits)))
		})
	}
}

func TestSmearRightmost1(t *testing.T) {
	cases := []struct {
		bits     string
		expected string
	}{
		{bits: "11101000", expected: "00001111"},
		{bits: "01111111", expected: "00000001"},
		{bits: "00000000", expected: "111111111111111111111111111111111111111111111111111111111111111111111111"},
	}

	for _, tt := range cases {
		title := fmt.Sprintf("input: %s, expected: %s", tt.bits, tt.expected)
		t.Run(title, func(t *testing.T) {
			assert.Equal(t, bitsToUint32(tt.expected), smearRightmost1(bitsToUint32(tt.bits)))
		})
	}
}

func TestBuildCharacterBitmap(t *testing.T) {
	cases := []struct {
		input        string
		ch           byte
		expectedBits []string
	}{
		{input: `{"id":"id:\"a\"","reviews":50,"a`, ch: '\\', expectedBits: []string{"00000000000000000010010000000000"}},
		{input: `{"id":"id:\"a\"","reviews":50,"a`, ch: '"', expectedBits: []string{"01000010000000101100100001010010"}},
		{input: `{"id":"id:\"a\"","reviews":50,"a`, ch: ':', expectedBits: []string{"00000100000000000000001000100000"}},
		{input: `{"id":"id:\"a\"","reviews":50,"a`, ch: '{', expectedBits: []string{"00000000000000000000000000000001"}},
		{input: `{"id":"id:\"a\"","reviews":50,"a`, ch: '}', expectedBits: []string{"00000000000000000000000000000000"}},
		{input: `{"id":"id:\"a\"","reviews":50,"attributes":{"breakfast":false,"l`, ch: '"', expectedBits: []string{"01000010000000101100100001010010", "01000000010000000001001000000000"}},
		{input: `{"id":"id:\"a\"","reviews":50,"attributes"`, ch: '"', expectedBits: []string{"01000010000000101100100001010010", "00000000000000000000001000000000"}},
	}

	for _, tt := range cases {
		title := fmt.Sprintf("%s, %c", tt.input, tt.ch)
		t.Run(title, func(t *testing.T) {
			expected := make([]uint32, len(tt.expectedBits))
			for i, bits := range tt.expectedBits {
				expected[i] = bitsToUint32(bits)
			}
			actual := buildCharacterBitmap(tt.input, tt.ch)
			msg := fmt.Sprintf("%s != %s", Uint32SliceToBits(actual), tt.expectedBits)
			assert.Equal(t, expected, actual, msg)
		})
	}
}
