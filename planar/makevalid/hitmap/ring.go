package hitmap

import (
	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/internal/rtreego"
	"github.com/go-spatial/geom/planar"
)

type segRect struct {
	cx     float64
	deltax float64
	seg    [2][2]float64
	isvert bool
	m, b   float64
	rect   *rtreego.Rect
}

func (sr *segRect) Bounds() *rtreego.Rect {
	if sr == nil {
		return nil
	}
	return sr.rect
}

func (sr *segRect) y4x(x float64) float64 { return sr.m*x + sr.b }

const smallep = 0.00001

func minmaxbyx(seg [2][2]float64) [2][2]float64 {
	if seg[0][0] == seg[1][0] {
		if seg[0][1] > seg[1][1] {
			return [2][2]float64{seg[1], seg[0]}
		}
		return seg
	}
	if seg[0][0] > seg[1][0] {
		return [2][2]float64{seg[1], seg[0]}
	}
	return seg
}

func newSegRects(segs ...[2][2]float64) (segRects []*segRect) {
	for i := range segs {
		minx, maxx := segs[i][0][0], segs[i][1][0]
		if minx > maxx {
			minx, maxx = maxx, minx
		}
		deltax := maxx - minx
		cx := minx + (deltax / 2)
		if deltax <= 0 {
			// https://github.com/dhconnelly/rtreego/issues/18
			deltax = smallep
		}

		m, b, defined := planar.Slope(segs[i])

		rect, _ := rtreego.NewRect(rtreego.Point{cx}, []float64{deltax})
		segRects = append(segRects, &segRect{
			cx:     cx,
			deltax: deltax,
			// We store the segment with the smaller x point first.
			seg:    minmaxbyx(segs[i]),
			m:      m,
			b:      b,
			isvert: !defined,
			rect:   rect,
		})
	}
	return segRects

}

func createSegments(ls [][2]float64, isClosed bool) (segs [][2][2]float64, err error) {
	if len(ls) <= 1 {
		return nil, ErrInvalidLineString

	}
	i := 0
	for j := 1; j < len(ls); j++ {
		segs = append(segs, [2][2]float64{ls[i], ls[j]})
		i = j
	}
	if isClosed {
		segs = append(segs, [2][2]float64{ls[len(ls)-1], ls[0]})
	}
	return segs, nil
}

type Ring struct {
	tree  *rtreego.Rtree
	bbox  *geom.Extent
	verts map[[2]float64]struct{}
	Label Label
}

// Contains returns weather the point is contained by the ring, if the point is on the border it is considered not contained.
func (r Ring) Contains(pt [2]float64) bool {
	if r.tree == nil {
		return false
	}
	if !r.bbox.ContainsPoint(pt) {
		return false
	}
	bb, _ := rtreego.NewRect(rtreego.Point{pt[0]}, []float64{r.bbox.XSpan()})
	results := r.tree.SearchIntersect(bb)
	if len(results) == 0 {
		return false
	}

	var segs []*segRect
	for i := range results {
		seg, ok := results[i].(*segRect)
		// don't know how to deal with this rect, ignore.
		if !ok {
			continue
		}
		// Is the x of the point even on the line?
		if seg.seg[0][0] > pt[0] || pt[0] > seg.seg[1][0] {
			continue
		}
		if seg.isvert {
			if seg.cx == pt[0] && seg.seg[0][1] <= pt[1] && pt[1] <= seg.seg[1][1] {
				// We ended up on a boundry, assume it's not contained.
				return false
			}
			// ignore this vertical line.
			continue
		}
		ly := seg.y4x(pt[0])

		// This may need a different comparison method due to floating point precision issues.
		if pt[1] == ly {
			// We ended up on a boundry, assume it's not contained.
			return false
		}
		if pt[1] < ly {
			continue
		}
		// Save line to process later.
		segs = append(segs, seg)
	}

	found := 0
	// seen := map[float64]struct{}{}
	for _, seg := range segs {
		if _, vok := r.verts[[2]float64{pt[0], seg.y4x(pt[0])}]; vok {
			return r.Contains([2]float64{pt[0] + 0.0001, pt[1]})
		}
		found++
	}
	// if it's odd the point is contained.
	return found%2 != 0
}

func NewRing(ring [][2]float64, label Label) (*Ring, error) {
	bb := geom.NewExtent(ring...)
	verts := make(map[[2]float64]struct{})
	for i := range ring {
		verts[ring[i]] = struct{}{}
	}
	segs, err := createSegments(ring, true)
	if err != nil {
		return nil, err
	}
	rects := newSegRects(segs...)
	var rs = make([]rtreego.Spatial, len(rects))
	for i := range rects {
		rs[i] = rects[i]
	}
	rtree := rtreego.NewTree(1, 2, 5, rs...)
	return &Ring{tree: rtree, bbox: bb, verts: verts, Label: label}, nil
}

type bySmallestBBArea []*Ring

// Sort Interface

func (rs bySmallestBBArea) Len() int { return len(rs) }
func (rs bySmallestBBArea) Less(i, j int) bool {
	ia, ja := rs[i].bbox.Area(), rs[j].bbox.Area()
	if ia == ja {
		return rs[i].Label == Outside
	}
	return ia < ja
}
func (rs bySmallestBBArea) Swap(i, j int) { rs[i], rs[j] = rs[j], rs[i] }
