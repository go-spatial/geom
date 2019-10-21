//go:generate stringer -type=Orientation -linecomment

package planar

import (
	"github.com/go-spatial/geom"
	"log"
)

type Orientation int8

const (

	//ClockwiseOrientation represents a clockwise orientation
	ClockwiseOrientation Orientation = -1 // clockwise
	//CoLinearOrientation represents a colinear orientation
	CoLinearOrientation Orientation = 0 // colinear
	//CounterClockwiseOrientation represents a coutner clockwise orientation
	CounterClockwiseOrientation Orientation = 1 // counter clockwise

)

// Orient will return the orientation of the three given points
func Orient(a, b, c geom.Point) Orientation {
	area := geom.Triangle{[2]float64(a), [2]float64(b), [2]float64(c)}.Area()

	/*
		sum := 0.0
		li := len(pts) - 1
		for i := range pts[:li] {
			sum += (pts[i][0] * pts[i+1][1]) - (pts[i+1][0] * pts[i][1])
		}

	*/

	if debug {
		log.Printf("area(%v,%v,%v): %v", a, b, c, area)
	}
	if area == 0 {
		return CoLinearOrientation
	}
	if area < 0 {
		return ClockwiseOrientation
	}
	return CounterClockwiseOrientation
}

// IsCCW will check if the given points are in counter-clockwise order.
func IsCCW(a, b, c geom.Point) bool {
	//log.Printf("windingorder %v",windingorder.OfPoints([2]float64(a), [2]float64(b), [2]float64(c)) == windingorder.CounterClockwise)
	return Orient(a, b, c) == CounterClockwiseOrientation
}

// OrientationInRegardsTo will return the orientation of the points a,b,c around the origin point
// This will make use of two triangle and the following truth table.
//  ccw (counter clockwise) area of triangle > 0
//  cw  (clockwise) area of triangle < 0
//  cl  (colinear) area of triangle == 0
//  +---+-----+-----+-----------+-------+
//  | # | abo | bco | ccw/cw/cl | IsCCW |
//  +---+-----+-----+-----------+-------+
//  | 0 | ccw | ccw |    ccw    | yes   |
//  | 1 | cl  | cl  |    cl     | no    |
//  | 2 | cw  | cw  |    cw     | no    |
//  | 3 | cl  | ccw |    ccw    | yes   |
//  | 4 | cl  | cw  |    cw     | no    |
//  | 5 | ccw | cl  |    ccw    | yes   |
//  | 6 | cw  | cl  |    cw     | no    |
//  | 7 | ccw | cw  |    cw     | no    |
//  | 8 | cw  | ccw |    cw     | no    |
//  +---+-----+-----+-----------+-------+
//
func OrientationInRegardsTo(origin, a, b, c geom.Point) Orientation {

	abo := Orient(a, b, origin)
	bco := Orient(b, c, origin)

	// cases 0-2
	if abo == bco {
		return abo
	}

	// cases 3,4
	if abo == CoLinearOrientation {
		return bco
	}

	// cases 5,6
	if bco == CoLinearOrientation {
		return abo
	}

	// last two cases (7,8) always clockwise
	return ClockwiseOrientation
}

// IsCCWWithRegardsTo will calculate is the three points a,b,c are counter clockwise around the origin
func IsCCWWithRegardsTo(origin, a, b, c geom.Point) bool {
	return OrientationInRegardsTo(origin, a, b, c) == CounterClockwiseOrientation
}
