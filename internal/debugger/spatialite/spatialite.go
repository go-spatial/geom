package spatialite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-spatial/geom/internal/debugger/recorder"
	_ "github.com/go-spatial/geom/internal/debugger/spatialite/go-spatialite"
)

type DB struct {
	*sql.DB
}

func New(outputDir, filename string) (*DB, string, error) {

	dbFilename := filepath.Join(outputDir, filename+".sqlite3")

	os.Remove(dbFilename)

	db, err := sql.Open("spatialite", dbFilename)
	if err != nil {
		return nil, dbFilename, fmt.Errorf("dbfile: %v err: %v", dbFilename, err)
	}

	if _, err = db.Exec("SELECT InitSpatialMetadata()"); err != nil {
		return nil, dbFilename, fmt.Errorf("dbfile: %v err: %v", dbFilename, err)
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
		if _, err = db.Exec(sql); err != nil {
			return nil, dbFilename, err
		}
	}
	return &DB{DB: db}, dbFilename, nil
}

const insertQueryFormat = `
INSERT INTO test_%v
  ( function_name, filename, line, name, description, category, geometry             )
VALUES
  ( ?            , ?       , ?   , ?   , ?          , ?       , GeomFromText(?,4326) )
`

func (db *DB) Record(geom interface{}, ffl recorder.FuncFileLineType, tblTest recorder.TestDescription) error {
	if db == nil {
		return nil
	}

	type_, wktStr := recorder.TypeAndWKT(geom)
	insertQuery := fmt.Sprintf(insertQueryFormat, type_)
	_, err := db.Exec(insertQuery,

		ffl.Func,
		ffl.File,
		ffl.LineNumber,

		tblTest.Name,
		tblTest.Description,
		tblTest.Category,

		wktStr,
	)
	return err
}
