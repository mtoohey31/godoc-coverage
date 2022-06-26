package main

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"strings"

	"github.com/barweiss/go-tuple"
	"mtoohey.com/iter"
)

func main() {
	// start by iterating through the arguments
	if iter.FlatMap(iter.Elems(os.Args[1:]), func(path string) iter.Iter[string] {
		// parse each package, ignoring test files, since we don't need to enforce
		// godocs in them
		fset := token.NewFileSet()
		pkgs, err := parser.ParseDir(fset, path, func(f fs.FileInfo) bool {
			return !strings.HasSuffix(f.Name(), "_test.go")
		}, parser.ParseComments)

		// if the package fails to parse, return a single-element iterator containing
		// the error
		if err != nil {
			iter.Elems([]string{err.Error()})
		}

		// flat map each package found in the path to an iterator of errors
		return iter.FlatMap(iter.KVZip(pkgs),
			func(t tuple.T2[string, *ast.Package]) iter.Iter[string] {
				// create the package doc for each package
				pkgDoc := doc.New(t.V2, t.V1, 0)

				// create an iterator of value docs
				valueDocs := iter.FlatMap(iter.Elems(pkgDoc.Types),
					func(t *doc.Type) iter.Iter[*doc.Value] {
						return iter.Elems(t.Consts).Chain(iter.Elems(t.Vars))
					}).Chain(iter.Elems(pkgDoc.Consts).Chain(iter.Elems(pkgDoc.Vars)))

				// create an iterator of function docs
				funcDocs := iter.FlatMap(iter.Elems(pkgDoc.Types),
					func(t *doc.Type) iter.Iter[*doc.Func] {
						return iter.Elems(t.Methods).Chain(iter.Elems(t.Funcs))
					}).Chain(iter.Elems(pkgDoc.Funcs))

				// the result is the chain of the elements of the above iterators that
				// had no doc, mapped to an informative error
				return iter.Map(valueDocs.Filter(func(vd *doc.Value) bool {
					return vd.Doc == ""
				}), func(vd *doc.Value) string {
					return fmt.Sprintf("%v: missing doc for %s\n",
						fset.Position(vd.Decl.Pos()), vd.Names[0])
				}).Chain(iter.Map(funcDocs.Filter(func(fd *doc.Func) bool {
					return fd.Doc == ""
				}), func(fd *doc.Func) string {
					return fmt.Sprintf("%v: missing doc for %s\n",
						fset.Position(fd.Decl.Pos()), fd.Name)
				}))
			})
	}).Inspect(func(s string) {
		// print all errors to stderr
		os.Stderr.WriteString(s)
		// errors are ignored, cause what else are we going to do about them, spit
		// them into the same stderr that the write just failed on?
	}).Count() > 0 {
		// if there were any errors, exit with code 1
		os.Exit(1)
	}
}
