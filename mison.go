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
