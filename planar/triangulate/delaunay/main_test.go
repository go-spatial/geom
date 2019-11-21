package delaunay_test

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
	"github.com/go-spatial/geom/planar/triangulate/delaunay/quadedge"
	"github.com/go-spatial/geom/planar/triangulate/delaunay/subdivision"
	"github.com/go-spatial/geom/winding"
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

func Draw(t *testing.T, rec debugger.Recorder, name string, pts ...geom.Point) {

	var (
		badPoints  []geom.Point
		goodPoints []geom.Point
	)

	start := time.Now()
	sort.Sort(cmp.PointByXY(pts))

	tri := geom.NewTriangleContaining(pts...)
	ext := geom.NewExtentFromPoints(pts...)
	sd := subdivision.New(winding.Order{}, geom.Point(tri[0]), geom.Point(tri[1]), geom.Point(tri[2]))

	for i, pt := range pts {
		if i != 0 && pts[i-1][0] == pt[0] && pts[i-1][1] == pt[1] {
			continue
		}
		if !sd.InsertSite(pt) {
			badPoints = append(badPoints, pt)
			continue
		}
		goodPoints = append(goodPoints, pt)
	}
	t.Logf("triangulation setup took: %v", time.Since(start))

	start = time.Now()

	t.Logf("frame:triangle: %v", wkt.MustEncode(tri))
	t.Logf("frame:extent: %v", wkt.MustEncode(ext.AsPolygon()))

	if len(badPoints) > 0 {
		t.Logf("Failed to insert %v points\n", len(badPoints))
		for i, pt := range badPoints {
			t.Logf("initial:point:failed:%03v: %v", i, wkt.MustEncode(pt))
		}
	}
	for i, pt := range goodPoints {
		t.Logf("initial:point:good:%03v: %v", i, wkt.MustEncode(pt))
	}

	dumpSD(t, sd)
	t.Logf("writing out triangulation setup info: %v", time.Since(start))
	start = time.Now()
	triangles, err := sd.Triangles(true)
	if err != nil {
		t.Logf("Got an error: %v", err)
	}
	t.Logf("triangulation took: %v", time.Since(start))
	start = time.Now()
	for i, tri := range triangles {
		t.Logf("triangle:%03v: %v", i, wkt.MustEncode(tri))
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

func gettests(inputdir string, ts map[string][]geom.Point) {
	files, err := ioutil.ReadDir(inputdir)
	if err != nil {
		panic(
			fmt.Sprintf("Could not read dir %v: %v", inputdir, err),
		)
	}
	var filename string
	for _, file := range files {

		if file.IsDir() {
			gettests(filepath.Join(inputdir, file.Name()), ts)
			continue
		}
		filename = file.Name()

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
		var pts []geom.Point

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
			pts = append(pts, geom.Point{x, y})
		}
		ts[filename[:idx]] = pts
	}
}
func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	debugger.DefaultOutputDir = "output"
}

/*
const (
	inputdir                       = "testdata"
	triangulationTestCrateTableSQL = `

	CREATE TABLE IF NOT EXISTS test_triangulation_info (
		id INTEGER PRIMARY KEY AUTOINCREMENT
		, name TEXT
	);

	CREATE TABLE IF NOT EXISTS test_triangulation_input_point (
		test_id INTEGER
		, "order" INTEGER
		, geometry POINT
		, FOREIGN KEY(test_id) REFERENCES test_triangulation_info(id)
	);

	CREATE TABLE IF NOT EXISTS test_constrained_triangulation_linestring (
		test_id INTEGER
		, "order" INTEGER
		, geometry LINESTRING
		, FOREIGN KEY(test_id) REFERENCES test_triangulation_info(id)

	)

	CREATE TABLE IF NOt EXISTS "test_triangulation_expected_point" (
		test_id INTEGER
		, is_bad BOOLEAN DEFAULT false -- for good or bad points
		, "order" INTEGER
		, geometry POINT
		, FOREIGN KEY(test_id) REFERENCES test_triangulation_info(id)
	);

	CREATE TABLE IF NOT EXISTS "test_triangulation_expected_linestring" (
		test_id INTEGER
		, "order" INTEGER
		, geometry LINESTRING
		, FOREIGN KEY(test_id) REFERENCES test_triangulation_info(id)
	);

	CREATE TABLE IF NOT EXISTS "test_triangulation_expected_polygon" (
		test_id INTEGER
		, is_frame BOOLEAN DEFAULT false -- is part of the frame
		, type TEXT -- should be triangle, triangle:main, extent
		, "order" INTEGER
		, geometry LINESTRING
		, FOREIGN KEY(test_id) REFERENCES test_triangulation_info(id)
	);
	`
)

type gpkgTriangulationTestDB struct {
	Filename string
	init     bool
	h        *gpkg.Handle
	prepared map[string]*sql.Stmt
}

type gpkgTriangulationTestTableDescription struct {
	name    string
	geoType gpkg.GeometryType
}

func (tab gpkgTriangulationTestTableDescription) TableDescription() gpkg.TableDescription {
	return gpkg.TableDescription{
		Name:          tab.name,
		ShortName:     strings.Replace(tab.name, "_", " ", -1),
		Description:   fmt.Sprintf("triangualtion test table %v for %v", tab.name, tab.geoType),
		GeometryField: "geometry",
		GeometryType:  tab.geoType,
		Z:             gpkg.Prohibited,
		M:             gpkg.Prohibited,
	}
}

func (db *gpkgTriangulationTestDB) Load() error {
	tables := []gpkgTriangulationTestTableDescription{
		{
			name:    "test_triangulation_input_point",
			geoType: gpkg.Point,
		},
		{
			name:    "test_constrained_triangulation_linestring",
			geoType: gpkg.Linestring,
		},
		{
			name:    "test_triangulation_expected_point",
			geoType: gpkg.Point,
		},
		{
			name:    "test_triangulation_expected_linestring",
			geoType: gpkg.Linestring,
		},
		{
			name:    "test_triangulation_expected_polygon",
			geoType: gpkg.Polygon,
		},
	}
	if db.Filename == "" {
		return fmt.Errorf("need a filename")
	}
	h, err := gpkg.Open(db.Filename)
	if err != nil {
		return err
	}
	db.h = h
	if _, err = h.Exec(triangulationTestCrateTableSQL); err != nil {
		return err
	}
	for _, tbl := range tables {
		if err = h.AddGeometryTable(tbl.TableDescription()); err != nil {
			return err
		}
	}
	db.init = true
	return nil
}

func (db *gpkgTriangulationTestDB) Tests() []gpkgTriangulationTestDBRow {
	const (
		selectSQL = `
		SELECT
			id,
			name
		FROM
			test_triangulation_info
		;
		`
	)
	rows, err := db.h.Query(selectSQL)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var tests []gpkgTriangulationTestDBRow
	for rows.Next() {
		var test gpkgTriangulationTestDBRow
		if err = rows.Scan(&test.ID, &test.Name); err != nil {
			panic(err)
		}
		tests = append(tests, test)
	}
	return tests
}

type gpkgTriangulationTestDBRow struct {
	db   *gpkgTriangulationTestDB
	Name string
	ID   int
}

func (tcase gpkgTriangulationTestDBRow) Points() (pts []geom.Point) {
	const (
		selectSQL = `
	SELECT
		geometry
	FROM
		test_triangulation_input_point
	WHERE
		test_id = ?
	ORDER BY
		"order"
	;
	`
	)

	stmt := tcase.db.prepared["input_points"]
	if stmt == nil {
		stmt, err := tcase.db.h.Prepare(selectSQL)
		if err != nil {
			panic(err)
		}
		tcase.db.prepared["input_points"] = stmt
	}
	rows, err := stmt.Query(tcase.ID)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var sb gpkg.StandardBinary
		rows.Scan(&sb)
		pt, ok := sb.Geometry.(geom.Point)
		if !ok {
			panic(fmt.Sprintf("failed to decode point: %v %T", sb, sb.Geometry))
		}
		pts = append(pts, pt)
	}
	return pts
}
*/

func TestTriangulation(t *testing.T) {

	/*
			tests := map[string][]geom.Point{
				"First Test": {
					{516, 661}, {369, 793}, {426, 539}, {273, 525}, {204, 694}, {747, 750}, {454, 390},
				},
				"Second test": {
					{382, 302}, {382, 328}, {382, 205}, {623, 175}, {382, 188}, {382, 284}, {623, 87}, {623, 341}, {141, 227},
				},
			}

			type struct tcase {
				Name string
				Points []geom.Point `gpkgtest:"input"`
				Good []geom.Point `gpkgtest:"got"`
				Bad []geom.Point `gpkgtest:"got"`
				Edges []geom.Line `gpkgtest:"got"`
				Triangles []geom.Triangle `gpgktest:"got"`
			}

			test_triangulation_info {
				id integer
				name string -- Name
			}
			test_triangulation_points_point {
				test_id integer
				order integer
				geometry POINT
			}

			test_triangulation_good_point {
				test_id integer
				order integer
				geometry POINT
			}
			test_triangulation_good_point {
				test_id integer
				order integer
				geometry POINT
			}

			test_triangulation_expected_linestring {
				test_id
				order
				geometry LINESTRING
			}

			-- this will hold the triangles
			test_triangulation_expected_polygon {
				test_id
				is_frame
				type -- triangle, extent
				order
				geometry POLYGON
			}


		tests := make(map[string][]geom.Point)
		gettests(inputdir, tests)

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
	*/

}
