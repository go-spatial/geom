package planar

import (
	"math"
)

const Rad = math.Pi / 180

type PointLineDistanceFunc func(line [2][2]float64, point [2]float64) float64

// PerpendicularDistance  provides the distance between a line and a point in Euclidean space.
// ref: https://en.wikipedia.org/wiki/Distance_from_a_point_to_a_line#Line_defined_by_two_points
func PerpendicularDistance(line [2][2]float64, point [2]float64) float64 {

	deltaX := line[1][0] - line[0][0]
	deltaY := line[1][1] - line[0][1]
	deltaXSq := deltaX * deltaX
	deltaYSq := deltaY * deltaY

	num := math.Abs((deltaY * point[0]) - (deltaX * point[1]) + (line[1][0] * line[0][1]) - (line[1][1] * line[0][0]))
	denom := math.Sqrt(deltaYSq + deltaXSq)
	if denom == 0 {
		return 0
	}
	return num / denom
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
