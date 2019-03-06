package makevalid

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/encoding/wkt"

	_ "github.com/shaxbee/go-spatialite"
)

/*
The purpose of this file is to be a holding ground
for usefull function that would otherwise clutter up
the test case. Generally these should be functions
that are going to be wrapped in debug blocks and
dump to the terminal helpful debugging information.


The general format for these functions should be the
following:

func dumpDescriptiveName(...Paramaters...) string
*/

const (
	DebugContextTestName           = "debug_test_name"
	DebugContextTableTestName      = "debug_table_test_name"
	DebugContextTestOutputDatabase = "debug_test_output_database"
)

type debugDB struct {
	db *sql.DB
	sync.Mutex
	count uint
}

func fnFileLine() (string, string, int) {
	fnName := "unknown"
	file := "unknown"
	lineNo := -1

	pc, _, _, ok := runtime.Caller(2)
	details := runtime.FuncForPC(pc)
	if ok && details != nil {
		fnName = details.Name()
		file, lineNo = details.FileLine(pc)
	}

	vs := strings.Split(fnName, "/")
	fnName = vs[len(vs)-1]

	return fnName, file, lineNo
}

func debugContext(testFilename string, ctx context.Context) context.Context {
	if debug {
		if db := debugGetDatabaseFromContext(ctx); db != nil {
			db.Lock()
			db.count++
			db.Unlock()
			return ctx
		}

		if testFilename == "" {
			testFilename, _, _ = fnFileLine()
		}

		dbFilename := fmt.Sprintf("_test_output/%v.db", testFilename)
		log.Println("Writing to ", dbFilename)
		os.Remove(dbFilename)

		db, err := sql.Open("spatialite", dbFilename)
		if err != nil {
			log.Panic(err)
		}

		_, err = db.Exec("SELECT InitSpatialMetadata()")
		if err != nil {
			log.Println("dbfile:", dbFilename)
			log.Panic(err)
		}

		var sqls = make([]string, 0, 4*6)
		for _, gType := range []string{
			"POINT", "MULTIPOINT",
			"LINESTRING", "MULTILINESTRING",
			"POLYGON", "MULTIPOLYGON",
		} {
			lgType := strings.ToLower(gType)
			tblName := "test_" + lgType
			sqls = append(sqls,
				fmt.Sprintf("DROP TABLE IF EXISTS %v", tblName),
				fmt.Sprintf(
					`CREATE TABLE %v 
		        ( id INTEGER PRIMARY KEY AUTOINCREMENT 
		        , name CHAR(255)
		        , function_name CHAR(255)
			, filename CHAR(255)
		        , line INTEGER
		        , category CHAR(255)
		        , description CHAR(255)
	                );
		        `, tblName,
				),
				fmt.Sprintf(
					`SELECT AddGeometryColumn('%v', 'geometry', 4326, '%v', 2); `,
					tblName,
					gType,
				),
				fmt.Sprintf("SELECT CreateSpatialIndex('%v', 'geometry');", tblName),
			)

		}

		for _, sql := range sqls {
			_, err = db.Exec(sql)
			if err != nil {
				log.Println("Running sql ", sql)
				log.Panic(err)
			}
		}

		return context.WithValue(
			ctx,
			DebugContextTestOutputDatabase,
			&debugDB{
				db:    db,
				count: 1,
			},
		)
	}
	return ctx
}

func debugGetDatabaseFromContext(ctx context.Context) *debugDB {
	if debug {
		if v := ctx.Value(DebugContextTestOutputDatabase); v != nil {
			db, ok := v.(*debugDB)
			if ok && db != nil {
				return db
			}
		}
	}
	return nil
}

func debugClose(ctx context.Context) {
	if debug {
		db := debugGetDatabaseFromContext(ctx)
		if db != nil {
			db.Lock()
			db.count--
			if db.count <= 0 {
				log.Println("Closing db")
				db.db.Close()
			}
			db.Unlock()
		}
	}
}

func debugAddTestName(ctx context.Context, name string) (string, context.Context) {
	if debug {
		return name, context.WithValue(
			ctx,
			DebugContextTableTestName,
			name,
		)
	}
	return name, ctx
}

func debugTypeAndWKT(geom interface{}) (string, string) {
	wktGeom := wkt.MustEncode(geom)
	switch {
	case strings.HasPrefix(wktGeom, "POINT"):
		return "point", wktGeom
	case strings.HasPrefix(wktGeom, "MULTIPOINT"):
		return "multipoint", wktGeom
	case strings.HasPrefix(wktGeom, "LINESTRING"):
		return "linestring", wktGeom
	case strings.HasPrefix(wktGeom, "MULTILINESTRING"):
		return "multilinestring", wktGeom
	case strings.HasPrefix(wktGeom, "POLYGON"):
		return "polygon", wktGeom
	case strings.HasPrefix(wktGeom, "MULTIPOLYGON"):
		return "multipolygon", wktGeom
	default:
		panic(fmt.Sprintf("Unknown wkt type: %v", wktGeom))

	}
	return "", ""

}

// debugRecordEntity will records into the debug data base a line for the given gomes
// it will fill out the TestFunctionName and the TableTestName
func debugRecordEntity(ctx context.Context, description string, category string, geoms ...interface{}) {
	if debug {
		var (
			db            *sql.DB
			ok            bool
			tableTestName string
		)
		insertQueryFormat := `INSERT INTO test_%v 
		( name, function_name, filename, line, description, category, geometry)  VALUES 
		( '%v'  , '%v'  , '%v'  , %v  , '%v'  , '%v'  , GeomFromText('%v',4326))
		`
		if ddb := debugGetDatabaseFromContext(ctx); ddb != nil {
			db = ddb.db
		}

		fnName, filename, lineno := fnFileLine()
		{
			v := ctx.Value(DebugContextTableTestName)
			if v == nil {
				// don't do anything. because we don't have a good db
				return
			}

			tableTestName, ok = v.(string)
			if !ok || tableTestName == "" {
				tableTestName = "[[NONE]]"
			}
		}

		for _, geom := range geoms {
			gType, wktGeom := debugTypeAndWKT(geom)
			insertQuery := fmt.Sprintf(insertQueryFormat,
				gType,
				tableTestName,
				fnName,
				filename,
				lineno,
				description,
				category,
				wktGeom,
			)
			_, err := db.Exec(insertQuery)
			if err != nil {
				panic(err)
			}
		}
	}
}

// dumpWKTLineSegments will dump out both provided lineSegments next to each other.
func dumpWKTLineSegments(format string, lineSegment1, lineSegment2 []geom.Line) string {

	var b strings.Builder

	if format == "" {
		format = "%04d : %s | %s\n"
	}
	l1Larger := true
	maxi := len(lineSegment1)
	smax := len(lineSegment2)
	if len(lineSegment2) > maxi {
		maxi, smax = smax, maxi
		l1Larger = false
	}

	for i := 0; i < maxi; i++ {

		v1, v2 := "\t", "\t"
		if i < smax {
			v1, v2 = wkt.MustEncode(lineSegment1[i]), wkt.MustEncode(lineSegment2[i])
		} else if l1Larger {
			v1 = wkt.MustEncode(lineSegment1[i])
		} else {
			v2 = wkt.MustEncode(lineSegment2[i])
		}
		fmt.Fprintf(&b, format, i, v1, v2)
	}
	return b.String()
}

// dumpWKTLineSegment will dump the given lineSegments encodeing each lineSegment as WKT
func dumpWKTLineSegment(title, format string, lineSegments []geom.Line) string {
	var b strings.Builder

	if format == "" {
		format = "%04d : %s\n"
	}

	b.WriteString(title)
	b.WriteString("\n")

	for i := 0; i < len(lineSegments); i++ {
		fmt.Fprintf(&b, format, i, wkt.MustEncode(lineSegments[i]))
	}
	return b.String()
}
