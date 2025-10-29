package patukek_ast

import (
	"fmt"
	"strings"

	"patukek/internal/patukek_code"
	"patukek/internal/patukek_compiler"
)

type List []Node

func NewList(elements ...Node) Node {
	var ret List

	for _, e := range elements {
		ret = append(ret, e)
	}
	return ret
}

func (l List) String() string {
	var elements []string

	for _, e := range l {
		if s, ok := e.(String); ok {
			elements = append(elements, s.Quoted())
		} else {
			elements = append(elements, e.String())
		}
	}
	return fmt.Sprintf("[%s]", strings.Join(elements, ", "))
}

func (l List) Compile(c *patukek_compiler.Compiler) (position int, err error) {
	for _, n := range l {
		if position, err = n.Compile(c); err != nil {
			return
		}
	}
	position = c.Emit(patukek_code.OpList, len(l))
	return
}

func (l List) IsConstExpression() bool {
	return false
}