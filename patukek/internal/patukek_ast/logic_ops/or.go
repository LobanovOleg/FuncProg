package logic_ops

import (
	"patukek/internal/patukek_ast"
	"fmt"

	"patukek/internal/patukek_code"
	"patukek/internal/patukek_compiler"
)

type Or struct {
	l   patukek_ast.Node
	r   patukek_ast.Node
	pos int
}

func NewOr(l, r patukek_ast.Node, pos int) patukek_ast.Node {
	return Or{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (o Or) String() string {
	return fmt.Sprintf("(%v || %v)", o.l, o.r)
}

func (o Or) Compile(c *patukek_compiler.Compiler) (position int, err error) {
	if position, err = o.l.Compile(c); err != nil {
		return
	}
	if position, err = o.r.Compile(c); err != nil {
		return
	}
	position = c.Emit(patukek_code.OpOr)
	c.Bookmark(o.pos)
	return
}

func (o Or) IsConstExpression() bool {
	return o.l.IsConstExpression() && o.r.IsConstExpression()
}