A small utility to break down a given wkt and z/x/y into various parts for the makevalid algo.

This tool is very much a work in progress. 

#Quick Start:

```
$ cmd z/x/y input.wkt
```

Options:

* simplify [true]  -- simplify the geom 
* tag -- create an additional directory for the output files
* buffer [64] -- buffer to expand the tile boundry by
* extent [4096] -- the extent of mvt tile. 
