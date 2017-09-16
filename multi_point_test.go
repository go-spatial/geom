package geom

import "testing"

func TestMultiPoint(t *testing.T) {
	var (
		mp MultiPointSetter
	)
	points := [][2]float64{[2]float64{15, 20}, [2]float64{35, 40}, [2]float64{-15, -5}}
	mp = &MultiPoint{{10, 20}, {30, 40}, {-10, -5}}
	mp.SetPoints(points)
	x0 := mp.Points()[0][0]
	if x0 != points[0][0] {
		t.Errorf("Expected x0 to be %v, found %v.", points[0][0], x0)
	}
}
