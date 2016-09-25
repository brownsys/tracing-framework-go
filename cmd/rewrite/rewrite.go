package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"reflect"
	"text/template"

	"golang.org/x/tools/go/ast/astutil"
)

func rewriteGos(fset *token.FileSet, info types.Info, qual types.Qualifier, f *ast.File) (changed bool, err error) {
	rname := runtimeName(f)
	err = mapStmts(f, func(s ast.Stmt) ([]ast.Stmt, error) {
		if g, ok := s.(*ast.GoStmt); ok {
			stmts, err := rewriteGoStmt(fset, info, qual, rname, g)
			if stmts != nil {
				changed = true
			}
			return stmts, err
		}
		return nil, nil
	})
	if changed {
		// astutil.AddNamedImport(fset, f, rname, "runtime")
		astutil.AddImport(fset, f, "github.com/brownsys/tracing-framework-go/local")
		// astutil.AddImport(fset, f, "runtime")
	}
	return changed, err
}

// runtimeName searches through f's imports to find whether
// the "runtime" package has been imported, and if not, whether
// another package whose name is also "runtime" has been
// imported (which would conflict if we were to add "runtime"
// as an import). It returns the name that should be used to
// identify the "runtime" package.
func runtimeName(f *ast.File) string {
	for _, imp := range f.Imports {
		if imp.Path.Value == `"runtime"` {
			if imp.Name != nil {
				return imp.Name.Name
			}
			return "runtime"
		}
	}
	return "__runtime"
}

// nameForPackage searches through f's imports to find
// whether the package identified by the given path has
// been imported, and if not, whether another package
// whose name is the same has been imported (which would
// conflict if we were to add the given path as an
// import). It returns the name that should be used to
// identify the given package.
//
// TODO: does this actually implement the spec?
// func nameForPackage(f *ast.File, path, name string) string {
// 	path = '"' + path + '"'
// 	for _, imp := range f.Imports {
// 		if imp.Path.Value == path {
// 			if imp.Name != nil {
// 				return imp.Name.Name
// 			}
// 			return name
// 		}
// 	}
// 	return "__" + name
// }

// mapStmts walks v, searching for values of type []ast.Stmt
// or [x]ast.Stmt. After recurring into such values, it loops
// over the slice or array, and for each element, calls f.
// If f returns nil, the value is left as is. If f returns a
// non-nil slice (of any length, including 0), the contents
// of this slice replace the original value in the slice or array.
//
// If f ever returns a non-nil error, it is immediately returned.
func mapStmts(v ast.Node, f func(s ast.Stmt) ([]ast.Stmt, error)) error {
	var blocks []*ast.BlockStmt
	ast.Inspect(v, func(n ast.Node) bool {
		if b, ok := n.(*ast.BlockStmt); ok {
			blocks = append(blocks, b)
		}
		return true
	})

	// make sure to process blocks backwards
	// so that children are processed before parents
	for i := len(blocks) - 1; i >= 0; i-- {
		b := blocks[i]
		var newStmts []ast.Stmt
		for _, s := range b.List {
			new, err := f(s)
			if err != nil {
				return err
			}
			if new == nil {
				newStmts = append(newStmts, s)
			} else {
				newStmts = append(newStmts, new...)
			}
		}
		b.List = newStmts
	}

	return nil
}

func rewriteGoStmt(fset *token.FileSet, info types.Info, qual types.Qualifier, rname string, g *ast.GoStmt) ([]ast.Stmt, error) {
	ftyp := info.TypeOf(g.Call.Fun)

	if ftyp == nil {
		return nil, fmt.Errorf("%v: could not determine type of function",
			fset.Position(g.Call.Fun.Pos()))
	}
	sig := ftyp.(*types.Signature)

	var arg struct {
		Runtime                       string
		Func                          string
		Typ                           string
		DefArgs, InnerArgs, OuterArgs []string
	}

	arg.Runtime = rname
	arg.Func = nodeString(fset, g.Call.Fun)
	arg.Typ = types.TypeString(ftyp, qual)

	params := sig.Params()
	for i := 0; i < params.Len(); i++ {
		name := fmt.Sprintf("arg%v", i)
		if sig.Variadic() && i == params.Len()-1 {
			typ := types.TypeString(params.At(i).Type().(*types.Slice).Elem(), qual)
			arg.DefArgs = append(arg.DefArgs, name+" ..."+typ)
			arg.InnerArgs = append(arg.InnerArgs, name+"...")
		} else {
			typ := types.TypeString(params.At(i).Type(), qual)
			arg.DefArgs = append(arg.DefArgs, name+" "+typ)
			arg.InnerArgs = append(arg.InnerArgs, name)
		}
	}

	for i, a := range g.Call.Args {
		if g.Call.Ellipsis.IsValid() && i == len(g.Call.Args)-1 {
			// g.Call.Ellipsis.IsValid() is true if g is variadic
			arg.OuterArgs = append(arg.OuterArgs, nodeString(fset, a)+"...")
		} else {
			arg.OuterArgs = append(arg.OuterArgs, nodeString(fset, a))
		}
	}

	var buf bytes.Buffer
	err := goTmpl.Execute(&buf, arg)
	if err != nil {
		panic(fmt.Errorf("internal error: %v", err))
	}
	return parseStmts(string(buf.Bytes())), nil
}

var goTmpl = template.Must(template.New("").Parse(`
go func(__f1 func(), __f2 {{.Typ}} {{range .DefArgs}},{{.}}{{end}}){
	__f1()
	__f2({{range .InnerArgs}}{{.}},{{end}})
}(local.GetSpawnCallback(), {{.Func}}{{range .OuterArgs}},{{.}}{{end}})
`))

func parseStmts(src string) []ast.Stmt {
	src = `package main
	func a() {` + src + `}`
	fset := token.NewFileSet()
	a, err := parser.ParseFile(fset, "", src, parser.ParseComments|parser.DeclarationErrors)
	if err != nil {
		panic(fmt.Errorf("internal error: %v", err))
	}
	stmts := a.Decls[0].(*ast.FuncDecl).Body.List
	zeroPos(&stmts)
	return stmts
}

// walk v and zero all values of type token.Pos
func zeroPos(v interface{}) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		panic("internal error")
	}
	zeroPosHelper(rv)
}

var posTyp = reflect.TypeOf(token.Pos(0))

func zeroPosHelper(rv reflect.Value) {
	if rv.Type() == posTyp {
		rv.SetInt(0)
		return
	}
	switch rv.Kind() {
	case reflect.Ptr:
		if !rv.IsNil() {
			zeroPosHelper(rv.Elem())
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < rv.Len(); i++ {
			zeroPosHelper(rv.Index(i))
		}
	case reflect.Map:
		keys := rv.MapKeys()
		for _, k := range keys {
			zeroPosHelper(rv.MapIndex(k))
		}
	case reflect.Struct:
		for i := 0; i < rv.NumField(); i++ {
			zeroPosHelper(rv.Field(i))
		}
	}
}

func nodeString(fset *token.FileSet, node interface{}) string {
	var buf bytes.Buffer
	err := format.Node(&buf, fset, node)
	if err != nil {
		panic(fmt.Errorf("unexpected internal error: %v", err))
	}
	return string(buf.Bytes())
}

func qualifierForFile(pkg *types.Package, f *ast.File) types.Qualifier {
	pathToPackage := make(map[string]*types.Package)
	for _, pkg := range pkg.Imports() {
		pathToPackage[pkg.Path()] = pkg
	}

	m := make(map[*types.Package]string)
	for _, imp := range f.Imports {
		if imp.Path.Value == `"unsafe"` {
			continue
		}
		// slice out quotation marks
		l := len(imp.Path.Value)
		pkg, ok := pathToPackage[imp.Path.Value[1:l-1]]
		if !ok {
			panic(fmt.Errorf("package %v (imported in %v) not in (*loader.Program).AllPackages", imp.Path.Value, f.Name.Name))
		}
		name := ""
		if imp.Name == nil {
			name = pkg.Name()
		} else {
			name = imp.Name.Name
		}
		m[pkg] = name
	}
	return func(p *types.Package) string { return m[p] }
}
