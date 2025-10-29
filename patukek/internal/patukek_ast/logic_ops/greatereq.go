package logic_ops

import (
	"patukek/internal/patukek_ast"
	"fmt"

	"patukek/internal/patukek_code"
	"patukek/internal/patukek_compiler"
)

type GreaterEq struct {
	l   patukek_ast.Node
	r   patukek_ast.Node
	pos int
}

func NewGreaterEq(l, r patukek_ast.Node, pos int) patukek_ast.Node {
	return GreaterEq{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (g GreaterEq) String() string {
	return fmt.Sprintf("(%v >= %v)", g.l, g.r)
}

func (g GreaterEq) Compile(c *patukek_compiler.Compiler) (position int, err error) {
	if position, err = g.l.Compile(c); err != nil {
		return
	}
	if position, err = g.r.Compile(c); err != nil {
		return
	}
	position = c.Emit(patukek_code.OpGreaterThanEqual)
	c.Bookmark(g.pos)
	return
}

func (g GreaterEq) IsConstExpression() bool {
	return g.l.IsConstExpression() && g.r.IsConstExpression()
}