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

func bitsSliceToUint32(slice []string) []uint32 {
	ret := make([]uint32, len(slice))
	for i, bits := range slice {
		ret[i] = bitsToUint32(bits)
	}

	return ret
}

func uint32ToBits(x uint32) string {
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

func uint32SliceToBits(slice []uint32) []string {
	ret := make([]string, len(slice))
	for i, x := range slice {
		ret[i] = uint32ToBits(x)
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
			expected := bitsSliceToUint32(tt.expectedBits)
			actual := buildCharacterBitmap(tt.input, tt.ch)
			msg := fmt.Sprintf("%s != %s", uint32SliceToBits(actual), tt.expectedBits)
			assert.Equal(t, expected, actual, msg)
		})
	}
}

type foo struct {
	bar []int
	baz []string
}

func TestBuildStructualCharacterBitmaps(t *testing.T) {
	cases := []struct {
		text            string
		backSlashBits   []string
		doubleQuoteBits []string
		colonBits       []string
		lBraceBits      []string
		rBraceBits      []string
	}{
		{
			text:            `{"id":"id:\"a\"","reviews":50,"a`,
			backSlashBits:   []string{"00000000000000000010010000000000"},
			doubleQuoteBits: []string{"01000010000000101100100001010010"},
			colonBits:       []string{"00000100000000000000001000100000"},
			lBraceBits:      []string{"00000000000000000000000000000001"},
			rBraceBits:      []string{"00000000000000000000000000000000"}},
	}

	for _, tt := range cases {
		title := fmt.Sprintf("input: %s", tt.text)
		t.Run(title, func(t *testing.T) {
			expected := &structualCharacterBitmaps{
				backslashes:  bitsSliceToUint32(tt.backSlashBits),
				doubleQuotes: bitsSliceToUint32(tt.doubleQuoteBits),
				colons:       bitsSliceToUint32(tt.colonBits),
				lBraces:      bitsSliceToUint32(tt.lBraceBits),
				rBraces:      bitsSliceToUint32(tt.rBraceBits),
			}

			assert.Equal(t, expected, buildStructualCharacterBitmaps(tt.text))
		})
	}
}
