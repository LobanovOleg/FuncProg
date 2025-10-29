package patukek_ast

import (
	"patukek/internal/patukek_compiler"
)

type parseFn func(string, string) (Node, []error)

type Node interface {
	String() string
	patukek_compiler.Compilable
}