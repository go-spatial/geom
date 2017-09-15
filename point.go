package geom

// Point describes a simple 2d point
type Point struct {
	X float64
	Y float64
}

// XY returns an array of 2D coordinates
func (p *Point) XY() [2]float64 {
	return [2]float64{p.X, p.Y}
}

// SetXY sets an array of 2D coordinates
func (p *Point) SetXY(xy [2]float64) {
	p.X = xy[0]
	p.Y = xy[1]
}
