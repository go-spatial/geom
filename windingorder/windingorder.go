// Package windingorder provides primitives for determining the winding order of a
// set of points
package windingorder

import (
	"github.com/go-spatial/geom"
	"log"
)

// WindingOrder is the clockwise direction of a set of points.
type WindingOrder uint8

const (

	// Clockwise indicates that the winding order is in the clockwise direction
	Clockwise        WindingOrder = 0
	// Colinear indicates that the points are colinear to each other
	Colinear         WindingOrder = 1
	// CounterClockwise indicates that the winding order is in the counter clockwise direction
	CounterClockwise WindingOrder = 2

	// Collinear alternative spelling of Colinear
	Collinear = Colinear
)

// String implements the stringer interface
func (w WindingOrder) String() string {
	switch w {
	case Clockwise:
		return "clockwise"
	case Colinear:
		return "colinear"
	case CounterClockwise:
		return "counter clockwise"
	default:
		return "unknown"
	}
}

// IsClockwise checks if winding is clockwise
func (w WindingOrder) IsClockwise() bool { return w == Clockwise }

// IsCounterClockwise checks if winding is counter clockwise
func (w WindingOrder) IsCounterClockwise() bool { return w == CounterClockwise }

// IsColinear check if the points are colinear
func (w WindingOrder) IsColinear() bool { return w == Colinear }

// Not returns the inverse of the winding, clockwise <-> counter-clockwise, colinear is it's own
// inverse
func (w WindingOrder) Not() WindingOrder {
	switch w {
	case Clockwise:
		return CounterClockwise
	case CounterClockwise:
		return Clockwise
	default:
		return w
	}
}

// Orient will take the points and calculate the Orientation of the points. by
// summing the normal vectors. It will return 0 of the given points are colinear
// or 1, or -1 for clockwise and counter clockwise depending on the direction of
// the y axis. If the y axis increase as you go up on the graph then clockwise will
// be -1, otherwise it will be 1; vice versa for counter-clockwise.
func Orient(pts ...[2]float64) int8 {
	sum := 0.0

	if len(pts) < 3 {
		return 0
	}
	li := len(pts) - 1

	for i := range pts {
		prd := (pts[li][0] * pts[i][1]) - (pts[i][0] * pts[li][1])
		if debug {
			log.Printf("\t%v : %v x %v : %v",i,pts[li], pts[i], prd)
		}
		sum += prd
		li = i
	}
	if debug {
		log.Printf("sum: %v", sum)
	}
	switch {
	case sum == 0:
		return 0
	case sum < 0:
		return -1
	default:
		return 1
	}
}

// Orientation returns the clockwise orientation of the set of the points given the
// direction of the positive values of the y axis
func Orientation(yPositiveDown bool, pts ...[2]float64) WindingOrder {
	mul := int8(1)
	if !yPositiveDown {
		mul = -1
	}
	switch mul * Orient(pts...) {
	case 0:
		return Colinear
	case 1:
		return Clockwise
	default: // -1
		return CounterClockwise
	}
}

// OfPoints returns the winding order of the given points
func OfPoints(pts ...[2]float64) WindingOrder {
	return Orientation(true, pts...)
}

// OfGeomPoints is the same as OfPoints, just a convenience to unwrap geom.Point
func OfGeomPoints(points ...geom.Point) WindingOrder {
	pts := make([][2]float64, len(points))
	for i := range points {
		pts[i] = [2]float64(points[i])
	}
	return OfPoints(pts...)
}
