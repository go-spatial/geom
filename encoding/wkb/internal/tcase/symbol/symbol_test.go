package symbol

import (
	"reflect"
	"strings"
	"testing"

	"github.com/go-spatial/geom/internal/parsing"
)

func TestSymbol(t *testing.T) {

	type tcase struct {
		input    string
		expected [][]byte
		err      error
	}
	fn := func(t *testing.T, tc tcase) {
		r := strings.NewReader(tc.input)
		s := parsing.NewScanner(r, SplitFn)
		i := 0
		for s.Scan() {
			if i >= len(tc.expected) {
				t.Errorf("number of entries, expected %v got %v", i, len(tc.expected))
			}
			b := s.RawBytes()
			if !reflect.DeepEqual(tc.expected[i], b) {
				t.Errorf("bytes, expected %v got %v", tc.expected[i], b)
			}
			f, m := s.RawPeek()
			i++
			if m {
				if i >= len(tc.expected) {
					t.Errorf("number of entries, expected %v got %v", len(tc.expected), i)
					break
				}
				if !reflect.DeepEqual(tc.expected[i], f) {
					t.Errorf("M, expected %v got %v", tc.expected[i], f)
				}
			} else {
				if !reflect.DeepEqual([]byte{parsing.EOF}, f) {
					t.Errorf("!M, expected %v got %v", []byte{parsing.EOF}, f)
				}
			}
		}
		if i < len(tc.expected) {
			t.Errorf("number of entries, expected %v got %v", len(tc.expected), i)
			for j := i; j < len(tc.expected); j++ {
				t.Errorf("\tmissing: %v:%v:%v", tc.expected[j][0], string(tc.expected[j][1:]), tc.expected[j])
			}
		}
	}
	tests := map[string]tcase{
		"1": {
			input: ` this is a test.`,
			expected: [][]byte{
				parsing.EncodeSymbol(Space, 1, []byte{' '}),
				parsing.EncodeSymbol(Letter, 4, []byte{'t', 'h', 'i', 's'}),
				parsing.EncodeSymbol(Space, 1, []byte{' '}),
				parsing.EncodeSymbol(Letter, 2, []byte{'i', 's'}),
				parsing.EncodeSymbol(Space, 1, []byte{' '}),
				parsing.EncodeSymbol(Letter, 1, []byte{'a'}),
				parsing.EncodeSymbol(Space, 1, []byte{' '}),
				parsing.EncodeSymbol(Letter, 4, []byte{'t', 'e', 's', 't'}),
				parsing.EncodeSymbol(Dot, 1, []byte{'.'}),
			},
		},
		"2": {
			input: ` { this is a test.}`,
			expected: [][]byte{
				parsing.EncodeSymbol(Space, 1, []byte{' '}),
				parsing.EncodeSymbol(Brace, 1, []byte{'{'}),
				parsing.EncodeSymbol(Space, 1, []byte{' '}),
				parsing.EncodeSymbol(Letter, 4, []byte{'t', 'h', 'i', 's'}),
				parsing.EncodeSymbol(Space, 1, []byte{' '}),
				parsing.EncodeSymbol(Letter, 2, []byte{'i', 's'}),
				parsing.EncodeSymbol(Space, 1, []byte{' '}),
				parsing.EncodeSymbol(Letter, 1, []byte{'a'}),
				parsing.EncodeSymbol(Space, 1, []byte{' '}),
				parsing.EncodeSymbol(Letter, 4, []byte{'t', 'e', 's', 't'}),
				parsing.EncodeSymbol(Dot, 1, []byte{'.'}),
				parsing.EncodeSymbol(Cbrace, 1, []byte{'}'}),
			},
		},
		"3": {
			input: `}`,
			expected: [][]byte{
				parsing.EncodeSymbol(Cbrace, 1, []byte{'}'}),
			},
		},
		"4": {
			input: `{`,
			expected: [][]byte{
				parsing.EncodeSymbol(Brace, 1, []byte{'{'}),
			},
		},
		"5": {
			input: `[`,
			expected: [][]byte{
				parsing.EncodeSymbol(Bracket, 1, []byte{'['}),
			},
		},
		"6": {
			input: `}}`,
			expected: [][]byte{
				parsing.EncodeSymbol(Cdbrace, 1, []byte{'}', '}'}),
			},
		},
		"7": {
			input: `-12.000`,
			expected: [][]byte{
				parsing.EncodeSymbol(Dash, 1, []byte{'-'}),
				parsing.EncodeSymbol(Digit, 2, []byte{'1', '2'}),
				parsing.EncodeSymbol(Dot, 1, []byte{'.'}),
				parsing.EncodeSymbol(Digit, 3, []byte{'0', '0', '0'}),
			},
		},
	}
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}
