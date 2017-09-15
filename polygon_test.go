package geom

import "testing"

func TestPolygon(t *testing.T) {
	var (
		polygon Polygoner
	)
	polygon = &Polygon{[][2]float64{[2]float64{10, 20},
		[2]float64{30, 40},
		[2]float64{-10, -5},
		[2]float64{10, 20}}}
	polygon.SubLineStrings()
}
