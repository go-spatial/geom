package geom

// Point describes a simple 2d point
type Point struct {
	X float64
	Y float64
}

func (p *Point) XY() [2]float64 {
	return [2]float64{p.X, p.Y}
}
