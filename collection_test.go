package geom

import (
	"testing"
)

func TestCollection(t *testing.T) {
	var (
		c Collectioner
	)
	c = &Collection{Point{10, 20}, LineString{[2]float64{30, 40}, [2]float64{50, 60}}}
	c.Geometries()
}
