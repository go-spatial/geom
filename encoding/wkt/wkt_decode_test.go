package wkt

import (
	"strings"
	"testing"
	"errors"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
)

func TestDecode(t *testing.T) {
	type tcase struct {
		in  string
		out geom.Geometry
		err error
	}

	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {
			decoder := NewDecoder(strings.NewReader(tc.in))
			out, err := decoder.Decode()
			if (err == nil) != (tc.err == nil) {
				t.Errorf("incorrect error %v, expected %v", err, tc.err)
				return
			} else if err != nil {
				if err.Error() != tc.err.Error() {
					t.Errorf("incorrect error %v, expected %v", err, tc.err)
				}
				return
			}
			if !cmp.GeometryEqual(out, tc.out) {
				t.Errorf("incorrect geometry %v, expected %v", out, tc.out)
			}
		}
	}

	tcases := map[string]tcase{
		"point 0": {
			in:  "POINT(0 0)",
			out: geom.Point{0, 0},
			err: nil,
		},
		"point 1": {
			in: "POINT(0 99.99)",
			out: geom.Point{0, 99.99},
			err: nil,
		},
		"point 2": {
			in: "POINT(99.99 0)",
			out: geom.Point{99.99, 0},
			err: nil,
		},
		"point 3": {
			in: "POINT(99.99 42.0)",
			out: geom.Point{99.99, 42.0},
			err: nil,
		},
		"point 4": {
			in: "POINT()",
			err: errors.New("syntax error (1:8): POINT cannot be empty"),
		},
		"point 5": {
			in: "POINT ( 1 1 )",
			out: geom.Point{1, 1},
			err: nil,
		},
		"point 6": {
			in: "POINT(0 0, 1 1)",
			err: errors.New("syntax error (1:16): too many points in POINT, 2"),
		},
		"point 7": {
			in: "point  \t(\n0 \t\n 0 \f \r   )  ",
			out: geom.Point{0, 0},
		},
		"point 8": {
			in: "POINT(1.3E100 2.3E-35)",
			out: geom.Point{1.3e100, 2.3e-35},
		},
		"multipoint 0": {
			in: "MULTIPOINT()",
			out: geom.MultiPoint{},
		},
		"multipoint 1": {
			in: "MULTIPOINT(0 0, 2 3)",
			out: geom.MultiPoint{{0, 0}, {2, 3}},
		},
		"multipoint 2": {
			in: "MULTIPOINT(0 0, 1 1, 2 3)",
			out: geom.MultiPoint{{0, 0}, {1, 1}, {2, 3}},
		},
		"linestring 0": {
			in: "LINESTRING(0 0,1 1)",
			out: geom.LineString{{0, 0}, {1, 1}},
			err: nil,
		},
		"linestring 1": {
			in: "LINESTRING ( 0\t0\t,\t1\n1 \n \t )",
			out: geom.LineString{{0, 0}, {1, 1}},
			err: nil,
		},
		"linestring 2": {
			in: "LINESTRING(0 0)",
			err: errors.New("syntax error (1:16): not enough points in LINESTRING, 1"),
		},
		"linestring 3": {
			in: "LINESTRING()",
			err: errors.New("syntax error (1:13): not enough points in LINESTRING, 0"),
		},
		"linestring 4": {
			in: "LINESTRING(0 0, 1 1, 2 3)",
			out: geom.LineString{{0, 0}, {1, 1}, {2, 3}},
		},
		"multilinestring 0": {
			in: "MULTILINESTRING((0 0, 1 1))",
			out: geom.MultiLineString{{{0, 0}, {1, 1}}},
		},
		"multilinestring 1": {
			in: "MULTILINESTRING((0 0, 1 1), (2 2, 1 3))",
			out: geom.MultiLineString{{{0, 0}, {1, 1}}, {{2, 2}, {1, 3}}},
		},
		"multilinestring 2": {
			in: "MULTILINESTRING((0 0, 1 1), (1 3))",
			err: errors.New("syntax error (1:35): not enough points in MULTILINESTRING[1], 1"),
		},
		"multilinestring 3": {
			in: "MULTILINESTRING()",
			err: errors.New("syntax error (1:18): not enough lines in MULTILINESTRING, 0"),
		},
		"polygon 0": {
			in: "POLYGON((0 0, 1 1, 1 0, 0 0))",
			out: geom.Polygon{{{0, 0}, {1, 1}, {1, 0}}},
		},
		"polygon 1": {
			in: "POLYGON((0 0, 0 1, 1 1, 1 0))",
			err: errors.New("syntax error (1:30): first and last point of POLYGON[0] not equal"),
		},
		"polygon 2": {
			in: "POLYGON()",
			err: errors.New("syntax error (1:10): not enough lines in POLYGON, 0"),
		},
		"polygon 3": {
			in: "POLYGON((0 0, 1 1, 0 1))",
			err: errors.New("syntax error (1:25): not enough points in POLYGON[0], 3"),
		},
		"polygon 4": {
			in: "POLYGON ((35 10, 45 45, 15 40, 10 20, 35 10),(20 30, 35 35, 30 20, 20 30))",
			out: geom.Polygon{{{35, 10}, {45, 45}, {15, 40}, {10, 20}}, {{20, 30}, {35, 35}, {30, 20}}},
		},
		"multipolygon 0": {
			in: "multipolygon(((0 0, 1 1, 1 0, 0 0)))",
			out: geom.MultiPolygon{{{{0, 0}, {1, 1}, {1, 0}}}},
		},
		"multipolygon 1": {
			in: "multipolygon ( ( ( 0 0, 1 1, 1 0, 0 0 ) ) )",
			out: geom.MultiPolygon{{{{0, 0}, {1, 1}, {1, 0}}}},
		},
		"multipolygon 2": {
			in: "MULTIPOLYGON()",
			err: errors.New("syntax error (1:15): not enough polys in MULTIPOLYGON, 0"),
		},
		"collection 0": {
			in: "geometrycollection(point(0 0))",
			out: geom.Collection{geom.Point{0, 0}},
		},
		"collection 1": {
			in: "geometrycollection ( point ( 0 0 ) )",
			out: geom.Collection{geom.Point{0, 0}},
		},
		"collection 2": {
			in: "geometrycollection(multipolygon(((0 0, 1 1, 1 0, 0 0))))",
			out: geom.Collection{geom.MultiPolygon{{{{0, 0}, {1, 1}, {1, 0}}}}},
		},
		"collection 3": {
			in: "geometrycollection(MULTIPOLYGON())",
			err: errors.New("syntax error (1:34): not enough polys in MULTIPOLYGON, 0"),
		},
	}

	for k, v := range tcases {
		t.Run(k, fn(v))
	}
}
