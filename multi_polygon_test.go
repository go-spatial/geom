package geom

import (
	"reflect"
	"testing"
)

func TestMultiPolygon(t *testing.T) {
	var (
		mp, mp2 MultiPolygonSetter
	)
	mp = &MultiPolygon{{{{10, 20}, {30, 40}, {-10, -5}, {10, 20}}}}
	mp2 = &MultiPolygon{{{{15, 20}, {30, 45}, {-15, -5}, {10, 25}}}}
	mp.SetPolygons(mp2.Polygons())
	if !reflect.DeepEqual(mp, mp2) {
		t.Errorf("Output (%+v) does not match expected (%+v).", mp, mp2)
	}
	mp.Points()
}
