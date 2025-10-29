package logic_ops

import (
	"patukek/internal/patukek_ast"
	"fmt"

	"patukek/internal/patukek_code"
	"patukek/internal/patukek_compiler"
)

type And struct {
	l   patukek_ast.Node
	r   patukek_ast.Node
	pos int
}

func NewAnd(l, r patukek_ast.Node, pos int) patukek_ast.Node {
	return And{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (a And) String() string {
	return fmt.Sprintf("(%v && %v)", a.l, a.r)
}

func (a And) Compile(c *patukek_compiler.Compiler) (p int, err error) {
	if p, err = a.l.Compile(c); err != nil {
		return
	}
	if p, err = a.r.Compile(c); err != nil {
		return
	}
	p = c.Emit(patukek_code.OpAnd)
	c.Bookmark(a.pos)
	return
}

func (a And) IsConstExpression() bool {
	return a.l.IsConstExpression() && a.r.IsConstExpression()
}