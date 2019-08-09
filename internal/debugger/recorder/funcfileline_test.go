package recorder_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-spatial/geom/internal/debugger/recorder"
)

func levelTwoCall(lvl uint) recorder.FuncFileLineType {
	return levelOneCall(lvl)
}

func levelOneCall(lvl uint) recorder.FuncFileLineType {
	return recorder.FuncFileLine(lvl)
}

func TestFuncFileLine(t *testing.T) {
	/*
		It's is important not to move this bit of
		code as it will change the result of the
		the test.
	*/
	testFFL(t, levelOneCall(0), "TestFuncFileLine", "funcfileline_test.go", 25)
	testFFL(t, levelTwoCall(1), "TestFuncFileLine", "funcfileline_test.go", 26)

	testFFL(t, levelTwoCall(0), "levelTwoCall", "funcfileline_test.go", 12)

	testFFL(t, levelOneCall(1), "tRunner", "testing.go", -1)
}

func testFFL(t *testing.T, ffl recorder.FuncFileLineType, funcName, filename string, ln int) {
	funcNameStrs := strings.Split(ffl.Func, ".")
	gotFuncName := funcNameStrs[len(funcNameStrs)-1]
	if gotFuncName != funcName {
		t.Errorf("Func, expected %v got %v", funcName, gotFuncName)
	}
	if ln != -1 {
		if ffl.LineNumber != ln {
			t.Errorf("LineNumber, expected %v got %v", ln, ffl.LineNumber)
		}
	}
	file := filepath.Base(ffl.File)
	if file != filename {
		t.Errorf("LineNumber, expected %v got %v", filename, file)
	}
}
