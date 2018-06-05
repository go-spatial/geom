package tegola

import (
	"context"
	"errors"
	"log"
	"math"
	"sort"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
	"github.com/go-spatial/geom/planar/intersect"
	"github.com/go-spatial/geom/planar/makevalid/hitmap"
)

// asSegments calls the AsSegments functions and flattens the array of segments that are returned.
func asSegments(g geom.Geometry) (segs []geom.Line, err error) {
	switch g := g.(type) {
	case geom.LineString:
		return g.AsSegments()
	case geom.MultiLineString:
		s, err := g.AsSegments()
		if err != nil {
			return nil, err
		}
		for i := range s {
			segs = append(segs, s[i]...)
		}
		return segs, nil
	case geom.Polygon:
		s, err := g.AsSegments()
		if err != nil {
			return nil, err
		}
		for i := range s {
			segs = append(segs, s[i]...)
		}
		return segs, nil
	case geom.MultiPolygon:
		s, err := g.AsSegments()
		if err != nil {
			return nil, err
		}
		for i := range s {
			for j := range s[i] {
				segs = append(segs, s[i][j]...)
			}
		}
		return segs, nil
	}
	return nil, errors.New("Unsupported")
}

type ByXYPoint [][2]float64

func (a ByXYPoint) Len() int           { return len(a) }
func (a ByXYPoint) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByXYPoint) Less(i, j int) bool { return cmp.PointLess(a[i], a[j]) }

// destructure will take a multipolygon, break up the polygon into a set of segments that have the following characteristics:
// 1. no segment will intersect with another segment, other then at the end points; or colinear and partial-coliner lines.
// 2. normalize direction of line segments to left to right
// 3. line segments are generally unique.
// 4. line segments outside of the clipbox will be clipped
func destructure(ctx context.Context, clipbox *geom.Extent, multipolygon *geom.MultiPolygon) ([]geom.Line, error) {

	segments, err := asSegments(*multipolygon)
	if err != nil {
		if debug {
			log.Printf("asSegments returned error: %v", err)
		}
		return nil, err
	}
	gext, err := geom.NewExtentFromGeometry(multipolygon)
	if err != nil {
		return nil, err
	}

	// Let's see if our clip box is bigger then our polygon.
	// if it is we don't need the clip box.
	hasClipbox := clipbox != nil && !clipbox.Contains(gext)
	// Let's get the edges of our clipbox; as segments and add it to the begining.
	if hasClipbox {
		edges := clipbox.Edges(nil)
		segments = append([]geom.Line{
			geom.Line(edges[0]), geom.Line(edges[1]),
			geom.Line(edges[2]), geom.Line(edges[3]),
		}, segments...)
	}
	ipts := make(map[int][][2]float64)

	// Lets find all the places we need to split the lines on.
	eq := intersect.NewEventQueue(segments)
	eq.FindIntersects(ctx, true, func(src, dest int, pt [2]float64) error {
		ipts[src] = append(ipts[src], pt)
		ipts[dest] = append(ipts[dest], pt)
		return nil
	})

	// Time to start splitting lines. if we have a clip box we can ignore the first 4 (0,1,2,3) lines.

	nsmap := make(map[geom.Line]struct{})

	for i := 0; i < len(segments); i++ {
		pts := append([][2]float64{segments[i][0], segments[i][1]}, ipts[i]...)

		// Normalize the direction of the points.
		sort.Sort(ByXYPoint(pts))

		for j := 1; j < len(pts); j++ {
			if cmp.PointEqual(pts[j-1], pts[j]) {
				continue
			}
			nl := geom.Line{pts[j-1], pts[j]}
			if hasClipbox && !clipbox.ContainsLine(nl) {
				// Not in clipbox discard segment.
				continue
			}
			nsmap[nl] = struct{}{}
		}
		if ctx.Err() != nil {
			return nil, err
		}
	}

	nsegs := make([]geom.Line, len(nsmap))
	{
		i := 0
		for s := range nsmap {
			nsegs[i] = s
			i++
		}
	}
	return nsegs, nil
}

func snapToGrid(tolerance float64, segments []geom.Line) []geom.Line {
	// We first need to get all the unque x values, and track how they relate to the segments and their points.
	type pt struct {
		ptidx int
		// index to the segment
		index int
		// left point of the segment, otherwise it's the right point
		isLeft   bool
		modified bool
	}

	xs := make(map[float64][]*pt)
	ys := make(map[float64][]*pt)

	shiftX := func(from, to float64) {
		for j := range xs[from] {
			idx, ptidx := xs[from][j].index, xs[from][j].ptidx
			segments[idx][ptidx][0] = to
			xs[from][j].modified = true
			xs[to] = append(xs[to], xs[from][j])
		}
		delete(xs, from)
	}
	shiftY := func(from, to float64) {
		for j := range ys[from] {
			idx, ptidx := xs[from][j].index, xs[from][j].ptidx
			segments[idx][ptidx][1] = to
			ys[from][j].modified = true
			ys[to] = append(ys[to], ys[from][j])
		}
		delete(ys, from)
	}

	for i := range segments {
		pt1, pt2 := segments[i][0], segments[i][1]
		isLeft := cmp.PointEqual(pt1, pt2)
		ppt1 := &pt{
			ptidx:  0,
			index:  i,
			isLeft: isLeft,
		}
		xs[pt1[0]] = append(xs[pt1[0]], ppt1)
		ys[pt1[1]] = append(ys[pt1[1]], ppt1)
		ppt2 := &pt{
			ptidx:  1,
			index:  i,
			isLeft: !isLeft,
		}
		xs[pt2[0]] = append(xs[pt2[0]], ppt2)
		ys[pt2[1]] = append(ys[pt2[1]], ppt2)

	}
	xkeys, ykeys := make([]float64, 0, len(xs)), make([]float64, 0, len(ys))
	for k, _ := range xs {
		xkeys = append(xkeys, k)
	}
	sort.Float64s(xkeys)

	for i := 1; i < len(xkeys)-1; {
		dist1 := math.Abs(xkeys[i] - xkeys[i-1])
		dist2 := math.Abs(xkeys[i+1] - xkeys[i])

		if dist1 >= tolerance && dist2 >= tolerance {
			// no change.
			i++
			continue
		}

		switch {
		case dist1 < tolerance && dist2 < tolerance:
			if dist1 <= dist2 {
				shiftX(xkeys[i], xkeys[i-1])
			} else {
				shiftX(xkeys[i], xkeys[i+1])
			}
		case dist1 < tolerance:
			shiftX(xkeys[i], xkeys[i-1])
		default:
			shiftX(xkeys[i], xkeys[i+1])
		}
		i = i + 2
	}

	sort.Float64s(ykeys)
	for i := 1; i < len(ykeys)-1; {
		dist1 := math.Abs(ykeys[i] - ykeys[i-1])
		dist2 := math.Abs(ykeys[i+1] - ykeys[i])

		if dist1 >= tolerance && dist2 >= tolerance {
			// no change.
			i++
			continue
		}

		switch {
		case dist1 < tolerance && dist2 < tolerance:
			if dist1 <= dist2 {
				shiftY(ykeys[i], ykeys[i-1])
			} else {
				shiftY(ykeys[i], ykeys[i+1])
			}
		case dist1 < tolerance:
			shiftY(ykeys[i], ykeys[i-1])
		default:
			shiftY(ykeys[i], ykeys[i+1])
		}
		i = i + 2
	}

	segmap := make(map[geom.Line]int)

	for i := range segments {
		if cmp.PointEqual(segments[i][0], segments[i][1]) {
			// Both points are equal, we can ignore it.
			continue
		}
		// We have not seen this line, so we can add it to our map.
		if _, ok := segmap[segments[i]]; !ok {
			segmap[segments[i]] = i
		}
	}

	idxs := make([]int, 0, len(segmap))
	for _, v := range segmap {
		idxs = append(idxs, v)
	}
	sort.Ints(idxs)
	segs := make([]geom.Line, len(idxs))
	for i := range idxs {
		segs[i] = segments[idxs[i]]
	}
	return segs

}

func (mv *Makevalid) makevalidPolygon(ctx context.Context, clipbox *geom.Extent, multipolygon *geom.MultiPolygon) (*geom.MultiPolygon, error) {
	if debug {
		log.Printf("*Step  1 : Destructure the geometry into segments w/ the clipbox applied.")
	}
	segs, err := destructure(ctx, clipbox, multipolygon)
	if err != nil {
		if debug {
			log.Printf("Destructure returned err %v", err)
		}
		return nil, err
	}
	if len(segs) == 0 {
		if debug {
			log.Printf("Step   1a: Segments are zero.")
			log.Printf("\t multiPolygon: %+v", multipolygon)
			log.Printf("\n clipbox:      %+v", clipbox)
		}
		return nil, nil
	}
	if debug {
		log.Printf("Step   2 : Convert segments to linestrings to use in triangleuation.")
	}
	geomSegments := make(geom.MultiLineString, len(segs))
	for i := range segs {
		geomSegments[i] = segs[i][:]
	}
	hm, err := hitmap.NewFromPolygons(nil, (*multipolygon)...)
	if err != nil {
		return nil, err
	}
	if debug {
		log.Printf("Step   3 : generate triangles")
	}
	edix, err := newEdgeIndexTriangles(ctx, hm, geomSegments)
	if err != nil {
		return nil, err
	}
	if debug {
		log.Printf("Step   4 : generate multipolygon from triangles")
	}
	mplygs := geom.MultiPolygon(edix.MultiPolygon(ctx))

	return &mplygs, nil

}

func (mv *Makevalid) Makevalid(ctx context.Context, geo geom.Geometry, clipbox *geom.Extent) (geometry geom.Geometry, didClip bool, err error) {

	switch g := geo.(type) {

	case geom.LineStringer, geom.MultiLineStringer, geom.Pointer, geom.MultiPointer:
		if mv.Clipper != nil {
			gg, err := mv.Clipper.Clip(ctx, geo, clipbox)
			if err != nil {
				return nil, false, err
			}
			return gg, true, nil
		}
		return geo, false, nil
	case geom.Polygoner:
		if debug {
			log.Printf("Working on Polygoner: %v", geo)
		}
		mp := geom.MultiPolygon{g.LinearRings()}
		vmp, err := mv.makevalidPolygon(ctx, clipbox, &mp)
		if err != nil {
			return nil, false, err
		}
		if debug {
			log.Printf("Returning on Polygon: %T", vmp)
		}
		return vmp, true, nil
	case geom.MultiPolygoner:
		if debug {
			log.Printf("Working on MultiPolygoner: %v", geo)
		}
		mp := geom.MultiPolygon(g.Polygons())
		vmp, err := mv.makevalidPolygon(ctx, clipbox, &mp)
		if err != nil {
			return nil, false, err
		}
		if debug {
			log.Printf("Returning on MultiPolygon: %T", vmp)
		}
		return vmp, true, nil
	}
	if debug {
		log.Printf("Got an unknown geometry %T", geo)
	}
	return nil, false, geom.ErrUnknownGeometry{geo}

}
