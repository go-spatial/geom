package slippy

import (
	"fmt"
	"math"
)

// TileGrid contains the tile layout, including ability to get WGS84 coordinates for tile extents
type TileGrid interface {
	ContainsIndex(zoom, x, y uint) bool
	GridSize(zoom uint) (x, y uint)
	MaxXY(zoom uint) (maxx, maxy uint)
	Lat2YIndex(zoom uint, lat float64) (gridy uint)
	Lon2XIndex(zoom uint, lon float64) (gridx uint)
	XIndex2Lon(zoom, x uint) (lon float64)
	YIndex2Lat(zoom, y uint) (lat float64)
}

func GetGrid(srid uint) TileGrid {
	switch srid {
	case 4326:
		return &grid{tileExtentRatio: 2, srid: srid}
	case 3857:
		return &grid{tileExtentRatio: 1, srid: srid}
	default:
		panic(fmt.Sprintf("unsupported srid: %v", srid))
	}
}

type grid struct {
	tileExtentRatio float64
	srid            uint
}

func (g *grid) ContainsIndex(zoom, x, y uint) bool {
	xsize, ysize := g.GridSize(zoom)
	if x < xsize && y < ysize {
		return true
	}

	return false
}

func (g *grid) GridSize(zoom uint) (x, y uint) {
	dim := uint(math.Exp2(float64(zoom)))
	return uint(float64(dim) * g.tileExtentRatio), dim
}

func (g *grid) MaxXY(zoom uint) (maxx, maxy uint) {
	xsize, ysize := g.GridSize(zoom)

	return xsize - 1, ysize - 1
}

func (g *grid) Lat2YIndex(zoom uint, lat float64) (gridy uint) {
	switch g.srid {
	case 3857:
		latRad := lat * math.Pi / 180
		return uint(math.Exp2(float64(zoom))*
			(1.0-math.Log(
				math.Tan(latRad)+
					(1/math.Cos(latRad)))/math.Pi)) /
			2.0
	case 4326:
		return uint(math.Exp2(float64(zoom)) * -(lat - 90.0) / (360.0 / g.tileExtentRatio))
	default:
		panic(fmt.Sprintf("unsupported srid: %v", g.srid))
	}
}

func (g *grid) Lon2XIndex(zoom uint, lon float64) (gridx uint) {
	switch g.srid {
	case 3857:
		fallthrough
	case 4326:
		return uint(math.Exp2(float64(zoom)) * (lon + 180.0) / (360.0 / g.tileExtentRatio))
	default:
		panic(fmt.Sprintf("unsupported srid: %v", g.srid))
	}
}

func (g *grid) XIndex2Lon(zoom, x uint) (lon float64) {
	switch g.srid {
	case 3857:
		fallthrough
	case 4326:
		return float64(x)/math.Exp2(float64(zoom))*(360.0/g.tileExtentRatio) - 180.0
	default:
		panic(fmt.Sprintf("unsupported srid: %v", g.srid))
	}
}

func (g *grid) YIndex2Lat(zoom, y uint) (lat float64) {
	switch g.srid {
	case 3857:
		var n = math.Pi
		if y != 0 {
			n = math.Pi - 2.0*math.Pi*float64(y)/math.Exp2(float64(zoom))
		}

		return 180.0 / math.Pi * math.Atan(0.5*(math.Exp(n)-math.Exp(-n)))
	case 4326:
		return -(180.0/math.Exp2(float64(zoom))*float64(y) - 90.0)
	default:
		panic(fmt.Sprintf("unsupported srid: %v", g.srid))
	}
}
