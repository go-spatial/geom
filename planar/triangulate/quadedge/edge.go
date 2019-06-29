package quadedge

import (
	"github.com/go-spatial/geom/planar/triangulate/geometry"
)

const (
	precision = 6
)

type Edge struct {
	num  int
	next *Edge
	qe   *QuadEdge
	v    *geometry.Point
}

// NewEdge will return a new edge that is part of an QuadEdge
func New() *Edge {
	ql := NewQEdge()
	return &ql.e[0]
}

func NewWithEndPoints(a, b *geometry.Point) *Edge {
	e := New()
	e.EndPoints(a, b)
	return e
}

func (e *Edge) QEdge() *QuadEdge {
	if e == nil {
		return nil
	}
	return e.qe
}

func (e *Edge) Orig() *geometry.Point {
	if e == nil {
		return nil
	}
	return e.v
}

func (e *Edge) Dest() *geometry.Point {
	return e.Sym().Orig()
}

func (e *Edge) EndPoints(org, dest *geometry.Point) {
	e.v = org
	e.Sym().v = dest
}

/******** Edge Algebra *********************************************************/

// Rot returns the dual of the current edge, directed from its right
// to its left.
func (e *Edge) Rot() *Edge {
	if e == nil {
		return nil
	}
	if e.num == 3 {
		return &(e.qe.e[0])
	}
	return &(e.qe.e[e.num+1])
}

// InvRot returns the dual of the current edge, directed from its left
// to its right.
func (e *Edge) InvRot() *Edge {
	if e == nil {
		return nil
	}
	if e.num == 0 {
		return &(e.qe.e[3])
	}
	return &(e.qe.e[e.num-1])
}

// Sym returns the edge from the destination to the origin of the current edge.
func (e *Edge) Sym() *Edge {
	if e == nil {
		return nil
	}
	if e.num < 2 {
		return &(e.qe.e[e.num+2])
	}
	return &(e.qe.e[e.num-2])
}

// ONext returns the next ccw edge around (from) the origin of the current edge
func (e *Edge) ONext() *Edge {
	if e == nil {
		return nil
	}
	return e.next
}

// OPrev returns the next cw edge around (from) the origin of the currect edge.
func (e *Edge) OPrev() *Edge {
	return e.Rot().ONext().Rot()
}

// DNext returns the next ccw edge around (into) the destination of the current edge.
func (e *Edge) DNext() *Edge {
	return e.Sym().ONext().Sym()
}

// DPrev returns the next cw edge around (into) the destination of the current edge.
func (e *Edge) DPrev() *Edge {
	return e.InvRot().ONext().InvRot()
}

// LNext returns the ccw edge around the left face following the current edge.
func (e *Edge) LNext() *Edge {
	return e.InvRot().ONext().Rot()
}

// LPrev returns the ccw edge around the left face before the current edge.
func (e *Edge) LPrev() *Edge {
	return e.ONext().Sym()
}

// RNext returns the edge around the right face ccw following the current edge.
func (e *Edge) RNext() *Edge {
	return e.Rot().ONext().InvRot()
}

// RPrev returns the edge around the right face ccw before the current edge.
func (e *Edge) RPrev() *Edge {
	return e.Sym().ONext()
}

/*****************************************************************************/
/*         Convenience functions to find edges                                 */
/*****************************************************************************/

// FindONextDest will look for and return a ccw edge the given dest point, if it
// exists.
func (e *Edge) FindONextDest(dest geometry.Point) *Edge {
	if e == nil {
		return nil
	}
	if geometry.ArePointsEqual(dest, *e.Dest()) {
		return e
	}
	for ne := e.ONext(); ne != e; ne = ne.ONext() {
		if geometry.ArePointsEqual(dest, *ne.Dest()) {
			return ne
		}
	}
	return nil
}
