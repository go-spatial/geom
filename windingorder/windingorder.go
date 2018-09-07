package windingorder

// WindingOrder is the clockwise direction of a set of points.
type WindingOrder bool

const (
	Clockwise        WindingOrder = false // false is the zero value of bool. We want clockwise to be the default.
	CounterClockwise WindingOrder = true
)

func (w WindingOrder) String() string {
	if w {
		return "counter clockwise"
	}
	return "clockwise"
}

// IsClockwise checks if winding is clockwise
func (w WindingOrder) IsClockwise() bool { return w == Clockwise }

// IsCounterClockwise checks if winding is counter clockwise
func (w WindingOrder) IsCounterClockwise() bool { return w == CounterClockwise }

// Not returns the inverse of the winding
func (w WindingOrder) Not() WindingOrder { return !w }

// OfPoints returns the winding order of the given points
func OfPoints(pts ...[2]float64) WindingOrder {
	sum := 0.0
	li := len(pts) - 1
	for i := range pts[:li] {
		sum += (pts[i][0] * pts[i+1][1]) - (pts[i+1][0] * pts[i][1])
	}
	if sum < 0 {
		return CounterClockwise
	}
	return Clockwise
}
