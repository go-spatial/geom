package subdivision

import (
	"fmt"

	"github.com/go-spatial/geom/cmp"

	"github.com/go-spatial/geom"
)

type QType uint

const (
	LEFT = QType(iota)
	RIGHT
	BEYOND
	BEHIND
	BETWEEN
	ORIGIN
	DESTINATION
)

func (q QType) String() string {
	switch q {
	case LEFT:
		return "LEFT"
	case RIGHT:
		return "RIGHT"
	case BEYOND:
		return "BEYOND"
	case BEHIND:
		return "BEHIND"
	case BETWEEN:
		return "BETWEEN"
	case ORIGIN:
		return "ORIGIN"
	case DESTINATION:
		return "DESTINATION"
	default:
		return fmt.Sprintf("UNKNOWN(%v)", int(q))
	}
}

// Classify returns where b is in realation to a and c.
func Classify(a, b, c geom.Point) QType {
	aa := c.Subtract(b)
	bb := a.Subtract(b)
	sa := aa.CrossProduct(bb)
	ab := aa.Multiply(bb)

	switch {
	case sa > 0.0:
		return LEFT
	case sa < 0.0:
		return RIGHT
	case ab[0] < 0.0 || ab[1] < 0.0:
		return BEHIND
	case aa.Magnitude() < bb.Magnitude():
		return BEYOND
	case cmp.GeomPointEqual(a, b):
		return ORIGIN
	case cmp.GeomPointEqual(a, c):
		return DESTINATION
	default:
		return BETWEEN
	}
}
