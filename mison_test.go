package mison

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func bitsToUint32(slice ...string) []uint32 {
	ret := make([]uint32, len(slice))
	for i, bits := range slice {
		ret[i] = 0
		n := len(bits)
		for j := 0; j < n; j++ {
			ret[i] <<= 1
			if bits[j] == '1' {
				ret[i] |= 1
			}
		}
	}

	return ret
}

func uint32ToBits(x uint32) string {
	return fmt.Sprintf("%032b", x)
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
			assert.Equal(t, bitsToUint32(tt.expected)[0], removeRightmost1(bitsToUint32(tt.bits)[0]))
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
			assert.Equal(t, bitsToUint32(tt.expected)[0], extractRightmost1(bitsToUint32(tt.bits)[0]))
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
			assert.Equal(t, bitsToUint32(tt.expected)[0], smearRightmost1(bitsToUint32(tt.bits)[0]))
		})
	}
}

func TestBuildStructualCharacterBitmaps(t *testing.T) {
	cases := []struct {
		text          string
		backSlashBits []string
		quoteBits     []string
		colonBits     []string
		lBraceBits    []string
		rBraceBits    []string
	}{
		{
			text:          `{"id":"id:\"a\"","reviews":50,"a`,
			backSlashBits: []string{"00000000000000000010010000000000"},
			quoteBits:     []string{"01000010000000101100100001010010"},
			colonBits:     []string{"00000100000000000000001000100000"},
			lBraceBits:    []string{"00000000000000000000000000000001"},
			rBraceBits:    []string{"00000000000000000000000000000000"},
		},
		{
			text: `      {"id":"id:\"a\"","reviews"` +
				`:50,"a`,
			backSlashBits: []string{
				"00000000000010010000000000000000",
				"00000000000000000000000000000000",
			},
			quoteBits: []string{
				"10000000101100100001010010000000",
				"00000000000000000000000000010000",
			},
			colonBits: []string{
				"00000000000000001000100000000000",
				"00000000000000000000000000000001",
			},
			lBraceBits: []string{
				"00000000000000000000000001000000",
				"00000000000000000000000000000000",
			},
			rBraceBits: []string{
				"00000000000000000000000000000000",
				"00000000000000000000000000000000",
			},
		},
	}

	for _, tt := range cases {
		title := fmt.Sprintf("input: %s", tt.text)
		t.Run(title, func(t *testing.T) {
			expected := &structualCharacterBitmaps{
				backslashes: bitsToUint32(tt.backSlashBits...),
				quotes:      bitsToUint32(tt.quoteBits...),
				colons:      bitsToUint32(tt.colonBits...),
				lBraces:     bitsToUint32(tt.lBraceBits...),
				rBraces:     bitsToUint32(tt.rBraceBits...),
			}

			actual := buildStructualCharacterBitmaps([]byte(tt.text))
			assert.Equal(t, expected, actual)
		})
	}
}

func TestBuildStructualQuoteBitmap(t *testing.T) {
	cases := []struct {
		bitmaps  *structualCharacterBitmaps
		expected []uint32
	}{
		{
			bitmaps: &structualCharacterBitmaps{
				backslashes: bitsToUint32("00000000000000000010010000000000"),
				quotes:      bitsToUint32("01000010000000101100100001010010"),
			},
			expected: bitsToUint32("01000010000000101000000001010010"),
		},
		{
			bitmaps: &structualCharacterBitmaps{
				backslashes: bitsToUint32(
					"00000000000000000000000000000000",
					"00000000000000000000000000000001",
				),
				quotes: bitsToUint32(
					"00000000000000000000000000000000",
					"00000000000000000000000000000010",
				),
			},
			expected: bitsToUint32(
				"00000000000000000000000000000000",
				"00000000000000000000000000000000",
			),
		},
		{
			bitmaps: &structualCharacterBitmaps{
				backslashes: bitsToUint32(
					"10000000000000000000000000000000",
					"00000000000000000000000000000001",
				),
				quotes: bitsToUint32(
					"00000000000000000000000000000000",
					"00000000000000000000000000000010",
				),
			},
			expected: bitsToUint32(
				"00000000000000000000000000000000",
				"00000000000000000000000000000010",
			),
		},
		{
			bitmaps: &structualCharacterBitmaps{
				backslashes: bitsToUint32(
					"10000000000000000000000000000000",
					"11111111111111111111111111111111",
					"00000000000000000000000000000001",
				),
				quotes: bitsToUint32(
					"00000000000000000000000000000000",
					"00000000000000000000000000000000",
					"00000000000000000000000000000010",
				),
			},
			expected: bitsToUint32(
				"00000000000000000000000000000000",
				"00000000000000000000000000000000",
				"00000000000000000000000000000010",
			),
		},
	}

	for i, tt := range cases {
		t.Run(fmt.Sprintf("case%d", i+1), func(t *testing.T) {
			actual := buildStructualQuoteBitmap(tt.bitmaps)
			assert.Equalf(t, tt.expected, actual, "expected: %s, actual: %s", uint32SliceToBits(tt.expected), uint32SliceToBits(actual))
		})
	}
}

func TestBuildStringMaskBitmap(t *testing.T) {
	cases := []struct {
		quoteBitmap []uint32
		expected    []uint32
	}{
		{
			quoteBitmap: bitsToUint32("01000010000000101000000001010010"),
			expected:    bitsToUint32("10000011111111001111111110011100"),
		},
		{
			quoteBitmap: bitsToUint32(
				"00010100100000000000000000000000",
				"00000000010100001000000010100000",
			),
			expected: bitsToUint32(
				"11100111000000000000000000000000",
				"00000000011000001111111100111111",
			),
		},
	}

	for i, tt := range cases {
		t.Run(fmt.Sprintf("case%d", i+1), func(t *testing.T) {
			actual := buildStringMaskBitmap(tt.quoteBitmap)
			assert.Equalf(t, tt.expected, actual, "expected: %s, actual: %s", uint32SliceToBits(tt.expected), uint32SliceToBits(actual))
		})
	}
}

func TestBuildLeveledColonBitmaps(t *testing.T) {
	cases := []struct {
		bitmaps    *structualCharacterBitmaps
		stringMask []uint32
		level      int
		expected   [][]uint32
	}{
		{
			// {"a":1,"b":{"c":2}}
			// {{2:"c"}:"b",1:"a"}
			bitmaps: &structualCharacterBitmaps{
				colons: bitsToUint32(
					"00000000000000001000010000010000",
				),
				lBraces: bitsToUint32("00000000000000000000100000000001"),
				rBraces: bitsToUint32("00000000000001100000000000000000"),
			},
			stringMask: bitsToUint32("00000000000000000110001100001100"),
			level:      2,
			expected: [][]uint32{
				bitsToUint32("00000000000000000000010000010000"),
				bitsToUint32("00000000000000001000010000010000"),
			},
		},
		{
			// {"a":1,"b":{"c":{"d":2},"e":3}}
			// }}3:"e",}2:"d"{:"c"{:"b",1:"a"{
			bitmaps: &structualCharacterBitmaps{
				colons:  bitsToUint32("00001000000100001000010000010000"),
				lBraces: bitsToUint32("00000000000000010000100000000001"),
				rBraces: bitsToUint32("01100000010000000000000000000000"),
			},
			stringMask: bitsToUint32("00000110000011000110001100001100"),
			level:      3,
			expected: [][]uint32{
				bitsToUint32("00000000000000000000010000010000"),
				bitsToUint32("00001000000000001000010000010000"),
				bitsToUint32("00001000000100001000010000010000"),
			},
		},
		{
			//                       {"a":1,"b"
			// :{"c":{"d":2},"e":3}}
			// "b",1:"a"{
			//            }}3:"e",}2:"d"{:"c"{:
			bitmaps: &structualCharacterBitmaps{
				colons: bitsToUint32(
					"00000100000000000000000000000000",
					"00000000000000100000010000100001",
				),
				lBraces: bitsToUint32(
					"00000000010000000000000000000000",
					"00000000000000000000000001000010",
				),
				rBraces: bitsToUint32(
					"00000000000000000000000000000000",
					"00000000000110000001000000000000",
				),
			},
			stringMask: bitsToUint32(
				"11000011000000000000000000000000",
				"00000000000000011000001100011000",
			),
			level: 3,
			expected: [][]uint32{
				bitsToUint32(
					"00000100000000000000000000000000",
					"00000000000000000000000000000001",
				),
				bitsToUint32(
					"00000100000000000000000000000000",
					"00000000000000100000000000100001",
				),
				bitsToUint32(
					"00000100000000000000000000000000",
					"00000000000000100000010000100001",
				),
			},
		},
	}

	for i, tt := range cases {
		t.Run(fmt.Sprintf("case%d", i+1), func(t *testing.T) {
			actual, err := buildLeveledColonBitmaps(tt.bitmaps, tt.stringMask, tt.level)
			if assert.NoError(t, err) {
				assert.Equal(t, tt.expected, actual)
			}
		})
	}
}

func TestGenerateColonPositions(t *testing.T) {
	cases := []struct {
		index    [][]uint32
		start    int
		end      int
		level    int
		expected []int
	}{
		{
			// {"a":1,"b":{"c":{"d":2},"e":3}}
			// }}3:"e",}2:"d"{:"c"{:"b",1:"a"{
			index: [][]uint32{
				bitsToUint32("00000000000000000000010000010000"),
				bitsToUint32("00001000000000001000010000010000"),
				bitsToUint32("00001000000100001000010000010000"),
			},
			start:    0,
			end:      31,
			level:    0,
			expected: []int{4, 10},
		},
		{
			index: [][]uint32{
				bitsToUint32("00000000000000000000010000010000"),
				bitsToUint32("00001000000000001000010000010000"),
				bitsToUint32("00001000000100001000010000010000"),
			},
			start:    11,
			end:      30,
			level:    1,
			expected: []int{15, 27},
		},
	}

	for i, tt := range cases {
		t.Run(fmt.Sprintf("case%d", i+i), func(t *testing.T) {
			actual := generateColonPositions(tt.index, tt.start, tt.end, tt.level)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestBuildStructualIndex(t *testing.T) {
	cases := []struct {
		input               string
		level               int
		stringMaskBitmap    []uint32
		leveledColonBitmaps [][]uint32
	}{
		{
			input:            `{"a":1,"b":{"c":2}}`,
			level:            2,
			stringMaskBitmap: bitsToUint32("00000000000000000110001100001100"),
			leveledColonBitmaps: [][]uint32{
				bitsToUint32("00000000000000000000010000010000"),
				bitsToUint32("00000000000000001000010000010000"),
			},
		},
		{
			input:            `{"a":1,"b":{"c":2}}`,
			level:            1,
			stringMaskBitmap: bitsToUint32("00000000000000000110001100001100"),
			leveledColonBitmaps: [][]uint32{
				bitsToUint32("00000000000000000000010000010000"),
			},
		},
		{
			input: `                      {"a":1,"b":{"c":{"d":2},"e":3}}`,
			level: 3,
			stringMaskBitmap: bitsToUint32(
				"11000011000000000000000000000000",
				"00000000000000011000001100011000",
			),
			leveledColonBitmaps: [][]uint32{
				bitsToUint32(
					"00000100000000000000000000000000",
					"00000000000000000000000000000001",
				),
				bitsToUint32(
					"00000100000000000000000000000000",
					"00000000000000100000000000100001",
				),
				bitsToUint32(
					"00000100000000000000000000000000",
					"00000000000000100000010000100001",
				),
			},
		},
	}

	for i, tt := range cases {
		t.Run(fmt.Sprintf("case%d", i), func(t *testing.T) {
			actual, err := buildStructualIndex([]byte(tt.input), tt.level)
			if assert.NoError(t, err) {
				expected := &structualIndex{
					json:                []byte(tt.input),
					level:               tt.level,
					stringMaskBitmap:    tt.stringMaskBitmap,
					leveledColonBitmaps: tt.leveledColonBitmaps,
				}
				assert.Equal(t, expected, actual)
			}
		})
	}
}

func TestRetrieveFieldName(t *testing.T) {
	cases := []struct {
		json             []byte
		stringMaskBitmap []uint32
		colon            int
		expected         string
	}{
		{
			json:             []byte(`{"abc":1}`),
			stringMaskBitmap: bitsToUint32("00000000000000000000000000111100"),
			colon:            6,
			expected:         "abc",
		},
		{
			json:             []byte(`{"\\\"abc\"\\":1}`),
			stringMaskBitmap: bitsToUint32("00000000000000000011111111111100"),
			colon:            14,
			expected:         `\"abc"\`,
		},
		{
			json: []byte(`{                         "abc" ` + ` : 1}`),
			stringMaskBitmap: bitsToUint32(
				"01111000000000000000000000000000",
				"00000000000000000000000000000000",
			),
			colon:    33,
			expected: "abc",
		},
		{
			json: []byte(`{                            "ab` + `c":1}`),
			stringMaskBitmap: bitsToUint32(
				"11000000000000000000000000000000",
				"00000000000000000000000000000011",
			),
			colon:    34,
			expected: "abc",
		},
		{
			json: []byte(`{                            "ab` + `                                ` + `c":1}`),
			stringMaskBitmap: bitsToUint32(
				"11000000000000000000000000000000",
				"11111111111111111111111111111111",
				"00000000000000000000000000000011",
			),
			colon:    66,
			expected: "ab                                c",
		},
		{
			json: []byte(`{                              "` + `abc":1}`),
			stringMaskBitmap: bitsToUint32(
				"00000000000000000000000000000000",
				"00000000000000000000000000001111",
			),
			colon:    36,
			expected: "abc",
		},
	}

	for i, tt := range cases {
		t.Run(fmt.Sprintf("case%d", i), func(t *testing.T) {
			actual, err := retrieveFieldName(tt.json, tt.stringMaskBitmap, tt.colon)
			if assert.NoError(t, err) {
				assert.Equalf(t, tt.expected, actual, "expected: %q, actual: %q", tt.expected, actual)
			}
		})
	}
}

func TestBuildQueriedFieldTable(t *testing.T) {
	cases := []struct {
		queriedFields []string
		table         queriedFieldTable
		level         int
	}{
		{
			queriedFields: []string{"abc", "def"},
			table:         queriedFieldTable{"abc": &queriedFieldEntry{id: 0}, "def": &queriedFieldEntry{id: 1}},
			level:         1,
		},
		{
			queriedFields: []string{"abc.def", "abc.ghi", "jkl"},
			table: queriedFieldTable{
				"abc": &queriedFieldEntry{children: queriedFieldTable{"def": &queriedFieldEntry{id: 0}, "ghi": &queriedFieldEntry{id: 1}}},
				"jkl": &queriedFieldEntry{id: 2}},
			level: 2,
		},
		{
			queriedFields: []string{`abc\.def`},
			table: queriedFieldTable{
				"abc.def": &queriedFieldEntry{id: 0},
			},
			level: 1,
		},
		{
			queriedFields: []string{`abc\\.def`},
			table: queriedFieldTable{
				`abc\`: &queriedFieldEntry{children: queriedFieldTable{"def": &queriedFieldEntry{id: 0}}},
			},
			level: 2,
		},
	}

	for i, tt := range cases {
		t.Run(fmt.Sprintf("case%d", i), func(t *testing.T) {
			table, level, err := buildQueriedFieldTable(tt.queriedFields)
			if assert.NoError(t, err) {
				assert.Equal(t, tt.table, table)
				assert.Equal(t, tt.level, level)
			}
		})
	}

	errCases := []struct {
		queriedFields []string
	}{
		{
			queriedFields: []string{"abc", "abc.def"},
		},
		{
			queriedFields: []string{"abc", "abc"},
		},
	}

	for i, tt := range errCases {
		t.Run(fmt.Sprintf("case%d", i), func(t *testing.T) {
			_, _, err := buildQueriedFieldTable(tt.queriedFields)
			assert.Error(t, err)
		})
	}
}

func TestStartParse(t *testing.T) {
	cases := []struct {
		structualIndex *structualIndex
		table          queriedFieldTable
		expected       []*KeyValue
	}{
		{
			structualIndex: &structualIndex{
				json:             []byte(`{"b":2,"c":3,"a":1,}`),
				level:            1,
				stringMaskBitmap: bitsToUint32("00000000000000001100001100001100"),
				leveledColonBitmaps: [][]uint32{
					bitsToUint32("00000000000000010000010000010000"),
				},
			},
			table:    queriedFieldTable{"a": &queriedFieldEntry{id: 0}, "c": &queriedFieldEntry{id: 1}},
			expected: []*KeyValue{{1, 3.0, "3", JSONNumber, nil}, {0, 1.0, "1", JSONNumber, nil}},
		},
		{
			structualIndex: &structualIndex{
				json:             []byte(`{"a":1,"b":{"c":2}}`),
				level:            2,
				stringMaskBitmap: bitsToUint32("00000000000000000110001100001100"),
				leveledColonBitmaps: [][]uint32{
					bitsToUint32("00000000000000000000010000010000"),
					bitsToUint32("00000000000000001000010000010000"),
				},
			},
			table: queriedFieldTable{
				"a": &queriedFieldEntry{id: 0},
				"b": &queriedFieldEntry{children: queriedFieldTable{"c": &queriedFieldEntry{id: 1}}},
			},
			expected: []*KeyValue{{0, 1.0, "1", JSONNumber, nil}, {1, 2.0, "2", JSONNumber, nil}},
		},
		{
			structualIndex: &structualIndex{
				json:             []byte(`{"a":true,"b":false,"c":null}`),
				level:            1,
				stringMaskBitmap: bitsToUint32("00000000011000000001100000001100"),
				leveledColonBitmaps: [][]uint32{
					bitsToUint32("00000000100000000010000000010000"),
				},
			},
			table: queriedFieldTable{
				"a": &queriedFieldEntry{id: 0},
				"b": &queriedFieldEntry{id: 1},
				"c": &queriedFieldEntry{id: 2},
			},
			expected: []*KeyValue{{0, true, "true", JSONBool, nil}, {1, false, "false", JSONBool, nil}, {2, nil, "null", JSONNull, nil}},
		},
		{
			structualIndex: &structualIndex{
				json:             []byte(`{"a":"foo","b":"bar\"\\"}`),
				level:            1,
				stringMaskBitmap: bitsToUint32("00000000111111110011001111001100"),
				leveledColonBitmaps: [][]uint32{
					bitsToUint32("00000000000000000100000000010000"),
				},
			},
			table:    queriedFieldTable{"a": &queriedFieldEntry{id: 0}, "b": &queriedFieldEntry{id: 1}},
			expected: []*KeyValue{{0, "foo", `"foo"`, JSONString, nil}, {1, `bar"\`, `"bar\"\\"`, JSONString, nil}},
		},
		{
			structualIndex: &structualIndex{
				json:             []byte(`{"a":0,"b":1}`),
				level:            2,
				stringMaskBitmap: bitsToUint32("00000000000000000000001100001100"),
				leveledColonBitmaps: [][]uint32{
					bitsToUint32("00000000000000000000010000010000"),
					bitsToUint32("00000000000000000000010000010000"),
				},
			},
			table: queriedFieldTable{
				"a": &queriedFieldEntry{children: queriedFieldTable{"b": &queriedFieldEntry{id: 0}}},
			},
			expected: []*KeyValue{},
		},
		{
			structualIndex: &structualIndex{
				json:             []byte(`{"a":{"b":0}}`),
				level:            1,
				stringMaskBitmap: bitsToUint32("00000000000000000000000110001100"),
				leveledColonBitmaps: [][]uint32{
					bitsToUint32("00000000000000000000000000010000"),
				},
			},
			table:    queriedFieldTable{"a": &queriedFieldEntry{id: 0}},
			expected: []*KeyValue{},
		},
	}

	for i, tt := range cases {
		t.Run(fmt.Sprintf("case%d", i), func(t *testing.T) {
			ch := startParse(tt.structualIndex, tt.table)
			actual := make([]*KeyValue, 0)
			for {
				kv, ok := <-ch
				if !ok {
					break
				}
				actual = append(actual, kv)
			}
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestParserParse(t *testing.T) {
	cases := []struct {
		json          []byte
		queriedFields []string
		expected      []*KeyValue
	}{
		{
			json:          []byte(`{"b":2,"c":3,"a":1,}`),
			queriedFields: []string{"a", "c"},
			expected:      []*KeyValue{{1, 3.0, "3", JSONNumber, nil}, {0, 1.0, "1", JSONNumber, nil}},
		},
		{
			json:          []byte(`{"a":1.0,"b":{"c":2}}`),
			queriedFields: []string{"a", "b.c"},
			expected:      []*KeyValue{{0, 1.0, "1.0", JSONNumber, nil}, {1, 2.0, "2", JSONNumber, nil}},
		},
		{
			json:          []byte(`{"a":true,"b":false,"c":null}`),
			queriedFields: []string{"a", "b", "c"},
			expected:      []*KeyValue{{0, true, "true", JSONBool, nil}, {1, false, "false", JSONBool, nil}, {2, nil, "null", JSONNull, nil}},
		},
		{
			json:          []byte(`{"a":"foo","b":"bar\"\\"}`),
			queriedFields: []string{"a", "b"},
			expected:      []*KeyValue{{0, "foo", `"foo"`, JSONString, nil}, {1, `bar"\`, `"bar\"\\"`, JSONString, nil}},
		},
		{
			json:          []byte(`{"a":0,"b":1}`),
			queriedFields: []string{"a.b"},
			expected:      []*KeyValue{},
		},
		{
			json:          []byte(`{"a":{"b":0}}`),
			queriedFields: []string{"a"},
			expected:      []*KeyValue{},
		},
	}

	for i, tt := range cases {
		t.Run(fmt.Sprintf("case%d", i), func(t *testing.T) {
			p, err := NewParser(tt.queriedFields)
			if assert.NoError(t, err) {
				ch, err := p.Parse(tt.json)
				if assert.NoError(t, err) {
					actual := make([]*KeyValue, 0)
					for {
						kv, ok := <-ch
						if !ok {
							break
						}
						actual = append(actual, kv)
					}
					assert.Equal(t, tt.expected, actual)
				}
			}
		})
	}
}
