// +build !instrument

package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"
	"text/template"
)

func main() {
	processDir(".", func(f *ast.FuncDecl) bool { return true })
}

func processDir(path string, filter func(*ast.FuncDecl) bool) {
	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, path, nil, parser.ParseComments|parser.DeclarationErrors)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not parse package: %v\n", err)
		os.Exit(2)
	}

	if len(pkgs) > 2 {
		fmt.Fprintln(os.Stderr, "found multiple packages")
		os.Exit(2)
	}

	if len(pkgs) == 0 {
		os.Exit(0)
	}

	type tmplEntry struct {
		Fname string
		Flag  string
		Args  []string
	}

	var entries []tmplEntry
	var pkgname string
	var pkg *ast.Package

	for name, p := range pkgs {
		pkgname = name
		pkg = p
	}

	for fname, file := range pkg.Files {
		_ = fname
		for _, fnctmp := range file.Decls {
			fnc, ok := fnctmp.(*ast.FuncDecl)
			if !ok || !filter(fnc) {
				continue
			}
			entry := tmplEntry{
				Fname: fnc.Name.String(),
				Flag:  "__instrument_" + fnc.Name.String(),
			}
			for _, arg := range fnc.Type.Params.List {
				for _, name := range arg.Names {
					entry.Args = append(entry.Args, name.Name)
				}
			}
			entries = append(entries, entry)
			var buf bytes.Buffer
			err = shimTmpl.Execute(&buf, entry)
			if err != nil {
				fmt.Fprintf(os.Stderr, "unexpected internal error: %v\n", err)
				os.Exit(3)
			}
			stmt := parseStmt(string(buf.Bytes()))
			if len(stmt.List) != 1 {
				panic("internal error")
			}
			fnc.Body.List = append([]ast.Stmt{stmt.List[0]}, fnc.Body.List...)
		}

		origHasBuildTag := false

		for _, c := range file.Comments {
			for _, c := range c.List {
				// fmt.Println(c.Text)
				if c.Text == "// +build !instrument" {
					c.Text = "// +build instrument"
					origHasBuildTag = true
				}
			}
		}

		var buf bytes.Buffer
		if origHasBuildTag {
			printer.Fprint(&buf, fs, file)
		} else {
			buf.Write([]byte("// +build instrument\n\n"))
			printer.Fprint(&buf, fs, file)

			// prepend build comment to original file
			b, err := ioutil.ReadFile(fname)
			if err != nil {
				fmt.Fprintf(os.Stderr, "could not read source file: %v\n", err)
				os.Exit(2)
			}
			b = append([]byte("// +build !instrument\n\n"), b...)
			b, err = format.Source(b)
			if err != nil {
				fmt.Fprintf(os.Stderr, "could not format source file %v: %v\n", fname, err)
				os.Exit(2)
			}
			f, err := os.OpenFile(fname, os.O_WRONLY, 0)
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
			fmt.Fprintf(os.Stderr, "unexpected internal error: %v\n", err)
			os.Exit(3)
		}
		os.Stdout.Write(b)
	}

	// fmt.Println("=======")

	var buf bytes.Buffer
	err = initTmpl.Execute(&buf, entries)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unexpected internal error: %v\n", err)
		os.Exit(3)
	}

	b, err := format.Source([]byte("package " + pkgname + string(buf.Bytes())))
	if err != nil {
		fmt.Fprintf(os.Stderr, "unexpected internal error: %v\n", err)
		os.Exit(3)
	}
	_ = b
	// os.Stdout.Write(b)
}

func parseStmt(src string) *ast.BlockStmt {
	src = `package main
	func a() {` + src + `}`
	fset := token.NewFileSet()
	a, err := parser.ParseFile(fset, "", src, parser.ParseComments|parser.DeclarationErrors)
	if err != nil {
		panic(fmt.Errorf("internal error: %v", err))
	}
	return a.Decls[0].(*ast.FuncDecl).Body
}

var initTmpl *template.Template = template.Must(template.New("").Parse(`
import "local/research/instrument"

var (
	{{range .}}{{.Flag}} bool
{{end}})

func init() {
	{{range .}}instrument.RegisterFlag({{.Fname}}, &{{.Flag}})
{{end}}}
`))

var shimTmpl = template.Must(template.New("").Parse(`
if {{.Flag}} {
	callback, ok := instrument.GetCallback({{.Fname}})
	if ok {
		callback({{range .Args}}{{.}},{{end}})
	}
}
`))
