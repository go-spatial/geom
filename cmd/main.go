package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/go-spatial/geom/winding"

	"github.com/go-spatial/geom/planar/clip"

	"github.com/go-spatial/geom/cmp"
	"github.com/go-spatial/geom/planar"

	"github.com/go-spatial/geom/planar/makevalid"
	"github.com/go-spatial/geom/planar/makevalid/hitmap"
	"github.com/go-spatial/geom/planar/makevalid/walker"
	"github.com/go-spatial/geom/planar/simplify"
	"github.com/go-spatial/geom/planar/triangulate/delaunay"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/encoding/mvt"
	"github.com/go-spatial/geom/encoding/wkt"
	"github.com/go-spatial/geom/slippy"
)

var simplifyGeo = flag.Bool("simplify", true, "simplify the wkt before running makevalid")
var tag = flag.String("tag", "", "place in an additional directory")
var buffer = flag.Int("buffer", 64, "Buffer to place around the tile")
var help = flag.Bool("help", false, "print this message")
var mvtExtent = flag.Float64("extent", 4096, "extent of the mvt tile")

func usage() {
	fmt.Fprintf(
		os.Stderr,
		"%v takes input wkt file and z/x/y slippy tile and outputs triangles that make up the location\nusage %[1]v\n\t$ %[1]v [options] z/x/y input.wkt \noptions:\n",
		os.Args[0],
	)
	flag.PrintDefaults()
	os.Exit(1)
}

func readInputWKT(filename string) (geom.Geometry, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return wkt.Decode(file)
}

type outfile struct {
	tile   *slippy.Tile
	format string
}
type outfilefile struct {
	tag string
	*os.File
}

func (of outfile) NewFile(item string) *outfilefile {
	f, err := os.Create(fmt.Sprintf(of.format, item))
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to open %v file: %v", item, err)
		os.Exit(2)
	}
	return &outfilefile{File: f, tag: item}
}

func (off *outfilefile) WriteWKTGeom(geos ...geom.Geometry) *outfilefile {
	for _, g := range geos {
		if err := wkt.Encode(off, g); err != nil {
			log.Printf("failed to encode geo: %v: %v", off.tag, err)
			return off
		}
		off.WriteString("\n")
	}
	return off
}

func newOutFile(tile *slippy.Tile, tag string) outfile {
	path := fmt.Sprintf("%v/%v/%v", tile.Z, tile.X, tile.Y)
	if tag != "" {
		path = fmt.Sprintf("%v/%v", path, tag)
	}

	os.MkdirAll(path, os.ModePerm)
	return outfile{
		tile:   tile,
		format: fmt.Sprintf("%v/%%v.wkt", path),
	}
}

func main() {
	flag.Parse()
	if len(flag.Args()) < 2 || *help {
		usage()
	}
	if *help {
		usage()
		return
	}
	parts := strings.Split(flag.Args()[0], "/")
	if len(parts) < 3 {
		fmt.Fprintf(os.Stderr, "invalid first parameters expected slippy tile\n Got %v\n", flag.Args()[1])
		usage()
	}

	z, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unabled to parse z: %v", err)
		usage()
	}
	x, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unabled to parse x: %v", err)
		usage()
	}
	y, err := strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unabled to parse y: %v", err)
		usage()
	}
	tile := slippy.NewTile(uint(z), uint(x), uint(y))
	fileTemplate := newOutFile(tile, *tag)
	geo, err := readInputWKT(flag.Args()[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unabled to parse/open `%v` : %v", os.Args, err)
		usage()
	}
	ctx := context.Background()
	/*
		plywkt, err := wkt.EncodeString(geo)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Polygon:\n%v\n", plywkt)
	*/
	order := winding.Order{}
	grid3857, _ := slippy.NewGrid(3857)

	var clipRegion *geom.Extent
	{
		webs := slippy.PixelsToNative(grid3857, tile.Z, uint(*buffer))
		ext, _ := slippy.Extent(grid3857, tile)
		clipRegion = ext.ExpandBy(webs)
	}

	if *simplifyGeo {
		simp := simplify.DouglasPeucker{
			Tolerance: slippy.PixelsToNative(grid3857, tile.Z, 10.0),
		}

		var err error
		geo, err = planar.Simplify(ctx, simp, geo)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unabled to simplify geo : %v", err)
			usage()
		}
		sgeofile := fileTemplate.NewFile("simplified_geo")
		sgeofile.WriteWKTGeom(geo)
		sgeofile.Close()
	}

	var hm planar.HitMapper
	{
		hm, err = hitmap.New(clipRegion, geo)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unabled to create hm for geo : %v", err)
			usage()
		}
	}

	{
		mv := makevalid.Makevalid{
			Hitmap:  hm,
			Clipper: clip.Default,
			Order:   order,
		}

		mkvgeo, _, err := mv.Makevalid(ctx, geo, clipRegion)
		if err != nil {
			log.Printf("Got error using original makevalid: %v", err)

		} else {
			mvgeoFile := fileTemplate.NewFile("original_makevalid")
			mvgeoFile.WriteWKTGeom(mkvgeo)
			mvgeoFile.Close()
			ext, _ := slippy.Extent(grid3857, tile)
			if mvtgeo := mvt.PrepareGeo(mkvgeo, ext, *mvtExtent); mvtgeo != nil {
				mvtgeof := fileTemplate.NewFile("original_mvt_geo")
				mvtgeof.WriteWKTGeom(mvtgeo)
				mvtgeof.Close()
			}
		}
	}

	var mp geom.MultiPolygon
	switch g := geo.(type) {
	case geom.MultiPolygon:
		mp = g
	case *geom.MultiPolygon:
		if g == nil {
			return
		}
		mp = *g
	case geom.Polygon:
		mp = geom.MultiPolygon{g}
	case *geom.Polygon:
		if g == nil {
			return
		}
		mp = geom.MultiPolygon{*g}
	default:
		fmt.Fprintf(os.Stderr, "Unsupported geometry type: %t", geo)
		usage()
	}

	segs, err := makevalid.Destructure(context.Background(), cmp.HiCMP, clipRegion, &mp)
	if err != nil {
		log.Printf("Destructure returned err %v", err)
		return
	}
	if len(segs) == 0 {
		log.Printf("Step   1a: Segments are zero.")
		return
	}
	triangulator := delaunay.GeomConstrained{
		Constraints: segs,
	}
	allTriangles, err := triangulator.Triangles(ctx, false)
	if err != nil {
		log.Printf("triangulator returned err %v", err)
		return
	}

	sofile := fileTemplate.NewFile("outside_triangles_makevalid_steps")
	sifile := fileTemplate.NewFile("inside_triangles_makevalid_steps")
	defer sofile.Close()
	defer sifile.Close()

	sifile.WriteWKTGeom(clipRegion)
	sofile.WriteWKTGeom(clipRegion)

	var outsideTriangles, insideTriangles []geom.Triangle
	{
		inTrisFile := fileTemplate.NewFile("inside_triangles")
		outTrisFile := fileTemplate.NewFile("outside_triangles")
		fmt.Printf("Tagging triangles: %v\n#", len(allTriangles))
		numTri := len(allTriangles)
		size := int(math.Log10(float64(numTri)))

		for i := range allTriangles {
			center := allTriangles[i].Center()
			lbl := hm.LabelFor(center)
			if lbl == planar.Outside {
				sofile.WriteWKTGeom(center, allTriangles[i])
				outTrisFile.WriteWKTGeom(allTriangles[i])
				outsideTriangles = append(outsideTriangles, allTriangles[i])
				fmt.Printf("\rTagging triangle: % *d of %d as outside", size+1, i+1, numTri)
				continue
			}
			sifile.WriteWKTGeom(center, allTriangles[i])
			inTrisFile.WriteWKTGeom(allTriangles[i])
			insideTriangles = append(insideTriangles, allTriangles[i])
			fmt.Printf("\rTagging triangle: % *d of %d as  inside", size+1, i+1, numTri)
		}
		inTrisFile.Close()
		outTrisFile.Close()
	}

	ringFile := fileTemplate.NewFile("ring_from_inside_triangles")
	polygonFile := fileTemplate.NewFile("polygon_from_ring")

	var newMp geom.MultiPolygon
	triWalker := walker.New(insideTriangles)
	seen := make(map[int]bool, len(insideTriangles))
	for i := range insideTriangles {
		if seen[i] {
			continue
		}
		seen[i] = true
		ring := triWalker.RingForTriangle(ctx, i, seen)
		ringFile.WriteWKTGeom(geom.LineString(ring))
		ply := order.RectifyPolygon(walker.PolygonForRing(ctx, ring))
		polygonFile.WriteWKTGeom(geom.Polygon(ply))
		newMp = append(newMp, ply)
	}
	ringFile.Close()
	polygonFile.Close()

	{
		ext, _ := slippy.Extent(grid3857, tile)
		mvtgeo := mvt.PrepareGeo(newMp, ext, *mvtExtent)
		if mvtgeo != nil {
			mvtgeof := fileTemplate.NewFile("mvt_geo")
			mvtgeof.WriteWKTGeom(mvtgeo)
			mvtgeof.Close()
		}
	}

	/*
		triangles, err := makevalid.InsideTrianglesForMultiPolygon(context.Background(), extent, &mp, hm)
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to get triangles", err)
			os.Exit(1)
		}
	*/
}
