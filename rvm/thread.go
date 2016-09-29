package rvm

import (
	"errors"
	"fmt"
	"strconv"
)

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
	pc int64
	// constants that may be referenced by instructions
	consts []interface{}

	// NOTE: Consider adding a constant page-shifting instruction to handle constants outside a [0, 2047] range.
}

type stackFrame struct {
	returnTo int64 // continuation PC
	ebp      int   // starting ebp of this frame
	funcData
}

type Thread struct {
	pc     int64
	reg    [registerCount]interface{}
	stack  []interface{}
	frames []stackFrame
}

// NewThread allocates a new VM thread.
func NewThread() *Thread {
	const defaultStackSize = 512
	const defaultFrameSize = 16

	th := &Thread{
		stack:  make([]interface{}, 0, defaultStackSize),
		frames: make([]stackFrame, 0, defaultFrameSize),
	}
	return th
}

func (th *Thread) pushFrame(ebpOffset int, fn funcData) {
	th.frames = append(th.frames, stackFrame{
		returnTo: th.pc,
		ebp:      len(th.stack) + ebpOffset,
		funcData: fn,
	})
	// jump to call point
	th.pc = fn.pc
}

func (th *Thread) popFrame(keep int) {
	top := len(th.frames) - 1
	if top < 0 {
		panic(ErrUnderflow)
	}

	frame := &th.frames[top]
	th.pc, th.frames = frame.returnTo, th.frames[:top]
	if len(th.stack)-keep > frame.ebp {
		panic(ErrUnderflow)
	}
	copy(th.stack[frame.ebp:], th.stack[len(th.stack)-keep:])
	th.shrinkStack(frame.ebp + keep)
	*frame = stackFrame{}
}

func (th *Thread) Push(v interface{}) {
	th.stack = append(th.stack, v)
}

func (th *Thread) Pop() (v interface{}) {
	top := len(th.stack) - 1
	if top < 0 {
		panic(ErrUnderflow)
	}
	v = th.stack[top]
	th.shrinkStack(top)
	return v
}

func (th *Thread) At(i Index) interface{} {
	if i == nil {
		panic("nil index")
	}
	return i.load(th)
}

func (th *Thread) shrinkStack(top int) {
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
		load(th *Thread) interface{}
		store(th *Thread, v interface{})
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

func (i constIndex) load(th *Thread) interface{} {
	return th.frames[len(th.frames)-1].consts[int(i)]
}

func (constIndex) store(*Thread, interface{}) {
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

func (i StackIndex) load(th *Thread) interface{} {
	return th.stack[i.abs(th)]
}

func (i StackIndex) store(th *Thread, v interface{}) {
	th.stack[i.abs(th)] = v
}

func (i RegisterIndex) String() string {
	return "%" + strconv.Itoa(int(i))
}

func (i RegisterIndex) load(th *Thread) interface{} {
	if i < 0 || int(i) > len(th.reg) {
		panic(InvalidRegister(i))
	}
	return th.reg[int(i)]
}

func (i RegisterIndex) store(th *Thread, v interface{}) {
	if i < 0 || int(i) > len(th.reg) {
		panic(InvalidRegister(i))
	}
	th.reg[int(i)] = v
}

func (pcIndex) load(th *Thread) interface{} {
	return th.pc
}

func (pcIndex) String() string {
	return "%pc"
}

func (pcIndex) store(th *Thread, v interface{}) {
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
