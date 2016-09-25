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

func one(a int) {}

func two(a, b int) {}

func three(a, b, c int) {}

func variadic(a ...int) {}

func blank(_ int) {}

func nested(a int) {
	func(a int) {}(a)
}

func main() {
	go one(0)

	go two(0, 1)

	go three(0, 1, 2)

	go variadic()

	go variadic(0)

	go variadic(0, 1)

	go variadic([]int{1, 2, 3}...)

	go blank(0)

	go nested(0)
}
`

	expect := `package main

import "github.com/brownsys/tracing-framework-go/local"

func one(a int) {}

func two(a, b int) {}

func three(a, b, c int) {}

func variadic(a ...int) {}

func blank(_ int) {}

func nested(a int) {
	func(a int) {}(a)
}

func main() {
	go func(__f1 func(), __f2 func(a int), arg0 int) {
		__f1()
		__f2(arg0)
	}(local.GetSpawnCallback(), one, 0)
	go func(__f1 func(), __f2 func(a int, b int), arg0 int, arg1 int) {
		__f1()
		__f2(arg0, arg1)
	}(local.GetSpawnCallback(), two, 0, 1)
	go func(__f1 func(), __f2 func(a int, b int, c int), arg0 int, arg1 int, arg2 int) {
		__f1()
		__f2(arg0, arg1, arg2)
	}(local.GetSpawnCallback(), three, 0, 1,
		2)
	go func(__f1 func(), __f2 func(a ...int), arg0 ...int) {
		__f1()
		__f2(arg0...)
	}(local.
		GetSpawnCallback(), variadic,
	)
	go func(__f1 func(), __f2 func(a ...int), arg0 ...int) {
		__f1()
		__f2(arg0...)
	}(local.
		GetSpawnCallback(), variadic,

		0)
	go func(__f1 func(), __f2 func(a ...int), arg0 ...int) {
		__f1()
		__f2(arg0...)
	}(local.
		GetSpawnCallback(), variadic,

		0, 1)
	go func(__f1 func(), __f2 func(a ...int), arg0 ...int) {
		__f1()
		__f2(arg0...)
	}(local.
		GetSpawnCallback(), variadic,

		[]int{1, 2, 3}...,
	)
	go func(__f1 func(), __f2 func(_ int), arg0 int) {
		__f1()
		__f2(arg0)
	}(local.GetSpawnCallback(), blank, 0)
	go func(__f1 func(), __f2 func(a int), arg0 int) {
		__f1()
		__f2(arg0)
	}(local.GetSpawnCallback(), nested, 0)

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
