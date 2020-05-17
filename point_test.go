package geom_test

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
)

func TestPointSetter(t *testing.T) {
	type tcase struct {
		point    [2]float64
		setter   geom.PointSetter
		expected geom.PointSetter
		err      error
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			err := tc.setter.SetXY(tc.point)
			if tc.err == nil && err != nil {
				t.Errorf("error, expected nil got %v", err)
				return
			}
			if tc.err != nil {
				if tc.err.Error() != err.Error() {
					t.Errorf("error, expected %v got %v", tc.err, err)
				}
				return
			}

			// compare the results
			if !reflect.DeepEqual(tc.expected, tc.setter) {
				t.Errorf("setter, expected %v got %v", tc.expected, tc.setter)
				return
			}
			xy := tc.setter.XY()
			if !reflect.DeepEqual(tc.point, xy) {
				t.Errorf("XY, expected %v, got %v", tc.point, xy)
			}

		}
	}
	tests := []tcase{
		{
			point:    [2]float64{10, 20},
			setter:   &geom.Point{15, 20},
			expected: &geom.Point{10, 20},
		},
		{
			setter: (*geom.Point)(nil),
			err:    geom.ErrNilPoint,
		},
	}

	for i, tc := range tests {
		t.Run(strconv.FormatInt(int64(i), 10), fn(tc))
	}
}

func TestPoint(t *testing.T) {
	// This will test the basics of a point.
	fn := func(pt1, pt2 geom.Point, samePoint bool) (string, func(*testing.T)) {
		return fmt.Sprintf("%v", pt1),
			func(t *testing.T) {
				t.Run("x", func(t *testing.T) {
					x := pt1.X()
					if x != pt1[0] {
						t.Errorf("x, expected %v got %v", pt1[0], x)
					}
				})
				t.Run("maxx", func(t *testing.T) {
					x := pt1.MaxX()
					if x != pt1[0] {
						t.Errorf("x, expected %v got %v", pt1[0], x)
					}
				})
				t.Run("minx", func(t *testing.T) {
					x := pt1.MinX()
					if x != pt1[0] {
						t.Errorf("x, expected %v got %v", pt1[0], x)
					}
				})
				t.Run("y", func(t *testing.T) {
					y := pt1.Y()
					if y != pt1[1] {
						t.Errorf("y, expected %v got %v", pt1[1], y)
					}
				})
				t.Run("maxy", func(t *testing.T) {
					y := pt1.MaxY()
					if y != pt1[1] {
						t.Errorf("maxy, expected %v got %v", pt1[1], y)
					}
				})
				t.Run("miny", func(t *testing.T) {
					y := pt1.MinY()
					if y != pt1[1] {
						t.Errorf("miny, expected %v got %v", pt1[1], y)
					}
				})
				t.Run("miny", func(t *testing.T) {
					area := pt1.Area()
					if area != 0 {
						t.Errorf("area, expected 0 got %v", area)
					}
				})
				t.Run("subtract", func(t *testing.T) {
					pt3 := pt1.Subtract(pt2)
					x := pt1[0] - pt2[0]
					y := pt1[1] - pt2[1]

					if !(cmp.Float(pt3[0], x) && cmp.Float(pt3[1], y)) {
						t.Errorf("subtract, expected (%v,%v) got (%v,%v)", x, y, pt3[0], pt3[1])

					}

				})
				t.Run("multiply", func(t *testing.T) {
					pt3 := pt1.Multiply(pt2)
					x := pt1[0] * pt2[0]
					y := pt1[1] * pt2[1]

					if !(cmp.Float(pt3[0], x) && cmp.Float(pt3[1], y)) {
						t.Errorf("multiply, expected (%v,%v) got (%v,%v)", x, y, pt3[0], pt3[1])
					}
				})
			}
	}
	tests := []geom.Point{
		{0, 1}, {2, 2}, {1, 2},
	}
	for i, pt1 := range tests {
		for j, pt2 := range tests {
			t.Run(fn(pt1, pt2, i == j))
		}
	}

}
