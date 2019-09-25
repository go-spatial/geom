# testing

This package contains functions and variables for using pre made geometries.

The tile variables are very large and are initially stored as strings. In order
to access them, you must first compile compile them by calling `CompileTiles`.
Once compiled, access the tiles via the `Tiles` function (this does no processing
besides endusring that compilation has already occured).

```go
package my_test

import (
    "github.com/go-spatial/geom/encoding/wkt"
    gtesting "github.com/go-spatial/testing"
)

func init {
    // put this in init so benchmarks aren't skewed
    gtesting.CompileTiles(wkt.DecodeString)
}

func TestMy(t *testing) {
    tiles := gtesting.Tiles()
    ...
}
```
