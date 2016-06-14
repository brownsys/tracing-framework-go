package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/importer"
	"go/parser"
	"go/printer"
	"go/token"
	"go/types"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	fileSuffix = "_context.go"
	buildTag   = "context"
)

func main() {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not determine pwd:", err)
		os.Exit(2)
	}
	processDir(pwd)
}

func processDir(dir string) {
	fset := token.NewFileSet()
	filter := func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go") &&
			!strings.HasSuffix(fi.Name(), fileSuffix)
	}
	pkgs, err := parser.ParseDir(fset, dir, filter, parser.ParseComments|parser.DeclarationErrors)
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not parse source files:", err)
		os.Exit(2)
	}

	var pkgNames []string
	for name := range pkgs {
		pkgNames = append(pkgNames, name)
	}
	sort.Strings(pkgNames)

	info := types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Implicits:  make(map[ast.Node]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
		Scopes:     make(map[ast.Node]*types.Scope),
	}

	c := types.Config{
		Error:    func(err error) { fmt.Fprintln(os.Stderr, err) },
		Importer: importer.Default(),
	}

	for _, name := range pkgNames {
		pkg := pkgs[name]

		var files []*ast.File
		fpaths := make(map[*ast.File]string)
		for fname, f := range pkg.Files {
			files = append(files, f)
			fpaths[f] = fname
		}

		p, err := c.Check(pkg.Name, fset, files, &info)
		if err != nil {
			// it was already printed by c.Error
			os.Exit(2)
		}

		for _, f := range files {
			qual := qualifierForFile(p, f)

			err := rewriteGos(fset, info, qual, f)
			if err == nil {
				err = rewriteCalls(fset, info, qual, f)
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "could not rewrite %v: %v\n", f, err)
				os.Exit(2)
			}

			origHasBuildTag := false

			for _, c := range f.Comments {
				for _, c := range c.List {
					if c.Text == "// +build !"+buildTag {
						c.Text = "// +build " + buildTag
						origHasBuildTag = true
					}
				}
			}

			var buf bytes.Buffer
			fpath := fpaths[f]
			if origHasBuildTag {
				printer.Fprint(&buf, fset, f)
			} else {
				buf.Write([]byte("// +build " + buildTag + "\n\n"))
				printer.Fprint(&buf, fset, f)

				// prepend build comment to original file
				b, err := ioutil.ReadFile(fpath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "could not read source file: %v\n", err)
					os.Exit(2)
				}
				b = append([]byte("// +build !"+buildTag+"\n\n"), b...)
				b, err = format.Source(b)
				if err != nil {
					fmt.Fprintf(os.Stderr, "could not format source file %v: %v\n", filepath.Base(fpath), err)
					os.Exit(2)
				}
				f, err := os.OpenFile(fpath, os.O_WRONLY, 0)
				if err != nil {
					fmt.Fprintf(os.Stderr, "could not open source file for writing: %v\n", err)
					os.Exit(2)
				}
				if _, err = f.Write(b); err != nil {
					fmt.Fprintf(os.Stderr, "could not write to source file: %v\n", err)
					os.Exit(2)
				}
			}

			b, err := format.Source(buf.Bytes())
			if err != nil {
				panic(fmt.Errorf("unexpected internal error: %v", err))
			}
			fpath = fpath[:len(fpath)-3] + fileSuffix
			if err = ioutil.WriteFile(fpath, b, 0664); err != nil {
				fmt.Fprintf(os.Stderr, "could not create instrument source file: %v\n", err)
				os.Exit(2)
			}
		}
	}
}
