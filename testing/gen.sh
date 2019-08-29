#!/bin/bash
set -e

PARSER_FILE='script.go'
PARSER="go run $PARSER_FILE"


TABLE_LIST=(
	"ne_10m_admin_0_countries"
	"ne_10m_parks_and_protected_lands_line"
	"ne_10m_parks_and_protected_lands_area"
	"ne_110m_coastline"
	"ne_50m_lakes"
	"ne_50m_antarctic_ice_shelves_lines"
	"ne_50m_antarctic_ice_shelves_polys"
)

GEOMS_PER_TABLE=10
OUT_FILE='natural_earth_gen.go'
DB_CMD='psql -d natural_earth'

# from https://gist.github.com/ear7h/80d233451e90a31ec63aeea435f296b2
cat << EOF > $PARSER_FILE
package main

import (
	"fmt"
	"os"
	"strings"
	"io/ioutil"
)

func geomType(typ string) string {
	if strings.HasPrefix(typ, "MULTI") {
		return "Multi" + geomType(typ[len("MULTI"):])
	}

	switch typ {
		case "POINT":
			return "Point"

		case "LINESTRING":
			return "LineString"

		case "POLYGON":
			return "Polygon"

		case "GEOMETRYCOLLECTION":
			return "Collection"
	}
	panic("not found " + typ)
}

func main() {
	byt, _ := ioutil.ReadAll(os.Stdin)
	str := string(byt)
	n := strings.Index(str, "|")
	name := strings.Replace(str[:n], " ", "", -1)

	str = str[n+1:]
	n = strings.Index(str, "(")
	typ := strings.TrimSpace(str[:n])
	typ = geomType(typ)

	str = str[n:]

	str = strings.Replace(str, ",", "},{", -1)
	str = strings.Replace(str, " ", ", ", -1)
	str = strings.Replace(str, "(", "{", -1)
	str = strings.Replace(str, ")", "}", -1)
	str = strings.TrimSpace(str)
	fmt.Printf("var %s = geom.%s{%s}\n", name, typ, str)
}
EOF


echo 'package testing' > $OUT_FILE
echo '// this file was auto generated using gen.go' >> $OUT_FILE

for table in "${TABLE_LIST[@]}"; do

	if [[ $GEOMS_PER_TABLE -le 0 ]]; then
		echo "not writing any geoms"
		continue
	else
		echo "writing geoms from table $table"
	fi

	for i in $(seq 0 $(($GEOMS_PER_TABLE - 1))); do
		$DB_CMD -t \
			-c "select '_$table$i', ST_AsText(ST_Transform(wkb_geometry, 3857)) from  $table limit 1 offset $i;" | \
			$PARSER >> $OUT_FILE
	done
done

rm $PARSER_FILE

function varList() {

	cat $OUT_FILE natural_earth_picked.go | \
		sed -n 's/var \([_a-zA-Z0-9]*\) .*/\1/p' | \
		paste -s -d ',' -

}

echo 'var NaturalEarth = []geom.Geometry{' `varList` '}' >> $OUT_FILE

goimports -w $OUT_FILE

