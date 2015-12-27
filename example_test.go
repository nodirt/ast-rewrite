package rewrite

import (
	. "go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
)

func ExampleRewrite() {
	src := "func() { x, y := 0, 1 }"
	node, err := parser.ParseExpr(src)
	if err != nil {
		panic(err)
	}

	node = Rewrite(node, func(node Node) (Node, bool) {
		if assign, ok := node.(*AssignStmt); ok {
			if len(assign.Lhs) > 1 && assign.Tok == token.DEFINE && len(assign.Lhs) == len(assign.Rhs) {
				var block ExpandedBlockStmt
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
	// Output:
	// func() {
	// 	x := 0
	// 	y := 1
	// }
}
