package must

import (
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestParseLines(t *testing.T) {
	type tcase struct {
		Content   []byte
		Lines     []geom.Line
		WillPanic bool
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tc.WillPanic {
					t.Errorf("was not expecting panic: %v", r)
				}
			}()

			l := ParseLines(tc.Content)
			if tc.WillPanic {
				t.Errorf("expected panic!")
				return
			}
			if len(l) != len(tc.Lines) {
				t.Errorf("Expected number of lines to be the same.")
				return
			}
			// TODO(gdey): need to test the lines
			return
		}
	}

	tests := [...]tcase{
		{
			Content: []byte(`MULTILINESTRING ((0 10,0 20),(0 20,10 20),(10 20,0 10),(10 20,20 20),(20 20,20 10),(20 10,10 20),(20 10,20 0),(20 0,10 0),(10 0,20 10),(10 0,0 0),(0 0,0 10),(0 10,10 0),(0 10,10 20),(10 0,0 10),(10 0,10 20),(10 20,20 10),(20 10,10 0))`),
			Lines: []geom.Line{
				{{0, 10}, {0, 20}},
				{{0, 20}, {10, 20}},
				{{10, 20}, {0, 10}},
				{{10, 20}, {20, 20}},
				{{20, 20}, {20, 10}},
				{{20, 10}, {10, 20}},
				{{20, 10}, {20, 0}},
				{{20, 0}, {10, 0}},
				{{10, 0}, {20, 10}},
				{{10, 0}, {0, 0}},
				{{0, 0}, {0, 10}},
				{{0, 10}, {10, 0}},
				{{0, 10}, {10, 20}},
				{{10, 0}, {0, 10}},
				{{10, 0}, {10, 20}},
				{{10, 20}, {20, 10}},
				{{20, 10}, {10, 0}},
			},
		},
	}

	for i, tc := range tests {
		t.Run(strconv.FormatInt(int64(i), 10), fn(tc))
	}
}
