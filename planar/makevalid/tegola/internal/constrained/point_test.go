package constrained

import (
	"reflect"
	"strconv"
	"testing"
)

func pt(x, y float64) Point {
	return Point{Pt: [2]float64{x, y}}
}
func cpt(x, y float64) Point {
	return Point{Pt: [2]float64{x, y}, IsConstrained: true}
}

func TestRotatePointsToPos(t *testing.T) {

	type tcase struct {
		pts  []Point
		pos  int
		epts []Point
	}

	fn := func(t *testing.T, tc tcase) {
		RotatePointsToPos(tc.pts, tc.pos)
		if !reflect.DeepEqual(tc.pts, tc.epts) {
			t.Errorf("point seq, expected %v got %v", tc.epts, tc.pts)
		}
	}
	tests := [...]tcase{
		{
			pts:  []Point{pt(5, 4), pt(2, 2), pt(1, 1), pt(7, 2), pt(5, 3)},
			pos:  0,
			epts: []Point{pt(5, 4), pt(2, 2), pt(1, 1), pt(7, 2), pt(5, 3)},
		},
		{
			pts:  []Point{pt(5, 4), pt(2, 2), pt(1, 1), pt(7, 2), pt(5, 3)},
			pos:  6,
			epts: []Point{pt(5, 4), pt(2, 2), pt(1, 1), pt(7, 2), pt(5, 3)},
		},
		{
			pts:  []Point{pt(5, 4), pt(2, 2), pt(1, 1), pt(7, 2), pt(5, 3)},
			pos:  3,
			epts: []Point{pt(7, 2), pt(5, 3), pt(5, 4), pt(2, 2), pt(1, 1)},
		},
		{
			pts:  []Point{pt(5, 4), pt(2, 2), pt(1, 1), pt(7, 2), pt(5, 3)},
			pos:  2,
			epts: []Point{pt(1, 1), pt(7, 2), pt(5, 3), pt(5, 4), pt(2, 2)},
		},
	}
	for i := range tests {
		tc := tests[i]
		t.Run(strconv.Itoa(i), func(t *testing.T) { fn(t, tc) })
	}
}

func TestRotatePointsToLowestFirst(t *testing.T) {

	type tcase struct {
		pts  []Point
		epts []Point
	}

	fn := func(t *testing.T, tc tcase) {
		RotatePointsToLowestFirst(tc.pts)
		if !reflect.DeepEqual(tc.pts, tc.epts) {
			t.Errorf("point seq, expected %v got %v", tc.epts, tc.pts)
		}
	}
	tests := [...]tcase{
		{
			pts:  []Point{pt(5, 4), pt(2, 2), pt(1, 1), pt(7, 2), pt(5, 3)},
			epts: []Point{pt(1, 1), pt(7, 2), pt(5, 3), pt(5, 4), pt(2, 2)},
		},
		{
			pts:  []Point{pt(5, 4), pt(2, 2), pt(1, 1), pt(7, 2), pt(0, 0)},
			epts: []Point{pt(0, 0), pt(5, 4), pt(2, 2), pt(1, 1), pt(7, 2)},
		},
		{
			pts:  []Point{pt(5, 4)},
			epts: []Point{pt(5, 4)},
		},
		{
			pts:  []Point{},
			epts: []Point{},
		},
	}
	for i := range tests {
		tc := tests[i]
		t.Run(strconv.Itoa(i), func(t *testing.T) { fn(t, tc) })
	}
}
