package rvm

import (
	"errors"
	"fmt"
	"strconv"
)

// value is a general type for any type that can appear in the stack, registers, or constants table. It is currently the
// empty interface due to lack of specification around types while at least retaining concrete type information.
type Value interface{}

const (
	registerCount = 32
	registerMin   = -2147483648
	registerMax   = registerMin + registerCount

	maxInt = int(^uint(0) >> 1)
	minInt = -(maxInt - 1)
)

var (
	ErrRegisterRange = fmt.Errorf("register outside range 0..%d", registerCount-1)
	ErrPCRange       = fmt.Errorf("PC outside range 0..%d", maxInt)
	ErrStackRange    = errors.New("stack index out of range")
	ErrUnderflow     = errors.New("stack underflow")

	errConstStore = errors.New("attempt to write to constants table")
)

type funcData struct {
	// PC for the function
	pc   int64
	code []Instruction
	// constants that may be referenced by instructions
	consts []Value
	reg    [registerCount]Value

	// NOTE: Consider adding a constant page-shifting instruction to handle constants outside a [0, 2047] range.
}

type stackFrame struct {
	ebp int // starting ebp of this frame
	funcData
}

type Thread struct {
	stackFrame
	stack  []Value
	frames []stackFrame
}

// NewThread allocates a new VM thread.
func NewThread() *Thread {
	const defaultStackSize = 512
	const defaultFrameSize = 16

	th := &Thread{
		stack:  make([]Value, 0, defaultStackSize),
		frames: make([]stackFrame, 0, defaultFrameSize),
	}
	return th
}

// pushFrame pushes a new stack frame. This implicitly copies all registers for the new fn. ebpOffset may be <= 0; if
// less than 0, it can be used to mark a chunk from the top of the stack as belonging to the next frame.
func (th *Thread) pushFrame(ebpOffset int, fn funcData) {
	if ebpOffset > 0 {
		panic(InvalidStackIndex(len(th.stack) + ebpOffset))
	} else if th.ebp-ebpOffset > len(th.stack)+ebpOffset {
		panic(ErrUnderflow)
	}
	th.frames = append(th.frames, th.stackFrame)

	// Copy registers (may be used for argument passing)
	fn.reg = th.reg
	th.stackFrame = stackFrame{
		ebp:      len(th.stack) + ebpOffset,
		funcData: fn,
	}
}

func (th *Thread) replaceFrame(keep int, fn funcData) {
	th.copyAndResizeStack(th.ebp, keep)
	fn.reg = th.reg
	th.funcData = fn
}

func (th *Thread) popFrame(keep int) {
	top := len(th.frames) - 1
	if top < 0 {
		panic(ErrUnderflow)
	}

	frame := &th.frames[top]
	th.frames = th.frames[:top]
	th.copyAndResizeStack(th.ebp, keep)

	th.stackFrame = *frame
	*frame = stackFrame{}
}

// copyAndResizeStack resizes the stack to `newTop` plus `keep` elements from the top of the stack. The new stack top
// and the elements to keep may not overlap.
func (th *Thread) copyAndResizeStack(newTop, keep int) {
	if keep < 0 {
		panic(ErrUnderflow)
	}

	if newTop < 0 {
		panic(ErrUnderflow)
	} else if newTop+keep == len(th.stack) {
		return
	}

	// Take `keep` values off the top of the stack and move them down to the start.
	// We want to be able to use the stack to transfer values from a child frame to a parent frame, such as multiple
	// return values or other data that might be useful to a tailcall.
	if keep > 0 {
		oldTop := len(th.stack) - keep
		if newTop > oldTop {
			panic(ErrUnderflow)
		}

		copy(th.stack[newTop:], th.stack[oldTop:])
	}

	th.resizeStack(newTop + keep)
}

func (th *Thread) Run() {
	for th.pc < int64(len(th.code)) {
		pc := th.pc
		th.pc++
		instr := th.code[pc]
		instr.execer()(instr, th)
	}
}

func (th *Thread) Push(v Value) {
	th.stack = append(th.stack, v)
}

func (th *Thread) Pop() (v Value) {
	top := len(th.stack) - 1
	if top < 0 {
		panic(ErrUnderflow)
	}
	v = th.stack[top]
	th.resizeStack(top)
	return v
}

func (th *Thread) At(i Index) Value {
	if i == nil {
		panic("nil index")
	}
	return i.load(th)
}

// growStack grows the stack's capacity by at least elems entries. This does not resize the stack.
func (th *Thread) growStack(elems int) {
	var (
		pred = th.stack
		slen = len(pred)
		next = slen + elems
	)
	if next <= cap(th.stack) {
		return
	}

	dup := make([]Value, len(pred), next)
	copy(dup, th.stack)
	th.stack = dup

	for i := range pred {
		pred[i] = nil
	}
}

// resizeStack resizes the stack to the new top. If top is equal to or exceeds the current stack length, the call is
// a no-op.
func (th *Thread) resizeStack(top int) {
	curLen := len(th.stack)
	if curLen <= top {
		return
	}
	// Zero stack tail (optimized to mem zero)
	tail := th.stack[top:]
	for i := range tail {
		tail[i] = nil
	}
	// Truncate stack
	th.stack = th.stack[:top]
}

// Indices for accessing thread storage (registers, stack, constants, the PC)

type (
	Index interface {
		load(th *Thread) Value
		store(th *Thread, v Value)
	}

	pcIndex       struct{}
	StackIndex    int
	RegisterIndex int
	constIndex    int

	InvalidRegister   int
	InvalidStackIndex int
	InvalidConstIndex int
)

func (i InvalidRegister) Error() string {
	return fmt.Sprintf("register %d out of range 0..%d", i, registerCount-1)
}

func (i InvalidStackIndex) Error() string {
	return fmt.Sprintf("stack index %d out of range", i)
}

func (i InvalidConstIndex) Error() string {
	return fmt.Sprintf("constant index %d out of range", i)
}

func (i constIndex) String() string {
	return "const[" + strconv.Itoa(int(i)) + "]"
}

func (i constIndex) load(th *Thread) Value {
	return th.consts[int(i)]
}

func (constIndex) store(*Thread, Value) {
	panic(errConstStore)
}

func (i StackIndex) String() string {
	return "stack[" + strconv.Itoa(int(i)) + "]"
}

func (i StackIndex) abs(th *Thread) int {
	n := len(th.stack)
	ai := int(i)
	if ai < 0 {
		ai = n + ai
	}
	if ai < 0 || ai >= n {
		panic(InvalidStackIndex(i))
	}
	return ai
}

func (i StackIndex) Abs(th *Thread) StackIndex {
	return StackIndex(i.abs(th))
}

func (i StackIndex) load(th *Thread) Value {
	return th.stack[i.abs(th)]
}

func (i StackIndex) store(th *Thread, v Value) {
	th.stack[i.abs(th)] = v
}

func (i RegisterIndex) String() string {
	return "%" + strconv.Itoa(int(i))
}

func (i RegisterIndex) load(th *Thread) Value {
	if i < 0 || int(i) > len(th.reg) {
		panic(InvalidRegister(i))
	}
	return th.reg[int(i)]
}

func (i RegisterIndex) store(th *Thread, v Value) {
	if i < 0 || int(i) > len(th.reg) {
		panic(InvalidRegister(i))
	}
	th.reg[int(i)] = v
}

func (pcIndex) load(th *Thread) Value {
	return th.pc
}

func (pcIndex) String() string {
	return "%pc"
}

func (pcIndex) store(th *Thread, v Value) {
	var next int64
	switch i := v.(type) {
	case int:
		next = int64(i)
	case int32:
		next = int64(i)
	case int64:
		next = i
	case uint32:
		next = int64(i)
	case uint64:
		if i > uint64(maxInt) {
			panic(ErrPCRange)
		}
		next = int64(i)
	case uint:
		if uint64(i) > uint64(maxInt) {
			panic(ErrPCRange)
		}
		next = int64(i)
	default:
		panic(fmt.Errorf("invalid type for PC: %T, expected %T", v, int(0)))
	}

	if next < 0 {
		panic(ErrPCRange)
	}
	th.pc = next
}

var PC pcIndex
