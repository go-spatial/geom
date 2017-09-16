package geom

import (
	"testing"
)

func TestCollection(t *testing.T) {
	var (
		c Collectioner
	)
	c = &Collection{&Point{10, 20}, &LineString{{30, 40}, {50, 60}}}
	c.Geometries()
	c.Points()
}
