package geom

import (
	"reflect"
	"testing"
)

func TestPolygon(t *testing.T) {
	var (
		polygon, polygon2 PolygonSetter
	)
	polygon = &Polygon{{{10, 20}, {30, 40}, {-10, -5}, {10, 20}}}
	polygon2 = &Polygon{{{15, 20}, {35, 40}, {-15, -5}, {25, 20}}}
	polygon.SetLineStrings(polygon2.LineStrings())
	if !reflect.DeepEqual(polygon, polygon2) {
		t.Errorf("Output (%+v) does not match expected (%+v).", polygon, polygon2)
	}
	polygon.Points()
}
