package geom

import "testing"

func TestMultiPoint(t *testing.T) {
	var (
		mp MultiPointer
	)
	mp = &MultiPoint{{10, 20}, {30, 40}, {-10, -5}}
	mp.Points()
}
