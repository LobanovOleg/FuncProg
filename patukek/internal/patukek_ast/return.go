package patukek_ast

import (
	"fmt"

	"patukek/internal/patukek_code"
	"patukek/internal/patukek_compiler"
)

type Return struct {
	v   Node
	pos int
}

func NewReturn(n Node, pos int) Node {
	return Return{
		v:   n,
		pos: pos,
	}
}

func (r Return) String() string {
	return fmt.Sprintf("return %v", r.v)
}

func (r Return) Compile(c *patukek_compiler.Compiler) (position int, err error) {
	if position, err = r.v.Compile(c); err != nil {
		return
	}
	position = c.Emit(patukek_code.OpReturnValue)
	c.Bookmark(r.pos)
	return
}

func (r Return) IsConstExpression() bool {
	return false
}