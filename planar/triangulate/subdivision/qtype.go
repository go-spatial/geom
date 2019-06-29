package subdivision

import (
	"fmt"
	"log"

	"github.com/go-spatial/geom/planar/triangulate/geometry"
	"github.com/go-spatial/geom/planar/triangulate/quadedge"
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

func TestClassify(lbl string, a, b, c geometry.Point) QType {
	lc := Classify(a, b, c)
	log.Println(lbl, "a", a)
	log.Println(lbl, lc)
	va := quadedge.Vertex(geometry.UnwrapPoint(a))
	vb := quadedge.Vertex(geometry.UnwrapPoint(b))
	vc := quadedge.Vertex(geometry.UnwrapPoint(c))
	log.Println(lbl, "geomvertex classify", va.Classify(vb, vc))
	return lc
}

func Classify(a, b, c geometry.Point) QType {
	aa := geometry.Sub(c, b)
	bb := geometry.Sub(a, b)

	sa := geometry.CrossProduct(aa, bb)

	mab := geometry.Mul(aa, bb)

	mabuw := geometry.UnwrapPoint(mab)

	switch {
	case sa > 0.0:
		return LEFT
	case sa < 0.0:
		return RIGHT
	case mabuw[0] < 0.0 || mabuw[1] < 0.0:
		return BEHIND
	case geometry.Magn(aa) < geometry.Magn(bb):
		return BEYOND
	case geometry.ArePointsEqual(a, b):
		return ORIGIN
	case geometry.ArePointsEqual(a, c):
		return DESTINATION
	default:
		return BETWEEN
	}
}
