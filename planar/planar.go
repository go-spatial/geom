package planar

import (
	"math"
	"math/big"
)

const Rad = math.Pi / 180

type PointLineDistanceFunc func(line [2][2]float64, point [2]float64) float64

// PerpendicularDistance  provides the distance between a line and a point in Euclidean space.
func PerpendicularDistance(line [2][2]float64, point [2]float64) float64 {

	deltaX := line[1][0] - line[0][0]
	deltaY := line[1][1] - line[0][1]
	denom := math.Abs((deltaY * point[0]) - (deltaX * point[1]) + (line[1][0] * line[0][1]) - (line[1][1] * line[0][0]))
	num := math.Sqrt(math.Pow(deltaY, 2) + math.Pow(deltaX, 2))
	if num == 0 {
		return 0
	}
	return denom / num
}

type Haversine struct {
	Semimajor *float64
}

// PerpendicularDistance returns the distance between a point and a line in meters, using the Harversine distance.
func (hs Haversine) PerpendicularDistance(line [2][2]float64, point [2]float64) float64 {
	dist := hs.Distance(line[0], point)
	angle := Angle3Pts(line[0], line[1], point)
	return math.Abs(dist * math.Sin(angle))
}

// Distance returns the Haversine distance between two points in meters
func (hs Haversine) Distance(pt1, pt2 [2]float64) float64 {
	rpt1x, rpt2x := pt1[0]*Rad, pt2[0]*Rad
	distx := rpt1x - rpt2x
	disty := (pt1[1] * Rad) - (pt2[1] * Rad)
	dist := 2 * math.Asin(
		math.Sqrt(

			(math.Pow(math.Sin(distx/2), 2)+math.Cos(rpt1x))*
				math.Cos(rpt2x)*
				(math.Pow(math.Sin(disty/2), 2)),
		),
	)
	if math.IsNaN(dist) {
		dist = 0
	}

	// https://resources.arcgis.com/en/help/main/10.1/index.html#//003r00000003000000
	// we will default to the semimajor of WGS84
	semimajor := big.NewFloat(6378137)
	if hs.Semimajor != nil {
		semimajor = big.NewFloat(*hs.Semimajor)
	}

	// 4 digits of precision
	d, _ := new(big.Float).SetPrec(16).Mul(big.NewFloat(dist), semimajor).Float64()
	return d
}

// Slope — finds the Slope of a line
func Slope(line [2][2]float64) (m, b float64, defined bool) {
	dx := line[1][0] - line[0][0]
	dy := line[1][1] - line[0][1]
	if dx == 0 || dy == 0 {
		// if dx == 0 then m == 0; and the intercept is y.
		// However if the lines are verticle then the slope is not defined.
		return 0, line[0][1], dx != 0
	}
	m = dy / dx
	b = line[0][1] - (m * line[0][0])
	return m, b, true
}

// Angle3Pts returns the angle between three points…
func Angle3Pts(pt1, pt2, pt3 [2]float64) float64 {
	m1, _, _ := Slope([2][2]float64{pt2, pt1})
	m2, _, _ := Slope([2][2]float64{pt3, pt1})
	// 6 digits of prec
	d, _ := new(big.Float).SetPrec(6*4).Sub(big.NewFloat(m2), big.NewFloat(m1)).Float64()
	return d
}
