package patukek_vm

import (
	"patukek/internal/patukek_code"
	"patukek/internal/patukek_obj"
)

type Frame struct {
	cl          *patukek_obj.Closure
	ip          int
	basePointer int
}

func NewFrame(cl *patukek_obj.Closure, basePointer int) *Frame {
	return &Frame{
		cl:          cl,
		ip:          -1,
		basePointer: basePointer,
	}
}

func (f *Frame) Instructions() patukek_code.Instructions {
	return f.cl.Fn.Instructions
}