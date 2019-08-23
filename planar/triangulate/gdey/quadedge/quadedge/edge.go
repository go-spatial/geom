package quadedge

import (
	"fmt"
	"github.com/go-spatial/geom/windingorder"
	"log"

	"github.com/go-spatial/geom/planar"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
	"github.com/go-spatial/geom/encoding/wkt"
)

const (
	precision = 6
)

// Edge describes a directional edge in a quadedge
type Edge struct {
	num  int
	next *Edge
	qe   *QuadEdge
	v    *geom.Point
}

// New will return a new edge that is part of an QuadEdge
func New() *Edge {
	ql := NewQEdge()
	return &ql.e[0]
}

// NewWithEndPoints creates a new edge with the given end points
func NewWithEndPoints(a, b *geom.Point) *Edge {
	e := New()
	e.EndPoints(a, b)
	return e
}

// QEdge returns the quadedge this edge is part of
func (e *Edge) QEdge() *QuadEdge {
	if e == nil {
		return nil
	}
	return e.qe
}

// Orig returns the origin end point
func (e *Edge) Orig() *geom.Point {
	if e == nil {
		return nil
	}
	return e.v
}

// Dest returns the destination end point
func (e *Edge) Dest() *geom.Point {
	return e.Sym().Orig()
}

// EndPoints sets the end points of the Edge
func (e *Edge) EndPoints(org, dest *geom.Point) {
	e.v = org
	e.Sym().v = dest
}

// AsLine returns the Edge as a geom.Line
func (e *Edge) AsLine() geom.Line {
	porig, pdest := e.Orig(), e.Dest()
	orig, dest := geom.EmptyPoint, geom.EmptyPoint
	if porig != nil {
		orig = *porig
	}
	if pdest != nil {
		dest = *pdest
	}
	return geom.Line{[2]float64(orig), [2]float64(dest)}
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

// OPrev returns the next cw edge around (from) the origin of the current edge.
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
func (e *Edge) FindONextDest(dest geom.Point) *Edge {
	if e == nil {
		return nil
	}
	if cmp.GeomPointEqual(dest, *e.Dest()) {
		return e
	}
	for ne := e.ONext(); ne != e; ne = ne.ONext() {
		if cmp.GeomPointEqual(dest, *ne.Dest()) {
			return ne
		}
	}
	return nil
}

// DumpAllEdges dumps all the edges as a multiline string
func (e *Edge) DumpAllEdges() string {
	var ml geom.MultiLineString

	e.WalkAllONext(func(ee *Edge) bool {
		ln := ee.AsLine()
		ml = append(ml, [][2]float64{ln[0], ln[1]})
		return true
	})
	return wkt.MustEncode(ml)
}

func (e *Edge) WalkAllOPrev(fn func(*Edge) (loop bool)) {
	if !fn(e) {
		return
	}
	cwe := e.OPrev()
	for cwe != e {
		if !fn(cwe) {
			return
		}
		cwe = cwe.OPrev()
	}

	// for cwe := e.OPrev(); cwe != e && fn(cwe) ; cwe = e.OPrev(){}

}
func (e *Edge) WalkAllONext(fn func(*Edge) (loop bool)) {
	if !fn(e) {
		return
	}
	ccwe := e.ONext()
	count := 0
	for ccwe != e {
		if !fn(ccwe) {
			return
		}
		ccwe = ccwe.ONext()
		count++
		if count == 30 {
			panic("inifite loop")
		}
	}
}

// IsEqual checks to see if the edges are the same
func (e *Edge) IsEqual(e1 *Edge) bool {
	if e == nil {
		return e1 == nil
	}

	if e1 == nil {
		return e == nil
	}
	// first let's get the edge numbers the same
	return e == &e1.qe.e[e.num]
}

// Validate check to se if the edges in the edges are correctly
// oriented
func Validate(e *Edge) (err1 error) {

	const radius = 10
	var err ErrInvalid

	// TODO (gdey): fix this
	defer func() {
		if len(err) == 0 {
			err1 = nil
		}
	}()

	el := e.Rot()
	ed := el.Rot()
	er := ed.Rot()

	if ed.Sym() != e {
		// The Sym of Sym should be self
		err = append(err, "invalid Sym")
	}
	if ed != e.Sym() {
		err = append(err, fmt.Sprintf("invalid Rot: left.Rot != e.Sym %p : %p", el, e.Sym()))
	}
	if er != el.Sym() {
		err = append(err, fmt.Sprintf("invalid Rot: rot != e %p : %p", er, el.Sym()))

	}

	if e != el.InvRot() {
		err = append(err, "invalid Rot: rot != esym.InvRot")
	}

	if e.Orig() == nil {
		err = append(err, "expected edge to have origin")
		return err
	}
	if e.Dest() == nil {
		err = append(err, "expected edge to have dest")
		return err
	}

	npt := planar.PointOnLineAt(e.AsLine(), radius)
	normalizeDest := []geom.Point{npt}
	destPts := []geom.Point{*e.Dest()}

	{
		count := 0
		pts := make(map[geom.Point]bool)
		pts[*e.Dest()] = true
		pts[*e.Orig()] = true
		// Collect edges
		edges := []*Edge{e}
		ccwe := e.ONext()
		for e != ccwe {
			count++
			npt = planar.PointOnLineAt(ccwe.AsLine(), radius)

			if ccwe.Orig() == nil {
				err = append(err, "expected edge to have origin")
				return err
			}
			if ccwe.Dest() == nil {
				err = append(err, "expected edge to have dest")
				return err
			}

			if !cmp.GeomPointEqual(*e.Orig(), *ccwe.Orig()) {
				err = append(err, "orig not equal for edge")
				return err
			}
			if pts[*ccwe.Dest()] {
				err = append(err, fmt.Sprintf("dest (%v) not unique", wkt.MustEncode(*ccwe.Dest())))
				return err
			}
			pts[*ccwe.Dest()] = true

			// check ONext as well
			if edges[len(edges)-1] != ccwe.OPrev() {
				err = append(err, "expected oprev to be inverse of onext")
				return err
			}

			normalizeDest = append(normalizeDest, npt)
			destPts = append(destPts, *ccwe.Dest())
			edges = append(edges, ccwe)
			ccwe = ccwe.ONext()
		}
	}

	printErr := false
	switch len(normalizeDest) {
	case 1:
		// there is only one edge.
		if e.Sym() != e.LPrev() {
			err = append(err, "invalid single edge LPrev")
		}
		if e.Sym() != e.RPrev() {
			err = append(err, "invalid single edge RPrev")
		}
		if e.Sym() != e.RNext() {
			err = append(err, fmt.Sprintf("invalid single edge RNext : %p -- %p", e.RNext(), e))
		}
		if e.Sym() != e.LNext() {
			err = append(err, fmt.Sprintf("invalid single edge LNext : %p -- %p", e.LNext(), e))
		}

		if debug && err != nil {
			err = append(err, fmt.Sprintf("edges:\ne  %p\nel %p\ned %p\ner %p\n", e, el, ed, er))
			err = append(err, fmt.Sprintf("edges:\ne  %p\nel %p\ned %p\ner %p\n", e.next, el.next, ed.next, er.next))
			err = append(err, fmt.Sprintf("invalid edge: %v", wkt.MustEncode(e.AsLine())))
		}

	case 2:
		// Nothing left to test

	default:
		// only need to do one test.

		for i, j, k := len(normalizeDest)-2, len(normalizeDest)-1, 0; k < len(normalizeDest); i, j, k = j, k, k+1 {
			firstDest := normalizeDest[i]
			midDest := normalizeDest[j]
			lastDest := normalizeDest[k]

			if debug {
				log.Printf(
					" checking if counterclockwise: %v %v %v\n\tPts:%v %v %v",
					wkt.MustEncode(firstDest),
					wkt.MustEncode(midDest),
					wkt.MustEncode(lastDest),
					wkt.MustEncode(destPts[i]),
					wkt.MustEncode(destPts[j]),
					wkt.MustEncode(destPts[k]),
				)
			}

			if windingorder.OfGeomPoints(firstDest, midDest, lastDest).IsClockwise() {
				continue
			}

			/*
				if planar.IsCCW(firstDest, midDest, lastDest) {
					continue
				}
			*/

			err = append(err,
				fmt.Sprintf(
					" expected to be counterclockwise: %v %v %v\n%v %v %v",
					wkt.MustEncode(firstDest),
					wkt.MustEncode(midDest),
					wkt.MustEncode(lastDest),
					wkt.MustEncode(destPts[i]),
					wkt.MustEncode(destPts[j]),
					wkt.MustEncode(destPts[k]),
				),
			)
			printErr = true

		}
	}

	if debug && printErr {
		err = append(err, e.DumpAllEdges())
	}
	if len(err) == 0 {
		return nil
	}
	return err
}
