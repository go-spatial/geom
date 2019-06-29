// +build int64

package geometry

import "math"

type Point [2]int64

var Type = "int64"

func NewPoint(x, y float64) (pt Point) {
	pt[0] = int64(x)
	pt[1] = int64(y)
	return pt
}

func UnwrapPoint(pt Point) [2]float64 {
	x := float64(pt[0])
	y := float64(pt[1])
	return [2]float64{x, y}
}

func IsPointOn(l Line, pt Point) bool {

	x1, x2 := l[0][0], l[1][0]
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	y1, y2 := l[0][1], l[1][1]
	if y1 > y2 {
		y1, y2 = y2, y1
	}

	// Outside of the extent of the line
	if pt[0] < x1 || x2 < pt[0] || pt[1] < y1 || y2 < pt[1] {
		return false
	}

	// vertical line
	if x1 == x2 {
		return x1 == pt[0] &&
			y1 <= pt[1] && pt[1] <= y2
	}

	// horizontal line
	if y1 == y2 {
		return y1 == pt[1] &&
			x1 <= pt[0] && pt[0] <= x2
	}

	// Match the gradients
	return (x1-pt[0])*(y1-pt[1]) == (pt[0]-x2)*(pt[1]-y2)

}

// ArePointsEqual return if the two points are equal
func ArePointsEqual(a, b Point) bool {
	return a[0] == b[0] && a[1] == b[1]
}

// TriArea reaturns twice the area of the oriented triangle (a,b,c), i.e.
// the area is positive if the triangle is oriented counterclockwise.
func TriArea(a, b, c Point) float64 {
	return float64(triArea(a, b, c))
}

func triArea(a, b, c Point) int64 {
	return (b[0]-a[0])*(c[1]-a[1]) - (b[1]-a[1])*(c[0]-a[0])
}

// InCircle indicates weather the point d is inside the circle defined by the points
// a,b,c. See Guibas and Stolf (1985) p.107
func InCircle(a, b, c, d Point) bool {

	return (a[0]*a[0]+a[1]*a[1])*triArea(b, c, d)-
		(b[0]*b[0]+b[1]*b[1])*triArea(a, c, d)+
		(c[0]*c[0]+c[1]*c[1])*triArea(a, b, d)-
		(d[0]*d[0]+d[1]*d[1])*triArea(a, b, c) > 0

}

func CrossProduct(a, b Point) float64 {
	return float64((a[0] * b[0]) - (a[1] * b[1]))
}
func Sub(a, b Point) Point {
	return Point{
		a[0] - b[0],
		a[1] - b[1],
	}
}
func Add(a, b Point) Point {
	return Point{
		a[0] + b[0],
		a[1] + b[1],
	}
}

func Mul(a, b Point) Point {
	return Point{
		a[0] * b[0],
		a[1] * b[1],
	}
}

func Magn(a Point) float64 {
	return math.Sqrt(float64((a[0] * a[0]) + (a[1] * a[1])))
}
