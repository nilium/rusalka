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
	opRegMask = 0x3F

	instrOpMask Instruction = 0x1F << 1

	instrExtendedBit    Instruction = 0x1
	instrExtendedOpMask Instruction = 0xFFF << 1

	opBinOutStack  Instruction = 0x40
	opBinArgAStack Instruction = 0x2000
	opBinArgBConst Instruction = 0x100000
	opBinArgBStack Instruction = 0x80000000

	opCmpTestBit   Instruction = 0x200
	opCmpArgAConst Instruction = 0x400
	opCmpArgAStack Instruction = 0x100000
	opCmpArgBConst Instruction = 0x200000
	opCmpArgBStack Instruction = 0x80000000

	opJumpLiteral Instruction = 0x40 // In-range non-variable jumps
	opJumpConst   Instruction = 0x80 // For strange out of range jumps
	opJumpStack   Instruction = 0x80000000

	opLoadDstStack Instruction = 0x40
	opLoadSrcConst Instruction = 0x4000
	opLoadSrcStack Instruction = 0x8000

	opXloadDstStack Instruction = 0x2000
	opXloadSrcConst Instruction = 0x40000000
	opXloadSrcStack Instruction = 0x80000000
)

func (i Instruction) isExt() bool {
	return i&0x1 != 0
}

func (i Instruction) Opcode() Opcode {
	// +extended
	if i&instrExtendedBit != 0 {
		return Opcode((i & instrExtendedOpMask) >> 1)
	}
	return Opcode((i & instrOpMask) >> 1)
}

func (i Instruction) regOut() Index {
	if i&opBinOutStack != 0 {
		return StackIndex(int32(i<<19) >> 26)
	}
	return RegisterIndex((i >> 7) & opRegMask)
}

func (i Instruction) argA() Index {
	if i&opBinArgAStack != 0 {
		return StackIndex(int32(i<<12) >> 26)
	}
	return RegisterIndex((i >> 14) & opRegMask)
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
	return RegisterIndex(ix & opRegMask)
}

func (i Instruction) cmpOp() compareOp {
	return compareOp((i >> 6) & 0x7)
}

func (i Instruction) cmpWant() bool {
	return i&opCmpTestBit != 0
}

func (i Instruction) cmpArgA() Index {
	ix := uint32((i >> 11) & 0x3FF)
	if i&opCmpArgAConst != 0 {
		return constIndex(ix)
	} else if i&opCmpArgAStack != 0 {
		return StackIndex(int32(i<<12) >> 23)
	}
	return RegisterIndex(ix & opRegMask)
}

func (i Instruction) cmpArgB() Index {
	ix := uint32((i >> 22) & 0x3FF)
	if i&opCmpArgBConst != 0 {
		return constIndex(ix)
	} else if i&opCmpArgBStack != 0 {
		return StackIndex(int32(i<<1) >> 23)
	}
	return RegisterIndex(ix & opRegMask)
}

func (i Instruction) jumpOffset() (offset int64, index Index) {
	if i&opJumpLiteral != 0 {
		return int64(int32(i) >> 7), nil
	}

	if i&opJumpConst != 0 {
		return 0, constIndex((i >> 8) & 0xFFFFFF)
	} else if i&opJumpStack != 0 {
		return 0, StackIndex(int32(i<<1) >> 9)
	}

	return 0, RegisterIndex((i >> 8) & 0xFF)
}

func (i Instruction) loadDst() Index {
	var (
		stackF Instruction = 0x40
		stackL uint        = 50
		stackR uint        = 57
		regR   uint        = 7
	)

	if i&instrExtendedBit != 0 {
		stackF = 0x2000
		stackL, stackR = 34, 48
		regR = 14
	}

	if i&stackF == 0 {
		return RegisterIndex(uint32(i>>regR) & opRegMask)
	}

	return StackIndex(int64(i<<stackL) >> stackR)
}

func (i Instruction) loadSrc() Index {
	var (
		stackF Instruction = 0x8000
		stackL uint        = 32
		stackR uint        = 48
		constF Instruction = 0x4000
		uiR    uint        = 16
	)

	if i&instrExtendedBit != 0 {
		stackF = 0x80000000
		stackL, stackR = 0, 32
		constF = 0x40000000
		uiR = 32
	}

	if i&stackF != 0 {
		return StackIndex(int64(i<<stackL) >> stackR)
	} else if i&constF != 0 {
		return constIndex((i >> uiR))
	}
	return RegisterIndex((i >> uiR) & opRegMask)
}

func (i Instruction) String() string {
	xbit := ""
	if i.isExt() {
		xbit = "x"
	}

	switch op := i.Opcode(); op {
	// Binary
	case OpAdd, OpSub, OpDiv, OpMul, OpPow, OpMod,
		OpOr, OpAnd, OpXor, OpArithshift, OpBitshift:
		return fmt.Sprint(xbit, op, i.regOut(), i.argA(), i.argB())
		// Unary
	case OpReserve:
		return fmt.Sprint(xbit, op, i.argB())
	case OpLoad:
		return fmt.Sprint(xbit, op, i.loadDst(), i.loadSrc())
	case OpPop:
		return fmt.Sprint(xbit, op, i.regOut())
	case OpPush:
		return fmt.Sprint(xbit, op, i.argB())
	case OpNeg, OpNot, OpRound, OpDefer, OpJoin:
		// TODO: Fix per-unary string (e.g., load differs from neg)
		return fmt.Sprint(xbit, op, i.regOut(), i.argA(), i.argB())
	// Branch
	case OpJump:
		o, i := i.jumpOffset()
		if i == nil {
			return fmt.Sprint(xbit, op, o)
		}
		return fmt.Sprint(xbit, op, i)
	case OpTest:
		return fmt.Sprint(xbit, op, " (", i.cmpArgA(), i.cmpOp(), i.cmpArgB(), ") == ", i.cmpWant())
	// Frame
	case OpCall, OpReturn:
		return fmt.Sprint(xbit, op, i.regOut(), i.argA(), i.argB())
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

func (c compareOp) String() string {
	switch c {
	case cmpLess:
		return "<"
	case cmpLequal:
		return "<="
	case cmpEqual:
		return "=="
	case cmpNotEqual:
		return "<>"
	case cmpGreater:
		return ">"
	case cmpGequal:
		return ">="
	case cmpIncludes:
		return "includes"
	case cmpExcludes:
		return "excludes"
	default:
		return "{bad-test-op}"
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
