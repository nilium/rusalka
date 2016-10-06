package rvm

import (
	"fmt"
	"strconv"
)

// Binary / basic instruction format:
// 0:4  5:12  13:20  21  22:31  | MASK     | DESCRIPTION
// |====|=====|======|===|==============================
// |    |     |      |   |      |          |
// |    |     |      |   +------| FFC00000 | ArgB (register(8); const(10); stack(9; +-))
// |    |     |      +----------| 00200000 | ArgB is constant lookup if set
// |    |     +-----------------| 001FE000 | ArgA (register(8); may be used as argAX or argAU for flags
// |    +-----------------------| 00001FE0 | Output register(8)
// +----------------------------| 0000001F | Opcode (universal; 0x1F reserved for extensions)

type Instruction uint64

const (
	opBinOutStack  Instruction = 0x40
	opBinArgAStack Instruction = 0x2000
	opBinArgBConst Instruction = 0x100000
	opBinArgBStack Instruction = 0x80000000

	opCmpArgAConst Instruction = 0x200
	opCmpArgAStack Instruction = 0x80000
	opCmpArgBConst Instruction = 0x100000
	opCmpArgBStack Instruction = 0x40000000
)

func (i Instruction) isExt() bool {
	return i&0x1 != 0
}

func (i Instruction) Opcode() Opcode {
	// If the instruction is extended... (probably have specialized instruction types at that point)
	// if i&0x1 != 0 {
	// }
	return Opcode((i >> 1) & 0x1F)
}

func (i Instruction) regOut() Index {
	if i&opBinOutStack != 0 {
		return StackIndex(int32(i<<19) >> 26)
	}
	return RegisterIndex((i >> 7) & 0x3F)
}

func (i Instruction) argA() Index {
	if i&0x2000 != 0 {
		return StackIndex(int32(i<<12) >> 26)
	}
	return RegisterIndex((i >> 14) & 0x3F)
}

func (i Instruction) argAX() int {
	return int(int32(i<<12) >> 26)
}

func (i Instruction) argAU() uint {
	return uint(i>>13) & 0x7F
}

func (i Instruction) argB() Index {
	ix := uint32(i >> 21)
	if i&opBinArgBConst != 0 {
		return constIndex(ix & 0x7FF)
	} else if i&opBinArgBStack != 0 {
		return StackIndex(int32(i<<1) >> 22)
	}
	return RegisterIndex(ix & 0x3F)
}

func (i Instruction) cmpOp() compareOp {
	return compareOp((i >> 6) & 0x7)
}

func (i Instruction) cmpArgA() Index {
	ix := uint32((i >> 10) & 0x3FF)
	if i&opCmpArgAConst != 0 {
		return constIndex(ix)
	} else if ix&0x200 != 0 {
		return StackIndex(int32(i<<13) >> 23)
	}
	return RegisterIndex(ix & 0x3F)
}

func (i Instruction) cmpArgB() Index {
	ix := uint32((i >> 21) & 0x3FF)
	if i&opCmpArgBConst != 0 {
		return constIndex(ix)
	} else if i&opCmpArgBStack != 0 {
		return StackIndex(int32(i<<2) >> 23)
	}
	return RegisterIndex(ix & 0x3F)
}

func (i Instruction) String() string {
	switch op := i.Opcode(); op {
	// Binary
	case OpAdd, OpSub, OpDiv, OpMul, OpPow, OpMod,
		OpOr, OpAnd, OpXor, OpArithshift, OpBitshift:
		return fmt.Sprint(op, i.regOut(), i.argA(), i.argB())
		// Unary
	case OpReserve:
		return fmt.Sprint(op, i.argB())
	case OpLoad:
		return fmt.Sprint(op, i.regOut(), i.argB())
	case OpPop:
		return fmt.Sprint(op, i.regOut())
	case OpPush:
		return fmt.Sprint(op, i.argB())
	case OpNeg, OpNot, OpRound, OpJump, OpDefer, OpJoin:
		// TODO: Fix per-unary string (e.g., load differs from neg)
		return fmt.Sprint(op, i.regOut(), i.argA(), i.argB())
	// Cond
	case OpTest:
		return fmt.Sprint(op, i.regOut(), i.argA(), i.argB())
	// Frame
	case OpCall, OpReturn:
		return fmt.Sprint(op, i.regOut(), i.argA(), i.argB())
	default:
		return "<unknown opcode for instruction " + strconv.FormatUint(uint64(i), 16) + ">"
	}
}

func (i Instruction) execer() opFunc {
	return opFuncTable[int(i>>1)&0x1F]
}

type compareOp uint

const (
	cmpLess compareOp = iota
	cmpLequal
	cmpEqual
	cmpNotEqual
	cmpGreater
	cmpGequal
	cmpIncludes
	cmpExcludes
)

type (
	LessComparator interface {
		LessThan(Value) bool
	}

	LessEqualComparator interface {
		LessEqual(Value) bool
	}

	EqualComparator interface {
		EqualTo(Value) bool
	}

	Comparable interface {
		LessComparator
		LessEqualComparator
		EqualComparator
	}
)

func lessThan(lhs, rhs Value) bool {
	switch lhs := lhs.(type) {
	case LessComparator:
		return lhs.LessThan(rhs)
	default:
		return false
	}
}

func lessEqual(lhs, rhs Value) bool {
	type lessEqualFallback interface {
		LessComparator
		EqualComparator
	}

	switch lhs := lhs.(type) {
	case LessEqualComparator:
		return lhs.LessEqual(rhs)
	case lessEqualFallback:
		return lhs.LessThan(rhs) || lhs.EqualTo(rhs)
	default:
		return false
	}
}

func equalTo(lhs, rhs Value) bool {
	switch lhs := lhs.(type) {
	case EqualComparator:
		return lhs.EqualTo(rhs)
	default:
		return false
	}
}

func (c compareOp) comparator() (result bool, fn func(lhs, rhs Value) bool) {
	switch c {
	case cmpLess:
		return true, lessThan
	case cmpLequal:
		return true, lessEqual
	case cmpEqual:
		return true, equalTo
	case cmpNotEqual:
		return false, equalTo
	case cmpGreater:
		return false, lessEqual
	case cmpGequal:
		return false, lessThan
	case cmpIncludes:
		fallthrough
	case cmpExcludes:
		fallthrough
	default:
		return false, func(Value, Value) bool { panic(fmt.Errorf("bad comparator op: %d", c)) }
	}
}
