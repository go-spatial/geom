package slippy

import (
	"math"

	"github.com/go-spatial/geom"
)

func NewTile(z, x, y uint, buffer float64) *Tile {
	return &Tile{
		z:      z,
		x:      x,
		y:      y,
		Buffer: buffer,
	}
}

// Tile describes a slippy tile.
type Tile struct {
	// zoom
	z uint
	// column
	x uint
	// row
	y uint
	// buffer will add a buffer to the tile bounds. this buffer is expected to use the same units as the SRID
	// of the projected tile (i.e. WebMercator = pixels, 3395 = meters)
	Buffer float64
}

func NewTileLatLon(z uint, lat, lon, buffer float64, srid uint64) *Tile {
	x := Lon2Tile(z, lon)
	y := Lat2Tile(z, lat)

	return &Tile{
		z:      z,
		x:      x,
		y:      y,
		Buffer: buffer,
	}
}

func (t *Tile) ZXY() (uint, uint, uint) { return t.z, t.x, t.y }

func Lat2Tile(zoom uint, lat float64) (y uint) {
	lat_rad := lat * math.Pi / 180

	return uint(math.Exp2(float64(zoom)) *
		(1.0 - math.Log(
			math.Tan(lat_rad) +
				(1 / math.Cos(lat_rad))) / math.Pi)) /
		2.0

}

func Lon2Tile(zoom uint, lon float64) (x uint) {
	return uint(math.Exp2(float64(zoom)) * (lon + 180.0) / 360.0)
}

// Tile2Lon will return the west most longitude
func Tile2Lon(zoom, x uint) float64 { return float64(x)/math.Exp2(float64(zoom))*360.0 - 180.0 }

// Tile2Lat will return the east most Latitude
func Tile2Lat(zoom, y uint) float64 {
	var n float64 = math.Pi
	if y != 0 {
		n = math.Pi - 2.0*math.Pi*float64(y)/math.Exp2(float64(zoom))
	}

	return 180.0 / math.Pi * math.Atan(0.5*(math.Exp(n)-math.Exp(-n)))
}

// Bounds returns the bounds of the Tile as defined by the East most longitude, North most latitude, West most longitude, South most latitude.
func (t *Tile) Bounds() *geom.Extent {
	return t.Extent(geom.WGS84)
}

/*
	// Keep this comment as it is a guide for how we can take bounds and a srid and convert it to Extents and Buffereded Extents.
	// This is how we convert from the Bounds, and TileSize to Extent for Webmercator.
	bounds := t.Bounds()
	east,north,west, south := bounds[0],bounds[1],bounds[2],bounds[3]

	TileSize := 4096.0
	// Convert bounds to coordinates in webmercator.
	c, err := webmercator.PToXY(east, north, west, south)
	log.Println("c", c, "err", err)

	// Turn the Coordinates into an Extent (minx, miny, maxx, maxy)
	// Here is where the origin flip happens if there is one.
	extent := geom.NewBBox(
		[2]float64{c[0], c[1]},
		[2]float64{c[2], c[3]},
	)

	// A Span is just MaxX - MinX
	xspan := extent.XSpan()
	yspan := extent.YSpan()

	log.Println("Extent", extent, "MinX", extent.MinX(), "MinY", extent.MinY(), "xspan", xspan, "yspan", yspan)

	// To get the Buffered Extent, we just need the extent and the Buffer size.
	// Convert to tile coordinates. Convert the meters (WebMercator) into pixels of the tile..
	nx := float64(int64((c[0] - extent.MinX()) * TileSize / xspan))
	ny := float64(int64((c[1] - extent.MinY()) * TileSize / yspan))
	mx := float64(int64((c[2] - extent.MinX()) * TileSize / xspan))
	my := float64(int64((c[3] - extent.MinY()) * TileSize / yspan))

	// Expend by the that number of pixels. We could also do the Expand on the Extent instead, of the Bounding Box on the Pixel.
	mextent := geom.NewBBox([2]float64{nx, ny}, [2]float64{mx, my}).ExpandBy(64)
	log.Println("mxy[", nx, ny, mx, my, "]", "err", err, "mext", mextent)

	// Convert Pixel back to meters.
	bext := geom.NewBBox(
		[2]float64{
			(mextent.MinX() * xspan / TileSize) + extent.MinX(),
			(mextent.MinY() * yspan / TileSize) + extent.MinY(),
		},
		[2]float64{
			(mextent.MaxX() * xspan / TileSize) + extent.MinX(),
			(mextent.MaxY() * yspan / TileSize) + extent.MinY(),
		},
	)
	log.Println("bext", bext)
*/

var WebMercMax = 20037508.34

// tile to web mercator
func Tile2WebX(zoom uint, n uint) float64 {
	res := (WebMercMax * 2) / math.Exp2(float64(zoom))

	return -WebMercMax + float64(n)*res
}

func Tile2WebY(zoom uint, n uint) float64 {
	res := (WebMercMax * 2) / math.Exp2(float64(zoom))

	return WebMercMax - float64(n)*res
}

type Coord2ExtentFunc = func(z, n uint) float64

// TODO(arolek): support alternative SRIDs. Currently this assumes 3857
// Extent will return the tile extent excluding the tile's buffer and the Extent's SRID
func (t *Tile) Extent(srid uint64) (extent *geom.Extent) {

	var minX, minY, maxX, maxY float64

	switch srid {
	case geom.WGS84:
		minX = Tile2Lon(t.z, t.x)
		minY = Tile2Lat(t.z, t.y)
		maxX = Tile2Lon(t.z, t.x+1)
		maxY = Tile2Lat(t.z, t.y+1)
	case geom.WebMercator:
		minX = Tile2WebX(t.z, t.x)
		minY = Tile2WebY(t.z, t.y)
		maxX = Tile2WebX(t.z, t.x+1)
		maxY = Tile2WebY(t.z, t.y+1)
	}

	// unbuffered extent
	return geom.NewExtent(
		[2]float64{minX, minY},
		[2]float64{maxX, maxY},
	)
}

// BufferedExtent will return the tile extent including the tile's buffer and the Extent's SRID
func (t *Tile) BufferedExtent(srid uint64) (bufferedExtent *geom.Extent) {
	extent := t.Extent(srid)

	// TODO(arolek): the following value is hard coded for MVT, but this concept needs to be abstracted to support different projections
	mvtTileWidthHeight := 4096.0
	// the bounds / extent
	mvtTileExtent := [4]float64{
		0 - t.Buffer, 0 - t.Buffer,
		mvtTileWidthHeight + t.Buffer, mvtTileWidthHeight + t.Buffer,
	}

	xspan := extent.MaxX() - extent.MinX()
	yspan := extent.MaxY() - extent.MinY()

	bufferedExtent = geom.NewExtent(
		[2]float64{
			(mvtTileExtent[0] * xspan / mvtTileWidthHeight) + extent.MinX(),
			(mvtTileExtent[1] * yspan / mvtTileWidthHeight) + extent.MinY(),
		},
		[2]float64{
			(mvtTileExtent[2] * xspan / mvtTileWidthHeight) + extent.MinX(),
			(mvtTileExtent[3] * yspan / mvtTileWidthHeight) + extent.MinY(),
		},
	)
	return bufferedExtent
}

// TODO (ear7h): sibling support
// RangeFamilyAt calls f on every tile vertically related to t at the specified zoom
func (t *Tile) RangeFamilyAt(zoom uint, f func(*Tile) error) error {
	// handle ancestors and self
	if zoom <= t.z {
		mag := t.z - zoom
		arg := NewTile(zoom, t.x>>mag, t.y>>mag, t.Buffer)
		return f(arg)
	}

	// handle descendants
	mag := zoom - t.z
	delta := uint(math.Exp2(float64(mag)))

	leastX := t.x << mag
	leastY := t.y << mag

	for x := leastX; x < leastX+delta; x++ {
		for y := leastY; y < leastY+delta; y++ {
			err := f(NewTile(zoom, x, y, 0))
			if err != nil {
				return err
			}
		}
	}

	return nil
}
