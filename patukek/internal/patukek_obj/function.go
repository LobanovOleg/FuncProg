package patukek_obj

import (
	"fmt"
	"strings"

	"patukek/internal/patukek_code"
	"patukek/internal/patukek_err"
)

type Function struct {
	Body   any
	Env    *Env
	Params []string
}

func (f Function) Type() Type {
	return FunctionType
}

func (f Function) String() string {
	return fmt.Sprintf("fn(%s) { %v }", strings.Join(f.Params, ", "), f.Body)
}

type CompiledFunction struct {
	Instructions patukek_code.Instructions
	NumLocals    int
	NumParams    int
	Bookmarks    []patukek_err.Bookmark
}

func NewFunctionCompiled(i patukek_code.Instructions, nLocals, nParams int, bookmarks []patukek_err.Bookmark) Object {
	return &CompiledFunction{
		Instructions: i,
		NumLocals:    nLocals,
		NumParams:    nParams,
		Bookmarks:    bookmarks,
	}
}

func (c CompiledFunction) Type() Type {
	return FunctionType
}

func (c CompiledFunction) String() string {
	return "<compiled function>"
}