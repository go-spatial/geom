package geom_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestCollectionSetter(t *testing.T) {
	type tcase struct {
		geoms    []geom.Geometry
		setter   geom.CollectionSetter
		expected geom.CollectionSetter
		err      error
	}
	fn := func(t *testing.T, tc tcase) {
		err := tc.setter.SetGeometries(tc.geoms)
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
		}
		geos := tc.setter.Geometries()
		if !reflect.DeepEqual(tc.geoms, geos) {
			t.Errorf("geometries, expected %v got %v", tc.geoms, geos)
		}
	}
	tests := []tcase{
		{
			geoms: []geom.Geometry{
				&geom.Point{10, 20},
				&geom.LineString{
					{30, 40},
					{50, 60},
				},
			},
			setter: &geom.Collection{
				&geom.Point{15, 25},
				&geom.MultiPoint{
					{35, 45},
					{55, 65},
				},
			},
			expected: &geom.Collection{
				&geom.Point{10, 20},
				&geom.LineString{
					{30, 40},
					{50, 60},
				},
			},
		},
		{
			setter: (*geom.Collection)(nil),
			err:    geom.ErrNilCollection,
		},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}
