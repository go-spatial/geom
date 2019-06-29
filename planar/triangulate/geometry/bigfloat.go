// +build bigfloat

package geometry

import "math/big"

const (
	precision = 128
)

var Type = "bigfloat"

type Point [2]*big.Float

func NewPoint(x, y float64) (pt Point) {
	pt[0] = new(big.Float).SetFloat64(x).SetPrec(precision).SetMode(big.ToZero)
	pt[1] = new(big.Float).SetFloat64(y).SetPrec(precision).SetMode(big.ToZero)
	return pt
}

func UnwrapPoint(pt Point) [2]float64 {
	x, _ := pt[0].Float64()
	y, _ := pt[1].Float64()
	return [2]float64{x, y}
}

func IsPointOn(l Line, pt Point) bool {

	x1, x2 := l[0][0], l[1][0]
	if x1.Cmp(x2) > 0 {
		x1, x2 = x2, x1
	}
	y1, y2 := l[0][1], l[1][1]
	if y1.Cmp(y2) > 0 {
		y1, y2 = y2, y1
	}

	// Outside of the extent of the line
	if pt[0].Cmp(x1) < 0 || x2.Cmp(pt[0]) < 0 || pt[1].Cmp(y1) < 0 || y2.Cmp(pt[1]) < 0 {
		return false
	}

	// vertical line
	if x1.Cmp(x2) == 0 {
		return x1.Cmp(pt[0]) == 0 &&
			y1.Cmp(pt[1]) <= 0 && pt[1].Cmp(y2) <= 0
	}

	// horizontal line
	if y1.Cmp(y2) == 0 {
		return y1.Cmp(pt[1]) == 0 &&
			x1.Cmp(pt[0]) <= 0 && pt[0].Cmp(x2) <= 0
	}

	// Match the gradients
	side1 := new(big.Float).Mul(
		new(big.Float).Sub(
			x1,
			pt[0],
		),
		new(big.Float).Sub(
			y1,
			pt[1],
		),
	)
	side2 := new(big.Float).Mul(
		new(big.Float).Sub(
			pt[0],
			x2,
		),
		new(big.Float).Sub(
			pt[1],
			y2,
		),
	)
	return side1.Cmp(side2) == 0

	//	return (x1-pt[0])*(y1-pt[1]) == (pt[0]-x2)*(pt[1]-y2)

}

// ArePointsEqual return if the two points are equal
func ArePointsEqual(a, b Point) bool {

	return a[0].Cmp(b[0]) == 0 && a[1].Cmp(b[1]) == 0
	//	return a[0] == b[0] && a[1] == b[1]
}

// TriArea reaturns twice the area of the oriented triangle (a,b,c), i.e.
// the area is positive if the triangle is oriented counterclockwise.
func triArea(a, b, c Point) *big.Float {

	//return (b[0]-a[0])*(c[1]-a[1]) - (b[1]-a[1])*(c[0]-a[0])

	ba0 := new(big.Float).Sub(b[0], a[0])
	ca0 := new(big.Float).Sub(c[0], a[0])
	ba1 := new(big.Float).Sub(b[1], a[1])
	ca1 := new(big.Float).Sub(c[1], a[1])
	mulba0ca1 := new(big.Float).Mul(ba0, ca1)
	mulba1ca0 := new(big.Float).Mul(ba1, ca0)
	return new(big.Float).Sub(mulba0ca1, mulba1ca0)
}

// TriArea reaturns twice the area of the oriented triangle (a,b,c), i.e.
// the area is positive if the triangle is oriented counterclockwise.
func TriArea(a, b, c Point) float64 {
	result, _ := triArea(a, b, c).Float64()
	return result
}

// InCircle indicates weather the point d is inside the circle defined by the points
// a,b,c. See Guibas and Stolf (1985) p.107
func InCircle(a, b, c, d Point) bool {

	a2 := new(big.Float).Add(
		new(big.Float).Mul(a[0], a[0]),
		new(big.Float).Mul(a[1], a[1]),
	)
	b2 := new(big.Float).Add(
		new(big.Float).Mul(b[0], b[0]),
		new(big.Float).Mul(b[1], b[1]),
	)
	c2 := new(big.Float).Add(
		new(big.Float).Mul(c[0], c[0]),
		new(big.Float).Mul(c[1], c[1]),
	)
	d2 := new(big.Float).Add(
		new(big.Float).Mul(d[0], d[0]),
		new(big.Float).Mul(d[1], d[1]),
	)
	tabcd := new(big.Float).Mul(
		triArea(b, c, d),
		a2,
	)
	taacd := new(big.Float).Mul(
		triArea(a, c, d),
		b2,
	)
	taabd := new(big.Float).Mul(
		triArea(a, b, d),
		c2,
	)
	taabc := new(big.Float).Mul(
		triArea(a, b, c),
		d2,
	)
	result, _ := new(big.Float).Add(
		new(big.Float).Sub(
			tabcd,
			taacd,
		),
		new(big.Float).Sub(
			taabd,
			taabc,
		),
	).Float64()
	return result > 0

	/*
		return (a[0]*a[0]+a[1]*a[1])*TriArea(b, c, d)-
			(b[0]*b[0]+b[1]*b[1])*TriArea(a, c, d)+
			(c[0]*c[0]+c[1]*c[1])*TriArea(a, b, d)-
			(d[0]*d[0]+d[1]*d[1])*TriArea(a, b, c) > 0
	*/
}

func CrossProduct(a, b Point) float64 {
	f, _ := new(big.Float).Sub(
		new(big.Float).Mul(a[0], b[0]),
		new(big.Float).Mul(a[1], b[1]),
	).Float64()
	return f
}

func Dot(a, b Point) float64 {
	f, _ := new(big.Float).Add(
		new(big.Float).Mul(a[0], b[0]),
		new(big.Float).Mul(a[1], b[1]),
	).Float64()
	return f
}

func Sub(a, b Point) Point {
	return Point{
		new(big.Float).Sub(a[0], b[0]),
		new(big.Float).Sub(a[1], b[1]),
	}
}

func Add(a, b Point) Point {
	return Point{
		new(big.Float).Add(a[0], b[0]),
		new(big.Float).Add(a[1], b[1]),
	}
}

func Mul(a, b Point) Point {
	return Point{
		new(big.Float).Mul(a[0], b[0]),
		new(big.Float).Mul(a[1], b[1]),
	}
}

func Magn(a Point) float64 {
	f, _ := new(big.Float).Sqrt(
		new(big.Float).Add(
			new(big.Float).Mul(a[0], a[0]),
			new(big.Float).Mul(a[1], a[1]),
		),
	)
	return f
}

func DivideC(a Point, c float64) Point {
	fc = new(big.Float).SetFloat64(c).SetPrec(precision).SetMode(big.ToZero)
	return Point{
		new(big.Float).Quo(a[0], fc),
		new(big.Float).Quo(a[1], fc),
	}
}
