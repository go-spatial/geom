package qetriangulate_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
	"github.com/go-spatial/geom/encoding/wkt"
	"github.com/go-spatial/geom/internal/debugger"
	"github.com/go-spatial/geom/planar/triangulate/gdey/quadedge/quadedge"
	"github.com/go-spatial/geom/planar/triangulate/gdey/quadedge/subdivision"
)

func logEdges(sd *subdivision.Subdivision) {
	_ = sd.WalkAllEdges(func(e *quadedge.Edge) error {
		org := *e.Orig()
		dst := *e.Dest()

		str, err := wkt.EncodeString(
			geom.Line{
				[2]float64(org),
				[2]float64(dst),
			},
		)
		if err != nil {
			return err
		}

		fmt.Println(str)
		return nil
	})
}

func Draw(t *testing.T, rec debugger.Recorder, name string, pts ...[2]float64) {

	var (
		ffl        = debugger.FFL(0)
		badPoints  []geom.Point
		goodPoints []geom.Point
	)

	start := time.Now()
	sort.Sort(cmp.ByXY(pts))

	tri := geom.NewTriangleContainingPoints(pts...)
	ext := geom.NewExtent(pts...)
	sd := subdivision.New(geom.Point(tri[0]), geom.Point(tri[1]), geom.Point(tri[2]))

	for i, pt := range pts {
		if i != 0 && pts[i-1][0] == pt[0] && pts[i-1][1] == pt[1] {
			continue
		}
		bfpt := geom.Point(pt)
		if !sd.InsertSite(bfpt) {
			badPoints = append(badPoints, bfpt)
			continue
		}
		goodPoints = append(goodPoints, bfpt)
	}
	t.Logf("triangulation setup took: %v", time.Since(start))

	start = time.Now()

	rec.Record(
		tri,
		ffl,
		debugger.TestDescription{
			Category:    "frame:triangle",
			Description: "triangle frame.",
			Name:        name,
		},
	)

	rec.Record(
		ext,
		ffl,
		debugger.TestDescription{
			Category:    "frame:extent",
			Description: "extent frame.",
			Name:        name,
		},
	)

	if len(badPoints) > 0 {
		t.Logf("Failed to insert %v points\n", len(badPoints))
		for i, pt := range badPoints {
			rec.Record(
				pt,
				ffl,
				debugger.TestDescription{
					Category:    "initial:point:failed",
					Description: fmt.Sprintf("point:%v %v:failed", i, pt),
					Name:        name,
				},
			)
		}
	}
	for i, pt := range goodPoints {
		rec.Record(
			pt,
			ffl,
			debugger.TestDescription{
				Category:    "initial:point:failed",
				Description: fmt.Sprintf("point:%v %v:failed", i, pt),
				Name:        name,
			},
		)
	}

	count := 0
	_ = sd.WalkAllEdges(func(e *quadedge.Edge) error {
		org := *e.Orig()
		dst := *e.Dest()

		rec.Record(
			geom.Line{
				[2]float64(org),
				[2]float64(dst),
			},
			ffl,
			debugger.TestDescription{
				Category: fmt.Sprintf("edge:%v", count),
				Description: fmt.Sprintf(
					"edge:%v [%p]( %v %v, %v %v)",
					count, e,
					org[0], org[1],
					dst[0], dst[1],
				),
				Name: name,
			},
		)
		count++
		return nil
	})
	t.Logf("writing out triangulation setup info: %v", time.Since(start))
	start = time.Now()
	triangles, err := sd.Triangles(true)
	if err != nil {
		t.Logf("Got an error: %v", err)
	}
	t.Logf("triangulation took: %v", time.Since(start))
	start = time.Now()
	for i, tri := range triangles {
		rec.Record(
			geom.Triangle{
				[2]float64(tri[0]),
				[2]float64(tri[1]),
				[2]float64(tri[2]),
			},
			ffl,
			debugger.TestDescription{
				Category: fmt.Sprintf("triangle:%v", i),
				Description: fmt.Sprintf(
					"triangle:%v (%v)", i, tri,
				),
				Name: name,
			},
		)
	}
	t.Logf("writing out triangles took: %v", time.Since(start))
}

func cleanup(data []byte) (parts []string) {
	toreplace := []byte(`[]{}(),;`)
	for _, v := range toreplace {
		data = bytes.Replace(data, []byte{v}, []byte(" "), -1)
	}
	dparts := bytes.Split(data, []byte(` `))
	for _, dpt := range dparts {
		s := bytes.TrimSpace(dpt)
		if len(s) == 0 {
			continue
		}
		parts = append(parts, string(s))
	}
	return parts
}

func gettests(inputdir, mid string, ts map[string][][2]float64) {
	files, err := ioutil.ReadDir(inputdir)
	if err != nil {
		panic(
			fmt.Sprintf("Could not read dir %v: %v", inputdir, err),
		)
	}
	var filename string
	for _, file := range files {
		if mid != "" {
			filename = filepath.Join(mid, file.Name())
		} else {
			filename = file.Name()
		}
		if file.IsDir() {
			gettests(inputdir, filename, ts)
			continue
		}
		idx := strings.LastIndex(filename, ".points")
		if idx == -1 || filename[idx:] != ".points" {
			continue
		}
		data, err := ioutil.ReadFile(filepath.Join(inputdir, filename))
		if err != nil {
			panic(
				fmt.Sprintf("Could not read file %v: %v", filename, err),
			)
		}
		var pts [][2]float64

		// clean up file of { [ ( , ;
		parts := cleanup(data)
		if len(parts)%2 != 0 {
			panic(
				fmt.Sprintf("Badly formatted file %v:\n%v\n%s", filename, parts, data),
			)
		}
		for i := 0; i < len(parts); i += 2 {
			x, err := strconv.ParseFloat(parts[i], 64)
			if err != nil {
				panic(
					fmt.Sprintf("%v::%v: Badly formatted value {{%v}}:%v\n%s", filename, i, parts[i], err, data),
				)
			}
			y, err := strconv.ParseFloat(parts[i+1], 64)
			if err != nil {
				panic(
					fmt.Sprintf("%v::%v: Badly formatted value {{%v}}:%v\n%s", filename, i+1, parts[i], err, data),
				)
			}
			pts = append(pts, [2]float64{x, y})
		}
		ts[filename[:idx]] = pts
	}
}
func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	debugger.DefaultOutputDir = "output"
}

const inputdir = "testdata"

func TestTriangulation(t *testing.T) {

	/*
		tests := map[string][]geometry.Point{
			"First Test": {
				{516, 661}, {369, 793}, {426, 539}, {273, 525}, {204, 694}, {747, 750}, {454, 390},
			},
			"Second test": {
				{382, 302}, {382, 328}, {382, 205}, {623, 175}, {382, 188}, {382, 284}, {623, 87}, {623, 341}, {141, 227},
			},
		}
	*/
	tests := make(map[string][][2]float64)
	gettests(inputdir, "", tests)

	for name, pts := range tests {
		t.Run(name, func(t *testing.T) {

			var rec debugger.Recorder
			if cgo {
				// Only enable writing to log files if we have cgo enabled
				rec, _ = debugger.AugmentRecorder(rec, fmt.Sprintf("drawn_%v", name))
				t.Logf("writing entries to %v", rec.Filename)
			}
			Draw(t, rec, name, pts...)
			rec.CloseWait()
		})
	}

}
