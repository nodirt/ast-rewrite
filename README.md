# Package rewrite

[![GoDoc](https://godoc.org/github.com/nodirt/ast-rewrite?status.svg)](https://godoc.org/github.com/nodirt/ast-rewrite)

Package rewrite can recursively rewrite a Go AST given a rewriter function.

The following example splits an asignment `x, y := 0, 1` to two assignments `x := 0` and `y := 1`

```go
src := "func() { x, y := 0, 1 }"
node, err := parser.ParseExpr(src)
if err != nil {
	panic(err)
}

node = rewrite.Rewrite(node, func(node ast.Node) (ast.Node, bool) {
	if assign, ok := node.(*AssignStmt); ok {
		if len(assign.Lhs) > 1 && assign.Tok == token.DEFINE {
			var block rewrite.ExpandedBlockStmt
			for i := range assign.Lhs {
				block.List = append(block.List, &AssignStmt{
					Lhs: []Expr{assign.Lhs[i]},
					Tok: token.DEFINE,
					Rhs: []Expr{assign.Rhs[i]},
				})
			}
			node = &block
		}
	}
	return node, true
}).(Expr)

format.Node(os.Stdout, token.NewFileSet(), node)
```

Output:

```go
func() {
 	x := 0
 	y := 1
}
```
