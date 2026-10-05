// Command gostub rewrites Go files in place: every function and method body becomes
// `panic("TODO: see chapter 38")`, keeping signatures, doc comments, types, constants and vars.
// Test files (_test.go) are never touched. Used by tools/make_project_starter.py.
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

func stub(path string) (int, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil || fn.Name.Name == "main" {
			continue
		}
		msg := &ast.BasicLit{Kind: token.STRING, Value: `"TODO: see chapter 38"`}
		fn.Body.List = []ast.Stmt{&ast.ExprStmt{X: &ast.CallExpr{Fun: ast.NewIdent("panic"), Args: []ast.Expr{msg}}}}
		n++
	}
	// Comments inside removed bodies would float; drop every comment not attached to a declaration.
	keep := file.Comments[:0]
	for _, cg := range file.Comments {
		if insideBody(file, cg) {
			continue
		}
		keep = append(keep, cg)
	}
	file.Comments = keep
	pruneImports(file)
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, file); err != nil {
		return 0, err
	}
	return n, os.WriteFile(path, buf.Bytes(), 0o644)
}

func insideBody(file *ast.File, cg *ast.CommentGroup) bool {
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Body != nil && fn.Name.Name != "main" {
			if cg.Pos() > fn.Body.Lbrace && cg.Pos() < fn.Body.Rbrace {
				return true
			}
		}
	}
	return false
}

// pruneImports drops imports that no longer appear in the file once bodies are gone, so the starter compiles.
func pruneImports(file *ast.File) {
	used := map[string]bool{}
	ast.Inspect(file, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok {
			if id, ok := sel.X.(*ast.Ident); ok {
				used[id.Name] = true
			}
		}
		return true
	})
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.IMPORT {
			continue
		}
		specs := gen.Specs[:0]
		for _, s := range gen.Specs {
			spec := s.(*ast.ImportSpec)
			if spec.Name != nil && (spec.Name.Name == "_" || spec.Name.Name == ".") || used[importName(spec)] {
				specs = append(specs, spec)
			}
		}
		gen.Specs = specs
	}
	imports := file.Imports[:0]
	for _, spec := range file.Imports {
		if spec.Name != nil && (spec.Name.Name == "_" || spec.Name.Name == ".") || used[importName(spec)] {
			imports = append(imports, spec)
		}
	}
	file.Imports = imports
}

// importName is the local name of an import: its alias, or the last path element without a /vN or .vN suffix.
func importName(spec *ast.ImportSpec) string {
	if spec.Name != nil {
		return spec.Name.Name
	}
	parts := strings.Split(strings.Trim(spec.Path.Value, `"`), "/")
	last := parts[len(parts)-1]
	if len(parts) > 1 && len(last) > 1 && last[0] == 'v' && strings.Trim(last[1:], "0123456789") == "" {
		last = parts[len(parts)-2]
	}
	if i := strings.Index(last, ".v"); i > 0 {
		last = last[:i]
	}
	return strings.ReplaceAll(last, "-", "")
}

func main() {
	total := 0
	for _, path := range os.Args[1:] {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		n, err := stub(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, path, err)
			os.Exit(1)
		}
		total += n
	}
	fmt.Printf("gostub: %d function bodies stubbed in %d files\n", total, len(os.Args)-1)
}
