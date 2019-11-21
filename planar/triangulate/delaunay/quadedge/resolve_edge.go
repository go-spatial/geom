package quadedge

import (
	"log"

	"github.com/gdey/errors"
	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/winding"
)

const (

	// ErrInvalidStartingVertex is returned when the starting vertex is invalid
	ErrInvalidStartingVertex = errors.String("invalid starting vertex")

	// ErrInvalidEndVertex is returned when the ending vertex is invalid
	ErrInvalidEndVertex = errors.String("invalid ending vertex")

	// ErrCoincidentalEdges is returned when two edges are conincidental and not expected to be
	ErrCoincidentalEdges = errors.String("coincident edges")
)

type rEdge struct {
	orig geom.Point
	dest geom.Point

	e          *Edge
	ab, da, db winding.Winding

	err       error
	candidate *Edge
}

func (re *rEdge) CCWAB() bool { return re.ab.IsCounterClockwise() }
func (re *rEdge) CWAB() bool  { return re.ab.IsClockwise() }
func (re *rEdge) ZAB() bool   { return re.ab.IsColinear() }

func (re *rEdge) CCWDA() bool { return re.da.IsCounterClockwise() }
func (re *rEdge) CWDA() bool  { return re.da.IsClockwise() }
func (re *rEdge) ZDA() bool   { return re.da.IsColinear() }

func (re *rEdge) CCWDB() bool { return re.db.IsCounterClockwise() }
func (re *rEdge) CWDB() bool  { return re.db.IsClockwise() }
func (re *rEdge) ZDB() bool   { return re.db.IsColinear() }

func (re *rEdge) Next() {
	re.candidate = nil
	re.err = nil
}
func (re *rEdge) A() {
	re.candidate = re.e
}
func (re *rEdge) ErrA() {
	re.candidate = re.e
	re.err = geom.ErrPointsAreCoLinear
}
func (re *rEdge) ErrB() {
	re.candidate = re.e.ONext()
	re.err = geom.ErrPointsAreCoLinear
}
func (re *rEdge) ErrEdge() {
	re.candidate = re.e
	re.err = ErrCoincidentalEdges
}

// ContainsDest returns weather the edge constains the original dest
func (re *rEdge) ContainsDest() bool {
	return re.e.AsLine().ContainsPoint([2]float64(re.dest))
}

func resolveEdge(order winding.Order, gse *Edge, odest geom.Point, table func(*rEdge)) (*Edge, error) {
	orig := *gse.Orig()
	if cmp.GeomPointEqual(orig, odest) {
		return nil, ErrInvalidEndVertex

	}
	dest := geom.Point{odest[0] - orig[0], odest[1] - orig[1]}

	var re = rEdge{
		orig: orig,
		dest: odest,
	}

	gse.WalkAllONext(func(e *Edge) bool {
		apt := *e.Dest()
		bpt := *e.ONext().Dest()
		re.err = nil
		re.candidate = nil

		ao := [2]float64{apt[0] - orig[0], apt[1] - orig[1]}
		bo := [2]float64{bpt[0] - orig[0], bpt[1] - orig[1]}
		// calculate the cross product of the the dest line each of the edges
		//                                                     +---
		// ccw == 0,1 ->  1,0 == ( 0 * 0 ) - ( 1 * 1 ) == -1   |⟳
		//                                                     +--
		// cw  == 1,0 ->  0,1 == ( 1 * 1 ) - ( 0 * 0 ) ==  1   |⟲
		//                                                     +---
		// cl  == 1,0 -> -1,0 == ( 1 * 0 ) - (-1 * 0 ) ==  0   |——
		//                                                     +---
		oo := [2]float64{0, 0}
		re.ab, re.da, re.db = order.OfPoints(ao, bo, oo), order.OfPoints(dest, ao, oo), order.OfPoints(dest, bo, oo)
		re.e = e

		if debug {
			log.Printf("a: %v", wkt.MustEncode(re.e.AsLine()))
			log.Printf("b: %v", wkt.MustEncode(re.e.ONext().AsLine()))
			log.Printf("d: %v", wkt.MustEncode(odest))
			log.Printf("ab: %v %v da: %v %v db: %v %v", re.ab, re.ab.ShortString(), re.da, re.da.ShortString(), re.db, re.db.ShortString())
			log.Printf("ao: %v bo: %v dest: %v", ao, bo, dest)
		}

		table(&re)

		// continue if we don't have an error and no candidate
		return re.candidate == nil && re.err == nil
	})

	return re.candidate, re.err
}

func resolveEdgeYUp(re *rEdge) {
	switch {
	case re.CCWAB():
		switch {
		case re.CCWDA():
			re.Next()
		case re.CWDA() && re.CCWDB():
			re.A()
		case re.CWDA() && re.CWDB():
			re.Next()
		case re.CWDA() && re.ZDB():
			re.ErrB()
		case re.ZDA() && re.CCWDB():
			re.ErrA()
		case re.ZDA() && re.CWDB():
			re.Next()
		}
	case re.CWAB():
		switch {
		case re.CWDA():
			re.A()
		case re.CCWDA() && re.CCWDB():
			re.A()
		case re.CCWDA() && re.CWDB():
			re.Next()
		case re.CCWDA() && re.ZDB():
			re.ErrB()
		case re.ZDA() && re.CCWDB():
			re.A()
		case re.ZDA() && re.CWDB():
			re.ErrA()
		}
	case re.ZAB():
		switch {
		case re.CCWDA() && re.CWDB():
			re.Next()
		case re.CWDA() && re.CCWDB():
			re.A()
		case (re.CWDA() && re.CWDB()) || (re.CCWDA() && re.CCWDB()):
			re.A()
		case re.ZDA() && re.ZDB():
			if re.ContainsDest() {
				re.ErrA()
			} else {
				re.ErrB()
			}
		}
	default:
		re.ErrEdge()
	}
}
func resolveEdgeYDown(re *rEdge) {
	switch {
	// CCWAB
	case re.CCWAB() && re.CCWDA():
		re.Next()

	case re.CCWAB() && re.CWDA() && re.CWDB():
		re.Next()
	case re.CCWAB() && re.CWDA() && re.CCWDB():
		re.A()
	case re.CCWAB() && re.CWDA() && re.ZDB():
		re.ErrB()

	case re.CCWAB() && re.ZDA() && re.CCWDB():
		re.ErrA()
	case re.CCWAB() && re.ZDA() && re.CWDB():
		re.Next()

	// CWAB
	case re.CWAB() && re.CCWDA() && re.CCWDB():
		re.A()
	case re.CWAB() && re.CCWDA() && re.CWDB():
		re.Next()
	case re.CWAB() && re.CCWDA() && re.ZDB():
		re.ErrB()

	case re.CWAB() && re.CWDA():
		re.A()

	case re.CWAB() && re.ZDA() && re.CCWDB():
		re.A()
	case re.CWAB() && re.ZDA() && re.CWDB():
		re.ErrA()

	// ZAB
	case re.ZAB() && re.CCWDA() && re.CWDB():
		re.Next()
	case re.ZAB() && re.CWDA() && re.CCWDB():
		re.A()
	case re.ZAB() && re.ZDA() && re.ZDB():
		if re.ContainsDest() {
			re.ErrA()
		} else {
			re.ErrB()
		}

	case re.ZAB() && re.CCWDA() && re.CCWDB():
		re.A()
	case re.ZAB() && re.CWDA() && re.CWDB():
		re.A()

	default:
		re.ErrEdge()
	}
}

// ResolveEdge will find the edge such that dest lies between it and it's next edge.
// It does this using the following table:
//       ab -- orientation of a to b, (a being the edge of consideration)
//       da -- orientation of destPoint and a
//       db -- orientation of destPoint and b
//       ⟲ -- counter-clockwise
//       ⟳ -- clockwise
//        O -- colinear
//
//        +----+----+----+----+
//        |  # | ab | da | db | return - comment
//        +----+----+----+----+                                           8
//        |  1 | ⟲ | ⟲  | ⟲ | next                                   2  :  5
//        |  2 | ⟲ | ⟲  | ⟳ | next                                .3....+----6->b
//        |  3 | ⟲ | ⟲  | O  | next                                   1  |,,4,,,
//        |  4 | ⟲ | ⟳  | ⟲ | a                                         7,,,,,,       ab =  ⟲ == next orientation
//        |  5 | ⟲ | ⟳  | ⟳ | next                                      V
//        |  6 | ⟲ | ⟳  | O | b -- ErrColinearPoints                     a
//        |  7 | ⟲ | O   | ⟲ | a -- ErrColinearPoints
//        |  8 | ⟲ | O   | ⟳ | next
//        |  + | ⟲ | O   | O | point is at origin  : Err          ,,,,,,,14,,,,
//        |  9 | ⟳ | ⟲  | ⟲ | a                                  ,,,,12,:,,13,
//        | 10 | ⟳ | ⟲  | ⟳ | next                               .15....+----16>a
//        | 11 | ⟳ | ⟲  | O | b -- ErrColinearPoints              ,,,,9,,|  10
//        | 12 | ⟳ | ⟳  | ⟲ | a                                  ,,,,,,,11             ab = ⟳  == opposite of next orientation
//        | 13 | ⟳ | ⟳  | ⟳ | a                                         V
//        | 14 | ⟳ | ⟳  | O | a                                          b
//        | 15 | ⟳ | O   | ⟲ | a
//        | 16 | ⟳ | O   | ⟳ | a -- ErrColinearPoints
//        |  + | ⟳ | O   | O | point is at origin : Err            ,,,,,,,18,,,,,
//        | 17 | O  | ⟲  | ⟳ | next                              b-19----+---19->a
//        | 18 | O  | ⟳  | ⟲ | a                                         17
//        | 19 | O  | O   | O | a/b -- ErrColinearPoint a/b depending on which one contains dest
//        | 20 | O  | ⟲  | ⟲ | a -- ErrCoincidentalEdges                 21
//        | 21 | O  | ⟳  | ⟳ | a -- ErrCoincidentalEdges           .......+------>a,b
//        +----+----+-----+----+                                          20
//
//        if ab == O and da == O then db must be O
//
// Only errors returned are
//  * nil  // nothing is wrong
//  * ErrInvalidateEndVertex
//  * ErrConcidentalEdges
//  * geom.ErrColinearPoints
func ResolveEdge(order winding.Order, gse *Edge, odest geom.Point) (*Edge, error) {
	if order.YPositiveDown {
		return resolveEdge(order, gse, odest, resolveEdgeYDown)
	}
	return resolveEdge(order, gse, odest, resolveEdgeYUp)
}
