# Quad edge

## build notes

* must use > go 1.12 for `bytes.ReplaceAll`
* use the `gdey_delaunay` branch of go-spatial/geom
* tests must be run on linux, the default installation of
  sqlite on mac does not allow extensions.
  * on ubuntu `apt-get install libsqlite3-mod-spatialite`


