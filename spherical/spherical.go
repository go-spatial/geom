package spherical

import (
	"math"
)

const Rad = math.Pi / 180

// https://resources.arcgis.com/en/help/main/10.1/index.html#//003r00000003000000
// we will default to the semimajor of WGS84
const wgs84Semimajor = 6378137.0

// Angle3Pts returns the angle between three points…
func Angle3Pts(pt1, pt2, pt3 [2]float64) float64 {
	m1 := math.Atan2((pt1[1] - pt2[1]), (pt1[0] - pt2[0]))
	m2 := math.Atan2((pt1[1] - pt3[1]), (pt1[0] - pt3[0]))
	return m2 - m1
}

type Haversine struct {
	Semimajor *float64
}

// semimajor returns a default value if the main one is not set.
func (hs Haversine) semimajor() float64 {
	if hs.Semimajor == nil {
		return wgs84Semimajor
	}
	return *hs.Semimajor
}

// Distance returns the Haversine distance between two points in meters
// ref: https://en.wikipedia.org/wiki/Haversine_formula
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

	return dist * hs.semimajor()
}

// PerpendicularDistance returns the distance between a point and a line in meters, using the Harversine distance.
// cross track distance: https://www.movable-type.co.uk/scripts/latlong.html
func (hs Haversine) PerpendicularDistance(line [2][2]float64, point [2]float64) float64 {
	dist := hs.Distance(line[0], point)
	angle := Angle3Pts(line[0], line[1], point)
	return math.Abs(dist * math.Sin(angle))
}
