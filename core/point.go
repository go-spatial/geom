package core

// Point describes a simple 2d point
type Point struct {
	X float64
	Y float64
}

// X is the x coordinate
func (p Point) XY() (float64, float64) {
	return p.X, p.Y
}
