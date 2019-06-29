// +build float64 

package geometry

type Point [2]float64

var Type = "float64"

func NewPoint(x, y float64) (pt Point) {
	return Point{x,y}
}

func UnwrapPoint(pt Point) [2]float64{
	return [2]float64(pt)
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
	if cmp.Float(x1,x2) {
		return x1 == pt[0] &&
		y1 <= pt[1] && pt[1] <= y2
	}

	// horizontal line
	if cmp.Float(y1,y2) {
		return y1 == pt[1] &&
		x1 <= pt[0] && pt[0] <= x2
	}

	// Match the gradients
	return cmp.Float((x1-pt[0])*(y1-pt[1]) ,(pt[0]-x2)*(pt[1]-y2))

}

// ArePointsEqual return if the two points are equal
func ArePointsEqual(a, b Point) bool {
	return cmp.Float(a[0],b[0]) && cmp.Float(a[1],b[1])
}

// TriArea reaturns twice the area of the oriented triangle (a,b,c), i.e.
// the area is positive if the triangle is oriented counterclockwise.
func TriArea(a, b, c Point) float64 {
	return (b[0]-a[0])*(c[1]-a[1]) - (b[1]-a[1])*(c[0]-a[0])
}

// InCircle indicates weather the point d is inside the circle defined by the points
// a,b,c. See Guibas and Stolf (1985) p.107
func InCircle(a, b, c, d Point) bool {

		return (a[0]*a[0]+a[1]*a[1])*TriArea(b, c, d)-
			(b[0]*b[0]+b[1]*b[1])*TriArea(a, c, d)+
			(c[0]*c[0]+c[1]*c[1])*TriArea(a, b, d)-
			(d[0]*d[0]+d[1]*d[1])*TriArea(a, b, c) > 0

}


