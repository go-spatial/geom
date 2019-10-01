#!/bin/bash
set -e

PARSER_EXEC="./gen_exec"

go build -o $PARSER_EXEC gen.go

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
			$PARSER_EXEC >> $OUT_FILE
	done
done

rm $PARSER_EXEC

function varList() {

	cat $OUT_FILE natural_earth_picked.go | \
		sed -n 's/var \([_a-zA-Z0-9]*\) .*/\1/p' | \
		paste -s -d ',' -

}

echo
echo add/replace the following line to the end of natural_earth_lists.go:
echo
echo 'var NaturalEarth = []geom.Geometry{' `varList` '}'
echo

goimports -w $OUT_FILE

