package subdivision

import (
	"errors"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
)

var (
	ErrInvalidPseudoPolygonSize = errors.New("invalid polygon, not enough points.")
)

type byLength []geom.Line

func (l byLength) Len() int      { return len(l) }
func (l byLength) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l byLength) Less(i, j int) bool {
	lilen := l[i].LenghtSquared()
	ljlen := l[j].LenghtSquared()
	if lilen == ljlen {
		if cmp.PointEqual(l[i][0], l[j][0]) {
			return cmp.PointLess(l[i][1], l[j][1])
		}
		return cmp.PointLess(l[i][0], l[j][0])
	}
	return lilen < ljlen
}
