package logic_ops

import (
	"patukek/internal/patukek_ast"
	"fmt"

	"patukek/internal/patukek_code"
	"patukek/internal/patukek_compiler"
)

type LessEq struct {
	l   patukek_ast.Node
	r   patukek_ast.Node
	pos int
}

func NewLessEq(l, r patukek_ast.Node, pos int) patukek_ast.Node {
	return LessEq{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (l LessEq) String() string {
	return fmt.Sprintf("(%v <= %v)", l.l, l.r)
}

func (l LessEq) Compile(c *patukek_compiler.Compiler) (position int, err error) {
	if position, err = l.r.Compile(c); err != nil {
		return
	}
	if position, err = l.l.Compile(c); err != nil {
		return
	}
	position = c.Emit(patukek_code.OpGreaterThanEqual)
	c.Bookmark(l.pos)
	return
}

func (l LessEq) IsConstExpression() bool {
	return l.l.IsConstExpression() && l.r.IsConstExpression()
}