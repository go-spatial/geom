package simplify

import "github.com/go-spatial/geom/planar"

type DouglasPeucker struct {

	// Tolerance is the tolerance used to eliminate points, a tolerance of zero is not eliminate any points.
	Tolerance float64

	// Dist is the distance function to use, defaults to planar.PerpendicularDistance
	Dist planar.PointLineDistanceFunc
}

func (dp DouglasPeucker) Simplify(linestring [][2]float64, isClosed bool) ([][2]float64, error) {

	if dp.Tolerance <= 0 || len(linestring) <= 2 {
		return linestring, nil
	}

	dmax, idx := 0.0, 0
	dist := planar.PerpendicularDistance
	if dp.Dist != nil {
		dist = dp.Dist
	}

	line := [2][2]float64{linestring[0], linestring[len(linestring)-1]}

	// Find the point that is the furthest away.
	for i := 1; i <= len(linestring)-2; i++ {
		d := dist(line, linestring[i])
		if d > dmax {
			dmax, idx = d, i
		}
	}

	// If the furtherest point is greater then tolerance, we split at that point, and look again at each
	// subsections.
	if dmax > dp.Tolerance {
		if len(linestring) <= 3 {
			return linestring, nil
		}
		rec1, _ := dp.Simplify(linestring[0:idx], isClosed)
		rec2, _ := dp.Simplify(linestring[idx:], isClosed)
		return append(rec1, rec2...), nil
	}

	// Drop all points between the end points.
	return line[:], nil
}
