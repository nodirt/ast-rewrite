package rewrite

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"testing"

	. "go/ast"
	"go/format"
	"go/parser"
	"go/token"
)

const TEST_DATA_DIR = "./testdata"

func isBlank(expr Expr) bool {
	if ident, ok := expr.(*Ident); ok {
		return ident.Name == "_"
	}
	return false
}

type testCase struct {
	name     string
	rewriter Rewriter
}

var testCases = []testCase{
	{
		name: "simple_range",
		rewriter: func(node Node) (Node, bool) {
			if rng, ok := node.(*RangeStmt); ok {
				if (rng.Value == nil || isBlank(rng.Value)) && isBlank(rng.Key) {
					rng.Value = nil
					rng.Key = nil
				}
			}
			return node, true
		},
	},
}

func TestRewriters(t *testing.T) {
	var buf bytes.Buffer
	for _, testCase := range testCases {
		fset := token.NewFileSet()
		inFile := filepath.Join(TEST_DATA_DIR, testCase.name+".in.go")
		file, err := parser.ParseFile(fset, inFile, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		file = Rewrite(file, testCase.rewriter).(*File)

		buf.Reset()
		if err := format.Node(&buf, fset, file); err != nil {
			t.Fatal(err)
		}
		actual := buf.Bytes()

		outFile := filepath.Join(TEST_DATA_DIR, testCase.name+".out.go")
		expected, err := ioutil.ReadFile(outFile)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(actual, expected) {
			t.Logf("Test %s failed.\nExpected: %s.\nActual: %s", testCase.name, expected, actual)
			var astText bytes.Buffer
			Fprint(&astText, fset, file, NotNilFilter)
			t.Log(astText.String())
			t.FailNow()
		}
	}
}
