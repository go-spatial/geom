package simplify

import (
	"context"
	"strings"

	"github.com/go-spatial/geom/planar"
)

type DouglasPeucker struct {

	// Tolerance is the tolerance used to eliminate points, a tolerance of zero is not eliminate any points.
	Tolerance float64

	// Dist is the distance function to use, defaults to planar.PerpendicularDistance
	Dist planar.PointLineDistanceFunc
}

func (dp DouglasPeucker) Simplify(ctx context.Context, linestring [][2]float64, isClosed bool) ([][2]float64, error) {
	return dp.simplify(ctx, 0, linestring, isClosed)
}

func (dp DouglasPeucker) simplify(ctx context.Context, depth uint8, linestring [][2]float64, isClosed bool) ([][2]float64, error) {

	// helper function for debugging and tracing the code
	var printf = func(msg string, depth uint8, params ...interface{}) {
		if debug {
			ps := make([]interface{}, 1, len(params)+1)
			ps[0] = depth
			ps = append(ps, params...)
			logger.Printf(strings.Repeat(" ", int(depth*2))+"[%v]"+msg, ps...)
		}
	}

	if dp.Tolerance <= 0 || len(linestring) <= 2 {
		if debug {
			if dp.Tolerance <= 0 {
				printf("skipping due to Tolerance (%v) ≤ zero:", depth, dp.Tolerance)

			}
			if len(linestring) <= 2 {
				printf("skipping due to len(linestring) (%v) ≤ two:", depth, len(linestring))
			}
		}
		return linestring, nil
	}

	if debug {
		printf("starting linestring: %v ; tolerance: %v", depth, linestring, dp.Tolerance)
	}

	dmax, idx := 0.0, 0
	dist := planar.PerpendicularDistance
	if dp.Dist != nil {
		dist = dp.Dist
	}

	line := [2][2]float64{linestring[0], linestring[len(linestring)-1]}

	if debug {
		printf("starting dmax: %v ; idx %v ;  line : %v", depth, dmax, idx, line)
	}

	// Find the point that is the furthest away.
	for i := 1; i <= len(linestring)-2; i++ {
		d := dist(line, linestring[i])
		if d > dmax {
			dmax, idx = d, i
		}

		if debug {
			printf("looking at %v ; d : %v dmax %v ", depth, i, d, dmax)
		}
	}
	// If the furtherest point is greater then tolerance, we split at that point, and look again at each
	// subsections.
	if dmax > dp.Tolerance {
		if len(linestring) <= 3 {
			if debug {
				printf("returning linestring %v", depth, linestring)
			}
			return linestring, nil
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		rec1, _ := dp.simplify(ctx, depth+1, linestring[0:idx], isClosed)
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		rec2, _ := dp.simplify(ctx, depth+1, linestring[idx:], isClosed)
		if debug {
			printf("returning combined lines: %v %v", depth, rec1, rec2)
		}
		return append(rec1, rec2...), nil
	}

	// Drop all points between the end points.
	if debug {
		printf("dropping all points between the end points: %v", depth, line)
	}
	return line[:], nil
}
