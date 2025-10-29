package logic_ops

import (
	"patukek/internal/patukek_ast"
	"fmt"

	"patukek/internal/patukek_code"
	"patukek/internal/patukek_compiler"
)

type Equals struct {
	l   patukek_ast.Node
	r   patukek_ast.Node
	pos int
}

func NewEquals(l, r patukek_ast.Node, pos int) patukek_ast.Node {
	return Equals{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (e Equals) String() string {
	return fmt.Sprintf("(%v == %v)", e.l, e.r)
}

func (e Equals) Compile(c *patukek_compiler.Compiler) (position int, err error) {
	if position, err = e.l.Compile(c); err != nil {
		return
	}
	if position, err = e.r.Compile(c); err != nil {
		return
	}
	position = c.Emit(patukek_code.OpEqual)
	c.Bookmark(e.pos)
	return
}

func (e Equals) IsConstExpression() bool {
	return e.l.IsConstExpression() && e.r.IsConstExpression()
}