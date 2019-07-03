// +build !cgo

package spatialite

func init() {
	println("this is a non-cgo build, using the debugger will fail")
	ignoreOpenErr = true
}
