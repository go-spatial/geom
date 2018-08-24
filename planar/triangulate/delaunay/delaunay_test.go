package delaunay

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/encoding/wkt"
)

func TestDelaunayTriangulation(t *testing.T) {
	type tcase struct {
		points    []geom.Point
		withFrame bool

		triangles []geom.Triangle
		err       error
	}
	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {
			builder := New(0.0001, tc.points...)
			triangles, err := builder.Triangles(tc.withFrame)

			if tc.err != nil {
				if tc.err != err {
					t.Errorf("error, expected %v got %v", tc.err, err)
				}
				return
			}

			if err != nil {
				t.Errorf("error, expected %v got %v", tc.err, err)
				return
			}

			if !reflect.DeepEqual(tc.triangles, triangles) {
				t.Errorf("triangles,\n\t expected %v\n\t got      %v", wkt.MustEncode(tc.triangles), wkt.MustEncode(triangles))
			}

		}
	}
	tests := [...]tcase{
		{ // simple triangle
			points:    []geom.Point{{10, 10}, {10, 20}, {20, 20}},
			triangles: []geom.Triangle{{{10, 20}, {10, 10}, {20, 20}}},
		},
		{
			points: []geom.Point{{10, 20}, {20, 20}, {20, 10}, {20, 0}, {10, 0}, {0, 0}, {0, 10}, {0, 20}},
			triangles: []geom.Triangle{
				{{0, 20}, {0, 10}, {10, 20}},
				{{10, 20}, {0, 10}, {20, 10}},
				{{10, 20}, {20, 10}, {20, 20}},
				{{10, 0}, {20, 0}, {20, 10}},
				{{10, 0}, {20, 10}, {0, 10}},
				{{10, 0}, {0, 10}, {0, 0}},
			},
		},
	}
	for i, tc := range tests {
		t.Run(strconv.Itoa(i), fn(tc))
	}

}
