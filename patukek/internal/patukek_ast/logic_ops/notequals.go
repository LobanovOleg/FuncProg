package logic_ops

import (
	"patukek/internal/patukek_ast"
	"fmt"

	"patukek/internal/patukek_code"
	"patukek/internal/patukek_compiler"
)

type NotEquals struct {
	l   patukek_ast.Node
	r   patukek_ast.Node
	pos int
}

func NewNotEquals(l, r patukek_ast.Node, pos int) patukek_ast.Node {
	return NotEquals{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (n NotEquals) String() string {
	return fmt.Sprintf("(%v != %v)", n.l, n.r)
}

func (n NotEquals) Compile(c *patukek_compiler.Compiler) (position int, err error) {
	if position, err = n.l.Compile(c); err != nil {
		return
	}
	if position, err = n.r.Compile(c); err != nil {
		return
	}
	position = c.Emit(patukek_code.OpNotEqual)
	c.Bookmark(n.pos)
	return
}

func (n NotEquals) IsConstExpression() bool {
	return n.l.IsConstExpression() && n.r.IsConstExpression()
}