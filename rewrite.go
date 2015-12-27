package rewrite

import (
	"fmt"
	. "go/ast"
	"go/ast"
)

// Rewriter transforms a node.
// If continueRewriting is false, Rewrite does not rewrite children of the
// returned node.
type Rewriter func(Node) (n Node, continueRewriting bool)

// ExpandedBlockStmt is replaced with its children.
// See Rewriter's example.
type ExpandedBlockStmt struct {
	BlockStmt
}

// Rewrite recursively rewrites a node using rewriter r.
// It is like ast.Walk, but a rewriter returns a substitution node.
//
// If the returned node is *ExpandBlockNode and its destination is a list
// of statements, e.g. *ast.BlockStmt, the ExpandBlockNode is replaced with
// its substatements.
func Rewrite(node Node, r Rewriter) ast.Node {
	// return ast.Node instead of Node because of the bug
	// https://github.com/golang/go/issues/13742
	node, continueRewriting := r(node)
	if !continueRewriting {
		return node
	}

	// rewrite children
	// (the order of the cases matches the order
	// of the corresponding node types in go)
	switch n := node.(type) {
	// Comments and fields
	case *Comment:
	// nothing to do

	case *CommentGroup:
		for i, c := range n.List {
			n.List[i] = Rewrite(c, r).(*Comment)
		}

	case *Field:
		rewriteCommentGroup(&n.Doc, r)
		rewriteIdentList(n.Names, r)
		rewriteExpr(&n.Type, r)
		if n.Tag != nil {
			n.Tag = Rewrite(n.Tag, r).(*BasicLit)
		}
		rewriteCommentGroup(&n.Comment, r)

	case *FieldList:
		for i, f := range n.List {
			n.List[i] = Rewrite(f, r).(*Field)
		}

	// Expressions
	case *BadExpr, *Ident, *BasicLit:
	// nothing to do

	case *Ellipsis:
		rewriteExpr(&n.Elt, r)

	case *FuncLit:
		n.Type = Rewrite(n.Type, r).(*FuncType)
		rewriteBlockStmt(&n.Body, r)

	case *CompositeLit:
		rewriteExpr(&n.Type, r)
		rewriteExprList(n.Elts, r)

	case *ParenExpr:
		rewriteExpr(&n.X, r)

	case *SelectorExpr:
		rewriteExpr(&n.X, r)
		rewriteIdent(&n.Sel, r)

	case *IndexExpr:
		rewriteExpr(&n.X, r)
		rewriteExpr(&n.Index, r)

	case *SliceExpr:
		rewriteExpr(&n.X, r)
		rewriteExpr(&n.Low, r)
		rewriteExpr(&n.High, r)
		rewriteExpr(&n.Max, r)

	case *TypeAssertExpr:
		rewriteExpr(&n.X, r)
		rewriteExpr(&n.Type, r)

	case *CallExpr:
		rewriteExpr(&n.Fun, r)
		rewriteExprList(n.Args, r)

	case *StarExpr:
		rewriteExpr(&n.X, r)

	case *UnaryExpr:
		rewriteExpr(&n.X, r)

	case *BinaryExpr:
		rewriteExpr(&n.X, r)
		rewriteExpr(&n.Y, r)

	case *KeyValueExpr:
		rewriteExpr(&n.Key, r)
		rewriteExpr(&n.Value, r)

	// Types
	case *ArrayType:
		rewriteExpr(&n.Len, r)
		rewriteExpr(&n.Elt, r)

	case *StructType:
		rewriteFieldList(&n.Fields, r)

	case *FuncType:
		rewriteFieldList(&n.Params, r)
		rewriteFieldList(&n.Results, r)

	case *InterfaceType:
		rewriteFieldList(&n.Methods, r)

	case *MapType:
		rewriteExpr(&n.Value, r)
		rewriteExpr(&n.Key, r)

	case *ChanType:
		rewriteExpr(&n.Value, r)

	// Statements
	case *BadStmt:
	// nothing to do

	case *DeclStmt:
		n.Decl = Rewrite(n.Decl, r).(Decl)

	case *EmptyStmt:
	// nothing to do

	case *LabeledStmt:
		rewriteIdent(&n.Label, r)
		rewriteStmt(&n.Stmt, r)

	case *ExprStmt:
		rewriteExpr(&n.X, r)

	case *SendStmt:
		rewriteExpr(&n.Chan, r)
		rewriteExpr(&n.Value, r)

	case *IncDecStmt:
		rewriteExpr(&n.X, r)

	case *AssignStmt:
		rewriteExprList(n.Lhs, r)
		rewriteExprList(n.Rhs, r)

	case *GoStmt:
		n.Call = Rewrite(n.Call, r).(*CallExpr)

	case *DeferStmt:
		n.Call = Rewrite(n.Call, r).(*CallExpr)

	case *ReturnStmt:
		rewriteExprList(n.Results, r)

	case *BranchStmt:
		if n.Label != nil {
			rewriteIdent(&n.Label, r)
		}

	case *BlockStmt:
		rewriteStmtList(&n.List, r)
	case *ExpandedBlockStmt:
		rewriteStmtList(&n.List, r)

	case *IfStmt:
		rewriteStmt(&n.Init, r)
		rewriteExpr(&n.Cond, r)
		rewriteBlockStmt(&n.Body, r)
		rewriteStmt(&n.Else, r)

	case *CaseClause:
		rewriteExprList(n.List, r)
		rewriteStmtList(&n.Body, r)

	case *SwitchStmt:
		rewriteStmt(&n.Init, r)
		rewriteExpr(&n.Tag, r)
		rewriteBlockStmt(&n.Body, r)

	case *TypeSwitchStmt:
		rewriteStmt(&n.Init, r)
		rewriteStmt(&n.Assign, r)
		rewriteBlockStmt(&n.Body, r)

	case *CommClause:
		rewriteStmt(&n.Comm, r)
		rewriteStmtList(&n.Body, r)

	case *SelectStmt:
		rewriteBlockStmt(&n.Body, r)

	case *ForStmt:
		rewriteStmt(&n.Init, r)
		rewriteExpr(&n.Cond, r)
		rewriteStmt(&n.Post, r)
		rewriteBlockStmt(&n.Body, r)

	case *RangeStmt:
		rewriteExpr(&n.Key, r)
		rewriteExpr(&n.Value, r)
		rewriteExpr(&n.X, r)
		rewriteBlockStmt(&n.Body, r)

	// Declarations
	case *ImportSpec:
		rewriteCommentGroup(&n.Doc, r)
		if n.Name != nil {
			rewriteIdent(&n.Name, r)
		}
		n.Path = Rewrite(n.Path, r).(*BasicLit)
		rewriteCommentGroup(&n.Comment, r)

	case *ValueSpec:
		rewriteCommentGroup(&n.Doc, r)
		rewriteIdentList(n.Names, r)
		rewriteExpr(&n.Type, r)
		rewriteExprList(n.Values, r)
		rewriteCommentGroup(&n.Comment, r)

	case *TypeSpec:
		rewriteCommentGroup(&n.Doc, r)
		rewriteIdent(&n.Name, r)
		rewriteExpr(&n.Type, r)
		rewriteCommentGroup(&n.Comment, r)

	case *BadDecl:
	// nothing to do

	case *GenDecl:
		rewriteCommentGroup(&n.Doc, r)
		for i, s := range n.Specs {
			n.Specs[i] = Rewrite(s, r).(Spec)
		}

	case *FuncDecl:
		rewriteCommentGroup(&n.Doc, r)
		rewriteFieldList(&n.Recv, r)
		rewriteIdent(&n.Name, r)
		n.Type = Rewrite(n.Type, r).(*FuncType)
		rewriteBlockStmt(&n.Body, r)

	// Files and packages
	case *File:
		rewriteCommentGroup(&n.Doc, r)
		rewriteIdent(&n.Name, r)
		rewriteDeclList(n.Decls, r)
		// don't walk n.Comments - they have been
		// visited already through the individual
		// nodes

	case *Package:
		for i, f := range n.Files {
			n.Files[i] = Rewrite(f, r).(*File)
		}

	default:
		fmt.Printf("Rewrite: unexpected node type %T", n)
		panic("Rewrite")
	}

	r(nil)
	return node
}

func rewriteIdentList(list []*Ident, r Rewriter) {
	for i := range list {
		rewriteIdent(&list[i], r)
	}
}

func rewriteIdent(ident **Ident, r Rewriter) {
	if *ident != nil {
		*ident = Rewrite(*ident, r).(*Ident)
	}

}

func rewriteExprList(list []Expr, r Rewriter) {
	for i := range list {
		rewriteExpr(&list[i], r)
	}
}

func rewriteStmtList(list *[]Stmt, r Rewriter) {
	l := *list
	for i := 0; i < len(l); {
		rewriteStmt(&l[i], r)
		switch stmt := l[i].(type) {
		case *ExpandedBlockStmt:
			l = append(l[:i], append(stmt.List, l[i+1:]...)...)
			i += len(stmt.List)
		default:
			i++
		}
	}
	*list = l
}

func rewriteDeclList(list []Decl, r Rewriter) {
	for i, x := range list {
		list[i] = Rewrite(x, r).(Decl)
	}
}

func rewriteExpr(expr *Expr, r Rewriter) {
	if *expr != nil {
		*expr = Rewrite(*expr, r).(Expr)
	}
}

func rewriteStmt(stmt *Stmt, r Rewriter) {
	if *stmt != nil {
		*stmt = Rewrite(*stmt, r).(Stmt)
	}
}

func rewriteBlockStmt(stmt **BlockStmt, r Rewriter) {
	if *stmt != nil {
		*stmt = Rewrite(*stmt, r).(*BlockStmt)
	}
}

func rewriteCommentGroup(comment **CommentGroup, r Rewriter) {
	if *comment != nil {
		*comment = Rewrite(*comment, r).(*CommentGroup)
	}
}

func rewriteFieldList(list **FieldList, r Rewriter) {
	if *list != nil {
		*list = Rewrite(*list, r).(*FieldList)
	}
}
