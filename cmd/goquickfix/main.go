/*
The goquickfix command quick fixes Go source that is well typed but go refuses
to compile e.g. "foo imported but not used".

Run with -help flag for usage information.
*/
package main

import (
	"flag"
	"fmt"
	"os"

	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"

	"github.com/motemen/go-quickfix"
)

var (
	flagWrite = flag.Bool("w", false, "write result to (source) file instead of stdout")
)

func usage() {
	fmt.Fprintln(os.Stderr, `Usage:
  goquickfix [-w] <path>

Flags:`)
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
	}

	// list of files grouped by package.
	files := map[string][]*ast.File{}

	fset := token.NewFileSet()

	for i := 0; i < flag.NArg(); i++ {
		arg := flag.Arg(i)
		fi, err := os.Stat(arg)
		dieIf(err)

		if fi.IsDir() {
			if i != 0 {
				die("you can only specify exact one directory")
			}

			pkgs, err := parser.ParseDir(fset, arg, nil, parser.ParseComments)
			dieIf(err)

			for _, pkg := range pkgs {
				ff := make([]*ast.File, 0, len(pkg.Files))
				for _, f := range pkg.Files {
					ff = append(ff, f)
				}
				files[pkg.Name] = ff
			}
		} else {
			f, err := parser.ParseFile(fset, arg, nil, parser.ParseComments)
			dieIf(err)

			const adhocPkg = ""

			// *.go files are grouped as ad-hoc package.
			if files[adhocPkg] == nil {
				files[adhocPkg] = []*ast.File{}
			}

			files[adhocPkg] = append(files[adhocPkg], f)
		}

	}

	for _, ff := range files {
		err := quickfix.QuickFix(fset, ff)
		dieIf(err)

		for _, f := range ff {
			// TODO: do not write if no change
			out := os.Stdout
			if *flagWrite {
				out, err = os.Create(fset.File(f.Pos()).Name())
				dieIf(err)
			}

			conf := printer.Config{
				Tabwidth: 8,
				Mode:     printer.UseSpaces | printer.TabIndent,
			}
			err := conf.Fprint(out, fset, f)
			dieIf(err)
		}
	}
}

func dieIf(err error) {
	if err != nil {
		die(err)
	}
}

func die(msg interface{}) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
