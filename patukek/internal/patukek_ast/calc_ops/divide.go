package calc_ops

import (
	"patukek/internal/patukek_ast"
	"fmt"

	"patukek/internal/patukek_code"
	"patukek/internal/patukek_compiler"
)

type Divide struct {
	l   patukek_ast.Node
	r   patukek_ast.Node
	pos int
}

func NewDivide(l, r patukek_ast.Node, pos int) patukek_ast.Node {
	return Divide{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (d Divide) String() string {
	return fmt.Sprintf("(%v / %v)", d.l, d.r)
}

func (d Divide) Compile(c *patukek_compiler.Compiler) (position int, err error) {
	if position, err = d.l.Compile(c); err != nil {
		return
	}
	if position, err = d.r.Compile(c); err != nil {
		return
	}
	position = c.Emit(patukek_code.OpDiv)
	c.Bookmark(d.pos)
	return
}

func (d Divide) IsConstExpression() bool {
	return d.l.IsConstExpression() && d.r.IsConstExpression()
}