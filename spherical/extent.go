package spherical

import (
	"math"

	"github.com/go-spatial/geom"
)

// Hull returns the smallest extent from lon/lat points.
// The hull is defined in the following order [4]float64{ West, South, East, North}
func Hull(a, b [2]float64) *geom.Extent {
	// lat <=> y
	// lng <=> x

	// make a the westmost point
	if math.Abs(a[0]-b[0]) > 180.0 {
		// smallest longitudinal arc crosses the antimeridian
		if a[0] < b[0] {
			a[0], b[0] = b[0], a[0]
		}
	} else {
		if a[0] > b[0] {
			a[0], b[0] = b[0], a[0]
		}
	}

	return Extent(a, b)
}

// Segment of a sphere from two long/lat points, with a being the westmost point and b the eastmost point; in following format [4]float64{ West, South, East, North }
func Extent(westy, easty [2]float64) *geom.Extent {
	north, south := westy[1], easty[1]
	if north < south {
		south, north = north, south
	}

	return &geom.Extent{westy[0], south, easty[0], north}
}
