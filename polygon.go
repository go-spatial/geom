package geom

type Polygon [][][2]float64

// Sublines returns the lines that make up the polygon.
func (p Polygon) SubLineStrings() [][][2]float64 {
	return p
}
