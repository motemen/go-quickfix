// Package quickfix provides functions to rewrite Go AST
// that is typed well but "go build" fails to pass building.
package quickfix

import (
	"fmt"
	"regexp"
	"strings"

	"go/ast"
	"go/token"

	"golang.org/x/tools/go/ast/astutil"
	_ "golang.org/x/tools/go/gcimporter"
	"golang.org/x/tools/go/types"
)

var (
	declaredNotUsed        = regexp.MustCompile(`^([a-zA-Z0-9_]+) declared but not used$`)
	importedNotUsed        = regexp.MustCompile(`^(".+") imported but not used$`)
	noNewVariablesOnDefine = "no new variables on left side of :="
)

// QuickFix rewrites AST files of same package so that they pass go build.
// For example:
//   v declared but not used             -> append `_ = v`
//   "p" imported but not used           -> rewrite to `import _ "p"`
//   no new variables on left side of := -> rewrite `:=` to `=`
//
// TODO hardMode, which removes errorneous code rather than adding
func QuickFix(fset *token.FileSet, files []*ast.File) (err error) {
	const maxTries = 10
	for i := 0; i < maxTries; i++ {
		var foundError bool
		foundError, err = quickFix1(fset, files)
		if !foundError {
			return nil
		}
	}

	return
}

func quickFix1(fset *token.FileSet, files []*ast.File) (bool, error) {
	errs := []error{}
	config := &types.Config{
		Error: func(err error) {
			errs = append(errs, err)
		},
	}

	_, err := config.Check("_quickfix", fset, files, nil)
	if err == nil {
		return false, nil
	}

	// apply fixes on AST later so that we won't break funcs that inspect AST by positions
	fixes := map[error]func() bool{}
	unhandled := errorList{}

	foundError := len(errs) > 0

	for _, err := range errs {
		err, ok := err.(types.Error)
		if !ok {
			unhandled = append(unhandled, err)
			continue
		}

		f := findFile(files, err.Pos)
		if f == nil {
			err := fmt.Errorf("cannot find file for error %q: %s (%d)", err.Error(), fset.Position(err.Pos), err.Pos)
			unhandled = append(unhandled, err)
			continue
		}

		nodepath, _ := astutil.PathEnclosingInterval(f, err.Pos, err.Pos)

		var fix func() bool

		// - "%s declared but not used"
		// - "%q imported but not used"
		// - "label %s declared but not used" TODO
		// - "no new variables on left side of :="
		if m := declaredNotUsed.FindStringSubmatch(err.Msg); m != nil {
			identName := m[1]
			fix = func() bool {
				return fixDeclaredNotUsed(nodepath, identName)
			}
		} else if m := importedNotUsed.FindStringSubmatch(err.Msg); m != nil {
			pkgPath := m[1] // quoted string, but it's okay because this will be compared to ast.BasicLit.Value.
			fix = func() bool {
				return fixImportedNotUsed(nodepath, pkgPath)
			}
		} else if err.Msg == noNewVariablesOnDefine {
			fix = func() bool {
				return fixNoNewVariables(nodepath)
			}
		} else {
			unhandled = append(unhandled, err)
		}

		if fix != nil {
			fixes[err] = fix
		}
	}

	for err, fix := range fixes {
		if fix() == false {
			unhandled = append(unhandled, err)
		}
	}

	return foundError, unhandled.any()
}

func fixDeclaredNotUsed(nodepath []ast.Node, identName string) bool {
	// insert "_ = x" to supress "declared but not used" error
	stmt := &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent("_")},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{ast.NewIdent(identName)},
	}
	return appendStmt(nodepath, stmt)
}

func fixImportedNotUsed(nodepath []ast.Node, pkgPath string) bool {
	for _, node := range nodepath {
		if f, ok := node.(*ast.File); ok {
			for _, imp := range f.Imports {
				if imp.Path.Value == pkgPath {
					// make this import spec anonymous one
					imp.Name = ast.NewIdent("_")
					return true
				}
			}
		}
	}
	return false
}

func fixNoNewVariables(nodepath []ast.Node) bool {
	for _, node := range nodepath {
		switch node := node.(type) {
		case *ast.AssignStmt:
			if node.Tok == token.DEFINE {
				node.Tok = token.ASSIGN
				return true
			}

		case *ast.RangeStmt:
			if node.Tok == token.DEFINE {
				node.Tok = token.ASSIGN
				return true
			}
		}
	}
	return false
}

type errorList []error

func (errs errorList) any() error {
	if len(errs) == 0 {
		return nil
	}

	return errs
}

func (errs errorList) Error() string {
	s := []string{fmt.Sprintf("%d error(s):", len(errs))}
	for _, e := range errs {
		s = append(s, fmt.Sprintf("- %s", e))
	}
	return strings.Join(s, "\n")
}

func appendStmt(nodepath []ast.Node, stmt ast.Stmt) bool {
	for _, node := range nodepath {
		switch node := node.(type) {
		case *ast.BlockStmt:
			if node.List == nil {
				node.List = []ast.Stmt{}
			}
			node.List = append(node.List, stmt)

		case *ast.CaseClause:
			if node.Body == nil {
				node.Body = []ast.Stmt{}
			}
			node.Body = append(node.Body, stmt)

		case *ast.CommClause:
			if node.Body == nil {
				node.Body = []ast.Stmt{}
			}
			node.Body = append(node.Body, stmt)

		case *ast.RangeStmt:
			if node.Body == nil {
				node.Body = &ast.BlockStmt{}
			}
			if node.Body.List == nil {
				node.Body.List = []ast.Stmt{}
			}
			node.Body.List = append(node.Body.List, stmt)

		default:
			continue
		}

		return true
	}

	return false
}

func findFile(files []*ast.File, pos token.Pos) *ast.File {
	for _, f := range files {
		if f.Pos() <= pos && pos < f.End() {
			return f
		}
	}

	return nil
}
