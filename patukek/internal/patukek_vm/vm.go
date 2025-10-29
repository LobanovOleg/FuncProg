package patukek_vm

import (
	"patukek/internal/patukek_code"
	"patukek/internal/patukek_compiler"
	"patukek/internal/patukek_err"
	"patukek/internal/patukek_obj"
	"fmt"
	"path/filepath"
)

type State struct {
	Symbols *patukek_compiler.SymbolTable
	Consts  []patukek_obj.Object
	Globals []patukek_obj.Object
}

func NewState() *State {
	st := patukek_compiler.NewSymbolTable()
	for i, builtin := range patukek_obj.Builtins {
		st.DefineBuiltin(i, builtin.Name)
	}

	return &State{
		Consts:  []patukek_obj.Object{},
		Globals: make([]patukek_obj.Object, GlobalSize),
		Symbols: st,
	}
}

type VM struct {
	*State
	dir        string
	file       string
	stack      []patukek_obj.Object
	frames     []*Frame
	localTable []bool
	sp         int
	frameIndex int
}

const (
	StackSize  = 2048
	GlobalSize = 65536
	MaxFrames  = 1024
)

var (
	True  = patukek_obj.True
	False = patukek_obj.False
	Null  = patukek_obj.NullObj
)

func New(file string, bytecode *patukek_compiler.Bytecode) *VM {
	vm := &VM{
		stack:      make([]patukek_obj.Object, StackSize),
		frames:     make([]*Frame, MaxFrames),
		frameIndex: 1,
		localTable: make([]bool, GlobalSize),
		State:      NewState(),
	}

	vm.dir, vm.file = filepath.Split(file)
	vm.Consts = bytecode.Constants
	fn := &patukek_obj.CompiledFunction{
		Instructions: bytecode.Instructions,
		Bookmarks:    bytecode.Bookmarks,
	}
	vm.frames[0] = NewFrame(&patukek_obj.Closure{Fn: fn}, 0)
	return vm
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.frameIndex-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.frameIndex] = f
	vm.frameIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.frameIndex--
	return vm.frames[vm.frameIndex]
}

func (vm *VM) bookmark() patukek_err.Bookmark {
	var (
		frame     = vm.currentFrame()
		offset    = frame.ip
		bookmarks = frame.cl.Fn.Bookmarks
	)

	if len(bookmarks) == 0 {
		return patukek_err.Bookmark{}
	}

	prev := bookmarks[0]
	for _, cur := range bookmarks[1:] {
		if offset < prev.Offset {
			return prev
		} else if offset > prev.Offset && offset <= cur.Offset {
			return cur
		}
		prev = cur
	}
	return prev
}

func (vm *VM) errorf(s string, a ...any) error {
	return patukek_err.NewFromBookmark(
		filepath.Join(vm.dir, vm.file),
		vm.bookmark(),
		s,
		a...,
	)
}

func (vm *VM) execAdd() error {
	var (
		right = patukek_obj.Unwrap(vm.pop())
		left  = patukek_obj.Unwrap(vm.pop())
	)

	switch {
	case patukek_obj.AssertTypes(left, patukek_obj.IntType) && patukek_obj.AssertTypes(right, patukek_obj.IntType):
		l := left.(patukek_obj.Integer)
		r := right.(patukek_obj.Integer)
		return vm.push(l + r)

	case patukek_obj.AssertTypes(left, patukek_obj.StringType) && patukek_obj.AssertTypes(right, patukek_obj.StringType):
		l := left.(patukek_obj.String)
		r := right.(patukek_obj.String)
		return vm.push(l + r)

	default:
		return vm.errorf("unsupported operator '+' for types %v and %v", left.Type(), right.Type())
	}
}

func (vm *VM) execSub() error {
	var (
		right = patukek_obj.Unwrap(vm.pop())
		left  = patukek_obj.Unwrap(vm.pop())
	)

	switch {
	case patukek_obj.AssertTypes(left, patukek_obj.IntType) && patukek_obj.AssertTypes(right, patukek_obj.IntType):
		l := left.(patukek_obj.Integer)
		r := right.(patukek_obj.Integer)
		return vm.push(l - r)

	default:
		return vm.errorf("unsupported operator '-' for types %v and %v", left.Type(), right.Type())
	}
}

func (vm *VM) execMul() error {
	var (
		right = patukek_obj.Unwrap(vm.pop())
		left  = patukek_obj.Unwrap(vm.pop())
	)

	switch {
	case patukek_obj.AssertTypes(left, patukek_obj.IntType) && patukek_obj.AssertTypes(right, patukek_obj.IntType):
		l := left.(patukek_obj.Integer)
		r := right.(patukek_obj.Integer)
		return vm.push(l * r)

	default:
		return vm.errorf("unsupported operator '*' for types %v and %v", left.Type(), right.Type())
	}
}

func (vm *VM) execDiv() error {
	var (
		right = patukek_obj.Unwrap(vm.pop())
		left  = patukek_obj.Unwrap(vm.pop())
	)

	if !patukek_obj.AssertTypes(left, patukek_obj.IntType) || !patukek_obj.AssertTypes(right, patukek_obj.IntType) {
		return fmt.Errorf("unsupported operator '/' for types %v and %v", left.Type(), right.Type())
	}

	l := left.(patukek_obj.Integer)
	r := right.(patukek_obj.Integer)
	return vm.push(l / r)
}

func (vm *VM) execMod() error {
	var (
		right = patukek_obj.Unwrap(vm.pop())
		left  = patukek_obj.Unwrap(vm.pop())
	)

	if !patukek_obj.AssertTypes(left, patukek_obj.IntType) || !patukek_obj.AssertTypes(right, patukek_obj.IntType) {
		return fmt.Errorf("unsupported operator '%%' for types %v and %v", left.Type(), right.Type())
	}

	l := left.(patukek_obj.Integer)
	r := right.(patukek_obj.Integer)

	if r == 0 {
		return vm.errorf("can't divide by 0")
	}
	return vm.push(l % r)
}

func (vm *VM) execEqual() error {
	var (
		right = patukek_obj.Unwrap(vm.pop())
		left  = patukek_obj.Unwrap(vm.pop())
	)

	switch {
	case patukek_obj.AssertTypes(left, patukek_obj.BoolType, patukek_obj.NullType) || patukek_obj.AssertTypes(right, patukek_obj.BoolType, patukek_obj.NullType):
		return vm.push(patukek_obj.ParseBool(left == right))

	case patukek_obj.AssertTypes(left, patukek_obj.StringType) && patukek_obj.AssertTypes(right, patukek_obj.StringType):
		l := left.(patukek_obj.String)
		r := right.(patukek_obj.String)
		return vm.push(patukek_obj.ParseBool(l == r))

	case patukek_obj.AssertTypes(left, patukek_obj.IntType) && patukek_obj.AssertTypes(right, patukek_obj.IntType):
		l := left.(patukek_obj.Integer)
		r := right.(patukek_obj.Integer)
		return vm.push(patukek_obj.ParseBool(l == r))

	default:
		return vm.push(False)
	}
}

func (vm *VM) execNotEqual() error {
	var (
		right = patukek_obj.Unwrap(vm.pop())
		left  = patukek_obj.Unwrap(vm.pop())
	)

	switch {
	case patukek_obj.AssertTypes(left, patukek_obj.BoolType, patukek_obj.NullType) || patukek_obj.AssertTypes(right, patukek_obj.BoolType, patukek_obj.NullType):
		return vm.push(patukek_obj.ParseBool(left != right))

	case patukek_obj.AssertTypes(left, patukek_obj.StringType) && patukek_obj.AssertTypes(right, patukek_obj.StringType):
		l := left.(patukek_obj.String)
		r := right.(patukek_obj.String)
		return vm.push(patukek_obj.ParseBool(l != r))

	case patukek_obj.AssertTypes(left, patukek_obj.IntType) && patukek_obj.AssertTypes(right, patukek_obj.IntType):
		l := left.(patukek_obj.Integer)
		r := right.(patukek_obj.Integer)
		return vm.push(patukek_obj.ParseBool(l != r))

	default:
		return vm.push(True)
	}
}

func (vm *VM) execAnd() error {
	var (
		right = patukek_obj.Unwrap(vm.pop())
		left  = patukek_obj.Unwrap(vm.pop())
	)

	return vm.push(patukek_obj.ParseBool(patukek_obj.IsTruthy(left) && patukek_obj.IsTruthy(right)))
}

func (vm *VM) execOr() error {
	var (
		right = patukek_obj.Unwrap(vm.pop())
		left  = patukek_obj.Unwrap(vm.pop())
	)

	return vm.push(patukek_obj.ParseBool(patukek_obj.IsTruthy(left) || patukek_obj.IsTruthy(right)))
}

func (vm *VM) execGreaterThan() error {
	var (
		right = patukek_obj.Unwrap(vm.pop())
		left  = patukek_obj.Unwrap(vm.pop())
	)

	switch {
	case patukek_obj.AssertTypes(left, patukek_obj.IntType) && patukek_obj.AssertTypes(right, patukek_obj.IntType):
		l := left.(patukek_obj.Integer)
		r := right.(patukek_obj.Integer)
		return vm.push(patukek_obj.ParseBool(l > r))

	case patukek_obj.AssertTypes(left, patukek_obj.StringType) && patukek_obj.AssertTypes(right, patukek_obj.StringType):
		l := left.(patukek_obj.String)
		r := right.(patukek_obj.String)
		return vm.push(patukek_obj.ParseBool(l > r))

	default:
		return vm.errorf("unsupported operator '>' for types %v and %v", left.Type(), right.Type())
	}
}

func (vm *VM) execGreaterThanEqual() error {
	var (
		right = patukek_obj.Unwrap(vm.pop())
		left  = patukek_obj.Unwrap(vm.pop())
	)

	switch {
	case patukek_obj.AssertTypes(left, patukek_obj.IntType) && patukek_obj.AssertTypes(right, patukek_obj.IntType):
		l := left.(patukek_obj.Integer)
		r := right.(patukek_obj.Integer)
		return vm.push(patukek_obj.ParseBool(l >= r))

	case patukek_obj.AssertTypes(left, patukek_obj.StringType) && patukek_obj.AssertTypes(right, patukek_obj.StringType):
		l := left.(patukek_obj.String)
		r := right.(patukek_obj.String)
		return vm.push(patukek_obj.ParseBool(l >= r))

	default:
		return vm.errorf("unsupported operator '>=' for types %v and %v", left.Type(), right.Type())
	}
}

func (vm *VM) execReturnValue() error {
	retVal := patukek_obj.Unwrap(vm.pop())
	frame := vm.popFrame()
	vm.sp = frame.basePointer - 1

	return vm.push(retVal)
}

func (vm *VM) call(o patukek_obj.Object, numArgs int) error {
	switch fn := patukek_obj.Unwrap(o).(type) {
	case *patukek_obj.Closure:
		return vm.callClosure(fn, numArgs)
	case patukek_obj.Builtin:
		return vm.callBuiltin(fn, numArgs)
	default:
		return vm.errorf("calling non-function")
	}
}

func (vm *VM) execCall(numArgs int) error {
	return vm.call(vm.stack[vm.sp-1-numArgs], numArgs)
}

func (vm *VM) buildList(start, end int) patukek_obj.Object {
	var elements = make([]patukek_obj.Object, end-start)

	for i := start; i < end; i++ {
		elements[i-start] = vm.stack[i]
	}
	return patukek_obj.NewList(elements...)
}

func (vm *VM) callClosure(cl *patukek_obj.Closure, nargs int) error {
	if nargs != cl.Fn.NumParams {
		return vm.errorf("wrong number of arguments: expected %d, got %d", cl.Fn.NumParams, nargs)
	}

	frame := NewFrame(cl, vm.sp-nargs)
	vm.pushFrame(frame)
	vm.sp = frame.basePointer + cl.Fn.NumLocals
	return nil
}

func (vm *VM) callBuiltin(fn patukek_obj.Builtin, nargs int) error {
	args := vm.stack[vm.sp-nargs : vm.sp]
	res := fn(args...)
	vm.sp = vm.sp - nargs - 1

	if res == nil {
		return vm.push(Null)
	}
	return vm.push(res)
}

func (vm *VM) pushClosure(constIdx, numFree int) error {
	constant := vm.Consts[constIdx]
	fn, ok := constant.(*patukek_obj.CompiledFunction)
	if !ok {
		return vm.errorf("not a function: %+v", constant)
	}

	free := make([]patukek_obj.Object, numFree)
	for i := 0; i < numFree; i++ {
		free[i] = vm.stack[vm.sp-numFree+i]
	}
	vm.sp = vm.sp - numFree
	return vm.push(&patukek_obj.Closure{Fn: fn, Free: free})
}

func (vm *VM) Run() (err error) {
	var (
		ip  int
		ins patukek_code.Instructions
		op  patukek_code.Opcode
	)

	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 && err == nil {
		vm.currentFrame().ip++

		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()
		op = patukek_code.Opcode(ins[ip])

		switch op {
		case patukek_code.OpConstant:
			constIndex := patukek_code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			err = vm.push(vm.Consts[constIndex])

		case patukek_code.OpJump:
			pos := int(patukek_code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip = pos - 1

		case patukek_code.OpJumpNotTruthy:
			pos := int(patukek_code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			if cond := patukek_obj.Unwrap(vm.pop()); !patukek_obj.IsTruthy(cond) {
				vm.currentFrame().ip = pos - 1
			}

		case patukek_code.OpSetGlobal:
			globalIndex := patukek_code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			vm.localTable[globalIndex] = true
			vm.Globals[globalIndex] = vm.peek()

		case patukek_code.OpGetGlobal:
			globalIndex := patukek_code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			err = vm.push(vm.Globals[globalIndex])

		case patukek_code.OpGetLocal:
			localIndex := patukek_code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()
			err = vm.push(vm.stack[frame.basePointer+int(localIndex)])

		case patukek_code.OpList:
			nElements := int(patukek_code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			list := vm.buildList(vm.sp-nElements, vm.sp)
			vm.sp = vm.sp - nElements
			err = vm.push(list)

		case patukek_code.OpCall:
			numArgs := patukek_code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1
			err = vm.execCall(int(numArgs))

		case patukek_code.OpGetBuiltin:
			idx := patukek_code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1
			def := patukek_obj.Builtins[idx]
			err = vm.push(def.Builtin)

		case patukek_code.OpClosure:
			constIdx := patukek_code.ReadUint16(ins[ip+1:])
			numFree := patukek_code.ReadUint8(ins[ip+3:])
			vm.currentFrame().ip += 3
			err = vm.pushClosure(int(constIdx), int(numFree))

		case patukek_code.OpReturnValue:
			err = vm.execReturnValue()

		case patukek_code.OpNull:
			err = vm.push(Null)

		case patukek_code.OpAdd:
			err = vm.execAdd()

		case patukek_code.OpSub:
			err = vm.execSub()

		case patukek_code.OpMul:
			err = vm.execMul()

		case patukek_code.OpDiv:
			err = vm.execDiv()

		case patukek_code.OpMod:
			err = vm.execMod()

		case patukek_code.OpEqual:
			err = vm.execEqual()

		case patukek_code.OpNotEqual:
			err = vm.execNotEqual()

		case patukek_code.OpGreaterThan:
			err = vm.execGreaterThan()

		case patukek_code.OpGreaterThanEqual:
			err = vm.execGreaterThanEqual()

		case patukek_code.OpAnd:
			err = vm.execAnd()

		case patukek_code.OpOr:
			err = vm.execOr()

		case patukek_code.OpPop:
			vm.pop()
		}

	}
	return
}

func (vm *VM) push(o patukek_obj.Object) error {
	if vm.sp >= StackSize {
		return vm.errorf("stack overflow")
	}

	vm.stack[vm.sp] = o
	vm.sp++
	return nil
}

func (vm *VM) pop() patukek_obj.Object {
	vm.sp--
	o := vm.stack[vm.sp]
	return o
}

func (vm *VM) peek() patukek_obj.Object {
	return vm.stack[vm.sp-1]
}