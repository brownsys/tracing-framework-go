package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"go/token"
	"go/types"
	"io/ioutil"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/loader"
)

const (
	fileSuffix = "_local.go"
	buildTag   = "local"
)

func main() {
	path := "."
	switch {
	case len(os.Args) == 2:
		path = os.Args[1]
	case len(os.Args) > 2:
		fmt.Fprintf(os.Stderr, "Usage: %v [<path>]\n", os.Args[0])
		os.Exit(1)
	}

	packageDir(path)
}

func packageDir(dir string) {
	var conf loader.Config
	conf.Import(dir)

	prog, err := conf.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not parse source files:", err)
		os.Exit(2)
	}

	pi := prog.InitialPackages()[0]

	for _, f := range pi.Files {
		qual := qualifierForFile(pi.Pkg, f)

		fncs := []func(*token.FileSet, types.Info, types.Qualifier, *ast.File) (bool, error){
			rewriteGos,
		}

		var changed bool
		var err error
		for _, fnc := range fncs {
			var c bool
			c, err = fnc(prog.Fset, pi.Info, qual, f)
			if c {
				changed = true
			}
			if err != nil {
				break
			}
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "could not rewrite %v: %v\n", f, err)
			os.Exit(2)
		}

		if !changed {
			continue
		}

		fmt.Println(prog.Fset.Position(f.Pos()).Filename)

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
		fpath := prog.Fset.Position(f.Pos()).Filename
		if origHasBuildTag {
			printer.Fprint(&buf, prog.Fset, f)
		} else {
			buf.Write([]byte("// +build " + buildTag + "\n\n"))
			printer.Fprint(&buf, prog.Fset, f)

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
