// +build cgo

package main

import (
	"fmt"
	"log"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/encoding/gpkg"
)

const (
	TablePOISQL = `
	CREATE TABLE IF NOT EXISTS poi (
		id INTEGER NOT NULL PRIMARY KEY,
		name TEXT,
		geometry %v
	);
	`
	InsertSQL = `
	INSERT INTO poi(
		name,
		geometry
	)
	VALUES(?,?)
	`
)

type POIDesc struct {
	Name     string
	Location geom.Point
}

var sampleData = [...]POIDesc{
	{
		Name:     "San Diego",
		Location: geom.Point{13042545.682585, 3858464.750807},
	},
	{
		Name:     "Chula Vista",
		Location: geom.Point{-13033621.659533, 3848298.626045},
	},
}

func main() {
	h, err := gpkg.Open("cities.gpkg")
	if err != nil {
		log.Println("err:", err)
		return
	}
	defer h.Close()
	_, err = h.Exec(fmt.Sprintf(TablePOISQL, gpkg.Point.String()))
	if err != nil {
		log.Println("err:", err)
		return
	}
	err = h.AddGeometryTable(gpkg.TableDescription{
		Name:          "poi",
		ShortName:     "points of interest",
		Description:   "interesting points on the map",
		GeometryField: "geometry",
		GeometryType:  gpkg.Point,
		SRS:           3857,
		Z:             gpkg.Prohibited,
		M:             gpkg.Prohibited,
	})
	if err != nil {
		log.Println("err:", err)
	}
	var ext *geom.Extent
	stmt, err := h.Prepare(InsertSQL)
	if err != nil {
		log.Println("err:", err)
		return
	}
	for _, data := range sampleData {
		sb, err := gpkg.NewBinary(3857, data.Location)
		if err != nil {
			log.Println("err:", err)
			continue
		}
		_, err = stmt.Exec(data.Name, sb)
		if err != nil {
			log.Println("err:", err)
			continue
		}

		if ext == nil {
			ext, err = geom.NewExtentFromGeometry(data.Location)
			if err != nil {
				ext = nil
				log.Println("err:", err)
				continue
			}
		} else {
			ext.AddGeometry(data.Location)
		}
	}
	h.UpdateGeometryExtent("poi", ext)

}
