package wkt

import (
	"strings"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
)

func TestDecode(t *testing.T) {
	type tcase struct {
		in  string
		out geom.Geometry
		err error
	}

	testDecode := func(decoder *Decoder) func(tcase) func(*testing.T) {
		return func(tc tcase) func(*testing.T) {
			return func(t *testing.T) {
				out, err := decoder.Decode()
				if (err == nil) != (tc.err == nil) {
					t.Errorf("error, expected %v, got %v", tc.err, err)
					return
				}
				if err != nil {
					switch tcerr := tc.err.(type) {
					case ErrSyntax:
						eerr, ok := err.(ErrSyntax)
						if !ok {
							t.Errorf("error, expected %v, got %v", tc.err, err)
						}
						if eerr.Issue != tcerr.Issue || eerr.Type != tcerr.Type {
							t.Errorf("error, expected %v:%v got %v:%v", tcerr.Type, tcerr.Issue, eerr.Type, eerr.Issue)
						}

					default:
						if err.Error() != tc.err.Error() {
							t.Errorf("error,  expected %v, got %v", tc.err, err)
						}

					}
					return
				}
				if !cmp.GeometryEqual(out, tc.out) {
					t.Errorf("geometry, expected %v, got %v", tc.out, out)
				}
			}
		}
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			testDecode(NewDecoder(strings.NewReader(tc.in)))(tc)(t)
			t.Run(
				"extra spaces at front",
				testDecode(
					NewDecoder(strings.NewReader("   "+tc.in)),
				)(tc),
			)
			t.Run(
				"extra spaces at the end",
				testDecode(
					NewDecoder(strings.NewReader(tc.in+"   ")),
				)(tc),
			)
		}
	}

	tcases := map[string]tcase{
		"point 0": {
			in:  "POINT(0 0)",
			out: geom.Point{0, 0},
		},
		"point 1": {
			in:  "POINT(0 99.99)",
			out: geom.Point{0, 99.99},
		},
		"point 2": {
			in:  "POINT(99.99 0)",
			out: geom.Point{99.99, 0},
		},
		"point 3": {
			in:  "POINT(99.99 42.0)",
			out: geom.Point{99.99, 42.0},
		},
		"point 4": {
			in: "POINT()",
			err: ErrSyntax{
				Type:  "POINT",
				Issue: "cannot be empty",
			},
		},
		"point 5": {
			in:  "POINT ( 1 1 )",
			out: geom.Point{1, 1},
		},
		"point 6": {
			in: "POINT(0 0, 1 1)",
			err: ErrSyntax{
				Type:  "POINT",
				Issue: "too many points 2",
			},
		},
		"point 7": {
			in:  "point  \t(\n0 \t\n 0 \f \r   )  ",
			out: geom.Point{0, 0},
		},
		"point 8": {
			in:  "POINT(1.3E100 2.3E-35)",
			out: geom.Point{1.3e100, 2.3e-35},
		},
		"multipoint 0": {
			in:  "MULTIPOINT()",
			out: geom.MultiPoint{},
		},
		"multipoint 1": {
			in:  "MULTIPOINT(0 0, 2 3)",
			out: geom.MultiPoint{{0, 0}, {2, 3}},
		},
		"multipoint 2": {
			in:  "MULTIPOINT(0 0, 1 1, 2 3)",
			out: geom.MultiPoint{{0, 0}, {1, 1}, {2, 3}},
		},
		"linestring 0": {
			in:  "LINESTRING(0 0,1 1)",
			out: geom.LineString{{0, 0}, {1, 1}},
		},
		"linestring 1": {
			in:  "LINESTRING ( 0\t0\t,\t1\n1 \n \t )",
			out: geom.LineString{{0, 0}, {1, 1}},
		},
		"linestring 2": {
			in: "LINESTRING(0 0)",
			err: ErrSyntax{
				Type:  "LINESTRING",
				Issue: "not enough points 1",
			},
		},
		"linestring 3": {
			in: "LINESTRING()",
			err: ErrSyntax{
				Type:  "LINESTRING",
				Issue: "not enough points 0",
			},
		},
		"linestring 4": {
			in:  "LINESTRING(0 0, 1 1, 2 3)",
			out: geom.LineString{{0, 0}, {1, 1}, {2, 3}},
		},
		"multilinestring 0": {
			in:  "MULTILINESTRING((0 0, 1 1))",
			out: geom.MultiLineString{{{0, 0}, {1, 1}}},
		},
		"multilinestring 1": {
			in:  "MULTILINESTRING((0 0, 1 1), (2 2, 1 3))",
			out: geom.MultiLineString{{{0, 0}, {1, 1}}, {{2, 2}, {1, 3}}},
		},
		"multilinestring 2": {
			in: "MULTILINESTRING((0 0, 1 1), (1 3))",
			err: ErrSyntax{
				Type:  "MULTILINESTRING",
				Issue: "not enough points in LINESTRING[1], 1",
			},
		},
		"multilinestring 3": {
			in: "MULTILINESTRING()",
			err: ErrSyntax{
				Type:  "MULTILINESTRING",
				Issue: "not enough lines 0",
			},
		},
		"polygon 0": {
			in:  "POLYGON((0 0, 1 1, 1 0, 0 0))",
			out: geom.Polygon{{{0, 0}, {1, 1}, {1, 0}}},
		},
		"polygon 1": {
			in: "POLYGON((0 0, 0 1, 1 1, 1 0))",
			err: ErrSyntax{
				Type:  "POLYGON",
				Issue: "linear-ring[0] not closed",
			},
		},
		"polygon 2": {
			in: "POLYGON()",
			err: ErrSyntax{
				Type:  "POLYGON",
				Issue: "not enough lines 0",
			},
		},
		"polygon 3": {
			in: "POLYGON((0 0, 1 1, 0 1))",
			err: ErrSyntax{
				Type:  "POLYGON",
				Issue: "not enough points in linear-ring[0], 3",
			},
		},
		"polygon 4": {
			in:  "POLYGON ((35 10, 45 45, 15 40, 10 20, 35 10),(20 30, 35 35, 30 20, 20 30))",
			out: geom.Polygon{{{35, 10}, {45, 45}, {15, 40}, {10, 20}}, {{20, 30}, {35, 35}, {30, 20}}},
		},
		"multipolygon 0": {
			in:  "multipolygon(((0 0, 1 1, 1 0, 0 0)))",
			out: geom.MultiPolygon{{{{0, 0}, {1, 1}, {1, 0}}}},
		},
		"multipolygon 1": {
			in:  "multipolygon ( ( ( 0 0, 1 1, 1 0, 0 0 ) ) )",
			out: geom.MultiPolygon{{{{0, 0}, {1, 1}, {1, 0}}}},
		},
		"multipolygon 2": {
			in: "MULTIPOLYGON()",
			err: ErrSyntax{
				Type:  "MULTIPOLYGON",
				Issue: "not enough polygons 0",
			},
		},
		"collection 0": {
			in:  "geometrycollection(point(0 0))",
			out: geom.Collection{geom.Point{0, 0}},
		},
		"collection 1": {
			in:  "geometrycollection ( point ( 0 0 ) )",
			out: geom.Collection{geom.Point{0, 0}},
		},
		"collection 2": {
			in:  "geometrycollection(multipolygon(((0 0, 1 1, 1 0, 0 0))))",
			out: geom.Collection{geom.MultiPolygon{{{{0, 0}, {1, 1}, {1, 0}}}}},
		},
		"collection 3": {
			in: "geometrycollection(MULTIPOLYGON())",
			err: ErrSyntax{
				Type:  "MULTIPOLYGON",
				Issue: "not enough polygons 0",
			},
		},
		"collection 4": {
			in:  "geometrycollection(multipolygon(((0 0, 1 1, 1 0, 0 0))), point(1 1))",
			out: geom.Collection{geom.MultiPolygon{{{{0, 0}, {1, 1}, {1, 0}}}}, geom.Point{1, 1}},
		},
	}

	for k, v := range tcases {
		t.Run(k, fn(v))
	}
}
