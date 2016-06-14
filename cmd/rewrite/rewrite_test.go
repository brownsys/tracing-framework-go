package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"golang.org/x/tools/go/loader"
)

func TestGos(t *testing.T) {
	src := `
package main

import "golang.org/x/net/context"

func one(c context.Context) {}

func two(c context.Context, a int) {}

func three(c context.Context, a int, b int) {}

func variadic(c context.Context, a ...int) {}

func blank(_ context.Context) {}

func nested(c context.Context) {
	func(c context.Context) {}(c)
}

func main() {
	go one(nil)

	go two(nil, 0)

	go three(nil, 0, 1)

	go variadic(nil)

	go variadic(nil, 0)

	go variadic(nil, 0, 1)

	go variadic(nil, []int{1, 2, 3}...)

	go blank(nil)

	go nested(nil)
}
`

	expect := `package main

import "golang.org/x/net/context"

func one(c context.Context) {}

func two(c context.Context, a int) {}

func three(c context.Context, a int, b int) {}

func variadic(c context.Context, a ...int) {}

func blank(_ context.Context) {}

func nested(c context.Context) {
	func(c context.Context) {}(c)
}

func main() {
	go func(__f func(c context.
		Context), arg0 context.Context) {
		runtime.SetLocal(arg0)
		__f(arg0)

	}(one, nil)
	go func(__f func(c context.
		Context, a int), arg0 context.
		Context, arg1 int) {
		runtime.SetLocal(arg0)
		__f(arg0, arg1)
	}(two, nil, 0)
	go func(__f func(c context.
		Context, a int, b int), arg0 context.
		Context, arg1 int, arg2 int) {
		runtime.
			SetLocal(arg0)
		__f(arg0, arg1, arg2)
	}(three,

		nil, 0, 1)
	go func(__f func(c context.
		Context, a ...int), arg0 context.
		Context, arg1 ...[]int) {
		runtime.SetLocal(arg0)
		__f(arg0, arg1...)
	}(variadic, nil)
	go func(__f func(c context.
		Context, a ...int), arg0 context.
		Context, arg1 ...[]int) {
		runtime.SetLocal(arg0)
		__f(arg0, arg1...)
	}(variadic, nil,
		0,
	)
	go func(__f func(c context.
		Context, a ...int), arg0 context.
		Context, arg1 ...[]int) {
		runtime.SetLocal(arg0)
		__f(arg0, arg1...)
	}(variadic, nil,
		0,
		1)
	go func(__f func(c context.
		Context, a ...int), arg0 context.
		Context, arg1 ...[]int) {
		runtime.SetLocal(arg0)
		__f(arg0, arg1...)
	}(variadic, nil,
		[]int{1, 2, 3})
	go func(__f func(_ context.
		Context), arg0 context.Context) {
		runtime.SetLocal(arg0)
		__f(arg0)

	}(blank, nil)
	go func(__f func(c context.
		Context), arg0 context.Context) {
		runtime.SetLocal(arg0)
		__f(arg0)

	}(nested, nil)

}
`

	new := testHelper(t, src, rewriteGos)
	if new != expect {
		t.Errorf("unexpected output (see source for expected output):\n%v", new)
	}
}

type rewriter func(fset *token.FileSet, info types.Info, qual types.Qualifier, f *ast.File) error

func testHelper(t *testing.T, src string, r rewriter) string {
	c := loader.Config{
		Fset:        token.NewFileSet(),
		ParserMode:  parser.ParseComments | parser.DeclarationErrors,
		AllowErrors: true,
	}

	f, err := c.ParseFile("test.go", src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c.CreateFromFiles("main", f)
	p, err := c.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	info := p.Package("main").Info

	err = r(c.Fset, info, qualifierFromProg(p, f), f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	err = format.Node(&buf, c.Fset, f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b, err := format.Source(buf.Bytes())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return string(b)
}

// similar to qualifierForFile, but takes a *loader.Program
// instead of a *types.Package
func qualifierFromProg(p *loader.Program, f *ast.File) types.Qualifier {
	pathToPackage := make(map[string]*types.Package)
	for _, pkg := range p.AllPackages {
		pathToPackage[pkg.Pkg.Path()] = pkg.Pkg
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
