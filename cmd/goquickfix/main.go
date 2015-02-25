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
	fmt.Fprintln(os.Stderr, `
Usage:
  goquickfix [-w] <path>

Flags:`)
	flag.PrintDefaults()
	os.Exit(2)
}

// TODO: allow multiple args
func main() {
	flag.Usage = usage
	flag.Parse()

	arg := flag.Arg(0)
	if arg == "" {
		flag.Usage()
	}

	files := map[string][]*ast.File{}

	fset := token.NewFileSet()

	fi, err := os.Stat(arg)
	dieIf(err)

	if fi.IsDir() {
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

		files[""] = []*ast.File{f}
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
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
