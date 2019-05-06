package mison

import (
	"errors"
	"fmt"
	"io"
	"math"
	"math/bits"
	"strconv"
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
structualCharacterBitmaps represents set of bitmap for structual character.
*/
type structualCharacterBitmaps struct {
	backslashes []uint32
	quotes      []uint32
	colons      []uint32
	lBraces     []uint32
	rBraces     []uint32
}

/*
buildStructualCharacterBitmaps builda structual character bitmaps.

See section 4.2.1 (currently, SIMD is not used).
*/
func buildStructualCharacterBitmaps(r io.Reader) (*structualCharacterBitmaps, error) {
	indices := map[byte]int{'\\': 0, '"': 1, ':': 2, '{': 3, '}': 4}
	bitmaps := [][]uint32{
		make([]uint32, 0),
		make([]uint32, 0),
		make([]uint32, 0),
		make([]uint32, 0),
		make([]uint32, 0),
	}
	buf := make([]byte, 32)

	for i := 0; ; i++ {
		n, err := io.ReadFull(r, buf)

		if err == io.EOF {
			break
		} else if err == nil || err == io.ErrUnexpectedEOF {
			for c := 0; c < 5; c++ {
				bitmaps[c] = append(bitmaps[c], 0)
			}
			for _, x := range buf[0:n] {
				for _, bitmap := range bitmaps {
					bitmap[i] >>= 1
				}

				j, ok := indices[x]
				if ok {
					bitmaps[j][i] |= 1 << 31
				}
			}
			for _, bitmap := range bitmaps {
				bitmap[i] >>= uint(32 - n)
			}
		} else {
			return nil, err
		}
	}
	return &structualCharacterBitmaps{
		backslashes: bitmaps[indices['\\']],
		quotes:      bitmaps[indices['"']],
		colons:      bitmaps[indices[':']],
		lBraces:     bitmaps[indices['{']],
		rBraces:     bitmaps[indices['}']],
	}, nil
}

/*
buildStructualQuoteBitmaps builds structual quote bitmaps.

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
buildStringMaskBitmap builds string mask bitmap.

See section 4.2.3.
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

type maskStack struct {
	body []struct {
		index int
		mask  uint32
	}
	sp int
}

const stackInitialSize = 32

func newMaskStack() *maskStack {
	return &maskStack{
		body: make([]struct {
			index int
			mask  uint32
		}, stackInitialSize),
		sp: 0,
	}
}

func (s *maskStack) push(index int, mask uint32) error {
	if s.sp == len(s.body) {
		panic("stack overflow")
	}

	s.body[s.sp].index = index
	s.body[s.sp].mask = mask
	s.sp++
	return nil
}

func (s *maskStack) pop() (int, uint32, error) {
	if s.sp == 0 {
		return 0, 0, errors.New("attempt pop from empty stack")
	}

	s.sp--
	return s.body[s.sp].index, s.body[s.sp].mask, nil
}

func buildLeveledColonBitmaps(bitmaps *structualCharacterBitmaps, stringMaskBitmap []uint32, level int) ([][]uint32, error) {
	bitmapLen := len(stringMaskBitmap)
	colons := bitmaps.colons
	lBraces := bitmaps.lBraces
	rBraces := bitmaps.rBraces

	// make colons, lBraces, rBraces to be structual
	for i := 0; i < bitmapLen; i++ {
		stringMask := ^stringMaskBitmap[i]
		colons[i] &= stringMask
		lBraces[i] &= stringMask
		rBraces[i] &= stringMask
	}

	colonBitmaps := make([][]uint32, level)
	for i := 0; i < level; i++ {
		colonBitmaps[i] = make([]uint32, bitmapLen)
		copy(colonBitmaps[i], colons)
	}
	stack := newMaskStack()

	for i := 0; i < bitmapLen; i++ {
		mLeft := lBraces[i]
		mRight := rBraces[i]
		for {
			mLeftBit := extractRightmost1(mLeft)
			mRightBit := extractRightmost1(mRight)
			for mLeftBit != 0 && (mRightBit == 0 || mLeftBit < mRightBit) {
				stack.push(i, mLeftBit)
				mLeft = removeRightmost1(mLeft)
				mLeftBit = extractRightmost1(mLeft)
			}
			if mRightBit != 0 {
				var j int
				var err error
				j, mLeftBit, err = stack.pop()
				if err != nil {
					offset := bits.LeadingZeros32(mRightBit)
					return nil, fmt.Errorf("unexpected right bracement is found at position %d", i*32+32-offset)
				}
				if stack.sp > 0 && stack.sp <= level {
					if i == j {
						colonBitmaps[stack.sp-1][i] &= ^(mRightBit - mLeftBit)
					} else {
						colonBitmaps[stack.sp-1][j] &= mLeftBit - 1
						colonBitmaps[stack.sp-1][i] &= ^(mRightBit - 1)
						for k := j + 1; k < i; k++ {
							colonBitmaps[stack.sp][k] = 0
						}
					}
				}
			} else {
				break
			}
			mRight = removeRightmost1(mRight)
		}
	}

	return colonBitmaps, nil
}

func generateColonPositions(index [][]uint32, start, end, level int) []int {
	colons := make([]int, 0)
	for i := int(math.Floor(float64(start) / 32)); i <= int(math.Floor(float64(end)/32)); i++ {
		mColon := index[level][i]
		for mColon != 0 {
			mBit := extractRightmost1(mColon)
			offset := i*32 + popcnt(mBit-1)
			if offset >= start && offset <= end {
				colons = append(colons, offset)
			}
			mColon = removeRightmost1(mColon)
		}
	}
	return colons
}

func buildStructualIndex(r io.Reader, level int) ([][]uint32, error) {
	charactersBitmaps, err := buildStructualCharacterBitmaps(r)
	if err != nil {
		return nil, err
	}

	quoteBitmap := buildStructualQuoteBitmap(charactersBitmaps)
	stringMaskBitmap := buildStringMaskBitmap(quoteBitmap)
	return buildLeveledColonBitmaps(charactersBitmaps, stringMaskBitmap, level)
}

func retrieveFieldName(json []byte, stringMaskBitmap []uint32, colon int) ([]byte, error) {
	// find ending quote
	i := colon / 32
	mask := stringMaskBitmap[i] & (uint32(1)<<uint32(colon) - 1)
	if mask == uint32(0) {
		return nil, errors.New("not implemented")
	}

	leadingZeros := bits.LeadingZeros32(mask)
	endQuote := 32*i + 31 - leadingZeros

	leadingOnes := bits.LeadingZeros32(^(mask << uint32(leadingZeros)))

	if leadingOnes == 32-leadingZeros {
		return nil, errors.New("not implemented")
	}

	startQuote := endQuote - leadingOnes
	fieldName, err := strconv.Unquote(string(json[startQuote : endQuote+1]))
	if err != nil {
		return nil, err
	}

	return []byte(fieldName), nil
}
