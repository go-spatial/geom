package geom

import (
	"testing"
)

func TestMultiPolygon(t *testing.T) {
	var (
		mp MultiPolygoner
	)
	mp = &MultiPolygon{[][][2]float64{[][2]float64{[2]float64{10, 20}, [2]float64{30, 40}, [2]float64{-10, -5}, [2]float64{10, 20}}}}
	mp.Polygons()
	mp.Points()
}
