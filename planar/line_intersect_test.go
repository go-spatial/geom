package planar

import (
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
)

func TestLineIntersect(t *testing.T) {
	type tcase struct {
		l1, l2 geom.Line
		ok     bool
		pt     [2]float64
	}
	fn := func(t *testing.T, tc tcase) {
		pt, ok := LineIntersect(tc.l1, tc.l2)
		if ok != tc.ok {
			t.Errorf("ok, expected %v got %v", tc.ok, ok)
			return
		}
		if !tc.ok {
			return
		}
		if !cmp.PointEqual(pt, tc.pt) {
			t.Errorf("point, expected %v got %v", tc.pt, pt)
		}
	}

	tests := map[string]tcase{
		"simple": {
			l1: geom.Line{{-10, 0}, {10, 0}},
			l2: geom.Line{{0, 10}, {0, -10}},
			ok: true,
			pt: [2]float64{0, 0},
		},
		"simple line": {
			l1: geom.Line{{-10, 0}, {-1, 0}},
			l2: geom.Line{{0, 10}, {0, -10}},
			ok: true,
			pt: [2]float64{0, 0},
		},
		"parallel line": {
			l1: geom.Line{{-10, 0}, {10, 0}},
			l2: geom.Line{{-10, 1}, {10, 1}},
			ok: false,
		},
		"parallel line 1": {
			l1: geom.Line{{-10, 0}, {-1, 0}},
			l2: geom.Line{{-10, 1}, {10, 1}},
			ok: false,
		},
	}
	for key, tc := range tests {
		tc := tc
		t.Run(key, func(t *testing.T) { fn(t, tc) })
	}
}

func TestSegmentIntersect(t *testing.T) {
	type tcase struct {
		l1, l2 geom.Line
		ok     bool
		pt     [2]float64
	}
	fn := func(t *testing.T, tc tcase) {
		pt, ok := SegmentIntersect(tc.l1, tc.l2)
		if ok != tc.ok {
			t.Errorf("ok, expected %v got %v", tc.ok, ok)
			return
		}
		if !tc.ok {
			return
		}
		if !cmp.PointEqual(pt, tc.pt) {
			t.Errorf("point, expected %v got %v", tc.pt, pt)
		}
	}

	tests := map[string]tcase{
		"simple": {
			l1: geom.Line{{-10, 0}, {10, 0}},
			l2: geom.Line{{0, 10}, {0, -10}},
			ok: true,
			pt: [2]float64{0, 0},
		},
		"simple line not on line": {
			l1: geom.Line{{-10, 0}, {-1, 0}},
			l2: geom.Line{{0, 10}, {0, -10}},
			ok: false,
		},
		"parallel line": {
			l1: geom.Line{{-10, 0}, {10, 0}},
			l2: geom.Line{{-10, 1}, {10, 1}},
			ok: false,
		},
		"parallel line 1": {
			l1: geom.Line{{-10, 0}, {-1, 0}},
			l2: geom.Line{{-10, 1}, {10, 1}},
			ok: false,
		},
		"simple flipped": {
			l1: geom.Line{{10, 0}, {-10, 0}},
			l2: geom.Line{{0, -10}, {0, 10}},
			ok: true,
			pt: [2]float64{0, 0},
		},
		"simple flipped 1": {
			l1: geom.Line{{0, -10}, {0, 10}},
			l2: geom.Line{{10, 0}, {-10, 0}},
			ok: true,
			pt: [2]float64{0, 0},
		},
		"simple line not on line flipped": {
			l1: geom.Line{{-1, 0}, {-10, 0}},
			l2: geom.Line{{0, -10}, {0, 10}},
			ok: false,
		},
		"simple line not on line flipped 1": {
			l1: geom.Line{{0, -10}, {0, 10}},
			l2: geom.Line{{-1, 0}, {-10, 0}},
			ok: false,
		},
		"parallel line flipped": {
			l1: geom.Line{{10, 0}, {-10, 0}},
			l2: geom.Line{{10, 1}, {-10, 1}},
			ok: false,
		},
		"parallel line flipped 1": {
			l1: geom.Line{{10, 1}, {-10, 1}},
			l2: geom.Line{{10, 0}, {-10, 0}},
			ok: false,
		},
		"parallel line 1 flipped": {
			l1: geom.Line{{-1, 0}, {-10, 0}},
			l2: geom.Line{{10, 1}, {-10, 1}},
			ok: false,
		},
		"parallel line 1 flipped 1": {
			l1: geom.Line{{10, 1}, {-10, 1}},
			l2: geom.Line{{-1, 0}, {-10, 0}},
			ok: false,
		},
		"parallel line y 1 flipped 1": {
			l1: geom.Line{{1, 10}, {1, -10}},
			l2: geom.Line{{1, 0}, {-10, 0}},
			ok: true,
			pt: [2]float64{1, 0},
		},
		"triangle test cases for ringcolumns": {
			l1: geom.Line{{1, 1}, {10, 20}},
			l2: geom.Line{{-1, 5.928571428571428}, {13.333333333333334, 5.928571428571428}},
			ok: true,
			pt: [2]float64{3.334586, 5.92857142},
		},
	}
	for key, tc := range tests {
		tc := tc
		t.Run(key, func(t *testing.T) { fn(t, tc) })
	}
}

func TestAreLinesColinear(t *testing.T) {
	type tcase struct {
		l1, l2 geom.Line
		ok     bool
	}
	fn := func(t *testing.T, tc tcase) {
		fl1 := geom.Line{tc.l1[1], tc.l1[0]}
		fl2 := geom.Line{tc.l2[1], tc.l2[0]}
		tests := map[string][2]geom.Line{
			"normal":              {tc.l1, tc.l2},
			"flipped lines":       {tc.l2, tc.l1},
			"flipped l1":          {fl1, tc.l2},
			"filpped l1 lines":    {tc.l2, fl1},
			"flipped l2":          {tc.l1, fl2},
			"flipped l2 lines":    {fl2, tc.l1},
			"flipped l1 l2":       {fl1, fl2},
			"flipped l1 l2 lines": {fl2, fl1},
		}
		for k, v := range tests {
			k, v := k, v
			t.Run(k, func(t *testing.T) {
				ok := AreLinesColinear(v[0], v[1])
				if ok != tc.ok {
					t.Errorf("%v; ok, expected %v got %v", k, tc.ok, ok)
					return
				}
			})
		}
	}
	tests := map[string]tcase{
		"simple": {
			l1: geom.Line{{-10, 0}, {10, 0}},
			l2: geom.Line{{0, 10}, {0, -10}},
			ok: false,
		},
		"simple line not on line": {
			l1: geom.Line{{-10, 0}, {-1, 0}},
			l2: geom.Line{{0, 10}, {0, -10}},
			ok: false,
		},
		"parallel line": {
			l1: geom.Line{{-10, 0}, {10, 0}},
			l2: geom.Line{{-10, 1}, {10, 1}},
			ok: false,
		},
		"same lines": {
			l1: geom.Line{{-10, 1}, {10, 1}},
			l2: geom.Line{{-10, 1}, {10, 1}},
			ok: true,
		},
		"colinear horz lines": {
			l1: geom.Line{{-10, 1}, {5, 1}},
			l2: geom.Line{{1, 1}, {10, 1}},
			ok: true,
		},
		"colinear horz endpoints lines": {
			l1: geom.Line{{-10, 1}, {5, 1}},
			l2: geom.Line{{5, 1}, {10, 1}},
			ok: true,
		},
	}
	for key, tc := range tests {
		tc := tc
		t.Run(key, func(t *testing.T) { fn(t, tc) })
	}

}
