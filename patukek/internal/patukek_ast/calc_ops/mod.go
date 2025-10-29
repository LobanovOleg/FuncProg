package calc_ops

import (
	"patukek/internal/patukek_ast"
	"fmt"

	"patukek/internal/patukek_code"
	"patukek/internal/patukek_compiler"
)

type Mod struct {
	l   patukek_ast.Node
	r   patukek_ast.Node
	pos int
}

func NewMod(l, r patukek_ast.Node, pos int) patukek_ast.Node {
	return Mod{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (m Mod) String() string {
	return fmt.Sprintf("(%v %% %v)", m.l, m.r)
}

func (m Mod) Compile(c *patukek_compiler.Compiler) (position int, err error) {

	if position, err = m.l.Compile(c); err != nil {
		return
	}
	if position, err = m.r.Compile(c); err != nil {
		return
	}
	position = c.Emit(patukek_code.OpMod)
	c.Bookmark(m.pos)
	return
}

func (m Mod) IsConstExpression() bool {
	return m.l.IsConstExpression() && m.r.IsConstExpression()
}