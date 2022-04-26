//go:build cgo
// +build cgo

package gpkg

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
	"github.com/mattn/go-sqlite3"
)

const (
	// SPATIALITE database driver name
	SPATIALITE = "spatialite"

	// ApplicationID is the required application id for the file
	ApplicationID = 0x47504B47 // "GPKG"

	// UserVersion is the version of the GPKG file format. We support
	// 1.2.1, so the the decimal representation is 10201 (1 digit for the major
	// two digit for the minor and bug-fix).
	UserVersion = 0x000027D9 // 10201

	// TableSpatialRefSysSQL is the normative sql for the required spatial ref
	// table. http://www.geopackage.org/spec/#gpkg_spatial_ref_sys_sql
	TableSpatialRefSysSQL = `
	CREATE TABLE IF NOT EXISTS gpkg_spatial_ref_sys (
		srs_name TEXT NOT NULL,
		srs_id INTEGER NOT NULL PRIMARY KEY,
		organization TEXT NOT NULL,
		organization_coordsys_id INTEGER NOT NULL,
		definition  TEXT NOT NULL,
		description TEXT
	);
	`

	// TableContentsSQL is the normative sql for the required contents table.
	// http://www.geopackage.org/spec/#gpkg_contents_sql
	TableContentsSQL = `
	CREATE TABLE IF NOT EXISTS gpkg_contents (
		table_name TEXT NOT NULL PRIMARY KEY,
		data_type TEXT NOT NULL,
		identifier TEXT UNIQUE,
		description TEXT DEFAULT '',
		last_change DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
		min_x DOUBLE,
		min_y DOUBLE,
		max_x DOUBLE,
		max_y DOUBLE,
		srs_id INTEGER,
		CONSTRAINT fk_gc_r_srs_id FOREIGN KEY (srs_id) REFERENCES gpkg_spatial_ref_sys(srs_id)
	  );
	`

	// TableGeometryColumnsSQL is the normative sql for the geometry columns table that is
	// required if the contents table has at least one table with a data_type of features
	// http://www.geopackage.org/spec/#gpkg_geometry_columns_sql
	TableGeometryColumnsSQL = `
	CREATE TABLE IF NOT EXISTS gpkg_geometry_columns (
		table_name TEXT NOT NULL,
		column_name TEXT NOT NULL,
		geometry_type_name TEXT NOT NULL,
		srs_id INTEGER NOT NULL,
		z TINYINT NOT NULL,  -- 0: z values prohibited; 1: z values mandatory; 2: z values optional
		m TINYINT NOT NULL,  -- 0: m values prohibited; 1: m values mandatory; 2: m values optional
		CONSTRAINT pk_geom_cols PRIMARY KEY (table_name, column_name),
		CONSTRAINT uk_gc_table_name UNIQUE (table_name),
		CONSTRAINT fk_gc_tn FOREIGN KEY (table_name) REFERENCES gpkg_contents(table_name),
		CONSTRAINT fk_gc_srs FOREIGN KEY (srs_id) REFERENCES gpkg_spatial_ref_sys (srs_id)
	  );
	`

	// https://www.geopackage.org/spec/#gpkg_extensions_sql
	TableExtensionsSQL = `
	CREATE TABLE IF NOT EXISTS gpkg_extensions (
		table_name TEXT,
		column_name TEXT,
		extension_name TEXT NOT NULL,
		definition TEXT NOT NULL,
		scope TEXT NOT NULL,
		CONSTRAINT ge_tce UNIQUE (table_name, column_name, extension_name)
	  );
	`
)

// Organization names
const (
	// ORNone is for basic SRS
	ORNone = "none"
	OREPSG = "epsg"
)

var (
	initialSQL = fmt.Sprintf(
		`
		PRAGMA application_id = %d;
		PRAGMA user_version = %d ;
		PRAGMA foreign_keys = ON ;
		`,
		ApplicationID,
		UserVersion,
	)
)

const (
	DataTypeFeatures   = "features"
	DataTypeAttributes = "attributes"
	DataTypeTitles     = "titles"
)

// SpatialReferenceSystem describes the SRS
type SpatialReferenceSystem struct {
	Name                   string
	ID                     int
	Organization           string
	OrganizationCoordsysID int
	Definition             string
	Description            string
}

var KnownSRS = map[int32]SpatialReferenceSystem{
	-1: {
		Name:                   "any",
		ID:                     -1,
		Organization:           ORNone,
		OrganizationCoordsysID: -1,
		Definition:             "",
		Description:            "any",
	},
	0: {
		Name:                   "any",
		ID:                     0,
		Organization:           ORNone,
		OrganizationCoordsysID: 0,
		Definition:             "",
		Description:            "any",
	},
	4326: {
		Name:                   "WGS 84",
		ID:                     4326,
		Organization:           OREPSG,
		OrganizationCoordsysID: 4326,
		Definition: `
		GEOGCS["WGS 84",
		DATUM["WGS_1984",
			SPHEROID["WGS 84",6378137,298.257223563,
				AUTHORITY["EPSG","7030"]],
			AUTHORITY["EPSG","6326"]],
		PRIMEM["Greenwich",0,
			AUTHORITY["EPSG","8901"]],
		UNIT["degree",0.0174532925199433,
			AUTHORITY["EPSG","9122"]],
		AUTHORITY["EPSG","4326"]]
		`,
		Description: "World Geodetic System: WGS 84",
	},
	3857: {
		Name:                   "WebMercator",
		ID:                     3857,
		Organization:           OREPSG,
		OrganizationCoordsysID: 3857,
		Definition: `
		PROJCS["WGS 84 / Pseudo-Mercator",
    GEOGCS["WGS 84",
        DATUM["WGS_1984",
            SPHEROID["WGS 84",6378137,298.257223563,
                AUTHORITY["EPSG","7030"]],
            AUTHORITY["EPSG","6326"]],
        PRIMEM["Greenwich",0,
            AUTHORITY["EPSG","8901"]],
        UNIT["degree",0.0174532925199433,
            AUTHORITY["EPSG","9122"]],
        AUTHORITY["EPSG","4326"]],
    PROJECTION["Mercator_1SP"],
    PARAMETER["central_meridian",0],
    PARAMETER["scale_factor",1],
    PARAMETER["false_easting",0],
    PARAMETER["false_northing",0],
    UNIT["metre",1,
        AUTHORITY["EPSG","9001"]],
    AXIS["X",EAST],
    AXIS["Y",NORTH],
    EXTENSION["PROJ4","+proj=merc +a=6378137 +b=6378137 +lat_ts=0.0 +lon_0=0.0 +x_0=0.0 +y_0=0 +k=1.0 +units=m +nadgrids=@null +wktext  +no_defs"],
    AUTHORITY["EPSG","3857"]]
		`,
		Description: "WGS83 / Web Mercator",
	},
}

// nonZeroFileExists checks if a file exists, and has a size greater then Zero
// and is not a directory before we try using it to prevent further errors.
func nonZeroFileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	if info.IsDir() {
		return false
	}
	return info.Size() > 0
}

// Return spatialite drivename while loading the required libs
func drivername() string {
	// quick hotwired of https://github.com/shaxbee/go-spatialite/blob/master/spatialite.go
	type entrypoint struct {
		lib  string
		proc string
	}

	var libs = []entrypoint{
		{"mod_spatialite", "sqlite3_modspatialite_init"},
		{"mod_spatialite.dylib", "sqlite3_modspatialite_init"},
		{"libspatialite.so", "sqlite3_modspatialite_init"},
		{"libspatialite.so.5", "spatialite_init_ex"},
		{"libspatialite.so", "spatialite_init_ex"},
	}

	for _, s := range sql.Drivers() {
		if s == SPATIALITE {
			return SPATIALITE
		}
	}

	sql.Register(SPATIALITE, &sqlite3.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			for _, v := range libs {
				if err := conn.LoadExtension(v.lib, v.proc); err == nil {
					return nil
				}
			}
			return errors.New("spatialite extension not found")
		},
	})

	return SPATIALITE

}

// Open will open or create the sqlite file, and return a new db handle to it.
func Open(filename string) (*Handle, error) {
	var h = new(Handle)

	db, err := sql.Open(drivername(), filename)
	if err != nil {
		return nil, err
	}
	h.DB = db
	if err = initHandle(h); err != nil {
		return nil, err
	}
	return h, nil
}

// New will create a new gpkg file and return a new db handle
func New(filename string) (*Handle, error) {
	// First let's check to see if the file exists
	// if it does error. We will not overwrite an files
	if nonZeroFileExists(filename) {
		return nil, os.ErrExist
	}
	return Open(filename)
}

// initHandle will setup up all the required tables and metadata for
// a new gpkg file.
func initHandle(h *Handle) error {
	// Set the pragma's that we need to set for this file
	_, err := h.Exec(initialSQL)
	if err != nil {
		return err
	}
	// Make sure the required metadata tables are available
	for _, sql := range []string{TableSpatialRefSysSQL, TableContentsSQL, TableGeometryColumnsSQL, TableExtensionsSQL} {
		_, err := h.Exec(sql)
		if err != nil {
			return err
		}
	}

	srss := make([]SpatialReferenceSystem, 0, len(KnownSRS))
	// Now need to add SRS that we know about
	for _, srs := range KnownSRS {
		srss = append(srss, srs)
	}
	return h.UpdateSRS(srss...)
}

// AddGeometryTable will add the given features table to the metadata tables
// This should be called after creating the table.
func (h *Handle) AddGeometryTable(table TableDescription) error {

	const (
		validateSRSSQL = `
		SELECT Count(*) 
		FROM gpkg_spatial_ref_sys 
		WHERE 
			srs_id=?
		`
		validateTableFieldSQL = `
		SELECT "%v"
		FROM "%v"
		LIMIT 1
		`
		updateContentsTableSQL = `
		INSERT INTO gpkg_contents(
			table_name,
			data_type,
			identifier,
			description,
			srs_id
		)
		VALUES (?,?,?,?,?)
    	ON CONFLICT(table_name) DO NOTHING;
		`
		updateGeometryColumnsTableSQL = `
		INSERT INTO gpkg_geometry_columns(
			table_name,
			column_name,
			geometry_type_name,
			srs_id,
			z,
			m
		)
		VALUES(?,?,?,?,?,?)
    	ON CONFLICT(table_name) DO NOTHING;
		`

		updateExtensionTableSQL = `
		INSERT INTO gpkg_extensions(
			table_name,
			column_name,
			extension_name,
			definition,
			scope
		)
		VALUES(?,?,?,?,?)
    	ON CONFLICT(table_name, column_name, extension_name) DO NOTHING;		
		`

		// DDL and DML for the RTree -> http://www.geopackage.org/spec120/#extension_rtree
		// SQL statements based on the requeriments as of spec 1.2.0

		createRTreeTableSQL = `
		CREATE VIRTUAL TABLE "rtree_%v_%v" USING rtree(id, minx, maxx, miny, maxy);
		`

		selectPKfromTable = `SELECT name FROM pragma_table_info(("%v")) WHERE pk = 1;`

		tableTriggerTemplate = `
        /* Conditions: Insertion of non-empty geometry
           Actions   : Insert record into rtree */
        CREATE TRIGGER rtree_{{ .T }}_{{ .C }}_insert AFTER INSERT ON "{{ .T }}"
          WHEN (new."{{ .C }}" NOT NULL AND NOT ST_IsEmpty(NEW."{{ .C }}"))
        BEGIN
          INSERT OR REPLACE INTO rtree_{{ .T }}_{{ .C }} VALUES (
            NEW."{{ .I }}",
            ST_MinX(NEW."{{ .C }}"), ST_MaxX(NEW."{{ .C }}"),
            ST_MinY(NEW."{{ .C }}"), ST_MaxY(NEW."{{ .C }}")
          );
        END;
        
        /* Conditions: Update of geometry column to non-empty geometry
                       No row ID change
           Actions   : Update record in rtree */
        CREATE TRIGGER rtree_{{ .T }}_{{ .C }}_update1 AFTER UPDATE OF "{{ .C }}" ON "{{ .T }}"
          WHEN OLD."{{ .I }}" = NEW."{{ .I }}" AND
               (NEW."{{ .C }}" NOTNULL AND NOT ST_IsEmpty(NEW."{{ .C }}"))
        BEGIN
          INSERT OR REPLACE INTO rtree_{{ .T }}_{{ .C }} VALUES (
            NEW."{{ .I }}",
            ST_MinX(NEW."{{ .C }}"), ST_MaxX(NEW."{{ .C }}"),
            ST_MinY(NEW."{{ .C }}"), ST_MaxY(NEW."{{ .C }}")
          );
        END;
        
        /* Conditions: Update of geometry column to empty geometry
                       No row ID change
           Actions   : Remove record from rtree */
        CREATE TRIGGER rtree_{{ .T }}_{{ .C }}_update2 AFTER UPDATE OF "{{ .C }}" ON "{{ .T }}"
          WHEN OLD."{{ .I }}" = NEW."{{ .I }}" AND
               (NEW."{{ .C }}" ISNULL OR ST_IsEmpty(NEW."{{ .C }}"))
        BEGIN
          DELETE FROM rtree_{{ .T }}_{{ .C }} WHERE id = OLD."{{ .I }}";
        END;
        
        /* Conditions: Update of any column
                       Row ID change
                       Non-empty geometry
           Actions   : Remove record from rtree for old "{{ .I }}"
                       Insert record into rtree for new "{{ .I }}" */
        CREATE TRIGGER rtree_{{ .T }}_{{ .C }}_update3 AFTER UPDATE OF "{{ .C }}" ON "{{ .T }}"
          WHEN OLD."{{ .I }}" != NEW."{{ .I }}" AND
               (NEW."{{ .C }}" NOTNULL AND NOT ST_IsEmpty(NEW."{{ .C }}"))
        BEGIN
          DELETE FROM rtree_{{ .T }}_{{ .C }} WHERE id = OLD."{{ .I }}";
          INSERT OR REPLACE INTO rtree_{{ .T }}_{{ .C }} VALUES (
            NEW."{{ .I }}",
            ST_MinX(NEW."{{ .C }}"), ST_MaxX(NEW."{{ .C }}"),
            ST_MinY(NEW."{{ .C }}"), ST_MaxY(NEW."{{ .C }}")
          );
        END;
        
        /* Conditions: Update of any column
                       Row ID change
                       Empty geometry
           Actions   : Remove record from rtree for old and new "{{ .I }}" */
        CREATE TRIGGER rtree_{{ .T }}_{{ .C }}_update4 AFTER UPDATE ON "{{ .T }}"
          WHEN OLD."{{ .I }}" != NEW."{{ .I }}" AND
               (NEW."{{ .C }}" ISNULL OR ST_IsEmpty(NEW."{{ .C }}"))
        BEGIN
          DELETE FROM rtree_{{ .T }}_{{ .C }} WHERE id IN (OLD."{{ .I }}", NEW."{{ .I }}");
        END;
        
        /* Conditions: Row deleted
           Actions   : Remove record from rtree for old "{{ .I }}" */
        CREATE TRIGGER rtree_{{ .T }}_{{ .C }}_delete AFTER DELETE ON "{{ .T }}"
          WHEN old."{{ .C }}" NOT NULL
        BEGIN
          DELETE FROM rtree_{{ .T }}_{{ .C }} WHERE id = OLD."{{ .I }}";
        END;
		`
	)

	var (
		count int
		pk    string
	)

	// Validate that the value already exists in the data base.
	err := h.QueryRow(validateSRSSQL, table.SRS).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		// let's check known srs's to see if we have it and can add it.
		srsdef, ok := KnownSRS[table.SRS]
		if !ok {
			return fmt.Errorf("unknown srs: %v", table.SRS)
		}
		if err = h.UpdateSRS(srsdef); err != nil {
			return err
		}
	}
	rows, err := h.Query(fmt.Sprintf(validateTableFieldSQL, table.GeometryField, table.Name))
	if err != nil {
		return fmt.Errorf("unknown table %v or field %v : %v", table.Name, table.GeometryField, err)
	}
	rows.Close()
	_, err = h.Exec(updateContentsTableSQL, table.Name, DataTypeFeatures, table.ShortName, table.Description, table.SRS)
	if err != nil {
		return err
	}
	_, err = h.Exec(updateGeometryColumnsTableSQL, table.Name, table.GeometryField, table.GeometryType.String(), table.SRS, table.Z, table.M)
	if err != nil {
		return err
	}

	err = h.QueryRow(fmt.Sprintf(selectPKfromTable, table.Name)).Scan(&pk)
	if err != nil {
		return err
	}

	// Requirement 77
	type tableTriggerParameters struct {
		T string // <t>: The name of the feature table containing the geometry column
		C string // <c>: The name of the geometry column in <t> that is being indexed
		I string // <i>: The name of the integer primary key column in <t> as specified in Requirement 29
	}

	param := tableTriggerParameters{T: table.Name, C: table.GeometryField, I: pk}

	_, err = h.Exec(fmt.Sprintf(createRTreeTableSQL,
		param.T, param.C,
	))
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	template.Must(template.New("createtabletriggers").Parse(tableTriggerTemplate)).Execute(buf, param)

	_, err = h.Exec(buf.String())
	if err != nil {
		return err
	}

	_, err = h.Exec(updateExtensionTableSQL, table.Name, table.GeometryField, `gpkg_rtree_index`, `http://www.geopackage.org/spec120/#extension_rtree`, `write-only`)
	return err

}

// UpdateSRS will insert or update the srs table with the given srs
func (h *Handle) UpdateSRS(srss ...SpatialReferenceSystem) error {

	const (
		UpdateSQL = `
	INSERT INTO gpkg_spatial_ref_sys(
		srs_name,
		srs_id,
		organization,
		organization_coordsys_id,
		definition,
		description
	)
	VALUES %v
    ON CONFLICT(srs_id) DO NOTHING;
	`
		placeHolders = `(?,?,?,?,?,?) `
	)
	if len(srss) == 0 {
		return nil
	}

	valuePlaceHolder := strings.Join(
		strings.SplitN(
			strings.Repeat(placeHolders, len(srss)),
			" ",
			len(srss),
		),
		",",
	)
	updateSQL := fmt.Sprintf(UpdateSQL, valuePlaceHolder)
	values := make([]interface{}, 0, len(srss)*6)

	for _, srs := range srss {
		values = append(
			values,
			srs.Name,
			srs.ID,
			srs.Organization,
			srs.OrganizationCoordsysID,
			srs.Definition,
			srs.Description,
		)
	}
	_, err := h.Exec(updateSQL, values...)
	return err
}

// UpdateGeometryExtent will modify the extent for the given table by adding the passed
// in extent to the extent of the table. Growing the extent as necessary.
func (h *Handle) UpdateGeometryExtent(tablename string, extent *geom.Extent) error {
	if extent == nil {
		return nil
	}

	var (
		minx,
		miny,
		maxx,
		maxy *float64

		ext *geom.Extent
	)
	const (
		selectSQL = `
		SELECT
			min_x,
			min_y,
			max_x,
			max_y
		FROM 
			gpkg_contents
		WHERE
			table_name = ?
		`
		updateSQL = `
		UPDATE gpkg_contents
		SET
			min_x = ?,
			min_y = ?,
			max_x = ?,
			max_y = ?
		WHERE 
			table_name = ?
		`
	)
	err := h.QueryRow(selectSQL, tablename).Scan(&minx, &miny, &maxx, &maxy)
	if err != nil {
		return err
	}
	if minx == nil || miny == nil || maxx == nil || maxy == nil {
		ext = extent
	} else {
		ext = geom.NewExtent([2]float64{*minx, *miny}, [2]float64{*maxx, *maxy})
		ext.Add(extent)
	}
	_, err = h.Exec(updateSQL, ext.MinX(), ext.MinY(), ext.MaxX(), ext.MaxY(), tablename)

	return err
}

// CalculateGeometryExtent will grab all the geometries from the given table, use it
// to calculate the extent of all geometries in that table.
func (h *Handle) CalculateGeometryExtent(tablename string) (*geom.Extent, error) {
	const (
		selectGeomColSQL = `
		SELECT 
			column_name
		FROM 
			gpkg_geometry_columns
		WHERE
			table_name = ?
		`
		selectAllSQLFormat = ` SELECT "%v" FROM "%v"`
	)

	var (
		columnName string
		ext        *geom.Extent
		err        error
		rows       *sql.Rows
		sb         StandardBinary
	)
	// First get the geometry column for table.
	if err = h.QueryRow(selectGeomColSQL, tablename).Scan(&columnName); err != nil {
		return nil, err
	}
	if rows, err = h.Query(fmt.Sprintf(selectAllSQLFormat, columnName, tablename)); err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&sb)
		if cmp.IsEmptyGeo(sb.Geometry) {
			continue
		}
		if ext == nil {
			ext, err = geom.NewExtentFromGeometry(sb.Geometry)
			if err != nil {
				ext = nil
			}
			continue
		}
		ext.AddGeometry(sb.Geometry)
	}
	return ext, nil
}
