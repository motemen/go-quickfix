package quickfix

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"golang.org/x/tools/go/types"
	"testing"
)

func loadTestData(pkgName string) (*token.FileSet, []*ast.File, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, "testdata/"+pkgName, nil, parser.Mode(0))
	if err != nil {
		return nil, nil, err
	}

	pkg, ok := pkgs[pkgName]
	if !ok {
		return nil, nil, fmt.Errorf("package %s not found: %v", pkgName, pkgs)
	}

	files := make([]*ast.File, 0, len(pkg.Files))
	for _, f := range pkg.Files {
		files = append(files, f)
	}

	return fset, files, nil
}

func TestQuickFix_General(t *testing.T) {
	fset, files, err := loadTestData("eg1")
	if err != nil {
		t.Fatal(err)
	}

	err = QuickFix(fset, files)
	if err != nil {
		t.Fatalf("QuickFix(): %s", err)
	}

	logFiles(t, fset, files)

	_, err = types.Check("testdata/eg1", fset, files)
	if err != nil {
		t.Fatalf("should pass type checking: %s", err)
	}
}

func TestQuickFix_RangeStmt(t *testing.T) {
	fset, files, err := loadTestData("eg2")
	if err != nil {
		t.Fatal(err)
	}

	err = QuickFix(fset, files)
	if err != nil {
		t.Fatalf("QuickFix(): %s", err)
	}

	logFiles(t, fset, files)

	_, err = types.Check("testdata/eg2", fset, files)
	if err != nil {
		t.Fatalf("should pass type checking: %s", err)
	}
}

func logFiles(t *testing.T, fset *token.FileSet, files []*ast.File) {
	for _, f := range files {
		var buf bytes.Buffer
		printer.Fprint(&buf, fset, f)
		t.Log("#", fset.File(f.Pos()).Name())
		t.Log(buf.String())
	}
}
