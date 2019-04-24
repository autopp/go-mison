package mison

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

/*
buildCharacterBitmap builds bitmap for specified character
*/
func buildCharacterBitmap(text string, ch byte) []uint32 {
	jsonBytes := []byte(text)
	bitmap := make([]uint32, (len(jsonBytes)+31)/32)
	for i := range bitmap {
		sublen := len(jsonBytes) - i*32
		if sublen > 32 {
			sublen = 32
		}
		for _, x := range jsonBytes[i*32 : i*32+sublen] {
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
	backslashes  []uint32
	doubleQuotes []uint32
	colons       []uint32
	lBraces      []uint32
	rBraces      []uint32
}

/*
buildStructualCharacterBitmaps build structual character bitmaps

See section 4.2.1. (currently, SIMD is not used)
*/
func buildStructualCharacterBitmaps(text string) *structualCharacterBitmaps {
	return &structualCharacterBitmaps{
		backslashes:  buildCharacterBitmap(text, '\\'),
		doubleQuotes: buildCharacterBitmap(text, '"'),
		colons:       buildCharacterBitmap(text, ':'),
		lBraces:      buildCharacterBitmap(text, '{'),
		rBraces:      buildCharacterBitmap(text, '}'),
	}
}
