package geom

import (
	"testing"
)

func TestPoint(t *testing.T) {
	var (
		point PointSetter
	)
	point = &Point{X: 10, Y: 20}
	point.XY()
	point.SetXY([2]float64{30, 40})
	xy := point.XY()
	if xy[0] != 30 {
		t.Errorf("Expected 30, received %v", xy[0])
	}
}
