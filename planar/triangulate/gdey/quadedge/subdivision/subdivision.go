package subdivision

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"sync"

	"github.com/gdey/errors"

	"github.com/go-spatial/geom/planar"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
	"github.com/go-spatial/geom/encoding/wkt"
	"github.com/go-spatial/geom/internal/debugger"
	"github.com/go-spatial/geom/planar/intersect"
	"github.com/go-spatial/geom/planar/triangulate/gdey/quadedge/quadedge"
)

const RoundingFactor = 1000

// Subdivision describes a quadedge graph that is used for triangulation
type Subdivision struct {
	startingEdge *quadedge.Edge
	ptcount      int
	frame        [3]geom.Point

	vetexIndexLock   sync.RWMutex
	vertexIndexCache VertexIndex
}

// New initialize a subdivision to the triangle defined by the points a,b,c.
func New(a, b, c geom.Point) *Subdivision {
	ea := quadedge.New()
	ea.EndPoints(&a, &b)
	eb := quadedge.New()
	quadedge.Splice(ea.Sym(), eb)
	eb.EndPoints(&b, &c)

	ec := quadedge.New()
	ec.EndPoints(&c, &a)
	quadedge.Splice(eb.Sym(), ec)
	quadedge.Splice(ec.Sym(), ea)
	return &Subdivision{
		startingEdge: ea,
		ptcount:      3,
		frame:        [3]geom.Point{a, b, c},
	}
}

// NewForPoints creates a new subdivision for the given points, the points are
// sorted and duplicate points are not added
func NewForPoints(ctx context.Context, points [][2]float64) (*Subdivision, error) {
	// sort.Sort(cmp.ByXY(points))
	tri := geom.NewTriangleContainingPoints(points...)
	sd := New(tri[0], tri[1], tri[2])

	seen := make(map[geom.Point]bool)
	seen[tri[0]] = true
	seen[tri[1]] = true
	seen[tri[2]] = true

	for i, pt2f := range points {

		_ = i
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		pt := geom.Point{
			math.Round(pt2f[0]*RoundingFactor) / RoundingFactor,
			math.Round(pt2f[1]*RoundingFactor) / RoundingFactor,
		}
		if seen[pt] {
			continue
		}
		seen[pt] = true

		if !sd.InsertSite(pt) {
			// TODO(gdey): should this function return an error?
			log.Printf("Failed to insert point(%v) %v", i, wkt.MustEncode(pt))
			return nil, errors.String("Failed to insert point")
		}
	}
	if debug {
		//	log.Printf("Validating Subdivision (%v of %v", i, len(points))
		if err := sd.Validate(ctx); err != nil {
			if err1, ok := err.(quadedge.ErrInvalid); ok {
				var strBuf strings.Builder
				fmt.Fprintf(&strBuf, "Invalid subdivision:\n")
				for i, estr := range err1 {
					fmt.Fprintf(&strBuf, "\t%v : %v\n", i, estr)
				}
				fmt.Fprintf(&strBuf, "%v\n\n", wkt.MustEncode(geom.MultiPoint(points)))
				log.Printf(strBuf.String())
			}

			return sd, err
		}
	}
	return sd, nil
}

// locate returns an edge e, s.t. either x is on e, or e is an edge of
// a triangle containing x. The search starts from startingEdge
// and proceeds in the general direction of x. Based on the
// pseudocode in Guibas and Stolfi (1985) p.121
func (sd *Subdivision) locate(x geom.Point) (*quadedge.Edge, bool) {
	return locate(sd.startingEdge, x, sd.ptcount*2)
}

// InsertSite will insert a new point into a subdivision representing a Delaunay
// triangulation, and fixes the affected edges so that the result
// is  still a Delaunay triangulation. This is based on the pseudocode
// from Guibas and Stolfi (1985) p.120, with slight modifications and a bug fix.
func (sd *Subdivision) InsertSite(x geom.Point) bool {

	sd.ptcount++
	e, got := sd.locate(x)
	if !got {
		if debug {
			log.Println("did not find edge using normal walk")
		}
		// Did not find the edge using normal walk
		return false
	}
	if debug {
		log.Printf("insert %v found edge: %p %v", wkt.MustEncode(x), e, wkt.MustEncode(e.AsLine()))
		log.Printf("vertexs: %v", e.DumpAllEdges())
		log.Printf("subdivision")
		DumpSubdivision(sd)
	}

	if ptEqual(x, e.Orig()) || ptEqual(x, e.Dest()) {
		if debug {
			log.Printf("%v already in sd", wkt.MustEncode(x))
		}
		// Point is already in subdivision
		return true
	}

	if quadedge.OnEdge(x, e) {
		if debug {
			log.Printf("%v is on %v", wkt.MustEncode(x), wkt.MustEncode(e.AsLine()))
		}
		e = e.OPrev()
		// Check to see if this point is still already in subdivision.
		if ptEqual(x, e.Orig()) || ptEqual(x, e.Dest()) {
			if debug {
				log.Printf("%v already in sd", wkt.MustEncode(x))
			}
			// Point is already in subdivision
			return true
		}
		if debug {
			log.Printf("removing %v", wkt.MustEncode(e.ONext().AsLine()))
		}
		quadedge.Delete(e.ONext())
	}

	// Connect the new point to the vertices of the containing
	// triangle (or quadrilateral, if the new point fell on an
	// existing edge.)
	base := quadedge.NewWithEndPoints(e.Orig(), &x)
	quadedge.Splice(base, e)
	sd.startingEdge = base

	base = quadedge.Connect(e, base.Sym())
	e = base.OPrev()
	for e.LNext() != sd.startingEdge {
		base = quadedge.Connect(e, base.Sym())
		e = base.OPrev()
	}

	// Examine suspect edges to ensure that the Delaunay condition
	// is satisfied.
	for {
		t := e.OPrev()
		switch {
		case quadedge.RightOf(*t.Dest(), e) &&
			x.WithinCircle(*e.Orig(), *t.Dest(), *e.Dest()):
			quadedge.Swap(e)
			e = e.OPrev()

		case e.ONext() == sd.startingEdge: // no more suspect edges
			return true

		default: // pop a suspect edge
			e = e.ONext().LPrev()

		}
	}
}

// WalkAllEdges will call the provided function for each edge in the subdivision. The walk will
// be terminated if the function returns an error or ErrCancel. ErrCancel will not result in
// an error be returned by main function, otherwise the error will be passed on.
func (sd *Subdivision) WalkAllEdges(fn func(e *quadedge.Edge) error) error {

	if sd == nil || sd.startingEdge == nil {
		return nil
	}
	return WalkAllEdges(sd.startingEdge, fn)
}

// Triangles will return the triangles in the graph
func (sd *Subdivision) Triangles(includeFrame bool) (triangles [][3]geom.Point, err error) {

	if sd == nil {
		return nil, errors.String("subdivision is nil")
	}

	ctx := context.Background()
	WalkAllTriangles(ctx, sd.startingEdge, func(start, mid, end geom.Point) bool {
		if IsFramePoint(sd.frame, start, mid, end) && !includeFrame {
			return true
		}
		triangles = append(triangles, [3]geom.Point{start, mid, end})
		return true
	})

	return triangles, nil
}

// Validate will run a set of validation tests against the sd to insure
// the sd was built correctly. This process is very cpu and memory intensitive
func (sd *Subdivision) Validate(ctx context.Context) error {

	if cgo && debug {

		ctx = debugger.AugmentContext(ctx, "")
		defer debugger.Close(ctx)

	}

	var (
		lines []geom.Line
		err1  quadedge.ErrInvalid
	)

	if err := sd.WalkAllEdges(func(e *quadedge.Edge) error {
		l := e.AsLine()
		if debug {
			if err := quadedge.Validate(e); err != nil {
				return err
			}
		}
		l2 := l.LengthSquared()
		if l2 == 0 {
			if debug {
				debugger.Record(ctx,
					l,
					"ZeroLenght:Edge",
					"Line (%p) %v -- %v ", e, l2, l,
				)
			}
			err1 = append(err1, "zero length edge")
			return err1
		}
		lines = append(lines, l)
		return nil
	}); err != nil {
		return err
	}

	// Check for intersecting lines
	eq := intersect.NewEventQueue(lines)
	if err := eq.FindIntersects(ctx, true, func(i, j int, _ [2]float64) error {
		err1 = append(err1, "found intersecting lines: \n%v\n%v", wkt.MustEncode(lines[i]), wkt.MustEncode(lines[j]))
		return err1
	}); err != nil {
		return err
	}

	return nil
}

// IsValid will walk the graph making sure it is in a valid state
func (sd *Subdivision) IsValid(ctx context.Context) bool { return sd.Validate(ctx) == nil }

//
//*********************************************************************************************************
//  VertexIndex
//*********************************************************************************************************
//

// VertexIndex is an index of points to an quadedge in the graph
// this allows one to quickly jump to a group of edges by the origin
// point of that edge
type VertexIndex map[geom.Point]*quadedge.Edge

// VertexIndex will calculate and return a VertexIndex that can be used to
// quickly look up vertexies
func (sd *Subdivision) VertexIndex() VertexIndex {
	return NewVertexIndex(sd.startingEdge)
}

// NewVertexIndex will return a new vertex index given a starting edge.
func NewVertexIndex(startingEdge *quadedge.Edge) VertexIndex {
	vx := make(VertexIndex)
	WalkAllEdges(startingEdge, func(e *quadedge.Edge) error {
		vx.Add(e)
		return nil
	})
	return vx
}

// Add an edge to the graph
func (vx VertexIndex) Add(e *quadedge.Edge) {
	var (
		ok   bool
		orig = roundGeomPoint(*e.Orig())
		dest = roundGeomPoint(*e.Dest())
	)
	if _, ok = vx[orig]; !ok {
		vx[orig] = e
	}
	if _, ok = vx[dest]; !ok {
		vx[dest] = e.Sym()
	}
}

// Get retrives the edge
func (vx VertexIndex) Get(pt geom.Point) (*quadedge.Edge, bool) {
	pt = roundGeomPoint(pt)
	e, ok := vx[pt]
	return e, ok
}

// Remove an edge from the graph
func (vx VertexIndex) Remove(e *quadedge.Edge) {
	// Don't think I need e.Rot() and e.Rot().Sym() in this list
	// as they are face of the quadedge.
	toRemove := [4]*quadedge.Edge{e, e.Sym(), e.Rot(), e.Rot().Sym()}
	shouldRemove := func(e *quadedge.Edge) bool {
		for i := range toRemove {
			if toRemove[i] == e {
				return true
			}
		}
		return false
	}

	for _, v := range [...]geom.Point{*e.Orig(), *e.Dest()} {
		v = roundGeomPoint(v)
		ve := vx[v]
		if ve == nil || !shouldRemove(ve) {
			continue
		}
		delete(vx, v)
		// See if the ccw edge is the same as us, if it's isn't
		// then use that as the edge for our lookup.
		if ve != ve.ONext() {
			vx[v] = ve.ONext()
		}
	}
}

//
//*********************************************************************************************************
//  Helpers
//*********************************************************************************************************
//

func xprd(ao, bo [2]float64) float64 {
	// deal with yaxis downward positive
	return (ao[0] * bo[1]) - (ao[1] * bo[0])
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

// resolveEdge will find the edge such that dest lies between it and it's next edge.
// It does this using the following table:
//       ab -- orientation of a to b, (a being the edge of consideration)
//       da -- orientation of destPoint and a
//       db -- orientation of destPoint and b
//       ⟲ -- counter-clockwise
//       ⟳ -- clockwise
//        O -- colinear
//
// +----+----+----+----+
// |  # | ab | da | db | return - comment
// +----+----+----+----+
// |  1 | ⟲ | ⟲  | ⟲ | next
// |  2 | ⟲ | ⟲  | ⟳ | next
// |  3 | ⟲ | ⟲  | O  | nex
// |  4 | ⟲ | ⟳  | ⟲ | a
// |  5 | ⟲ | ⟳  | ⟳ | next
// |  6 | ⟲ | ⟳  | O | b -- ErrColinearPoints
// |  7 | ⟲ | O   | ⟲ | a -- ErrColinearPoints
// |  8 | ⟲ | O   | ⟳ | next
// |  9 | ⟳ | ⟲  | ⟲ | a
// | 10 | ⟳ | ⟲  | ⟳ | next
// | 11 | ⟳ | ⟲  | O | b -- ErrColinearPoints
// | 12 | ⟳ | ⟳  | ⟲ | a
// | 13 | ⟳ | ⟳  | ⟳ | a
// | 14 | ⟳ | ⟳  | O | a
// | 15 | ⟳ | O   | ⟲ | a
// | 16 | ⟳ | O   | ⟳ | a -- ErrColinearPoints
// | 17 | O  | O   | O | a/b -- ErrColinearPoint a/b depending on which one contains dest
// | 18 | O  | ⟲  | ⟳ | next
// | 19 | O  | ⟳  | ⟲ | a
//
// ab == 0 where a == b return nil and ErrCoincidentalEdges
// if ab == O and da == O then db must be O
// if ab != 0 then d can not be O to both a and b
//
func resolveEdge(gse *quadedge.Edge, odest geom.Point) (*quadedge.Edge, error) {

	var (
		candidate *quadedge.Edge
		err       error
	)

	orig := *gse.Orig()
	if cmp.GeomPointEqual(orig, odest) {
		return nil, ErrInvalidEndVertex

	}
	dest := geom.Point{odest[0] - orig[0], odest[1] - orig[1]}

	gse.WalkAllONext(func(e *quadedge.Edge) bool {

		a := *e.Dest()
		b := *e.ONext().Dest()

		ao := [2]float64{a[0] - orig[0], a[1] - orig[1]}
		bo := [2]float64{b[0] - orig[0], b[1] - orig[1]}

		// calculate the cross product of the the dest line each of the edges
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
		case zab && cmp.GeomPointEqual(a, b):
			candidate = a
			err = ErrCoincidentalEdges
			return false

		case zab && zda && zdb: // case 17
			candidate = e.ONext()
			if e.AsLine().ContainsPoint([2]float64(dest)) {
				candidate = e
			}
			err = geom.ErrPointsAreCoLinear
			return false

		case ccwab && zda && ccwdb: // case 7
			fallthrough
		case cwab && zda && cwdb: // case 16
			candidate = e
			err = geom.ErrPointsAreCoLinear
			return false

		case ccwab && cwda && zdb: // case 6
			fallthrough
		case cwab && ccwda && zdb: // case 11
			candidate = e.ONext()
			err = geom.ErrPointsAreCoLinear
			return false

		case zab && cwda && ccwdb:
			fallthrough
		case ccwab && cwda && ccwdb: // case 4
			fallthrough
		case cwab && ccwda && ccwdb: // case 9
			fallthrough
		case cwab && cwda && ccwdb: // case 12
			fallthrough
		case cwab && cwda && cwdb: // case 13
			fallthrough
		case cwab && cwda && zdb: // case 14
			fallthrough
		case cwab && zda && ccwdb: // case 15
			candidate = e
			err = nil
			return false

		default:
			/*
				these cases we move to the next entry
					case zab && ccwda && cwdb: // case 18
					case ccwab && ccwda && ccwdb: // case 1
					case ccwab && ccwda && cwdb: // case 2
					case ccwab && ccwda && zdb: // case 3
					case ccwab && cwda && cwdb: // case 5
					case ccwab && zda && cwdb: // case 8
					case cwab && ccwda && ccwdb: // case 10
			*/
			return true

		}

	})
	return candidate, err

}

// WalkAllEdges will call fn for each edge starting with se
func WalkAllEdges(se *quadedge.Edge, fn func(e *quadedge.Edge) error) error {
	if se == nil {
		return nil
	}
	var (
		toProcess quadedge.Stack
		visited   = make(map[*quadedge.Edge]bool)
	)
	toProcess.Push(se)
	for toProcess.Length() > 0 {
		e := toProcess.Pop()
		if visited[e] {
			continue
		}

		if err := fn(e); err != nil {
			if err == ErrCancelled {
				return nil
			}
			return err
		}

		sym := e.Sym()

		toProcess.Push(e.ONext())
		toProcess.Push(sym.ONext())

		visited[e] = true
		visited[sym] = true
	}
	return nil
}

// IsFrameEdge indicates if the edge is part of the given frame.
func IsFrameEdge(frame [3]geom.Point, es ...*quadedge.Edge) bool {
	for _, e := range es {
		o, d := *e.Orig(), *e.Dest()
		of := cmp.GeomPointEqual(o, frame[0]) || cmp.GeomPointEqual(o, frame[1]) || cmp.GeomPointEqual(o, frame[2])
		df := cmp.GeomPointEqual(d, frame[0]) || cmp.GeomPointEqual(d, frame[1]) || cmp.GeomPointEqual(d, frame[2])
		if of || df {
			return true
		}
	}
	return false
}

// IsHardFrameEdge indicates if the edge is part of the given frame where both vertexs are part of the frame.
func IsHardFrameEdge(frame [3]geom.Point, e *quadedge.Edge) bool {
	o, d := *e.Orig(), *e.Dest()
	of := cmp.GeomPointEqual(o, frame[0]) || cmp.GeomPointEqual(o, frame[1]) || cmp.GeomPointEqual(o, frame[2])
	df := cmp.GeomPointEqual(d, frame[0]) || cmp.GeomPointEqual(d, frame[1]) || cmp.GeomPointEqual(d, frame[2])
	return of && df
}

// IsFramePoint indicates if at least one of the points is equal to one of the frame points
func IsFramePoint(frame [3]geom.Point, pts ...geom.Point) bool {
	for _, pt := range pts {
		if cmp.GeomPointEqual(pt, frame[0]) ||
			cmp.GeomPointEqual(pt, frame[1]) ||
			cmp.GeomPointEqual(pt, frame[2]) {
			return true
		}
	}
	return false
}

func WalkAllTriangles(ctx context.Context, se *quadedge.Edge, fn func(start, mid, end geom.Point) (shouldContinue bool)) {
	if se == nil || fn == nil {
		return
	}
	var rcd debugger.Recorder

	if debug {
		rcd = debugger.GetRecorderFromContext(ctx)
	}

	var (
		// Hold the edges we still have to look at
		edgeStack []*quadedge.Edge

		startingEdge *quadedge.Edge
		workingEdge  *quadedge.Edge
		nextEdge     *quadedge.Edge

		// Hold points we have already seen and can ignore
		seenVerticies = make(map[geom.Point]bool)

		endPoint   geom.Point
		midPoint   geom.Point
		startPoint geom.Point

		count int
		loop  int
	)
	if debug {
		debugger.RecordOn(rcd, se.AsLine(), "WalkAllTriangles", "starting edge %v", se.AsLine())
	}

	edgeStack = append(edgeStack, se)

	for len(edgeStack) > 0 {
		if debug {
			count++
			loop = 0
		}

		// Pop of an edge to process
		startingEdge = edgeStack[len(edgeStack)-1]
		edgeStack = edgeStack[:len(edgeStack)-1]
		startPoint = *startingEdge.Orig()
		if seenVerticies[startPoint] {
			if debug {
				debugger.RecordOn(rcd, startPoint, "WalkAllTriangles:SkipVertex", "count:%v loop:%v vertex:%v", count, loop, startPoint)
			}
			// we have already processed this vertix
			continue
		}

		seenVerticies[startPoint] = true
		debugger.RecordOn(rcd, startPoint, "WalkAllTriangles:Vertex", "count:%v loop:%v vertex:%v", count, loop, startPoint)

		workingEdge = startingEdge
		nextEdge = startingEdge.ONext()
		if workingEdge == nextEdge {
			if debug {
				debugger.RecordOn(rcd, workingEdge.AsLine(), "WalkAllTriangles:SkipEdge:work==next", "count:%v loop:%v edge:%v", count, loop, workingEdge.AsLine())
			}
			continue
		}

		for {
			loop++
			endPoint = *nextEdge.Dest()
			midPoint = *workingEdge.Dest()
			if debug {
				debugger.RecordOn(
					rcd,
					geom.MultiPoint{
						[2]float64(startPoint),
						[2]float64(midPoint),
						[2]float64(endPoint),
					},
					"WalkAllTriangles:Vertex:Initial", "count:%v loop:%v initial verticies", count, loop,
				)
				debugger.RecordOn(
					rcd,
					geom.Triangle{
						[2]float64(startPoint),
						[2]float64(midPoint),
						[2]float64(endPoint),
					},
					"WalkAllTriangles:Triangle:Initial", "count:%v loop:%v prospective triangle", count, loop,
				)
				wln := workingEdge.AsLine()
				nln := nextEdge.AsLine()
				debugger.RecordOn(
					rcd,
					geom.MultiLineString{
						wln[:],
						nln[:],
					},
					"WalkAllTriangles:Edge:Initial", "count:%v loop:%v initial edges", count, loop,
				)
			}
			if seenVerticies[endPoint] || seenVerticies[midPoint] {
				if debug {
					skipPoint := midPoint
					if seenVerticies[endPoint] {
						skipPoint = endPoint
					}

					debugger.RecordOn(rcd, skipPoint, "WalkAllTriangles:SkipTriangle", "count:%v loop:%v vertex:%v(%v),%v(%v)", count, loop, midPoint, seenVerticies[midPoint], endPoint, seenVerticies[endPoint])
				}
				// we have already accounted for this triangle
				goto ADVANCE
			}

			// Add the working edge to the stack.
			edgeStack = append(edgeStack, workingEdge.Sym())
			if debug {
				debugger.RecordOn(rcd, workingEdge.AsLine(), "WalkAllTriangles:Edge", "count:%v loop:%v work-edge:%v", count, loop, workingEdge.AsLine())
			}

			if workingEdge.Sym().FindONextDest(endPoint) != nil {
				// found a triangle
				// *workingEdge.Orig(),*workingEdge.Dest(), *nextEdge.Dest()
				if debug {
					tri := geom.Triangle{[2]float64(startPoint), [2]float64(midPoint), [2]float64(endPoint)}
					debugger.RecordOn(rcd, tri, "WalkAllTriangles:Triangle", "count:%v loop:%v triangle:%v", count, loop, tri)
				}
				if !fn(startPoint, midPoint, endPoint) {
					return
				}
			} else if debug {
				debugger.RecordOn(rcd, endPoint, "WalkAllTriangles:Vertex", "count:%v loop:%v endPoint:%v not connected", count, loop, endPoint)
				debugger.RecordOn(rcd, workingEdge.Sym().AsLine(), "WalkAllTriangles:Edge", "count:%v loop:%v work-edge-sym:%v not connected", count, loop, workingEdge.Sym().AsLine())
				debugger.RecordOn(rcd, nextEdge.AsLine(), "WalkAllTriangles:Edge", "count:%v loop:%v next-edge:%v not connected", count, loop, nextEdge.AsLine())
			}

		ADVANCE:
			workingEdge = nextEdge
			nextEdge = workingEdge.ONext()
			if workingEdge == startingEdge {
				break
			}
		}

	}
}

// FindIntersectingEdges will find all edges in the graph that would be intersected by the origin of the starting edge and the
// dest of the endingEdge
func FindIntersectingEdges(startingEdge, endingEdge *quadedge.Edge) (edges []*quadedge.Edge, err error) {

	//const debug = true

	/*
					 Move starting edge so that the graph look like
					 ◌ .
		 se.ONext()╱ ┆ nse.Sym().ONext()   | \  ee
				  ╱  ┆					   |  \
		 (start) ● r ┆ l                 l | r ◍ (end)
				  ╲  ┆	 nee.Sym().ONext() |  /
		 	 	 se╲ ┆					   | /  ee.ONext()
					                       ◌

		right face of se is the triangle face, we want
		to go left, to find the next shared edge till
		we get to shared edge.
	*/

	if debug {

		log.Printf("\n\n FindIntersectingEdges \n\n\n")
		log.Printf("starting, %p\n%v\n", startingEdge, wkt.MustEncode(startingEdge.AsLine()))
		log.Printf("starting:ONext:Sym, %p\n%v\n", startingEdge.ONext().Sym(), wkt.MustEncode(startingEdge.ONext().Sym().AsLine()))
		log.Printf("ending, %p\n%v\n", endingEdge, wkt.MustEncode(endingEdge.AsLine()))
		log.Printf("ending:ONext:Sym, %p\n%v\n", endingEdge.ONext().Sym(), wkt.MustEncode(endingEdge.ONext().Sym().AsLine()))

	}

	if startingEdge == nil || endingEdge == nil {
		return edges, nil
	}

	start, end := *startingEdge.Orig(), *endingEdge.Orig()
	line := geom.Line{[2]float64(start), [2]float64(end)}
	if debug {
		log.Printf("line,\n%v\n", wkt.MustEncode(line))
	}
	if line.LengthSquared() == 0 {
		// nothing to do
		return edges, nil
	}

	startingEdge, _ = resolveEdge(startingEdge, end)
	endingEdge, _ = resolveEdge(endingEdge, start)

	if debug {
		log.Printf("\n\nAfter Resolve\n\n")

		log.Printf("starting, %p\n%v\n", startingEdge, wkt.MustEncode(startingEdge.AsLine()))
		log.Printf("starting:ONext:Sym, %p\n%v\n", startingEdge.ONext().Sym(), wkt.MustEncode(startingEdge.ONext().Sym().AsLine()))
		log.Printf("ending, %p\n%v\n", endingEdge, wkt.MustEncode(endingEdge.AsLine()))
		log.Printf("ending:ONext:Sym, %p\n%v\n", endingEdge.ONext().Sym(), wkt.MustEncode(endingEdge.ONext().Sym().AsLine()))
		log.Printf("line,\n%v\n", wkt.MustEncode(line))
	}
	sharedSE := startingEdge.ONext().Sym().ONext()
	sharedEE := endingEdge.ONext().Sym().ONext()

	if debug {
		log.Printf("shared starting, %p\n%v\n", sharedSE, wkt.MustEncode(sharedSE.AsLine()))
		log.Printf("shared end, %p\n%v\n", sharedEE, wkt.MustEncode(sharedEE.AsLine()))
	}

	count := 0
	workingEdge := sharedSE

	if debug {
		log.Printf("\n\nEdges\n\n")
	}

	for {
		count++
		if count > 21 {
			log.Printf("Failing with infint loop")
			log.Printf("starting, %p\n%v\n", startingEdge, wkt.MustEncode(startingEdge.AsLine()))
			log.Printf("starting:ONext:Sym, %p\n%v\n", startingEdge.ONext().Sym(), wkt.MustEncode(startingEdge.ONext().Sym().AsLine()))
			log.Printf("ending, %p\n%v\n", endingEdge, wkt.MustEncode(endingEdge.AsLine()))
			log.Printf("ending:ONext:Sym, %p\n%v\n", endingEdge.ONext().Sym(), wkt.MustEncode(endingEdge.ONext().Sym().AsLine()))
			log.Printf("line,\n%v\n", wkt.MustEncode(line))
			return edges, fmt.Errorf("infint loop")
		}

		wln := workingEdge.AsLine()
		nwln := workingEdge.ONext().AsLine()

		if debug {
			log.Printf("%3v working, %p\n%v\n", count, workingEdge, wkt.MustEncode(wln))
			log.Printf("%3v working:ONext, %p\n%v\n", count, workingEdge.ONext(), wkt.MustEncode(nwln))
		}

		if _, intersected := planar.SegmentIntersect(line, wln); intersected {
			if debug {
				log.Println("adding working edge to list of edges")
			}
			edges = append(edges, workingEdge)
		}

		if sharedEE.IsEqual(workingEdge) {
			// We have reached the end
			return edges, nil
		}

		if ipt, intersected := planar.SegmentIntersect(line, nwln); intersected {
			workingEdge = workingEdge.ONext()
			wln = workingEdge.AsLine()
			if debug {
				log.Printf("onext wln intersects line: %v\n%v\n%v", wkt.MustEncode(nwln), wkt.MustEncode(line), ipt)
				log.Printf("\nGoing to ONext()\n")
				log.Printf("working, %p\n%v\n", workingEdge, wkt.MustEncode(wln))
			}
			continue
		}

		workingEdge = workingEdge.ONext().Sym().ONext()
		if debug {
			log.Printf("\nGoing to ONext().Sym().ONext()\n")
			log.Printf("working, %p\n%v\n", workingEdge, wkt.MustEncode(wln))
			log.Printf("working:ONext, %p\n%v\n", workingEdge.ONext(), wkt.MustEncode(nwln))
		}

	}
}

func ptEqual(x geom.Point, a *geom.Point) bool {
	if a == nil {
		return false
	}
	return cmp.GeomPointEqual(x, *a)
}

func testEdge(x geom.Point, e *quadedge.Edge) (*quadedge.Edge, bool) {
	switch {
	case ptEqual(x, e.Orig()) || ptEqual(x, e.Dest()):
		return e, true
	case quadedge.RightOf(x, e):
		return e.Sym(), false
	case !quadedge.RightOf(x, e.ONext()):
		return e.ONext(), false
	case !quadedge.RightOf(x, e.DPrev()):
		return e.DPrev(), false
	default:
		return e, true
	}
}

func locate(se *quadedge.Edge, x geom.Point, limit int) (*quadedge.Edge, bool) {
	var (
		e     *quadedge.Edge
		ok    bool
		count int
	)
	for e, ok = testEdge(x, se); !ok; e, ok = testEdge(x, e) {
		if limit > 0 {

			count++
			if e == se || count > limit {
				log.Println("searching all edges for", x)
				e = nil

				WalkAllEdges(se, func(ee *quadedge.Edge) error {
					if _, ok = testEdge(x, ee); ok {
						e = ee
						return ErrCancelled
					}
					return nil
				})
				log.Printf(
					"Got back to starting edge after %v iterations, only have %v points ",
					count,
					limit,
				)
				return e, false
			}
		}
	}
	return e, true

}
