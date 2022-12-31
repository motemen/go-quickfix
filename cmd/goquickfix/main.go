/*
The goquickfix command quick fixes Go source that is well typed but go refuses
to compile e.g. "foo imported but not used".

Run with -help flag for usage information.
*/
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/motemen/go-quickfix"
)

var (
	flagWrite  = flag.Bool("w", false, "write result to (source) file instead of stdout")
	flagRevert = flag.Bool("revert", false, "try to revert possible quickfixes introduced by goquickfix")
)

func usage() {
	fmt.Fprintln(os.Stderr, `Usage:
  goquickfix [-w] [-revert] <path>...

Flags:`)
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
	}

	fileContents := map[string]string{}

	fset := token.NewFileSet()

	for i := 0; i < flag.NArg(); i++ {
		arg := flag.Arg(i)
		fi, err := os.Stat(arg)
		if err != nil {
			return err
		}

		if fi.IsDir() {
			if i != 0 {
				return errors.New("you can only specify exact one directory")
			}

			fis, err := os.ReadDir(arg)
			if err != nil {
				return err
			}

			for _, fi := range fis {
				if fi.IsDir() {
					continue
				}

				name := fi.Name()
				if !strings.HasSuffix(name, ".go") {
					continue
				}
				if name[0] == '_' || name[0] == '.' {
					continue
				}

				filename := filepath.Join(arg, name)
				b, err := os.ReadFile(filename)
				if err != nil {
					return err
				}

				fileContents[filename] = string(b)
			}
		} else {
			b, err := os.ReadFile(arg)
			if err != nil {
				return err
			}

			fileContents[arg] = string(b)
		}

	}

	ff, err := parseFiles(fset, fileContents)
	if err != nil {
		return err
	}

	if *flagRevert {
		err = quickfix.RevertQuickFix(fset, ff)
	} else {
		err = quickfix.QuickFix(fset, ff)
	}
	if err != nil {
		return err
	}

	for _, f := range ff {
		filename := fset.File(f.Pos()).Name()

		var buf bytes.Buffer
		conf := printer.Config{
			Tabwidth: 8,
			Mode:     printer.UseSpaces | printer.TabIndent,
		}
		if err := conf.Fprint(&buf, fset, f); err != nil {
			return err
		}

		if buf.String() == fileContents[filename] {
			// no change, skip this file
			continue
		}

		out := os.Stdout
		if *flagWrite {
			if out, err = os.Create(filename); err != nil {
				return err
			}
		}

		buf.WriteTo(out)
	}

	return nil
}

func parseFiles(fset *token.FileSet, fileContents map[string]string) ([]*ast.File, error) {
	files := make([]*ast.File, 0, len(fileContents))

	for filename, content := range fileContents {
		f, err := parser.ParseFile(fset, filename, content, parser.ParseComments)
		if err != nil {
			return nil, err
		}

		files = append(files, f)
	}

	return files, nil
}
