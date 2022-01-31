package main

import (
	"fmt"
	"go/doc"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"strings"
)

func main() {
	exitCode := 0
	errorf := func(format string, a ...interface{}) {
		fmt.Fprintf(os.Stderr, format, a...)
		exitCode = 1
	}
	for _, path := range os.Args[1:] {
		fset := token.NewFileSet()
		pkgs, err := parser.ParseDir(fset, path, func(f fs.FileInfo) bool {
			return !strings.HasSuffix(f.Name(), "_test.go")
		}, parser.ParseComments)
		if err != nil {
			errorf("%v\n", err)
		}
		for n, p := range pkgs {
			pkgDoc := doc.New(p, n, 0)
			if pkgDoc.Doc == "" {
				errorf("missing package doc for %s\n", n)
			}
			for _, c := range pkgDoc.Consts {
				if c.Doc == "" {
					errorf("%v: missing const doc for %s\n",
						fset.Position(c.Decl.Pos()), c.Names[0])
				}
			}
			for _, v := range pkgDoc.Vars {
				if v.Doc == "" {
					errorf("%v: missing var doc for %s\n",
						fset.Position(v.Decl.Pos()), v.Names[0])
				}
			}
			for _, t := range pkgDoc.Types {
				if t.Doc == "" {
					errorf("%v: missing type doc for %s\n",
						fset.Position(t.Decl.Pos()), t.Name)
				}
				for _, c := range t.Consts {
					if c.Doc == "" {
						errorf("%v: missing const doc for %s\n",
							fset.Position(c.Decl.Pos()), c.Names[0])
					}
				}
				for _, v := range t.Vars {
					if v.Doc == "" {
						errorf("%v: missing var doc for %s\n",
							fset.Position(v.Decl.Pos()), v.Names[0])
					}
				}
				for _, m := range t.Methods {
					if m.Doc == "" {
						errorf("%v: missing method doc for %s\n",
							fset.Position(m.Decl.Pos()), m.Name)
					}
				}
				for _, f := range t.Funcs {
					if f.Doc == "" {
						errorf("%v: missing func doc for %s\n",
							fset.Position(f.Decl.Pos()), f.Name)
					}
				}
			}
			for _, f := range pkgDoc.Funcs {
				if f.Doc == "" {
					errorf("%v: missing func doc for %s\n",
						fset.Position(f.Decl.Pos()), f.Name)
				}
			}
		}
	}
	os.Exit(exitCode)
}
