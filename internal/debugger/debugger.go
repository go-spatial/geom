package debugger

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-spatial/geom/internal/debugger/recorder/gpkg"
)

const (
	// ContextRecorderKey the key to store the recorder in the context
	ContextRecorderKey = "debugger_recorder_key"

	// ContextRecorderKey the key to store the testname in the context
	ContextRecorderTestnameKey = "debugger_recorder_testname_key"
)

// DefaultOutputDir is where the system will write the debugging db/files
// By default this will use os.TempDir() to write to the system temp directory
// set this in an init function to wite elsewhere.
var DefaultOutputDir = os.TempDir()

// AsString will create string contains the stringified items seperated by a ':'
func AsString(vs ...interface{}) string {
	var s strings.Builder
	var addc bool
	for _, v := range vs {
		if addc {
			s.WriteString(":")
		}
		fmt.Fprintf(&s, "%v", v)
		addc = true
	}
	return s.String()
}

// GetRecorderFromContext will return the recoder that is
// in the context. If there isn't a recorder, then an invalid
// recorder will be returned. This can be checked with the
// IsValid() function on the recorder.
func GetRecorderFromContext(ctx context.Context) Recorder {

	name, _ := ctx.Value(ContextRecorderTestnameKey).(string)
	r, _ := ctx.Value(ContextRecorderKey).(Recorder)

	r.Desc.Name = name

	return r
}

// AugmentContext is will add and configure the recorder used to track the
// debugging entries into the context.
// A Close call should be supplied along with the AugmentContext  call, this
// is usually done using a defer
// If the testFilename is "", then the function name of the calling function
// will be used as the filename for the database file.
func AugmentContext(ctx context.Context, testFilename string) context.Context {
	if rec := GetRecorderFromContext(ctx); rec.IsValid() {
		rec.IncrementCount()
		return ctx
	}

	if testFilename == "" {
		testFilename = funcFileLine().Func
	}

	err := os.MkdirAll(DefaultOutputDir, os.ModePerm)
	if err != nil {
		panic(fmt.Sprintf("Failed to created dir %v:%v", DefaultOutputDir, err))
	}

	rcd, filename, err := gpkg.New(DefaultOutputDir, testFilename, 0)
	if err != nil {
		panic(fmt.Sprintf("Failed to created gpkg db: %v", err))
	}
	log.Println("Write debugger output to", filename)
	return context.WithValue(
		ctx,
		ContextRecorderKey,
		Recorder{recorder: &recorder{Interface: rcd}},
	)
}

// Close allows the recorder to release any resources it as, each
// AugmentContext call should have a mirroring Close call that is
// called at the end of the function.
func Close(ctx context.Context) { GetRecorderFromContext(ctx).Close() }
func SetTestName(ctx context.Context, name string) context.Context {
	return context.WithValue(
		ctx,
		ContextRecorderTestnameKey,
		name,
	)
}

// Record records the geom and descriptive attributes into the debugging system
func Record(ctx context.Context, geom interface{}, category string, descriptionFormat string, data ...interface{}) {
	rec := GetRecorderFromContext(ctx)
	if !rec.IsValid() {
		return
	}

	description := fmt.Sprintf(descriptionFormat, data...)

	err := rec.Record(
		geom,
		funcFileLine(),
		TestDescription{
			Category:    category,
			Description: description,
		},
	)
	if err != nil {
		panic(fmt.Sprintf("failed to record entry: %v", err))
	}
}
