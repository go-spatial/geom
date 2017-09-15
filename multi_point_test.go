package geom

import "testing"

func TestMultiPoint(t *testing.T) {
	var (
		mp MultiPointer
	)
	mp = &MultiPoint{[2]float64{10, 20}, [2]float64{30, 40}, [2]float64{-10, -5}}
	mp.Points()
}
