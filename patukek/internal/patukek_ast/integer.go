package patukek_ast

import (
	"strconv"

	"patukek/internal/patukek_code"
	"patukek/internal/patukek_compiler"
	"patukek/internal/patukek_obj"
)

type Integer int64

func NewInteger(i int64) Node {
	return Integer(i)
}

func (i Integer) String() string {
	return strconv.FormatInt(int64(i), 10)
}

func (i Integer) Compile(c *patukek_compiler.Compiler) (position int, err error) {
	return c.Emit(patukek_code.OpConstant, c.AddConstant(patukek_obj.Integer(i))), nil
}

func (i Integer) IsConstExpression() bool {
	return true
}