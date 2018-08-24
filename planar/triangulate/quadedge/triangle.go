package quadedge

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/go-spatial/geom"
)

type ErrInvalidVertex struct {
	V Vertex
	T Triangle
}

func (e ErrInvalidVertex) Error() string {
	return fmt.Sprintf("invalid vertex: %v in %v", e.V, e.T)
}

type ErrNoMatchingEdgeFound struct {
	T1 Triangle
	T2 Triangle
}

func (e ErrNoMatchingEdgeFound) Error() string {
	return fmt.Sprintf("no matching edge found T1: %v T2: %v", e.T1, e.T2)
}

// Triangle provides operations on a triangle within a QuadEdgeSubdivision.
// 	the triangle referenced is to the right of this edge
type Triangle struct {
	QE *QuadEdge
}

// IntersectsPoint returns true if the vertex intersects the given triangle. This
// includes falling on an edge.
//
// If tri is nil a panic will occur.
func (tri Triangle) IntersectsPoint(v Vertex) bool {

	if tri.QE == nil {
		return false
	}
	e := tri.QE

	for i := 0; i < 3; i++ {
		lc := v.Classify(e.Orig(), e.Dest())
		switch lc {
		// return true if v is on the edge
		case ORIGIN:
			return true
		case DESTINATION:
			return true
		case BETWEEN:
			return true
		// return false if v is well outside the triangle
		case LEFT:
			return false
		case BEHIND:
			return false
		case BEYOND:
			return false
		}
		// go to the next edge of the triangle.
		e = e.RNext()
	}

	// if v is to the right of all edges, it is inside the triangle.
	return true
}

// GetStartEdge returns the 'starting' edge of this triangle. Unless Normalize
// has been called the edge is arbitrary.
func (tri Triangle) GetStartEdge() *QuadEdge { return tri.QE }

// IsValid returns true if the specified triangle has three sides that connect.
func (tri Triangle) IsValid() bool {
	if tri.QE == nil {
		return false
	}

	count := 0

	s := tri.QE
	e := s
	for {
		e = e.RPrev()
		count++
		if e.Orig().Equals(s.Orig()) {
			break
		}
		if count > 3 {
			return false
		}
	}
	return count == 3
}

//String returns a string representation of triangle.
//If IsValid is false you may get a "triangle" with more or less than three
//sides.
func (tri Triangle) String() string {
	pts := tri.Points()
	strPts := make([]string, len(pts))
	for i := range pts {
		strPts[i] = `[` + strconv.FormatFloat(pts[i][0], 'f', 2, 64) + " " + strconv.FormatFloat(pts[i][1], 'f', 2, 64) + `]`
	}
	return "[" + strings.Join(strPts, ",") + "]"
}

func (tri Triangle) Points() []geom.Point {
	s := tri.QE
	if s == nil {
		return nil
	}
	pts := make([]geom.Point, 0, 3)
	e := s
	for {
		pts = append(pts, geom.Point(e.Orig()))
		e = e.RPrev()
		if e.Orig().Equals(s.Orig()) {
			break
		}
	}
	return pts
}

// Opposed returns the triangle opposite to the vertex v
//        +
//       /|\
//      / | \
//     /  |  \
// v1 + a | b +
//     \  |  /
//      \ | /
//       \|/
//        +
//
// If this method is called on triangle a with v1 as the vertex, the result will be triangle b.
func (tri Triangle) OpposedTriangle(v Vertex) (Triangle, error) {
	s := tri.QE
	qe := s
	for !qe.Orig().Equals(v) {
		qe = qe.RNext()
		if qe == s {
			return Triangle{}, ErrInvalidVertex{v, tri}
		}
	}
	return Triangle{QE: qe.RNext().RNext().Sym()}, nil
}

// OpposedVertex returns the vertex opposite to this triangle.
//        +
//       /|\
//      / | \
//     /  |  \
// v1 + a | b + v2
//     \  |  /
//      \ | /
//       \|/
//        +
//
// If this method is called as a.opposedVertex(b), the result will be vertex v2.
func (tri Triangle) OpposedVertex(other Triangle) (Vertex, error) {
	ae, err := tri.SharedEdge(other)
	if err != nil {
		return Vertex{}, err
	}
	// using the matching edge in triangle a, find the opposed vertex in b.
	return ae.Sym().ONext().Dest(), nil
}

// SharedEdge returns the edge that is shared by both a and b. The edge is
// returned with triangle a on the left.
//
//        + l
//       /|\
//      / | \
//     /  |  \
//    + a | b +
//     \  |  /
//      \ | /
//       \|/
//        + r
//
// If this method is called as a.sharedEdge(b), the result will be edge lr.
//
func (tri Triangle) SharedEdge(other Triangle) (*QuadEdge, error) {

	ae := tri.QE
	be := other.QE

	for ai := 0; ai < 3; ai, ae = ai+1, ae.RNext() {
		for bi := 0; bi < 3; bi, be = bi+1, be.RNext() {
			if debug {
				log.Printf("Looking at %v %v === %v %v", ae.Orig(), ae.Dest(), be.Orig(), be.Dest())
			}
			if ae.Orig().Equals(be.Dest()) && ae.Dest().Equals(be.Orig()) {
				return ae, nil
			}
		}
	}
	return nil, ErrNoMatchingEdgeFound{T1: tri, T2: other}
}
