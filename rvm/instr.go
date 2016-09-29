package rvm

import (
	"fmt"
	"strconv"
)

type Instruction uint32

func (i Instruction) Opcode() Opcode {
	return Opcode(i & 0x7F)
}

// Basic operator instruction
// 0   7   8   13  18  29   MASK     - DESCRIPTION
// |   |   |   |   |   |
// |   |   |   |   |   +--- E0000000 - Flag bits (3): 1 - input is const; 2 - argB is stack; 3 - argA is stack.
// |   |   |   |   +------- 1FFC0000 - Input operand (11)
// |   |   |   +----------- 0003E000 - Input register / stack 0-31 (5) (relative stack if flag 3 set)
// |   |   +--------------- 00001F00 - Output register 0-31 (5)
// |   +------------------- 00000080 - Reserved (1)
// +----------------------- 0000007F - Opcode 0-127 (7)
//
// All instructions share the first 8 bits at a minimum, with the 8th bit reserved for signalling that an opcode is part
// an extended instruction set. How it'll work, in practice, isn't defined yet. Rusalka in C++ had room for user-defined
// instructions even if it didn't expose anything for working with them. rvm here does not yet have this sort of code
// built in, but having an extra bit to use may be useful.

func (i Instruction) regOut() RegisterIndex {
	return RegisterIndex((i >> 8) & 0x1F)
}

func (i Instruction) argA() Index {
	a := (i >> 13) & 0x1F
	if i.flags()&opBinArgAStack != 0 {
		return StackIndex(-int(a))
	}
	return RegisterIndex(a)
}

func (i Instruction) argB() Index {
	idx := uint16(i>>18) & 0x7FF
	flags := i.flags()
	switch {
	case flags&opBinArgBConst != 0:
		return constIndex(idx)
	case flags&opBinArgBStack != 0:
		return StackIndex(-int(idx))
	}
	return RegisterIndex(idx & 0x1F)
}

func (i Instruction) flags() uint32 {
	return uint32(i) & opBinArgMask
}

func (i Instruction) String() string {
	switch op := i.Opcode(); op {
	// Binary
	case OpAdd, OpSub, OpDiv, OpMul, OpPow, OpMod,
		OpOr, OpAnd, OpXor, OpArithshift, OpBitshift:
		return fmt.Sprint(op, i.regOut(), " ", i.argA(), " ", i.argB())
	// Unary
	case OpNeg, OpNot, OpFloor, OpCeil, OpRound, OpRint,
		OpJump, OpPush, OpPop, OpLoad, OpDefer, OpJoin:
		// TODO: Fix per-unary string (e.g., load differs from neg)
		return fmt.Sprint(op, i.regOut(), " ", i.argA(), " ", i.argB())
	// Cond
	case OpEq, OpLe, OpLt:
		return fmt.Sprint(op, i.regOut(), " ", i.argA(), " ", i.argB())
	// Frame
	case OpCall, OpReturn:
		return fmt.Sprint(op, i.regOut(), " ", i.argA(), " ", i.argB())
	default:
		return "unknown opcode for instruction " + strconv.FormatUint(uint64(i), 16)
	}
}

func (i Instruction) execer() opFunc {
	return opFuncTable[int(i&0x7F)]
}
