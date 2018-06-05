package intersect

import (
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestRingContains(t *testing.T) {
	type pt struct {
		pt        [2]float64
		contained bool
	}
	cpt := func(x, y float64) pt { return pt{pt: [2]float64{x, y}, contained: true} }
	opt := func(x, y float64) pt { return pt{pt: [2]float64{x, y}, contained: false} }
	type tcase struct {
		linestring [][2]float64
		pts        []pt
	}
	fn := func(t *testing.T, tc tcase) {
		var segs []geom.Line
		lp := len(tc.linestring) - 1
		for i := range tc.linestring {
			segs = append(segs, geom.Line{tc.linestring[lp], tc.linestring[i]})
			lp = i
		}
		ring := NewRing(segs)
		for i := range tc.pts {
			i := i
			t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) {
				c := ring.ContainsPoint(tc.pts[i].pt)
				if c != tc.pts[i].contained {
					t.Errorf("containes %v, expected %v got %v", tc.pts[i].pt, tc.pts[i].contained, c)
				}
			})
		}
	}
	tests := map[string]tcase{
		"simple 6 seg": {
			linestring: [][2]float64{{1, 1}, {4, -4}, {8, -4}, {8, 5}, {3, 5}, {1, 3}},
			pts: []pt{
				cpt(2, 2), cpt(2, 3), opt(2, 5), opt(2, -4), opt(6, -4), opt(8, 0),
				opt(20, -4), opt(2, -2), cpt(4, -1), cpt(3, 3)},
		},
		"complicated shape 20x20": {
			linestring: [][2]float64{
				{2, 3}, {4, 3}, {4, 4}, {6, 6}, {9, 6}, {8, 4}, {6, 4},
				{4, 2}, {10, 2}, {10, 4}, {12, 6}, {16, 3}, {16, 4},
				{18, 6}, {18, 8}, {16, 12}, {14, 10}, {16, 8}, {16, 6},
				{12, 11}, {10, 8}, {10, 7}, {8, 7}, {8, 10}, {6, 10},
				{6, 8}, {4, 8}, {4, 12}, {18, 18}, {8, 18}, {2, 12},
				{2, 8}, {4, 6}, {2, 4},
			},
			pts: []pt{
				opt(1, 1), opt(1, 2), opt(1, 3), opt(1, 4), opt(1, 5), opt(1, 6), opt(1, 7), opt(1, 8), opt(1, 9), opt(1, 10),
				opt(1, 11), opt(1, 12), opt(1, 13), opt(1, 14), opt(1, 15), opt(1, 16), opt(1, 17), opt(1, 18), opt(1, 19), opt(1, 20),

				opt(2, 1), opt(2, 2), opt(2, 3), opt(2, 4), opt(2, 5), opt(2, 6), opt(2, 7), opt(2, 8), opt(2, 9), opt(2, 10),
				opt(2, 11), opt(2, 12), opt(2, 13), opt(2, 14), opt(2, 15), opt(2, 16), opt(2, 17), opt(2, 18), opt(2, 19), opt(2, 20),

				opt(3, 1), opt(3, 2), opt(3, 3), cpt(3, 4), opt(3, 5), opt(3, 6), opt(3, 7), cpt(3, 8), cpt(3, 9), cpt(3, 10),
				cpt(3, 11), cpt(3, 12), opt(3, 13), opt(3, 14), opt(3, 15), opt(3, 16), opt(3, 17), opt(3, 18), opt(3, 19), opt(3, 20),

				opt(4, 1), opt(4, 2), opt(4, 3), opt(4, 4), cpt(4, 5), opt(4, 6), cpt(4, 7), opt(4, 8), opt(4, 9), opt(4, 10),
				opt(4, 11), opt(4, 12), cpt(4, 13), opt(4, 14), opt(4, 15), opt(4, 16), opt(4, 17), opt(4, 18), opt(4, 19), opt(4, 20),

				opt(5, 1), opt(5, 2), opt(5, 3), opt(5, 4), opt(5, 5), cpt(5, 6), cpt(5, 7), opt(5, 8), opt(5, 9), opt(5, 10),
				opt(5, 11), opt(5, 12), cpt(5, 13), cpt(5, 14), opt(5, 15), opt(5, 16), opt(5, 17), opt(5, 18), opt(5, 19), opt(5, 20),

				opt(6, 1), opt(6, 2), cpt(6, 3), opt(6, 4), opt(6, 5), opt(6, 6), cpt(6, 7), opt(6, 8), opt(6, 9), opt(6, 10),
				opt(6, 11), opt(6, 12), cpt(6, 13), cpt(6, 14), cpt(6, 15), opt(6, 16), opt(6, 17), opt(6, 18), opt(6, 19), opt(6, 20),

				opt(7, 1), opt(7, 2), cpt(7, 3), opt(7, 4), opt(7, 5), opt(7, 6), cpt(7, 7), cpt(7, 8), cpt(7, 9), opt(7, 10),
				opt(7, 11), opt(7, 12), opt(7, 13), cpt(7, 14), cpt(7, 15), cpt(7, 16), opt(7, 17), opt(7, 18), opt(7, 19), opt(7, 20),

				opt(8, 1), opt(8, 2), cpt(8, 3), opt(8, 4), opt(8, 5), opt(8, 6), opt(8, 7), opt(8, 8), opt(8, 9), opt(8, 10),
				opt(8, 11), opt(8, 12), opt(8, 13), cpt(8, 14), cpt(8, 15), cpt(8, 16), cpt(8, 17), opt(8, 18), opt(8, 19), opt(8, 20),

				opt(9, 1), opt(9, 2), cpt(9, 3), cpt(9, 4), cpt(9, 5), opt(9, 6), opt(9, 7), opt(9, 8), opt(9, 9), opt(9, 10),
				opt(9, 11), opt(9, 12), opt(9, 13), opt(9, 14), cpt(9, 15), cpt(9, 16), cpt(9, 17), opt(9, 18), opt(9, 19), opt(9, 20),

				opt(10, 1), opt(10, 2), opt(10, 3), opt(10, 4), cpt(10, 5), cpt(10, 6), opt(10, 7), opt(10, 8), opt(10, 9), opt(10, 10),
				opt(10, 11), opt(10, 12), opt(10, 13), opt(10, 14), cpt(10, 15), cpt(10, 16), cpt(10, 17), opt(10, 18), opt(10, 19), opt(10, 20),

				opt(11, 1), opt(11, 2), opt(11, 3), opt(11, 4), opt(11, 5), cpt(11, 6), cpt(11, 7), cpt(11, 8), cpt(11, 9), opt(11, 10),
				opt(11, 11), opt(11, 12), opt(11, 13), opt(11, 14), opt(11, 15), cpt(11, 16), cpt(11, 17), opt(11, 18), opt(11, 19), opt(11, 20),

				opt(12, 1), opt(12, 2), opt(12, 3), opt(12, 4), opt(12, 5), opt(12, 6), cpt(12, 7), cpt(12, 8), cpt(12, 9), cpt(12, 10),
				opt(12, 11), opt(12, 12), opt(12, 13), opt(12, 14), opt(12, 15), cpt(12, 16), cpt(12, 17), opt(12, 18), opt(12, 19), opt(12, 20),

				opt(13, 1), opt(13, 2), opt(13, 3), opt(13, 4), opt(13, 5), cpt(13, 6), cpt(13, 7), cpt(13, 8), cpt(13, 9), opt(13, 10),
				opt(13, 11), opt(13, 12), opt(13, 13), opt(13, 14), opt(13, 15), cpt(13, 16), cpt(13, 17), opt(13, 18), opt(13, 19), opt(13, 20),

				opt(14, 1), opt(14, 2), opt(14, 3), opt(14, 4), cpt(14, 5), cpt(14, 6), cpt(14, 7), cpt(14, 8), opt(14, 9), opt(14, 10),
				opt(14, 11), opt(14, 12), opt(14, 13), opt(14, 14), opt(14, 15), opt(14, 16), cpt(14, 17), opt(14, 18), opt(14, 19), opt(14, 20),

				opt(15, 1), opt(15, 2), opt(15, 3), cpt(15, 4), cpt(15, 5), cpt(15, 6), cpt(15, 7), opt(15, 8), opt(15, 9), cpt(15, 10),
				opt(15, 11), opt(15, 12), opt(15, 13), opt(15, 14), opt(15, 15), opt(15, 16), cpt(15, 17), opt(15, 18), opt(15, 19), opt(15, 20),

				opt(16, 1), opt(16, 2), opt(16, 3), opt(16, 4), cpt(16, 5), opt(16, 6), opt(16, 7), opt(16, 8), cpt(16, 9), cpt(16, 10),
				cpt(16, 11), opt(16, 12), opt(16, 13), opt(16, 14), opt(16, 15), opt(16, 16), opt(16, 17), opt(16, 18), opt(16, 19), opt(16, 20),

				opt(17, 1), opt(17, 2), opt(17, 3), opt(17, 4), opt(17, 5), cpt(17, 6), cpt(17, 7), cpt(17, 8), cpt(17, 9), opt(17, 10),
				opt(17, 11), opt(17, 12), opt(17, 13), opt(17, 14), opt(17, 15), opt(17, 16), opt(17, 17), opt(17, 18), opt(17, 19), opt(17, 20),

				opt(18, 1), opt(18, 2), opt(18, 3), opt(18, 4), opt(18, 5), opt(18, 6), opt(18, 7), opt(18, 8), opt(18, 9), opt(18, 10),
				opt(18, 11), opt(18, 12), opt(18, 13), opt(18, 14), opt(18, 15), opt(18, 16), opt(18, 17), opt(18, 18), opt(18, 19), opt(18, 20),

				opt(19, 1), opt(19, 2), opt(19, 3), opt(19, 4), opt(19, 5), opt(19, 6), opt(19, 7), opt(19, 8), opt(19, 9), opt(19, 10),
				opt(19, 11), opt(19, 12), opt(19, 13), opt(19, 14), opt(19, 15), opt(19, 16), opt(19, 17), opt(19, 18), opt(19, 19), opt(19, 20),

				opt(20, 1), opt(20, 2), opt(20, 3), opt(20, 4), opt(20, 5), opt(20, 6), opt(20, 7), opt(20, 8), opt(20, 9), opt(20, 10),
				opt(20, 11), opt(20, 12), opt(20, 13), opt(20, 14), opt(20, 15), opt(20, 16), opt(20, 17), opt(20, 18), opt(20, 19), opt(20, 20),
			},
		},
	}
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}
