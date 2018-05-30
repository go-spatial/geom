package constraineddelaunay

import (
	"errors"
	"fmt"

	"github.com/go-spatial/geom/planar/triangulate/quadedge"
)

var ErrInvalidVertex = errors.New("invalid vertex")
var ErrNoMatchingEdgeFound = errors.New("no matching edge found")

/*
Triangle provides operations on a triangle within a
quadedge.QuadEdgeSubdivision.

This is outside the quadedge package to avoid making changes to the original
JTS port.
*/
type Triangle struct {
	// the triangle referenced is to the right of this edge
	qe *quadedge.QuadEdge
}

/*
IntersectsPoint returns true if the vertex intersects the given triangle. This
includes falling on an edge.

If tri is nil a panic will occur.
*/
func (tri *Triangle) IntersectsPoint(v quadedge.Vertex) bool {
	e := tri.qe

	for i := 0; i < 3; i++ {
		lc := v.Classify(e.Orig(), e.Dest())
		switch lc {
		// return true if v is on the edge
		case quadedge.ORIGIN:
			return true
		case quadedge.DESTINATION:
			return true
		case quadedge.BETWEEN:
			return true
		// return false if v is well outside the triangle
		case quadedge.LEFT:
			return false
		case quadedge.BEHIND:
			return false
		case quadedge.BEYOND:
			return false
		}
		// go to the next edge of the triangle.
		e = e.RNext()
	}

	// if v is to the right of all edges, it is inside the triangle.
	return true
}

/*
opposedTriangle returns the triangle opposite to the vertex v.
       +
      /|\
     / | \
    /  |  \
v1 + a | b +
    \  |  /
     \ | /
      \|/
       +

If this method is called on triangle a with v1 as the vertex, the result will be triangle b.

If tri is nil a panic will occur.
*/
func (tri *Triangle) opposedTriangle(v quadedge.Vertex) (*Triangle, error) {
	qe := tri.qe
	for qe.Orig().Equals(v) == false {

		qe = qe.RNext()

		if qe == tri.qe {
			return nil, ErrInvalidVertex
		}
	}

	return &Triangle{qe.RNext().RNext().Sym()}, nil
}

/*
opposedVertex returns the vertex opposite to this triangle.
       +
      /|\
     / | \
    /  |  \
v1 + a | b + v2
    \  |  /
     \ | /
      \|/
       +

If this method is called as a.opposedVertex(b), the result will be vertex v2.

If tri is nil a panic will occur.
*/
func (tri *Triangle) opposedVertex(other *Triangle) (quadedge.Vertex, error) {
	ae, err := tri.sharedEdge(other)
	if err != nil {
		return quadedge.Vertex{}, err
	}

	// using the matching edge in triangle a, find the opposed vertex in b.
	return ae.Sym().ONext().Dest(), nil
}

/*
sharedEdge returns the edge that is shared by both a and b. The edge is
returned with triangle a on the left.

       + l
      /|\
     / | \
    /  |  \
   + a | b +
    \  |  /
     \ | /
      \|/
       + r

If this method is called as a.sharedEdge(b), the result will be edge lr.

If tri is nil a panic will occur.
*/
func (tri *Triangle) sharedEdge(other *Triangle) (*quadedge.QuadEdge, error) {
	ae := tri.qe
	be := other.qe
	foundMatch := false

	// search for the matching edge between both triangles
	for ai := 0; ai < 3; ai++ {
		for bi := 0; bi < 3; bi++ {
			if ae.Orig().Equals(be.Dest()) && ae.Dest().Equals(be.Orig()) {
				foundMatch = true
				break
			}
			be = be.RNext()
		}

		if foundMatch {
			break
		}
		ae = ae.RNext()
	}

	if foundMatch == false {
		// if there wasn't a matching edge
		return nil, ErrNoMatchingEdgeFound
	}

	// return the matching edge in triangle a
	return ae, nil
}

/*
String returns a string representation of triangle.

If tri is nil a panic will occur.
*/
func (tri *Triangle) String() string {
	str := "["
	e := tri.qe
	comma := ""
	for true {
		str += comma + fmt.Sprintf("%v", e.Orig())
		comma = ","
		e = e.RPrev()
		if e.Orig().Equals(tri.qe.Orig()) {
			break
		}
	}
	str = str + "]"
	return str
}
