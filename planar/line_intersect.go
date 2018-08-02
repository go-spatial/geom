package planar

import (
	"log"
	"math/big"

	"github.com/go-spatial/geom"
)

const (

	// Experimental testing produced this result.
	// For finding the intersect we need higher precision.
	// Then geom.PrecisionLevelBigFloat
	PrecisionLevelBigFloat = 110
)

func AreLinesColinear(l1, l2 geom.Line) bool {
	x1, y1 := l1.Point1().X(), l1.Point1().Y()
	x2, y2 := l1.Point2().X(), l1.Point2().Y()
	x3, y3 := l2.Point1().X(), l2.Point1().Y()
	x4, y4 := l2.Point2().X(), l2.Point2().Y()

	denom := ((x1 - x2) * (y3 - y4)) - ((y1 - y2) * (x3 - x4))
	// The lines are parallel or they overlap fi denom is 0.
	if denom != 0 {
		return false
	}

	// now we just need to see if one of the end points is on the other one.
	xmin, xmax := x1, x2
	if x1 > x2 {
		xmin, xmax = x2, x1
	}
	ymin, ymax := y1, y2
	if y1 > y2 {
		ymin, ymax = y2, y1
	}

	fn := func(x, y float64) bool { return xmin <= x && x <= xmax && ymin <= y && y <= ymax }
	return fn(x3, y3) || fn(x4, y4)
}

// LineIntersect find the intersection point (x,y) between two lines if there is one. Ok will be true if it found an interseciton point.
// ok being false, means there isn't just one intersection point, there could be zero, or more then one.
// ref: https://en.wikipedia.org/wiki/Line%E2%80%93line_intersection#Given_two_points_on_each_line
func LineIntersect(l1, l2 geom.Line) (pt [2]float64, ok bool) {

	x1, y1 := l1.Point1().X(), l1.Point1().Y()
	x2, y2 := l1.Point2().X(), l1.Point2().Y()
	x3, y3 := l2.Point1().X(), l2.Point1().Y()
	x4, y4 := l2.Point2().X(), l2.Point2().Y()

	denom := ((x1 - x2) * (y3 - y4)) - ((y1 - y2) * (x3 - x4))
	// The lines are parallel or they overlap. No single point.
	if denom == 0 {
		return pt, false
	}

	xnom := (((x1 * y2) - (y1 * x2)) * (x3 - x4)) - ((x1 - x2) * ((x3 * y4) - (y3 * x4)))
	ynom := (((x1 * y2) - (y1 * x2)) * (y3 - y4)) - ((y1 - y2) * ((x3 * y4) - (y3 * x4)))
	return [2]float64{xnom / denom, ynom / denom}, true

}

func bigFloat(f float64) *big.Float { return big.NewFloat(f).SetPrec(PrecisionLevelBigFloat) }

// LineIntersectBigFloat find the intersection point (x,y) between two lines if there is one. Ok will be true if it found an interseciton point. Internally uses math/big
// ok being false, means there isn't just one intersection point, there could be zero, or more then one.
// ref: https://en.wikipedia.org/wiki/Line%E2%80%93line_intersection#Given_two_points_on_each_line
func LineIntersectBigFloat(l1, l2 geom.Line) (pt [2]*big.Float, ok bool) {

	x1, y1 := bigFloat(l1.Point1().X()), bigFloat(l1.Point1().Y())
	x2, y2 := bigFloat(l1.Point2().X()), bigFloat(l1.Point2().Y())
	x3, y3 := bigFloat(l2.Point1().X()), bigFloat(l2.Point1().Y())
	x4, y4 := bigFloat(l2.Point2().X()), bigFloat(l2.Point2().Y())

	deltaX12 := bigFloat(0).Sub(x1, x2)
	deltaX34 := bigFloat(0).Sub(x3, x4)
	deltaY12 := bigFloat(0).Sub(y1, y2)
	deltaY34 := bigFloat(0).Sub(y3, y4)
	denom := bigFloat(0).Sub(
		bigFloat(0).Mul(deltaX12, deltaY34),
		bigFloat(0).Mul(deltaY12, deltaX34),
	)

	// The lines are parallel or they overlap. No single point.
	if d, _ := denom.Float64(); d == 0 {
		return pt, false
	}

	xnom := bigFloat(0).Sub(
		bigFloat(0).Mul(
			bigFloat(0).Sub(
				bigFloat(0).Mul(x1, y2),
				bigFloat(0).Mul(y1, x2),
			),
			deltaX34,
		),
		bigFloat(0).Mul(
			deltaX12,
			bigFloat(0).Sub(
				bigFloat(0).Mul(x3, y4),
				bigFloat(0).Mul(y3, x4),
			),
		),
	)
	ynom := bigFloat(0).Sub(
		bigFloat(0).Mul(
			bigFloat(0).Sub(
				bigFloat(0).Mul(x1, y2),
				bigFloat(0).Mul(y1, x2),
			),
			deltaY34,
		),
		bigFloat(0).Mul(
			deltaY12,
			bigFloat(0).Sub(
				bigFloat(0).Mul(x3, y4),
				bigFloat(0).Mul(y3, x4),
			),
		),
	)
	bx := bigFloat(0).Quo(xnom, denom)
	by := bigFloat(0).Quo(ynom, denom)
	return [2]*big.Float{bx, by}, true

}

// SegmentIntersect finds the intersection point (x,y) between two lines if there is one. Ok will be true if it found a point that is on both line segments, otherwise it will be false.
func SegmentIntersect(l1, l2 geom.Line) (pt [2]float64, ok bool) {
	bpt, ok := LineIntersectBigFloat(l1, l2)
	if !ok {
		if debug {
			log.Printf("Lines don't intersect: %v %v", l1, l2)
		}
		return pt, false
	}
	l1c, l2c := l1.ContainsPointBigFloat(bpt), l2.ContainsPointBigFloat(bpt)
	if debug {
		log.Printf("LineIntersect returns %v %v", bpt, ok)
		log.Printf("line (%v) contains point(%v) :%v ", l1, bpt, l1c)
		log.Printf("line (%v) contains point(%v) :%v ", l2, bpt, l2c)
	}
	x, _ := bpt[0].Float64()
	y, _ := bpt[1].Float64()
	if y == -0 {
		y = 0
	}
	if x == -0 {
		x = 0
	}
	// Check to see if the pt is on both line segments.
	return [2]float64{x, y}, l1c && l2c
}
