package subdivision

import (
	"context"
	"errors"
	"log"
	"sort"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
	"github.com/go-spatial/geom/encoding/wkt"
	"github.com/go-spatial/geom/internal/debugger"
	"github.com/go-spatial/geom/planar/triangulate/gdey/quadedge/quadedge"
)

var (
	ErrCancel           = errors.New("canceled walk")
	ErrCoincidentEdges  = errors.New("coincident edges")
	ErrDidNotFindToFrom = errors.New("did not find to and from edge")
)

type VertexIndex map[geom.Point]*quadedge.Edge

type Subdivision struct {
	startingEdge *quadedge.Edge
	ptcount      int
	frame        [3]geom.Point
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

func NewForPoints(ctx context.Context, points [][2]float64) *Subdivision {
	sort.Sort(cmp.ByXY(points))
	tri := geom.NewTriangleContainingPoints(points...)
	sd := New(tri[0], tri[1], tri[2])
	var oldPt geom.Point
	for i, pt := range points {
		if ctx.Err() != nil {
			return nil
		}
		bfpt := geom.Point(pt)
		if i != 0 && cmp.GeomPointEqual(oldPt, bfpt) {
			continue
		}
		oldPt = bfpt
		if !sd.InsertSite(bfpt) {
			log.Printf("Failed to insert point %v", bfpt)
		}
	}
	return sd
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
						return ErrCancel
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
func (vx VertexIndex) Add(e *quadedge.Edge) {
	var (
		ok   bool
		orig = *e.Orig()
		dest = *e.Dest()
	)
	if _, ok = vx[orig]; !ok {
		vx[orig] = e
	}
	if _, ok = vx[dest]; !ok {
		vx[dest] = e.Sym()
	}
}

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

// locate returns an edge e, s.t. either x is on e, or e is an edge of
// a triangle containing x. The search starts from startingEdge
// and proceeds in the general direction of x. Based on the
// pseudocode in Guibas and Stolfi (1985) p.121
func (sd *Subdivision) locate(x geom.Point) (*quadedge.Edge, bool) {
	return locate(sd.startingEdge, x, sd.ptcount*2)
}

func (sd *Subdivision) FindEdge(vertexIndex VertexIndex, start, end geom.Point) *quadedge.Edge {
	if vertexIndex == nil {
		vertexIndex = sd.VertexIndex()
	}
	return vertexIndex[start].FindONextDest(end)
}

// InsertSite will insert a new point into a subdivision representing a Delaunay
// triangulation, and fixes the affected edges so that the result
// is  still a Delaunay triangulation. This is based on the pseudocode
// from Guibas and Stolfi (1985) p.120, with slight modificatons and a bug fix.
func (sd *Subdivision) InsertSite(x geom.Point) bool {
	sd.ptcount++
	e, got := sd.locate(x)
	if !got {
		// Did not find the edge using normal walk
		return false
	}

	if ptEqual(x, e.Orig()) || ptEqual(x, e.Dest()) {
		// Point is already in subdivision
		return true
	}

	if quadedge.OnEdge(x, e) {
		e = e.OPrev()
		// Check to see if this point is still alreayd there.
		if ptEqual(x, e.Orig()) || ptEqual(x, e.Dest()) {
			// Point is already in subdivision
			return true
		}
		quadedge.Delete(e.ONext())
	}

	// Connect the new point to the vertices of the containing
	// triangle (or quadrilaterial, if the new point fell on an
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

func appendNonrepeat(pts []geom.Point, v geom.Point) []geom.Point {
	if len(pts) == 0 || cmp.GeomPointEqual(v, pts[len(pts)-1]) {
		return append(pts, v)
	}
	return pts
}

func selectCorrectEdges(from, to *quadedge.Edge) (cfrom, cto *quadedge.Edge) {
	orig := *from.Orig()
	dest := *to.Orig()
	cfrom, cto = from, to
	if debug {
		log.Printf("curr RightOf(dest)? %v", quadedge.RightOf(dest, cfrom))
		log.Printf("destedge.Sym RightOf(orig)? %v", quadedge.RightOf(orig, cto))
	}
	if !quadedge.RightOf(dest, cfrom) {
		cfrom = cfrom.OPrev()
	}
	if !quadedge.RightOf(orig, cto) {
		cto = cto.OPrev()
	}
	return cfrom, cto
}

func resolveEdge(gse *quadedge.Edge, dest geom.Point) *quadedge.Edge {

	// There aren't any other edges on this vertex.
	if gse == gse.ONext() {
		return gse
	}

	var lre *quadedge.Edge
	se := gse
	curr := se
	for {
		if quadedge.RightOf(dest, curr) {
			if lre == nil {
				// reset our starting edge.
				se = curr
			}
			lre = curr
			curr = curr.ONext()
			if curr == se {
				break
			}
			continue
		}
		// not right of
		if lre == nil {
			// We have not spotted an element right of us
			curr = curr.ONext()
			if curr == se {
				break
			}
			continue
		}
		return lre
	}
	if lre != nil {
		return lre
	}
	return se
}

func findImmediateRightOfEdges(se *quadedge.Edge, dest geom.Point) (*quadedge.Edge, *quadedge.Edge) {

	// We want the edge immediately left of the dest.

	orig := *se.Orig()
	if debug {
		log.Printf("Looking for orig fo %v to dest of %v", orig, dest)
	}
	curr := se
	for {
		if debug {
			log.Printf("top level looking at: %p (%v -> %v)", curr, *curr.Orig(), *curr.Dest())
		}
		if cmp.GeomPointEqual(*curr.Dest(), dest) {
			// edge already in the system.
			if debug {
				log.Printf("Edge already in system: %p", curr)
			}
			return curr, nil

		}

		// Need to see if the dest Next has the dest.
		for destedge := curr.Sym().ONext(); destedge != curr.Sym(); destedge = destedge.ONext() {
			if debug {
				log.Printf("\t looking at: %p (%v -> %v)", destedge, *destedge.Orig(), *destedge.Dest())
			}
			if cmp.GeomPointEqual(*destedge.Dest(), dest) {
				// found what we are looking for.
				if debug {
					log.Printf("Found the dest! %v -- %p %p", dest, curr, destedge.Sym())
				}

				return selectCorrectEdges(curr, destedge.Sym())
			}
			//log.Println("Next:", *destedge.Orig(), *curr.Sym().Orig(), *curr.Sym().Dest())

		}
		curr = curr.ONext()
		if curr == se {
			break
		}
	}
	return nil, nil
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

func (sd *Subdivision) Triangles(includeFrame bool) (triangles [][3]geom.Point, err error) {

	err = WalkAllTriangleEdges(
		sd.startingEdge,
		func(edges []*quadedge.Edge) error {
			if len(edges) != 3 {
				// skip this edge
				if debug {
					for i, e := range edges {
						log.Printf("got the following edge%v : %v", i, wkt.MustEncode(e.AsLine()))
					}
				}
				return nil
				//	return errors.New("Something Strange!")
			}

			pts := [3]geom.Point{*edges[0].Orig(), *edges[1].Orig(), *edges[2].Orig()}

			// Do we want to skip because the points are part of the frame and
			// we have been requested not to include triangles attached to the frame.
			if IsFramePoint(sd.frame, pts[:]...) && !includeFrame {
				return nil
			}

			triangles = append(triangles, pts)
			return nil
		},
	)
	return triangles, err
}

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
			if err == ErrCancel {
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

// IsFrameEdge indicates if the edge is part of the given frame where both vertexs are part of the frame.
func IsHardFrameEdge(frame [3]geom.Point, e *quadedge.Edge) bool {
	o, d := *e.Orig(), *e.Dest()
	of := cmp.GeomPointEqual(o, frame[0]) || cmp.GeomPointEqual(o, frame[1]) || cmp.GeomPointEqual(o, frame[2])
	df := cmp.GeomPointEqual(d, frame[0]) || cmp.GeomPointEqual(d, frame[1]) || cmp.GeomPointEqual(d, frame[2])
	return of && df
}

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

func constructTriangleEdges(
	e *quadedge.Edge,
	toProcess *quadedge.Stack,
	visited map[*quadedge.Edge]bool,
	fn func(edges []*quadedge.Edge) error,
) error {

	if visited[e] {
		return nil
	}

	curr := e
	var triedges []*quadedge.Edge
	for backToStart := false; !backToStart; backToStart = curr == e {

		// Collect edge
		triedges = append(triedges, curr)

		sym := curr.Sym()
		if !visited[sym] {
			toProcess.Push(sym)
		}

		// mark edge as visted
		visited[curr] = true

		// Move the ccw edge
		curr = curr.LNext()
	}
	return fn(triedges)
}

// WalkAllTriangleEdges will walk the subdivision starting from the starting edge (se) and return
// sets of edges that make make a triangle for each face.
func WalkAllTriangleEdges(se *quadedge.Edge, fn func(edges []*quadedge.Edge) error) error {
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
		err := constructTriangleEdges(e, &toProcess, visited, fn)
		if err != nil {
			if err == ErrCancel {
				return nil
			}
			return err
		}
	}
	return nil
}

func FindIntersectingTriangle(startingEdge *quadedge.Edge, end geom.Point) (*Triangle, error) {
	var (
		left  = startingEdge
		right *quadedge.Edge
	)

	for {
		right = left.OPrev()

		lc := Classify(end, *left.Orig(), *left.Dest())
		rc := Classify(end, *right.Orig(), *right.Dest())

		if (lc == RIGHT && rc == LEFT) ||
			lc == BETWEEN ||
			lc == DESTINATION ||
			lc == BEYOND {
			return &Triangle{left}, nil
		}

		if lc != RIGHT && lc != LEFT &&
			rc != RIGHT && rc != LEFT {
			return &Triangle{left}, ErrCoincidentEdges
		}
		left = right
		if left == startingEdge {
			// We have walked all around the vertex.
			break
		}

	}
	return nil, nil
}

func (sd *Subdivision) IsValid(ctx context.Context) bool {
	count := 0
	if debug {

		ctx = debugger.AugmentContext(ctx, "")
		defer debugger.Close(ctx)

	}
	_ = sd.WalkAllEdges(func(e *quadedge.Edge) error {
		l := e.AsLine()
		l2 := l.LenghtSquared()
		if l2 == 0 {
			count++
			if debug {
				debugger.Record(ctx,
					l,
					"ZeroLenght:Edge",
					"Line (%p) %v -- %v ", e, l2, l,
				)
			}
		}
		return nil
	})
	log.Println("Count", count)
	return count == 0
}
