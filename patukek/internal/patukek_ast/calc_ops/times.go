package calc_ops

import (
	"patukek/internal/patukek_ast"
	"fmt"

	"patukek/internal/patukek_code"
	"patukek/internal/patukek_compiler"
)

type Times struct {
	l   patukek_ast.Node
	r   patukek_ast.Node
	pos int
}

func NewTimes(l, r patukek_ast.Node, pos int) patukek_ast.Node {
	return Times{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (t Times) String() string {
	return fmt.Sprintf("(%v * %v)", t.l, t.r)
}

func (t Times) Compile(c *patukek_compiler.Compiler) (position int, err error) {
	if position, err = t.l.Compile(c); err != nil {
		return
	}
	if position, err = t.r.Compile(c); err != nil {
		return
	}
	position = c.Emit(patukek_code.OpMul)
	c.Bookmark(t.pos)
	return
}

func (t Times) IsConstExpression() bool {
	return t.l.IsConstExpression() && t.r.IsConstExpression()
}