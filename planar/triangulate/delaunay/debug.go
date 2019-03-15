package delaunay

import (
	"log"

	"github.com/go-spatial/geom/internal/debugger"
)

const debug = false

const (
	DebuggerCategoryConstrained = debugger.CategoryJoiner("constrained:")
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	debugger.DefaultOutputDir = "_test_output"
}
