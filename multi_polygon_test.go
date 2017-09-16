package geom

import (
	"testing"
)

func TestMultiPolygon(t *testing.T) {
	var (
		mp MultiPolygoner
	)
	mp = &MultiPolygon{{{{10, 20}, {30, 40}, {-10, -5}, {10, 20}}}}
	mp.Polygons()
	mp.Points()
}
