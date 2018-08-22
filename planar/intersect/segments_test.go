package intersect

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestSearchSegmentIdx(t *testing.T) {

	type tcase struct {
		segments []geom.Line
		seg      geom.Line
		filters  []SegmentFilterFn
		idxs     []int
	}
	fn := func(t *testing.T, tc tcase) {
		ss := NewSearchSegmentIdxs(tc.segments)
		idxs := ss.SearchIntersectIdxs(tc.seg, tc.filters...)
		if !reflect.DeepEqual(tc.idxs, idxs) {
			t.Errorf("idxs, expected %v got %v", tc.idxs, idxs)
			for _, idx := range idxs {
				t.Logf("idx: %v Line: %v\n", idx, tc.segments[idx])
			}
		}
	}
	testSegments := [...][]geom.Line{
		{ // 0
			{{1, 0}, {10, 9}},  // 00
			{{10, 9}, {14, 5}}, // 01
			{{14, 5}, {2, 5}},  // 02
			{{2, 5}, {5, 2}},   // 03
			{{5, 2}, {6, 3}},   // 04
			{{6, 3}, {9, 6}},   // 05
			{{9, 6}, {10, 2}},  // 06
			{{10, 2}, {10, 9}}, // 07
			{{10, 9}, {15, 9}}, // 08
			{{15, 9}, {9, 5}},  // 09
			{{9, 5}, {10, 5}},  // 10
			{{10, 5}, {10, 2}}, // 11
			{{10, 2}, {8, 7}},  // 12
			{{8, 7}, {11, 7}},  // 13
			{{11, 7}, {6, 5}},  // 14
			{{6, 5}, {3, 9}},   // 15
		},
		{ // 1
			{{1, 3}, {1, 1}},   // 01
			{{1, 1}, {4, -4}},  // 02
			{{4, -4}, {8, -4}}, // 03
			{{8, -4}, {8, 5}},  // 04
			{{8, 5}, {3, 5}},   // 05
			{{3, 5}, {1, 3}},   // 06
		},
	}
	tests := [...]tcase{
		{
			seg:      geom.Line{{2, 1}, {13, 3}},
			segments: testSegments[0],
			idxs:     []int{0, 3, 4, 5, 6, 7, 11, 12},
		},
		{
			seg:      geom.Line{{0, 0}, {0, 0}},
			segments: testSegments[0],
			idxs:     nil,
		},
		{
			seg:      geom.Line{{6, 11}, {6, 0}},
			segments: testSegments[0],
			idxs:     []int{0, 2, 4, 5, 14, 15},
		},
		{
			seg:      geom.Line{{10, 9}, {16, 9}},
			segments: testSegments[0],
			idxs:     []int{0, 1, 7, 8, 9},
		},
		{
			seg:      geom.Line{{-1, -4}, {6, -4}},
			segments: testSegments[1],
			idxs:     []int{1, 2},
		},
	}
	for i := range tests {
		tc := tests[i]
		t.Run(strconv.Itoa(i), func(t *testing.T) { fn(t, tc) })
	}

}
