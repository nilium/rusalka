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

type Instruction uint32

const (
	opBinArgBConst uint32 = 0x00200000
	opBinArgBStack uint32 = 0x80000000
)

func (i Instruction) Opcode() Opcode {
	return Opcode(i & 0x1F)
}

func (i Instruction) regOut() RegisterIndex {
	return RegisterIndex((i >> 5) & 0xFF)
}

func (i Instruction) argA() Index {
	return RegisterIndex((i >> 13) & 0xFF)
}

func (i Instruction) argAX() int {
	return int(int32(i<<11) >> 24)
}

func (i Instruction) argAU() uint {
	return uint(i>>13) & 0xFF
}

func (i Instruction) argB() Index {
	ix := uint32(i) >> 22
	if uint32(i)&opBinArgBConst != 0 {
		return constIndex(ix)
	} else if ix&0x200 != 0 {
		return StackIndex(int32(i<<1) >> 23)
	}
	return RegisterIndex(ix & 0xFF)
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
	return opFuncTable[int(i&0x1F)]
}
