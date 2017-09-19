package geom

import (
	"testing"
)

func TestLineString(t *testing.T) {
	var (
		ls LineStringSetter
	)
	points := [][2]float64{[2]float64{15, 20}, [2]float64{35, 40}, [2]float64{-15, -5}}
	ls = &LineString{{10, 20}, {30, 40}, {-10, -5}}
	ls.SetVertexes(points)
	x0 := ls.Vertexes()[0][0]
	if x0 != points[0][0] {
		t.Errorf("Expected x0 to be %v, found %v.", points[0][0], x0)
	}
}
