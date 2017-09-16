package geom

import (
	"testing"
)

func TestLineString(t *testing.T) {
	var (
		ls LineStringer
	)
	ls = &LineString{{10, 20}, {30, 40}, {-10, -5}}
	// ls = &LineString{[2]float64{10, 20}, [2]float64{30, 40}, [2]float64{-10, -5}}
	ls.Points()
}
