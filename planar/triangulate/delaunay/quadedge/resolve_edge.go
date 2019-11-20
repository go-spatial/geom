package quadedge

import (
	"log"
	"math"

	"github.com/gdey/errors"
	"github.com/go-spatial/geom"
)

const (

	// ErrInvalidStartingVertex is returned when the starting vertex is invalid
	ErrInvalidStartingVertex = errors.String("invalid starting vertex")

	// ErrInvalidEndVertex is returned when the ending vertex is invalid
	ErrInvalidEndVertex = errors.String("invalid ending vertex")

	// ErrCoincidentalEdges is returned when two edges are conincidental and not expected to be
	ErrCoincidentalEdges = errors.String("coincident edges")
)

func xprd(ao, bo [2]float64) float64 {
	return (ao[0] * bo[1]) - (ao[1] * bo[0])
}

func sign(f float64) float64 {
	if cmp.Float(f, 0.0) {
		return 0.0
	}
	if math.Signbit(f) {
		return -1.0
	}
	return 1.0
}

func toOrtStr(s float64) string {
	if s == 0 {
		return "O"
	}
	if s < 0 {
		return "⟲"
	}
	return "⟳"
}

type rEdge struct {
	orig  geom.Point
	dest  geom.Point
	yDown int

	e          *Edge
	ab, da, db float64

	err       error
	candidate *Edge
}

func (re *rEdge) CCWAB() bool { return re.ab > 0 }
func (re *rEdge) CWAB() bool  { return re.ab < 0 }
func (re *rEdge) ZAB() bool   { return re.ab == 0 }

func (re *rEdge) CCWDA() bool { return re.da > 0 }
func (re *rEdge) CWDA() bool  { return re.da < 0 }
func (re *rEdge) ZDA() bool   { return re.da == 0 }

func (re *rEdge) CCWDB() bool { return re.db > 0 }
func (re *rEdge) CWDB() bool  { return re.db < 0 }
func (re *rEdge) ZDB() bool   { return re.db == 0 }

func (re *rEdge) Next() {
	re.candidate = nil
	re.err = nil
	if debug {
		log.Printf("next: %v %v %v", wkt.MustEncode(re.e.AsLine()), wkt.MustEncode(re.e.ONext().AsLine()), wkt.MustEncode(re.dest))
	}
}
func (re *rEdge) A() {
	re.candidate = re.e
	if debug {
		log.Printf("a: %v %v %v", wkt.MustEncode(re.e.AsLine()), wkt.MustEncode(re.e.ONext().AsLine()), wkt.MustEncode(re.dest))
	}
}
func (re *rEdge) ErrA() {
	re.candidate = re.e
	re.err = geom.ErrPointsAreCoLinear
	if debug {
		log.Printf("erra: [%v] %v %v", wkt.MustEncode(re.e.AsLine()), wkt.MustEncode(re.e.ONext().AsLine()), wkt.MustEncode(re.dest))
	}
}
func (re *rEdge) ErrB() {
	re.candidate = re.e.ONext()
	re.err = geom.ErrPointsAreCoLinear
	if debug {
		log.Printf("errb: %v [%v] %v", wkt.MustEncode(re.e.AsLine()), wkt.MustEncode(re.e.ONext().AsLine()), wkt.MustEncode(re.dest))
	}
}
func (re *rEdge) ErrEdge() {
	re.candidate = re.e
	re.err = ErrCoincidentalEdges
	if debug {
		log.Printf("ConincidentalEdges: [%v] %v %v", wkt.MustEncode(re.e.AsLine()), wkt.MustEncode(re.e.ONext().AsLine()), wkt.MustEncode(re.dest))
	}
}

func resolveEdge(yDown bool, gse *Edge, odest geom.Point, table func(*rEdge)) (*Edge, error) {
	multi := 1.0
	if yDown {
		multi = -1.0
	}

	orig := *gse.Orig()
	if cmp.GeomPointEqual(orig, odest) {
		return nil, ErrInvalidEndVertex

	}
	dest := geom.Point{odest[0] - orig[0], odest[1] - orig[1]}

	var re rEdge

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
		re.ab, re.da, re.db = xprd(ao, bo)*multi, xprd(dest, ao)*multi, xprd(dest, bo)*multi
		re.e = e

		if debug {
			log.Printf("a: %v", wkt.MustEncode(re.e.AsLine()))
			log.Printf("b: %v", wkt.MustEncode(re.e.ONext().AsLine()))
			log.Printf("d: %v", wkt.MustEncode(odest))
			log.Printf("ab: %v %v da: %v %v db: %v %v", re.ab, toOrtStr(re.ab), re.da, toOrtStr(re.da), re.db, toOrtStr(re.db))
		}

		table(&re)

		// continue if we don't have an error and no candidate
		return re.candidate == nil && re.err == nil
	})

	return re.candidate, re.err
}

func resolveEdgeYUp(gse *Edge, odest geom.Point) (*Edge, error) {
	return resolveEdge(false, gse, odest, func(re *rEdge) {

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
			case re.ZDA() && re.ZDB():
				re.ErrEdge()
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
			case re.ZDA() && re.ZDB():
				re.ErrEdge()
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
				re.ErrEdge()
			}
		default:
			re.ErrEdge()
		}

	})
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
func ResolveEdge(gse *Edge, odest geom.Point) (*Edge, error) { return resolveEdgeYUp(gse, odest) }

/*
func ResolveEdge(gse *Edge, odest geom.Point) (*Edge, error) {

	var (
		candidate *Edge
		err       error = ErrInvalidEndVertex

		next = func(e *Edge) bool {
			if debug {
				log.Printf("next: %v %v %v", wkt.MustEncode(e.AsLine()), wkt.MustEncode(e.ONext().AsLine()), wkt.MustEncode(odest))
			}
			return true
		}
		a = func(e *Edge) bool {
			candidate = e
			err = nil
			if debug {
				log.Printf("a: %v %v %v", wkt.MustEncode(e.AsLine()), wkt.MustEncode(e.ONext().AsLine()), wkt.MustEncode(odest))
			}
			return false
		}
		errA = func(e *Edge) bool {
			candidate = e
			err = geom.ErrPointsAreCoLinear
			if debug {
				log.Printf("erra: %v %v %v", wkt.MustEncode(e.AsLine()), wkt.MustEncode(e.ONext().AsLine()), wkt.MustEncode(odest))
			}
			return false
		}
		errB = func(e *Edge) bool {
			candidate = e.ONext()
			err = geom.ErrPointsAreCoLinear
			if debug {
				log.Printf("errb: %v %v %v", wkt.MustEncode(e.AsLine()), wkt.MustEncode(e.ONext().AsLine()), wkt.MustEncode(odest))
			}
			return false
		}
		errEdge = func(e *Edge) bool {
			candidate = e
			err = ErrCoincidentalEdges
			if debug {
				log.Printf("ConincidentalEdges: %v %v %v", wkt.MustEncode(e.AsLine()), wkt.MustEncode(e.ONext().AsLine()), wkt.MustEncode(odest))
			}
			return false
		}
	)

	orig := *gse.Orig()
	if cmp.GeomPointEqual(orig, odest) {
		return nil, ErrInvalidEndVertex

	}
	dest := geom.Point{odest[0] - orig[0], odest[1] - orig[1]}

	gse.WalkAllONext(func(e *Edge) bool {

		apt := *e.Dest()
		bpt := *e.ONext().Dest()

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
		ab, da, db := xprd(ao, bo), xprd(dest, ao), xprd(dest, bo)
		ccwab, cwab, zab := ab > 0, ab < 0, ab == 0
		ccwda, cwda, zda := da > 0, da < 0, da == 0
		ccwdb, cwdb, zdb := db > 0, db < 0, db == 0

		if debug {
			log.Printf("a: %v", wkt.MustEncode(e.AsLine()))
			log.Printf("b: %v", wkt.MustEncode(e.ONext().AsLine()))
			log.Printf("d: %v", wkt.MustEncode(odest))
			log.Printf("ab: %v %v da: %v %v db: %v %v", ab, toOrtStr(ab), da, toOrtStr(da), db, toOrtStr(db))
		}

		switch {
		case ccwab && ccwda && ccwdb: // case 1
			return next(e)
		case ccwab && ccwda && cwdb: // case 2
			return next(e)
		case ccwab && ccwda && zdb: // case 3
			return next(e)
		case ccwab && cwda && ccwdb: // case 4
			return a(e)
		case ccwab && cwda && cwdb: // case 5
			return next(e)
		case ccwab && cwda && zdb: // case 6
			return errB(e)
		case ccwab && zda && ccwdb: // case 7
			return errA(e)
		case ccwab && zda && cwdb: // case 8
			return next(e)

		// +

		case cwab && ccwda && ccwdb: // case 9
			return a(e)
		case cwab && ccwda && cwdb: // case 10
			return next(e)
		case cwab && ccwda && zdb: // case 11
			return errB(e)
		case cwab && cwda && ccwdb: // case 12
			return a(e)
		case cwab && cwda && cwdb: // case 13
			return a(e)
		case cwab && cwda && zdb: // case 14
			return a(e)
		case cwab && zda && ccwdb: // case 15
			return a(e)
		case cwab && zda && cwdb: // case 16
			return errA(e)

		// +

		case zab && ccwda && cwdb: // case 17
			return next(e)
		case zab && cwda && ccwdb: // case 18
			return a(e)

		case zab && zda && zdb: // case 19
			if e.AsLine().ContainsPoint([2]float64(odest)) {
				return errA(e)
			}
			return errB(e)

		case zab && ccwda && ccwdb: // case 21
			return errEdge(e)
		case zab && cwda && cwdb: // case 20
			return errEdge(e)

		default:
			return true

		}
	})
	return candidate, err

}
*/
