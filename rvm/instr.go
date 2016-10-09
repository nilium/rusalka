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
	// Bits / masks for instruction flags

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

	opPushConst    Instruction = 0x1000
	opPushPopStack Instruction = 0x2000
)

const (
	// Offsets and bit lengths for instruction operands

	opBOpcodeOff = 1
	opBOpcodeLen = 5

	opXOpcodeOff = 1
	opXOpcodeLen = 12

	opBinOutOff       = 7
	opBinOutLen       = 6
	opBinArgAOff      = 14
	opBinArgALen      = 6
	opBinArgAXLen     = 7
	opBinArgBOff      = 21
	opBinArgBLen      = 11
	opBinArgBStackLen = 10

	opXloadDstOff = 14
	opXloadDstLen = 16
	opXloadSrcOff = 32
	opXloadSrcLen = 32

	opLoadDstOff = 7
	opLoadDstLen = 7
	opLoadSrcOff = 16
	opLoadSrcLen = 16

	opJumpLitOff   = 7
	opJumpLitLen   = 25
	opJumpRelOff   = opJumpLitOff + 1
	opJumpRelLen   = opJumpLitLen - 1
	opJumpStackOff = opJumpRelOff
	opJumpStackLen = opJumpRelLen - 1

	opTestOperOff      = 6
	opTestOperLen      = 3
	opTestArgAOff      = 11
	opTestArgALen      = 10
	opTestArgAStackLen = opTestArgALen - 1
	opTestArgBOff      = 22
	opTestArgBLen      = 10
	opTestArgBStackLen = opTestArgBLen - 1

	opPushPopRangeOff  = 6
	opPushPopRangeLen  = 6
	opPushPopTargetOff = 14
	opPushPopTargetLen = 18

	opBOpcodeMask       = (1<<opBOpcodeLen - 1) << opBOpcodeOff
	opXOpcodeMask       = (1<<opXOpcodeLen - 1) << opXOpcodeOff
	opBinOutMask        = (1<<opBinOutLen - 1) << opBinOutOff
	opBinArgAMask       = (1<<opBinArgALen - 1) << opBinArgAOff
	opBinArgAXMask      = (1<<opBinArgAXLen - 1) << opBinArgAOff
	opBinArgBMask       = (1<<opBinArgBLen - 1) << opBinArgBOff
	opBinArgBStackMask  = (1<<opBinArgBStackLen - 1) << opBinArgBOff
	opXloadDstMask      = (1<<opXloadDstLen - 1) << opXloadDstOff
	opXloadSrcMask      = (1<<opXloadSrcLen - 1) << opXloadSrcOff
	opLoadDstMask       = (1<<opLoadDstLen - 1) << opLoadDstOff
	opLoadSrcMask       = (1<<opLoadSrcLen - 1) << opLoadSrcOff
	opJumpLitMask       = (1<<opJumpLitLen - 1) << opJumpLitOff
	opJumpRelMask       = (1<<opJumpRelLen - 1) << opJumpRelOff
	opJumpStackMask     = (1<<opJumpStackLen - 1) << opJumpStackOff
	opTestOperMask      = (1<<opTestOperLen - 1) << opTestOperOff
	opTestArgAMask      = (1<<opTestArgALen - 1) << opTestArgAOff
	opTestArgAStackMask = (1<<opTestArgAStackLen - 1) << opTestArgAOff
	opTestArgBMask      = (1<<opTestArgBLen - 1) << opTestArgBOff
	opTestArgBStackMask = (1<<opTestArgBStackLen - 1) << opTestArgBOff
	opPushPopRangeMask  = (1<<opPushPopRangeLen - 1) << opPushPopRangeOff
	opPushPopTargetMask = (1<<opPushPopTargetLen - 1) << opPushPopTargetOff
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
	const l, r uint = 32 - (opBinOutOff + opBinOutLen), 32 - opBinOutLen
	if i&opBinOutStack != 0 {
		return StackIndex(int32(i<<l) >> r)
	}
	return RegisterIndex((i >> opBinOutOff) & opRegMask)
}

func (i Instruction) argA() Index {
	if i&opBinArgAStack != 0 {
		const l, r uint = 32 - (opBinArgAOff + opBinArgALen), 32 - opBinArgALen
		return StackIndex(int32((i&opBinArgAMask)<<l) >> r)
	}
	return RegisterIndex((i >> opBinArgAOff) & opRegMask)
}

func (i Instruction) argAX() int {
	const l, r uint = 32 - (opBinArgAOff + opBinArgAXLen), 32 - opBinArgAXLen
	return int(int32(i<<l) >> r)
}

func (i Instruction) argAU() uint {
	return uint(i&opBinArgAXMask) >> opBinArgAOff
}

func (i Instruction) argB() Index {
	ix := uint32(i >> opBinArgBOff)
	if i&opBinArgBConst != 0 {
		return constIndex((i & opBinArgBMask) >> opBinArgBOff)
	} else if i&opBinArgBStack != 0 {
		const l, r uint = 32 - (opBinArgBOff + opBinArgBStackLen), 32 - opBinArgBStackLen
		return StackIndex(int32((i&opBinArgBMask)<<l) >> r)
	}
	return RegisterIndex(ix & opRegMask)
}

func (i Instruction) pushPopRange() int {
	return 1 + int((i&opPushPopRangeMask)>>opPushPopRangeOff)
}

func (i Instruction) pushArg() Index {
	if i&opPushConst != 0 {
		return constIndex((i & opPushPopTargetMask) >> opPushPopTargetOff)
	} else if i&opPushPopStack != 0 {
		return StackIndex(int32(i&opPushPopTargetMask) >> opPushPopTargetOff)
	}

	return RegisterIndex(i>>opPushPopTargetOff) & opRegMask
}

func (i Instruction) popArg() Index {
	if i&opPushPopStack != 0 {
		return StackIndex(int32(i&opPushPopTargetMask) >> opPushPopTargetOff)
	}

	return RegisterIndex(i>>opPushPopTargetOff) & opRegMask
}

func (i Instruction) cmpOp() compareOp {
	return compareOp((i & opTestOperMask) >> opTestOperOff)
}

func (i Instruction) cmpWant() bool {
	return i&opCmpTestBit != 0
}

func (i Instruction) cmpArgA() Index {
	ix := uint32((i & opTestArgAMask) >> opTestArgAOff)
	if i&opCmpArgAConst != 0 {
		return constIndex(ix)
	} else if i&opCmpArgAStack != 0 {
		const l, r uint = 32 - (opTestArgAOff + opTestArgAStackLen), 32 - opTestArgAStackLen
		return StackIndex(int32(i<<l) >> r)
	}
	return RegisterIndex(ix & opRegMask)
}

func (i Instruction) cmpArgB() Index {
	ix := uint32((i & opTestArgBMask) >> opTestArgBOff)
	if i&opCmpArgBConst != 0 {
		return constIndex(ix)
	} else if i&opCmpArgBStack != 0 {
		const l, r uint = 32 - (opTestArgBOff + opTestArgBStackLen), 32 - opTestArgBStackLen
		return StackIndex(int32(i<<l) >> r)
	}
	return RegisterIndex(ix & opRegMask)
}

func (i Instruction) jumpOffset() (offset int64, index Index) {
	if i&opJumpLiteral != 0 {
		const l, r uint = 32 - (opJumpLitOff + opJumpLitLen), 32 - opJumpLitLen
		return int64(int32(i<<l) >> r), nil
	}

	if i&opJumpConst != 0 {
		return 0, constIndex((i & opJumpRelMask) >> opJumpRelOff)
	} else if i&opJumpStack != 0 {
		const l, r uint = 32 - (opJumpStackOff + opJumpStackLen), 32 - opJumpStackLen
		return 0, StackIndex(int32(i<<l) >> r)
	}

	return 0, RegisterIndex((i >> opJumpRelOff) & opRegMask)
}

func (i Instruction) loadDst() Index {
	var (
		stackF Instruction = opLoadDstStack
		stackL uint        = opLoadDstOff + opLoadDstLen
		stackR uint        = opLoadDstLen
		regR   uint        = opLoadDstOff
	)

	if i&instrExtendedBit != 0 {
		stackF = opXloadDstStack
		stackL, stackR = opXloadDstOff+opXloadDstLen, opXloadDstLen
		regR = opXloadDstOff
	}

	if i&stackF == 0 {
		return RegisterIndex(uint32(i>>regR) & opRegMask)
	}

	return StackIndex(int64(i<<(64-stackL)) >> (64 - stackR))
}

func (i Instruction) loadSrc() Index {
	var (
		stackF      = opLoadSrcStack
		constF      = opLoadSrcConst
		stackL uint = opLoadSrcOff + opLoadSrcLen
		stackR uint = opLoadSrcLen
		uiR    uint = opLoadSrcOff
	)

	if i&instrExtendedBit != 0 {
		stackF = opXloadSrcStack
		constF = opXloadSrcConst
		stackL, stackR = opXloadSrcOff+opXloadSrcLen, opXloadSrcLen
		uiR = opXloadSrcOff
	}

	if i&stackF != 0 {
		return StackIndex(int64(i<<(64-stackL)) >> (64 - stackR))
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
		return fmt.Sprint(xbit, op, i.pushPopRange(), i.popArg())
	case OpPush:
		return fmt.Sprint(xbit, op, i.pushPopRange(), i.pushArg())
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
		return "{bad-test-op: " + strconv.Itoa(int(c)) + "}"
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
