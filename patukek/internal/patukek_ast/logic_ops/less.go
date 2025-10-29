package logic_ops

import (
	"patukek/internal/patukek_ast"
	"fmt"

	"patukek/internal/patukek_code"
	"patukek/internal/patukek_compiler"
)

type Less struct {
	l   patukek_ast.Node
	r   patukek_ast.Node
	pos int
}

func NewLess(l, r patukek_ast.Node, pos int) patukek_ast.Node {
	return Less{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (l Less) String() string {
	return fmt.Sprintf("(%v < %v)", l.l, l.r)
}

func (l Less) Compile(c *patukek_compiler.Compiler) (position int, err error) {
	if position, err = l.r.Compile(c); err != nil {
		return
	}
	if position, err = l.l.Compile(c); err != nil {
		return
	}
	position = c.Emit(patukek_code.OpGreaterThan)
	c.Bookmark(l.pos)
	return
}

func (l Less) IsConstExpression() bool {
	return l.l.IsConstExpression() && l.r.IsConstExpression()
}