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
)

func rewriteGos(fset *token.FileSet, info types.Info, qual types.Qualifier, f *ast.File) error {
	return mapStmts(f, func(s ast.Stmt) ([]ast.Stmt, error) {
		if g, ok := s.(*ast.GoStmt); ok {
			return rewriteGoStmt(fset, info, qual, g)
		}
		return nil, nil
	})
}

func rewriteCalls(fset *token.FileSet, info types.Info, qual types.Qualifier, f *ast.File) error {
	return mapStmts(f, func(s ast.Stmt) ([]ast.Stmt, error) {
		if a, ok := s.(*ast.AssignStmt); ok {
			return rewriteCallStmt(fset, info, qual, a)
		}
		return nil, nil
	})
}

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

func rewriteCallStmt(fset *token.FileSet, info types.Info, qual types.Qualifier, a *ast.AssignStmt) ([]ast.Stmt, error) {
	// for the time being, we only handle
	// statements which have a single
	// function call on the RHS, like:
	//  a, b = f()

	if len(a.Rhs) != 1 {
		for _, aa := range a.Rhs {
			if _, ok := aa.(*ast.CallExpr); ok {
				return nil, fmt.Errorf("%v: unsupported statement format", fset.Position(a.Pos()))
			}
		}
		// none of the RHS expressions are function
		// calls, so we can just safely ignore this
		return nil, nil
	}

	c, ok := a.Rhs[0].(*ast.CallExpr)
	if !ok {
		return nil, nil
	}

	rettyp := info.TypeOf(c)
	if rettyp == nil {
		return nil, fmt.Errorf("%v: could not determine return type of function",
			fset.Position(c.Pos()))
	}

	var vname string

	// since the code has been type checked,
	// we can assume that the function has
	// at least one return value, and that
	// len(LHS) = len(RHS)
	if t, ok := rettyp.(*types.Tuple); ok {
		context := false
		for i := 0; i < t.Len(); i++ {
			switch v := a.Lhs[i].(type) {
			case *ast.Ident:
				if v.Name != "_" && isContext(t.At(i).Type()) {
					if context {
						// more than one context.Context variable
						return nil, fmt.Errorf("%v: unsupported statement format", fset.Position(a.Pos()))
					}
					context = true
					vname = v.Name
				}
			default:
				// TODO: handle LHS elements other than identifiers
				return nil, nil
				panic(fmt.Errorf("unexpected type %v", reflect.TypeOf(v)))
			}
		}
		if !context {
			return nil, nil
		}
	} else {
		switch v := a.Lhs[0].(type) {
		case *ast.Ident:
			if v.Name == "_" || !isContext(rettyp) {
				return nil, nil
			}
			vname = v.Name
		default:
			// TODO: handle LHS elements other than identifiers
			return nil, nil
			// panic(fmt.Errorf("unexpected type %v", reflect.TypeOf(v)))
		}
	}

	arg := struct{ Ctx string }{vname}

	var buf bytes.Buffer
	err := callTmpl.Execute(&buf, arg)
	if err != nil {
		panic(fmt.Errorf("internal error: %v", err))
	}
	return append([]ast.Stmt{a}, parseStmts(string(buf.Bytes()))...), nil
}

var callTmpl = template.Must(template.New("").Parse(`runtime.SetLocal({{.Ctx}})`))

func rewriteGoStmt(fset *token.FileSet, info types.Info, qual types.Qualifier, g *ast.GoStmt) ([]ast.Stmt, error) {
	ftyp := info.TypeOf(g.Call.Fun)

	if ftyp == nil {
		return nil, fmt.Errorf("%v: could not determine type of function",
			fset.Position(g.Call.Fun.Pos()))
	}
	sig := ftyp.(*types.Signature)

	// According to the context documentation:
	//
	// Do not store Contexts inside a struct type;
	// instead, pass a Context explicitly to each
	// function that needs it. The Context should
	// be the first parameter, typically named ctx.
	//
	// Thus, we only handle this case.
	if sig.Params().Len() == 0 || !isContext(sig.Params().At(0).Type()) {
		return nil, nil
	}

	var arg struct {
		Func                          string
		Typ                           string
		DefArgs, InnerArgs, OuterArgs []string
	}

	arg.Func = nodeString(fset, g.Call.Fun)
	arg.Typ = types.TypeString(ftyp, qual)

	params := sig.Params()
	for i := 0; i < params.Len(); i++ {
		typ := types.TypeString(params.At(i).Type(), qual)
		name := fmt.Sprintf("arg%v", i)
		if sig.Variadic() && i == params.Len()-1 {
			arg.DefArgs = append(arg.DefArgs, name+" ..."+typ)
			arg.InnerArgs = append(arg.InnerArgs, name+"...")
		} else {
			arg.DefArgs = append(arg.DefArgs, name+" "+typ)
			arg.InnerArgs = append(arg.InnerArgs, name)
		}
	}

	for _, a := range g.Call.Args {
		arg.OuterArgs = append(arg.OuterArgs, nodeString(fset, a))
	}

	var buf bytes.Buffer
	err := goTmpl.Execute(&buf, arg)
	if err != nil {
		panic(fmt.Errorf("internal error: %v", err))
	}
	return parseStmts(string(buf.Bytes())), nil
}

var goTmpl = template.Must(template.New("").Parse(`
go func(__f {{.Typ}} {{range .DefArgs}},{{.}}{{end}}){
	runtime.SetLocal(arg0)
	__f({{range .InnerArgs}}{{.}},{{end}})
}({{.Func}}{{range .OuterArgs}},{{.}}{{end}})
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

func isContext(t types.Type) bool {
	return t.String() == "golang.org/x/net/context.Context" ||
		t.String() == "context.Context"
}
