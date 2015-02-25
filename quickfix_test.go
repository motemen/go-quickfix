package quickfix

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"golang.org/x/tools/go/types"
	"testing"
)

func TestQuickFix(t *testing.T) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, "testdata/eg1", nil, parser.Mode(0))
	if err != nil {
		t.Fatalf("ParsseDir(): %s", err)
	}

	pkg, ok := pkgs["eg1"]
	if !ok {
		t.Fatalf("package eg1 not found: %v", pkgs)
	}

	files := make([]*ast.File, 0, len(pkg.Files))
	for _, f := range pkg.Files {
		files = append(files, f)
	}

	err = QuickFix(fset, files)
	if err != nil {
		t.Fatalf("QuickFix(): %s", err)
	}

	_, err = types.Check("testdata/ex1", fset, files)
	if err != nil {
		t.Fatalf("should pass type checking: %s", err)
	}

	var buf bytes.Buffer
	printer.Fprint(&buf, fset, files[0])
	t.Log(buf.String())
}
