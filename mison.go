package mison

import (
	"errors"
	"fmt"
	"math"
	"math/bits"
	"regexp"
	"strconv"
)

type structualIndex struct {
	json                []byte
	level               int
	stringMaskBitmap    []uint32
	leveledColonBitmaps [][]uint32
}

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
func buildStructualCharacterBitmaps(json []byte) *structualCharacterBitmaps {
	indices := map[byte]int{'\\': 0, '"': 1, ':': 2, '{': 3, '}': 4}
	jsonLen := len(json)
	bitmapLen := (jsonLen-1)/32 + 1
	bitmaps := [][]uint32{
		make([]uint32, bitmapLen),
		make([]uint32, bitmapLen),
		make([]uint32, bitmapLen),
		make([]uint32, bitmapLen),
		make([]uint32, bitmapLen),
	}

	for i := 0; i < bitmapLen; i++ {
		sublen := jsonLen - i*32
		if sublen > 32 {
			sublen = 32
		}
		for _, x := range json[i*32 : i*32+sublen] {
			for _, bitmap := range bitmaps {
				bitmap[i] >>= 1
			}

			j, ok := indices[x]
			if ok {
				bitmaps[j][i] |= 1 << 31
			}
		}

		for _, bitmap := range bitmaps {
			bitmap[i] >>= uint(32 - sublen)
		}
	}

	return &structualCharacterBitmaps{
		backslashes: bitmaps[indices['\\']],
		quotes:      bitmaps[indices['"']],
		colons:      bitmaps[indices[':']],
		lBraces:     bitmaps[indices['{']],
		rBraces:     bitmaps[indices['}']],
	}
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
		s.body = append(s.body, struct {
			index int
			mask  uint32
		}{})
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
					return nil, fmt.Errorf("unexpected right curry blace is found at position %d", i*32+32-offset)
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

func buildStructualIndex(json []byte, level int) (*structualIndex, error) {
	charactersBitmaps := buildStructualCharacterBitmaps(json)
	quoteBitmap := buildStructualQuoteBitmap(charactersBitmaps)
	stringMaskBitmap := buildStringMaskBitmap(quoteBitmap)
	leveledColonBitmaps, err := buildLeveledColonBitmaps(charactersBitmaps, stringMaskBitmap, level)

	if err != nil {
		return nil, err
	}

	return &structualIndex{
		json:                json,
		level:               level,
		stringMaskBitmap:    stringMaskBitmap,
		leveledColonBitmaps: leveledColonBitmaps,
	}, nil
}

func retrieveFieldName(json []byte, stringMaskBitmap []uint32, colon int) (string, error) {
	// find ending quote
	i := colon / 32
	mask := stringMaskBitmap[i] & (uint32(1)<<uint32(colon) - 1)
	if mask == uint32(0) {
		for i--; i >= 0 && stringMaskBitmap[i] == 0; i-- {
		}

		if i < 0 {
			return "", fmt.Errorf("ending quote for colon at %d is not found", colon)
		}

		mask = stringMaskBitmap[i]
	}

	leadingZeros := bits.LeadingZeros32(mask)
	endQuote := 32*i + 31 - leadingZeros

	leadingOnes := bits.LeadingZeros32(^(mask << uint32(leadingZeros)))

	if leadingOnes == 32-leadingZeros {
		for i--; i >= 0; i-- {
			l := bits.LeadingZeros32(^stringMaskBitmap[i])
			leadingOnes += l
			if l != 32 {
				break
			}
		}

		if i < 0 {
			return "", fmt.Errorf("starting quote for colon at %d is not found", colon)
		}
	}

	startQuote := endQuote - leadingOnes
	fieldName, err := strconv.Unquote(string(json[startQuote : endQuote+1]))
	if err != nil {
		return "", err
	}

	return fieldName, nil
}

type queriedFieldTable map[string]*queriedFieldEntry

type queriedFieldEntry struct {
	id       int
	children queriedFieldTable
}

func findStructualDot(queriedField string) int {
	for i, c := range queriedField {
		if c == '.' {
			escaped := false
			for j := i - 1; j >= 0; j-- {
				if queriedField[j] == '\\' {
					escaped = !escaped
				} else {
					break
				}
			}
			if !escaped {
				return i
			}
		}
	}

	return -1
}

func buildQueriedFieldTableFromSingleField(t queriedFieldTable, queriedField, fullField string, nextID int, level int) (int, error) {
	maxLevel := level
	r := regexp.MustCompile(`\\(.)`)
	if dot := findStructualDot(queriedField); dot >= 0 {
		parent := r.ReplaceAllString(queriedField[0:dot], `$1`)
		child := queriedField[dot+1:]

		if _, ok := t[parent]; !ok {
			t[parent] = &queriedFieldEntry{children: make(queriedFieldTable)}
		} else if t[parent].children == nil {
			return -1, fmt.Errorf("duplicated field %q", fullField)
		}
		l, err := buildQueriedFieldTableFromSingleField(t[parent].children, child, fullField, nextID, level+1)
		if err != nil {
			return -1, err
		}
		if l > maxLevel {
			maxLevel = l
		}
	} else {
		queriedField = r.ReplaceAllString(queriedField, `$1`)
		if _, ok := t[queriedField]; ok {
			return -1, fmt.Errorf("duplicated field %q", fullField)
		}
		t[queriedField] = &queriedFieldEntry{id: nextID}
	}
	return maxLevel, nil
}

func buildQueriedFieldTable(queriedFields []string) (queriedFieldTable, int, error) {
	t := make(queriedFieldTable)
	level := 0

	for i, field := range queriedFields {
		l, err := buildQueriedFieldTableFromSingleField(t, field, field, i, 1)
		if err != nil {
			return nil, -1, err
		}
		if l > level {
			level = l
		}
	}

	return t, level, nil
}

/*
JSONType represents type of the field
*/
type JSONType int

const (
	// JSONUnknown is initial value of JSONType and represnents no value
	JSONUnknown JSONType = iota
	// JSONNull represents null in JSON
	JSONNull
	// JSONBool represents boolean value in JSON
	JSONBool
	// JSONNumber represents number value in JSON
	JSONNumber
	// JSONString represents string value in JSON
	JSONString
)

// KeyValue represents found key-value in JSON
type KeyValue struct {
	FieldID  int
	RawValue string
	Type     JSONType
	Err      error
}

var errUnexpectedObject = errors.New("unexpected object")
var errUnexpectedArray = errors.New("unexpected array")

func parseLiteral(json []byte, colon int) (string, JSONType, error) {
	i := colon + 1
	size := len(json)
	// skip blanks
	for ; i < size; i++ {
		if !(json[i] == ' ' || json[i] == '\t' || json[i] == '\n') {
			break
		}
	}

	if i == size {
		return "", JSONUnknown, errors.New("value is not found")
	}

	if json[i] == '{' {
		return "", JSONUnknown, errUnexpectedObject
	} else if json[i] == '[' {
		return "", JSONUnknown, errUnexpectedArray
	}

	// Now parse literal
	r := regexp.MustCompile(`\A(true|false|null|[0-9]+(\.[0-9]+)?|"([^\\\n"]|\\[\\"])*")`)
	literal := r.Find(json[i:size])
	if literal == nil {
		return "", JSONUnknown, fmt.Errorf("value is not found at %d", i)
	}

	var t JSONType
	switch literal[0] {
	case 't', 'f':
		t = JSONBool
	case 'n':
		t = JSONNull
	case '"':
		t = JSONString
	default:
		t = JSONNumber
	}

	return string(literal), t, nil
}

func startParse(index *structualIndex, table queriedFieldTable) <-chan *KeyValue {
	var parse func(*structualIndex, queriedFieldTable, int, int, int, chan<- *KeyValue)
	parse = func(index *structualIndex, table queriedFieldTable, start, end, level int, ch chan<- *KeyValue) {
		json := index.json
		colons := generateColonPositions(index.leveledColonBitmaps, start, end, level)
		for i, colon := range colons {
			name, err := retrieveFieldName(json, index.stringMaskBitmap, colon)
			if err != nil {
				ch <- &KeyValue{Err: err}
			}

			if entry, ok := table[name]; ok {
				if entry.children == nil {
					// field is atomic value
					// parse value
					v, t, err := parseLiteral(index.json, colon)
					if err == errUnexpectedObject || err == errUnexpectedArray {
						// skip
					} else if err != nil {
						ch <- &KeyValue{Err: err}
					} else {
						ch <- &KeyValue{FieldID: entry.id, Type: t, RawValue: v}
					}
				} else {
					// field is objetct value
					var innerEnd int
					if i < len(colons)-1 {
						innerEnd = colons[i+1] - 1
					} else {
						innerEnd = end - 1
					}
					parse(index, entry.children, colon+1, innerEnd, level+1, ch)
				}
			}
		}
	}

	ch := make(chan *KeyValue)
	go func() {
		parse(index, table, 0, len(index.json), 0, ch)
		close(ch)
	}()

	return ch
}

// Parser is stream provider for specified queried fields
type Parser struct {
	queriedFieldTable queriedFieldTable
	level             int
}

func NewParser(queriedFields []string) (*Parser, error) {
	t, level, err := buildQueriedFieldTable(queriedFields)
	if err != nil {
		return nil, err
	}
	return &Parser{queriedFieldTable: t, level: level}, nil
}

func (p *Parser) Parse(json []byte) (<-chan *KeyValue, error) {
	index, err := buildStructualIndex(json, p.level)
	if err != nil {
		return nil, err
	}

	return startParse(index, p.queriedFieldTable), nil
}
