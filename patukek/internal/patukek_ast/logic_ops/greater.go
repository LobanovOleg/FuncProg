package logic_ops

import (
	"patukek/internal/patukek_ast"
	"fmt"

	"patukek/internal/patukek_code"
	"patukek/internal/patukek_compiler"
)

type Greater struct {
	l   patukek_ast.Node
	r   patukek_ast.Node
	pos int
}

func NewGreater(l, r patukek_ast.Node, pos int) patukek_ast.Node {
	return Greater{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (g Greater) String() string {
	return fmt.Sprintf("(%v > %v)", g.l, g.r)
}

func (g Greater) Compile(c *patukek_compiler.Compiler) (position int, err error) {
	if position, err = g.l.Compile(c); err != nil {
		return
	}
	if position, err = g.r.Compile(c); err != nil {
		return
	}
	position = c.Emit(patukek_code.OpGreaterThan)
	c.Bookmark(g.pos)
	return
}

func (g Greater) IsConstExpression() bool {
	return g.l.IsConstExpression() && g.r.IsConstExpression()
}