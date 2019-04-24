package mison

/*
RemoveRightmost1 removes the rightmost 1 in x.

E.g.
	11101000 -> 11100000
*/
func RemoveRightmost1(x uint32) uint32 {
	return x & (x - 1)
}

/*
ExtractRightmost1 extract the rightmost 1 in x.

E.g.
	11101000 -> 00001000
*/
func ExtractRightmost1(x uint32) uint32 {
	return x & -x
}

/*
SmearRightmost1 extract the rightmost 1 in x and smear it to the right.

E.g.
	11101000 -> 00001111
*/
func SmearRightmost1(x uint32) uint32 {
	return x ^ (x - 1)
}

/*
BuildCharacterBitmap builds bitmap for specified character
*/
func BuildCharacterBitmap(text string, ch byte) []uint32 {
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
