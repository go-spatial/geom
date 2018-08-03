package tegola

import (
	"context"
	"log"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
	"github.com/go-spatial/geom/planar"
	"github.com/go-spatial/geom/planar/triangulate/constraineddelaunay"
)

func sortedEdge(pt1, pt2 [2]float64) [2][2]float64 {
	if cmp.PointLess(pt1, pt2) {
		return [2][2]float64{pt1, pt2}
	}
	return [2][2]float64{pt2, pt1}
}

func edgeMapFromTriangles(triangles ...triangle) map[[2][2]float64][]int {
	// an edge can have at most two triangles associated with it.
	em := make(map[[2][2]float64][]int, 2*len(triangles))
	for i, tri := range triangles {
		for _, edg := range tri.SortedEdges() {
			if _, ok := em[edg]; !ok {
				em[edg] = make([]int, 0, 2)
			}
			em[edg] = append(em[edg], i)
		}
	}
	return em
}

func triangulateGeometry(g geom.Geometry) (geom.MultiPolygon, error) {
	uut := new(constraineddelaunay.Triangulator)
	uut.InsertSegments(g)

	if debug {
		err := uut.Validate()
		if err != nil {
			log.Printf("Triangulator is not validate for the given segments %v : %v", g, err)
			return nil, err
		}
	}
	// TODO(gdey): We need to insure that GetTriangles does not dup the first point to the
	//              last point. It may be better if it returned triangles and we moved triangles to Geom.
	return uut.GetTriangles()
}

func newEdgeIndexTriangles(ctx context.Context, hm planar.HitMapper, g geom.Geometry) (*edgeIndexTriangles, error) {

	var geomTriangles geom.MultiPolygon
	var err error

	if geomTriangles, err = triangulateGeometry(g); err != nil {
		return nil, err
	}

	triangles := make([]triangle, 0, len(geomTriangles))

	for _, ply := range geomTriangles {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		tri := newTriangleFromPolygon(ply)
		if hm.LabelFor(tri.Center()) == planar.Outside {
			continue
		}
		triangles = append(triangles, tri)
	}

	t := edgeIndexTriangles{
		triangles: make([]triangle, len(triangles)),
		edgeMap:   edgeMapFromTriangles(triangles...),
	}

	copy(t.triangles, triangles)
	return &t, nil
}

type edgeIndexTriangles struct {
	triangles []triangle
	edgeMap   map[[2][2]float64][]int
}

func (eit *edgeIndexTriangles) MultiPolygon(ctx context.Context) (mplyg [][][][2]float64) {
	if eit == nil {
		return mplyg
	}
	seen := make(map[int]bool, len(eit.triangles))
	for i := range eit.triangles {
		if ctx.Err() != nil {
			return nil
		}
		if seen[i] {
			continue
		}
		seen[i] = true
		plyg := eit.PolygonForTriangle(ctx, i, seen)
		if len(plyg) > 0 {
			mplyg = append(mplyg, plyg)
		}
	}
	return mplyg
}

func (eit *edgeIndexTriangles) indexForEdge(p1, p2 [2]float64, defaultIdx int, seen map[int]bool) (idx int, ok bool) {
	for _, idx := range eit.edgeMap[sortedEdge(p1, p2)] {
		if seen[idx] || idx == defaultIdx {
			continue
		}
		return idx, true
	}
	return defaultIdx, false
}

// ringForTriangle will walk the set of triangles starting at the given triangle index. As it walks the triangles it will
// mark them as seen on the seen map. The function will return the outside ring of the walk
func (eit *edgeIndexTriangles) ringForTriangle(ctx context.Context, idx int, seen map[int]bool) (rng [][2]float64) {

	var ok bool

	if debug {
		log.Printf("getting ring for triangle %v", idx)
	}

	seen[idx] = true

	// This tracks the start of the ring.
	// The segment we are adding a point will be between the endpoint and the begining of the ring.
	// This track the original begining of the ring.
	var headIdx int

	rng = append(rng, eit.triangles[idx][:]...)
	cidxs := []int{idx, idx, idx}
	cidx := cidxs[len(cidxs)-1]

RING_LOOP:
	for {
		// A few sanity checks, were we cancelled, or reached the end of our walk.
		if ctx.Err() != nil || // We were told to come home.
			headIdx >= len(rng) || len(cidxs) == 0 { // we returned home.
			return rng
		}

		if debug {
			log.Printf("headIdx: %v -- len(rng): %v", headIdx, len(rng))
			log.Printf("ring: %v | %v", rng[:headIdx], rng[headIdx:])
			log.Printf("cidxs: %v", cidxs)
		}

		if cidx, ok = eit.indexForEdge(rng[0], rng[len(rng)-1], cidxs[len(cidxs)-1], seen); !ok {
			// We don't have a neighbor to walk to here. Let's move back one and see if there is a path we need to go down.
			headIdx += 1
			lpt := rng[len(rng)-1]
			copy(rng[1:], rng)
			rng[0] = lpt
			cidxs = cidxs[:len(cidxs)-1]
			continue
		}

		if cidx == idx {
			// We go back to our starting triangle. We need to stop.
			return rng
		}

		if debug {
			log.Printf("check to see if we have seen the triangle we are going to jump to.")
		}

		// Check to see if we have reached the triangle before.
		for i, pcidx := range cidxs {
			if pcidx != cidx {
				continue
			}
			if debug {
				log.Printf("we have encountered idx (%v) before at %v", cidx, i)
			}
			// need to move all the points over
			tlen := len(rng) - (i + 1)
			tpts := make([][2]float64, tlen)
			copy(tpts, rng[i+1:])
			copy(rng[tlen:], rng[:i+1])
			copy(rng, tpts)
			headIdx += tlen

			cidxs = cidxs[:i+1]
			continue RING_LOOP
		}

		rng = append(rng, eit.triangles[cidx].ThirdPoint(rng[0], rng[len(rng)-1]))

		cidxs[len(cidxs)-1] = cidx
		cidxs = append(cidxs, cidx)
		seen[cidx] = true

	} // for loop
	return rng
}

// polygonForRing returns a polygon for the given ring, this will destroy the ring.
func polygonForRing(ctx context.Context, rng [][2]float64) (plyg [][][2]float64) {
	if debug {
		log.Printf("turn ring into polygon.")
	}

	if len(rng) <= 2 {
		return nil
	}

	// normalize ring
	cmp.RotateToLeftMostPoint(rng)

	pIdx := func(i int) int {
		if i == 0 {
			return len(rng) - 1
		}
		return i - 1
	}
	nIdx := func(i int) int {
		if i == len(rng)-1 {
			return 0
		}
		return i + 1
	}

	// Allocate space for the initial ring.
	plyg = make([][][2]float64, 1)

	// Remove bubbles. There are two types of bubbles we have to look for.
	// 1. ab … bc, in which case we need to hold on to b.
	//    It is possible that b is absolutely not necessary. It could lie on the line between a and c, in which case
	//    we should remove the extra point.
	// 2. ab … ba, which case we do not need to have b in the ring.

	// let's build an index of where the points that we are walking are. That way when we encounter the same
	// point we are able to “jump” to that point.
	ptIndex := map[[2]float64]int{}
	var ok bool
	var idx int

	// Let's walk the points
	for i := 0; i < len(rng); i++ {
		// Context has been cancelled.
		if ctx.Err() != nil {
			return nil
		}

		// check to see if we have already seen this point.
		if idx, ok = ptIndex[rng[i]]; !ok {
			ptIndex[rng[i]] = i
			continue
		}

		// We need to figure out which type of bubble this is.
		pidx, nidx := pIdx(idx), nIdx(i)

		// Clear out ptIndex of the values we are going to cut.
		for j := idx; j <= i; j++ {
			delete(ptIndex, rng[j])
		}

		// ab…ba ring. So we need to remove all the way to a.
		if nidx != pidx && cmp.PointEqual(rng[pidx], rng[nidx]) {
			if debug {
				log.Printf("bubble type ab…ba: (% 5v)(% 5v) … (% 5v)(% 5v)", pidx, idx, i, nidx)
			}

			// Delete the a points as well.
			delete(ptIndex, rng[pidx])

			sliver := cut(&rng, pidx, nidx)
			// remove ther start ab
			sliver = sliver[2:]
			if len(sliver) >= 3 { // make a copy to free up memory.
				ps := make([][2]float64, len(sliver))
				copy(ps, sliver)
				cmp.RotateToLeftMostPoint(ps)
				plyg = append(plyg, ps)
			}

			if i = idx - 1; i < 0 {
				i = 0
			}
			continue
		}

		// do a quick check to see if b is on ac
		removeB := planar.IsPointOnLine(rng[i], rng[pidx], rng[nidx])

		// ab … bc
		if debug {
			log.Printf("bubble type ab…bc: (% 5v)(% 5v) … (% 5v)(% 5v) == %t", pidx, idx, i, nidx, removeB)
		}

		sliver := cut(&rng, idx, i)
		if len(sliver) >= 3 {
			cmp.RotateToLeftMostPoint(sliver)
			plyg = append(plyg, sliver)
		}

		if removeB {
			cut(&rng, idx, idx+1)
			if idx == 0 {
				break
			}
			i = idx - 1
		}
	}

	if len(rng) <= 2 {
		if debug {
			log.Println("rng:", rng)
			log.Println("plyg:", plyg)
			panic("main ring is not correct!")
		}
		return nil
	}

	plyg[0] = make([][2]float64, len(rng))
	copy(plyg[0], rng)
	return plyg
}

func (eit *edgeIndexTriangles) PolygonForTriangle(ctx context.Context, idx int, seen map[int]bool) (plyg [][][2]float64) {
	// Get the external ring for the given triangle.
	return polygonForRing(ctx, eit.ringForTriangle(ctx, idx, seen))
}
