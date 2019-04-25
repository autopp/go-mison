package mison

import (
	"math/bits"
)

/*
removeRightmost1 removes the rightmost 1 in x.

E.g.
	11101000 -> 11100000
*/
func removeRightmost1(x uint32) uint32 {
	return x & (x - 1)
}

/*
extractRightmost1 extract the rightmost 1 in x.

E.g.
	11101000 -> 00001000
*/
func extractRightmost1(x uint32) uint32 {
	return x & -x
}

/*
smearRightmost1 extract the rightmost 1 in x and smear it to the right.

E.g.
	11101000 -> 00001111
*/
func smearRightmost1(x uint32) uint32 {
	return x ^ (x - 1)
}

func popcnt(x uint32) int {
	return bits.OnesCount32(x)
}

/*
buildCharacterBitmap builds bitmap for specified character
*/
func buildCharacterBitmap(text []byte, ch byte) []uint32 {
	bitmap := make([]uint32, (len(text)+31)/32)
	for i := range bitmap {
		sublen := len(text) - i*32
		if sublen > 32 {
			sublen = 32
		}
		for _, x := range text[i*32 : i*32+sublen] {
			bitmap[i] >>= 1
			if x == ch {
				bitmap[i] |= 1 << 31
			}
		}

		bitmap[i] >>= uint(32 - sublen)
	}
	return bitmap
}

/*
structualCharacterBitmaps represents set of bitmap for structual character
*/
type structualCharacterBitmaps struct {
	backslashes []uint32
	quotes      []uint32
	colons      []uint32
	lBraces     []uint32
	rBraces     []uint32
}

/*
buildStructualCharacterBitmaps build structual character bitmaps

See section 4.2.1. (currently, SIMD is not used)
*/
func buildStructualCharacterBitmaps(text string) *structualCharacterBitmaps {
	jsonBytes := []byte(text)
	return &structualCharacterBitmaps{
		backslashes: buildCharacterBitmap(jsonBytes, '\\'),
		quotes:      buildCharacterBitmap(jsonBytes, '"'),
		colons:      buildCharacterBitmap(jsonBytes, ':'),
		lBraces:     buildCharacterBitmap(jsonBytes, '{'),
		rBraces:     buildCharacterBitmap(jsonBytes, '}'),
	}
}

/*
buildStructualQuoteBitmaps builds structual quote bitmaps

See section 4.2.2.
*/
func buildStructualQuoteBitmap(bitmaps *structualCharacterBitmaps) []uint32 {
	backslashes := bitmaps.backslashes
	quotes := bitmaps.quotes
	bitmapLen := len(backslashes)
	backsalashedQuotes := make([]uint32, bitmapLen)
	for i := 0; i < bitmapLen-1; i++ {
		backsalashedQuotes[i] = ((quotes[i] >> 1) | (quotes[i+1] << 31)) & backslashes[i]
	}
	backsalashedQuotes[bitmapLen-1] = (quotes[bitmapLen-1] >> 1) & backslashes[bitmapLen-1]

	unstructualQuotes := make([]uint32, bitmapLen)
	for i := 0; i < bitmapLen; i++ {
		var unstructualQuote uint32
		backsalashedQuote := backsalashedQuotes[i]
		for backsalashedQuote != 0 {
			mask := smearRightmost1(backsalashedQuote)
			numberOfOnes := popcnt(mask)
			backslashOnLeft := (backslashes[i] & mask) << uint(32-numberOfOnes)
			numberOfLeadingOnes := bits.LeadingZeros32(^backslashOnLeft)
			if numberOfLeadingOnes == numberOfOnes {
				for j := i - 1; j >= 0; j-- {
					numberOfLeadingOneInWord := bits.LeadingZeros32(^backslashes[j])
					if numberOfLeadingOneInWord < 32 {
						numberOfLeadingOnes += numberOfLeadingOneInWord
						break
					}
				}
			} else if numberOfLeadingOnes > numberOfOnes {
				panic("illigal state about of bitmask")
			}
			if numberOfLeadingOnes&1 == 1 {
				unstructualQuote = unstructualQuote | extractRightmost1(backsalashedQuote)
			}
			backsalashedQuote = removeRightmost1(backsalashedQuote)
		}
		unstructualQuotes[i] = ^unstructualQuote
	}

	structualQuotes := make([]uint32, bitmapLen)
	structualQuotes[0] = quotes[0] & (unstructualQuotes[0] << 1)
	for i := 1; i < bitmapLen; i++ {
		structualQuotes[i] = quotes[i] & ((unstructualQuotes[i] << 1) | (unstructualQuotes[i-1] >> 31))
	}
	return structualQuotes
}

/*
buildStringMaskBitmap builds string mask bitmap
*/
func buildStringMaskBitmap(quoteBitmaps []uint32) []uint32 {
	// return []uint32{}
	bitmapLen := len(quoteBitmaps)
	n := 0
	stringBitmap := make([]uint32, bitmapLen)
	for i := 0; i < bitmapLen; i++ {
		quoteMask := quoteBitmaps[i]
		var stringMask uint32
		for quoteMask != 0 {
			mask := smearRightmost1(quoteMask)
			stringMask ^= mask
			quoteMask = removeRightmost1(quoteMask)
			n++
		}
		if n%2 == 1 {
			stringMask = ^stringMask
		}
		stringBitmap[i] = stringMask
	}
	return stringBitmap
}
