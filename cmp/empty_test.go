package cmp

import (
	"math"
	"testing"

	"github.com/go-spatial/geom"
)

func TestIsEmptyGeo(t *testing.T) {
	type tcase struct {
		geo     geom.Geometry
		isEmpty bool
		err     string
	}

	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {
			isEmpty, err := IsEmptyGeo(tc.geo)
			if err != nil {
				if err.Error() != tc.err {
					t.Errorf("expected error %v, got %v", tc.err, err)
				}
				return
			} else if tc.err != "" {
				t.Errorf("expected error %v, got nil", tc.err)
				return
			}

			if isEmpty != tc.isEmpty {
				t.Errorf("expected isEmpty %v, got %v", tc.isEmpty, isEmpty)
			}
		}
	}

	tcases := map[string]tcase{
		"point": {
			geo:     geom.Point{},
			isEmpty: false,
		},
		"empty point": {
			geo:     geom.Point{math.NaN(), math.NaN()},
			isEmpty: true,
		},
		"nil point": {
			geo: (*geom.Point)(nil),
			isEmpty: true,
		},
		"non-nil point": {
			geo: &geom.Point{},
			isEmpty: false,
		},
		"multipoint": {
			geo: geom.MultiPoint{geom.Point{}},
			isEmpty: false,
		},
		"empty multipoint": {
			geo: geom.MultiPoint{},
			isEmpty: true,
		},
		"empty multipoint 1": {
			geo: geom.MultiPoint{geom.Point{math.NaN(), math.NaN()}},
			isEmpty: true,
		},
		"non empty multipoint": {
			geo: geom.MultiPoint{
				{},
				geom.Point{math.NaN(), math.NaN()},
			},
			isEmpty: false,
		},
		"nil multipoint": {
			geo: (*geom.MultiPoint)(nil),
			isEmpty: true,
		},
		"linestring": {
			geo: geom.LineString{geom.Point{}},
			isEmpty: false,
		},
		"empty linestring": {
			geo: geom.LineString{},
			isEmpty: true,
		},
		"empty linestring 1": {
			geo: geom.LineString{geom.Point{math.NaN(), math.NaN()}},
			isEmpty: true,
		},
		"non empty linestring": {
			geo: geom.LineString{
				{},
				geom.Point{math.NaN(), math.NaN()},
			},
			isEmpty: false,
		},
		"nil linestring": {
			geo: (*geom.LineString)(nil),
			isEmpty: true,
		},
		"multilinestring": {
			geo: geom.MultiLineString{
				geom.LineString{
					geom.Point{},
				},
			},
			isEmpty: false,
		},
		"empty multilinestring": {
			geo: geom.MultiLineString{},
			isEmpty: true,
		},
		"empty multilinestring 1": {
			geo: geom.MultiLineString{
				geom.LineString{},
			},
			isEmpty: true,
		},
		"empty multilinestring 2": {
			geo: geom.MultiLineString{
				geom.LineString{
					geom.Point{math.NaN(), math.NaN()},
				},
			},
			isEmpty: true,
		},
		"empty multilinestring 3": {
			geo: geom.MultiLineString{
				geom.LineString{
					geom.Point{math.NaN(), math.NaN()},
					geom.Point{math.NaN(), math.NaN()},
				},
				geom.LineString{
					geom.Point{math.NaN(), math.NaN()},
					geom.Point{math.NaN(), math.NaN()},
				},
			},
			isEmpty: true,
		},
		"non empty multilinestring": {
			geo: geom.MultiLineString{
				geom.LineString{
					{},
					geom.Point{math.NaN(), math.NaN()},
				},
			},
			isEmpty: false,
		},
		"non empty multilinestring 1": {
			geo: geom.MultiLineString{
				geom.LineString{},
				geom.LineString{
					{},
					geom.Point{math.NaN(), math.NaN()},
				},
			},
			isEmpty: false,
		},
		"nil multilinestring": {
			geo: (*geom.MultiLineString)(nil),
			isEmpty: true,
		},
		"polygon": {
			geo: geom.Polygon{
				geom.LineString{
					geom.Point{},
				},
			},
			isEmpty: false,
		},
		"empty polygon": {
			geo: geom.Polygon{},
			isEmpty: true,
		},
		"empty polygon 1": {
			geo: geom.Polygon{
				geom.LineString{},
			},
			isEmpty: true,
		},
		"empty polygon 2": {
			geo: geom.Polygon{
				geom.LineString{
					geom.Point{math.NaN(), math.NaN()},
				},
			},
			isEmpty: true,
		},
		"empty polygon 3": {
			geo: geom.Polygon{
				geom.LineString{
					geom.Point{math.NaN(), math.NaN()},
					geom.Point{math.NaN(), math.NaN()},
				},
				geom.LineString{
					geom.Point{math.NaN(), math.NaN()},
					geom.Point{math.NaN(), math.NaN()},
				},
			},
			isEmpty: true,
		},
		"non empty polygon": {
			geo: geom.Polygon{
				geom.LineString{
					{},
					geom.Point{math.NaN(), math.NaN()},
				},
			},
			isEmpty: false,
		},
		"non empty polygon 1": {
			geo: geom.Polygon{
				geom.LineString{},
				geom.LineString{
					{},
					geom.Point{math.NaN(), math.NaN()},
				},
			},
			isEmpty: false,
		},
		"nil polygon": {
			geo: (*geom.Polygon)(nil),
			isEmpty: true,
		},
		"multipolygon": {
			geo: geom.MultiPolygon{
				geom.Polygon{
					geom.LineString{
						geom.Point{},
					},
				},
			},
			isEmpty: false,
		},
		"empty multipolygon": {
			geo: geom.MultiPolygon{},
			isEmpty: true,
		},
		"empty multipolygon 1": {
			geo: geom.MultiPolygon{
				geom.Polygon{
					geom.LineString{},
				},
			},
			isEmpty: true,
		},
		"empty multipolygon 2": {
			geo: geom.MultiPolygon{
				geom.Polygon{
					geom.LineString{
						geom.Point{math.NaN(), math.NaN()},
					},
				},
			},
			isEmpty: true,
		},
		"empty multipolygon 3": {
			geo: geom.MultiPolygon{
				geom.Polygon{
					geom.LineString{
						geom.Point{math.NaN(), math.NaN()},
						geom.Point{math.NaN(), math.NaN()},
					},
					geom.LineString{
						geom.Point{math.NaN(), math.NaN()},
						geom.Point{math.NaN(), math.NaN()},
					},
				},
			},
			isEmpty: true,
		},
		"empty multipolygon 4": {
			geo: geom.MultiPolygon{
				geom.Polygon{
					geom.LineString{
						geom.Point{math.NaN(), math.NaN()},
						geom.Point{math.NaN(), math.NaN()},
					},
					geom.LineString{
						geom.Point{math.NaN(), math.NaN()},
						geom.Point{math.NaN(), math.NaN()},
					},
				},
				geom.Polygon{
					geom.LineString{
						geom.Point{math.NaN(), math.NaN()},
						geom.Point{math.NaN(), math.NaN()},
					},
					geom.LineString{
						geom.Point{math.NaN(), math.NaN()},
						geom.Point{math.NaN(), math.NaN()},
					},
				},
			},
			isEmpty: true,
		},
		"non empty multipolygon": {
			geo: geom.MultiPolygon{
				geom.Polygon{
					geom.LineString{
						{},
						geom.Point{math.NaN(), math.NaN()},
					},
				},
			},
			isEmpty: false,
		},
		"non empty multipolygon 1": {
			geo: geom.MultiPolygon{
				geom.Polygon{
					geom.LineString{},
					geom.LineString{
						geom.Point{math.NaN(), math.NaN()},
						{},
					},
				},
			},
			isEmpty: false,
		},
		"non empty multipolygon 2": {
			geo: geom.MultiPolygon{
				geom.Polygon{
					geom.LineString{
						geom.Point{math.NaN(), math.NaN()},
						geom.Point{math.NaN(), math.NaN()},
					},
					geom.LineString{
						geom.Point{math.NaN(), math.NaN()},
						geom.Point{math.NaN(), math.NaN()},
					},
				},
				geom.Polygon{
					geom.LineString{},
					geom.LineString{
						geom.Point{math.NaN(), math.NaN()},
						{},
					},
				},
			},
			isEmpty: false,
		},
		"nil multipolygon": {
			geo: (*geom.Polygon)(nil),
			isEmpty: true,
		},
		"collection": {
			geo: geom.Collection{
				geom.Point{},
			},
			isEmpty: false,
		},
		"empty collection": {
			geo: geom.Collection{},
			isEmpty: true,
		},
		"empty collection 1": {
			geo: geom.Collection{
				geom.Point{math.NaN(), math.NaN()},
				geom.MultiPoint{},
				geom.LineString{},
				geom.MultiLineString{},
				geom.Polygon{},
				geom.MultiPolygon{},
				geom.Collection{
					geom.Point{math.NaN(), math.NaN()},
				},
			},
			isEmpty: true,
		},
		"non-empty collection": {
			geo: geom.Collection{
				geom.Collection{},
				geom.Collection{
					geom.Point{},
				},
			},
			isEmpty: false,
		},
		"non-empty collection 1": {
			geo: geom.Collection{
				geom.Point{math.NaN(), math.NaN()},
				geom.MultiPoint{},
				geom.LineString{},
				geom.MultiLineString{},
				geom.Polygon{},
				geom.MultiPolygon{},
				geom.Collection{
					geom.Point{math.NaN(), math.NaN()},
					geom.Point{},
				},
			},
			isEmpty: false,
		},
		"nil collection": {
			geo: (*geom.Collection)(nil),
			isEmpty: true,
		},
		"type check": {
			geo: int(0),
			err: "unknown geometry int",
		},
		// non-nil pointers
		"*point": {
			geo: &geom.Point{},
			isEmpty: false,
		},
		"empty *point": {
			geo: &geom.Point{math.NaN(), math.NaN()},
			isEmpty: true,
		},
		"*multipoint": {
			geo: &geom.MultiPoint{
				geom.Point{},
			},
			isEmpty: false,
		},
		"empty *multipoint": {
			geo: &geom.MultiPoint{
				geom.Point{math.NaN(), math.NaN()},
			},
			isEmpty: true,
		},
		"*linestring": {
			geo: &geom.LineString{
				geom.Point{},
			},
			isEmpty: false,
		},
		"empty *linestring": {
			geo: &geom.LineString{
				geom.Point{math.NaN(), math.NaN()},
			},
			isEmpty: true,
		},
		"*multilinestring": {
			geo: &geom.MultiLineString{
				geom.LineString{
					geom.Point{},
				},
			},
			isEmpty: false,
		},
		"empty *multilinestring": {
			geo: &geom.MultiLineString{
				geom.LineString{
					geom.Point{math.NaN(), math.NaN()},
				},
			},
			isEmpty: true,
		},
		"*polygon": {
			geo: &geom.Polygon{
				geom.LineString{
					geom.Point{},
				},
			},
			isEmpty: false,
		},
		"empty *polygon": {
			geo: &geom.Polygon{
				geom.LineString{
					geom.Point{math.NaN(), math.NaN()},
				},
			},
			isEmpty: true,
		},
		"*multipolygon": {
			geo: &geom.MultiPolygon{
				geom.Polygon{
					geom.LineString{
						geom.Point{},
					},
				},
			},
			isEmpty: false,
		},
		"empty *multipolygon": {
			geo: &geom.MultiPolygon{
				geom.Polygon{
					geom.LineString{
						geom.Point{math.NaN(), math.NaN()},
					},
				},
			},
			isEmpty: true,
		},
		"*collection": {
			geo: &geom.Collection{
				&geom.Polygon{
					geom.LineString{
						geom.Point{},
					},
				},
			},
			isEmpty: false,
		},
		"empty *collection": {
			geo: &geom.Collection{
				&geom.Polygon{
					geom.LineString{
						geom.Point{math.NaN(), math.NaN()},
					},
				},
			},
			isEmpty: true,
		},
	}

	for k, v := range tcases {
		t.Run(k, fn(v))
	}
}
