package geom

import (
	"reflect"
	"testing"
)

func TestCollection(t *testing.T) {
	var (
		c1, c2 CollectionSetter
	)
	c1 = &Collection{&Point{10, 20}, &LineString{{30, 40}, {50, 60}}}
	c2 = &Collection{&Point{15, 25}, &MultiPoint{{35, 45}, {55, 65}}}
	c1.SetGeometries(c2.Geometries())
	if !reflect.DeepEqual(c1, c2) {
		t.Errorf("Output (%+v) does not match expected (%+v).", c1, c2)
	}
	c1.Points()
}
