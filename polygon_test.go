package geom

import "testing"

func TestPolygon(t *testing.T) {
	var (
		polygon Polygoner
	)
	polygon = &Polygon{{{10, 20}, {30, 40}, {-10, -5}, {10, 20}}}
	polygon.SubLineStrings()
	polygon.Points()
}
