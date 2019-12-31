package quickfix

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"go/types"
	"path/filepath"
	"strings"
	"testing"
)

func TestQuickFix_General(t *testing.T) {
	checkCorrectness(t, "general")
}

func TestQuickFix_RangeStmt(t *testing.T) {
	checkCorrectness(t, "rangestmt")
}

func TestRevertQuickFix_BlankAssign(t *testing.T) {
	fset, files, _, err := loadTestData("revert")
	if err != nil {
		t.Fatal(err)
	}

	err = RevertQuickFix(fset, files)
	if err != nil {
		t.Fatalf("RevertQuickFix(): %s", err)
	}

	if strings.Contains(fileContent(fset, files[0]), `_ = `) {
		t.Fatal("assignments to blank identifiers should be removed")
	}

	if !strings.Contains(fileContent(fset, files[0]), `import "fmt"`) {
		t.Fatal("quickfixes to blank imports should be reverted")
	}

	if !strings.Contains(fileContent(fset, files[0]), `import _ "image/png"`) {
		t.Fatal("imports of packages with side effects should not be considered as quickfixed")
	}

	logFiles(t, fset, files)
}

func TestImportName(t *testing.T) {
	checkCorrectness(t, "importname")
}

func loadTestData(pkgName string) (*token.FileSet, []*ast.File, string, error) {
	fset := token.NewFileSet()
	dir := filepath.Join("testdata", pkgName)
	pkgs, err := parser.ParseDir(fset, dir, nil, parser.Mode(0))
	if err != nil {
		return nil, nil, "", err
	}

	pkg, ok := pkgs[pkgName]
	if !ok {
		return nil, nil, "", fmt.Errorf("package %s not found: %v", pkgName, pkgs)
	}

	files := make([]*ast.File, 0, len(pkg.Files))
	for _, f := range pkg.Files {
		files = append(files, f)
	}

	return fset, files, dir, nil
}

func checkCorrectness(t *testing.T, testName string) {
	fset, files, dir, err := loadTestData(testName)
	if err != nil {
		t.Fatal(err)
	}

	qfconfig := Config{
		Fset:     fset,
		Files:    files,
		Dir:      dir,
		MaxTries: 10,
	}
	err = qfconfig.QuickFix()
	if err != nil {
		t.Fatalf("QuickFix(): %s", err)
	}

	logFiles(t, fset, files)

	config := &types.Config{Importer: pkgsImporter{dir: dir}}
	_, err = config.Check(dir, fset, files, nil)
	if err != nil {
		t.Fatalf("should pass type checking: %s", err)
	}
}

func logFiles(t *testing.T, fset *token.FileSet, files []*ast.File) {
	for _, f := range files {
		t.Log("#", fset.File(f.Pos()).Name())
		t.Log(fileContent(fset, f))
	}
}

func fileContent(fset *token.FileSet, f *ast.File) string {
	var buf bytes.Buffer
	printer.Fprint(&buf, fset, f)
	return buf.String()
}
