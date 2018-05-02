package hitmap

import (
	"sort"

	"github.com/go-spatial/geom"
)

// PolygonHM implements a basic hit map that gives the label for a point based on the order of the rings.
type PolygonHM struct {
	// clipBox this is going to be either the clipping area or the bouding box of all the rings.
	// This allows us to quickly determine if a point is outside the set of rings.
	clipBox *geom.Extent
	// These are the rings
	rings []*Ring
}

// NewFromPolygons assumes that the outer ring of each polygon is inside, and each inner ring is inside.
func NewFromPolygons(clipbox *geom.Extent, plys ...[][][2]float64) (*PolygonHM, error) {

	hm := &PolygonHM{
		clipBox: new(geom.Extent),
	}
	for i := range plys {
		if len(plys[i]) == 0 {
			continue
		}
		{
			ring, err := NewRing(plys[i][0], Inside)
			if err != nil {
				return nil, err
			}
			if clipbox == nil {
				// add to the bb of ring to the hm clipbox
				hm.clipBox.Add(ring.bbox)
			}
			hm.rings = append(hm.rings, ring)
		}
		if len(plys[i]) <= 1 {
			continue
		}
		for j := range plys[i][1:] {
			// plys we assume the first ring is inside, and all other rings are outside.
			ring, err := NewRing(plys[i][j+1], Outside)
			if err != nil {
				return nil, err
			}
			if clipbox == nil {
				// add to the bb of ring to the hm clipbox
				hm.clipBox.Add(ring.bbox)
			}
			hm.rings = append(hm.rings, ring)
		}
	}
	sort.Sort(bySmallestBBArea(hm.rings))
	return hm, nil
}

// LabelFor returns the label for the given point.
func (hm *PolygonHM) LabelFor(pt [2]float64) Label {
	// nil clipBox contains all points.
	if hm == nil || !hm.clipBox.ContainsPoint(pt) {
		return Outside
	}

	// TODO(gdey): See if it make sense to change data structures here.
	// For now we iterate through all the rings, but maybe an r-tree would make
	// sense here, or some "smart" container that would use an r-tree or iterate
	// through all the points depending on the number of things.

	// We assume the []*Rings are sorted in from smallest area to largest area.
	for i := range hm.rings {
		if hm.rings[i].Contains(pt) {
			return hm.rings[i].Label
		}
	}
	return Outside
}

// Extent returns the extent of the hitmap.
func (hm *PolygonHM) Extent() [4]float64 { return hm.clipBox.Extent() }

// Area returns the area covered by the hitmap.
func (hm *PolygonHM) Area() float64 { return hm.clipBox.Area() }
