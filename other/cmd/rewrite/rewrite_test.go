package main

import (
	"bytes"
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

import (
	"golang.org/x/net/context"
	__runtime "runtime"
)

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
		__runtime.SetLocal(arg0)
		__f(arg0)
	}(one, nil)
	go func(__f func(c context.
		Context, a int), arg0 context.
		Context, arg1 int) {
		__runtime.SetLocal(
			arg0)
		__f(arg0, arg1)
	}(two, nil, 0)
	go func(__f func(c context.
		Context, a int, b int), arg0 context.
		Context, arg1 int, arg2 int) {
		__runtime.
			SetLocal(arg0)
		__f(arg0, arg1, arg2)
	}(three,

		nil, 0, 1)
	go func(__f func(c context.
		Context, a ...int), arg0 context.
		Context, arg1 ...[]int) {
		__runtime.
			SetLocal(arg0)
		__f(arg0, arg1...)
	}(variadic, nil)
	go func(__f func(c context.
		Context, a ...int), arg0 context.
		Context, arg1 ...[]int) {
		__runtime.
			SetLocal(arg0)
		__f(arg0, arg1...)
	}(variadic, nil,

		0)
	go func(__f func(c context.
		Context, a ...int), arg0 context.
		Context, arg1 ...[]int) {
		__runtime.
			SetLocal(arg0)
		__f(arg0, arg1...)
	}(variadic, nil,

		0, 1)
	go func(__f func(c context.
		Context, a ...int), arg0 context.
		Context, arg1 ...[]int) {
		__runtime.
			SetLocal(arg0)
		__f(arg0, arg1...)
	}(variadic, nil,

		[]int{1, 2, 3})
	go func(__f func(_ context.
		Context), arg0 context.Context) {
		__runtime.SetLocal(arg0)
		__f(arg0)
	}(blank, nil)
	go func(__f func(c context.
		Context), arg0 context.Context) {
		__runtime.SetLocal(arg0)
		__f(arg0)
	}(nested, nil)

}
`

	new := testHelper(t, src, rewriteGos)
	if new != expect {
		var idx int
		for i, c := range []byte(new) {
			if c != expect[i] {
				idx = i
				break
			}
		}
		t.Errorf("unexpected output (see source for expected output) at character %v:\n%v", idx, new)
	}
}

type rewriter func(fset *token.FileSet, info types.Info, qual types.Qualifier, f *ast.File) (bool, error)

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
	pi := p.Package("main")

	_, err = r(c.Fset, pi.Info, qualifierForFile(pi.Pkg, f), f)
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
