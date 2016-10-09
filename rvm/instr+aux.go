package rvm

import "fmt"

type codeTable []uint32

func (c codeTable) binaryOp(op Opcode, out, argA, argB Index) codeTable {
	return append(c, mkBinaryInstr(op, out, argA, argB))
}

func (c codeTable) test(op compareOp, want bool, lhs, rhs Index) codeTable {
	return append(c, mkTestInstr(op, want, lhs, rhs))
}

func (c codeTable) load(dst, src Index) codeTable {
	return append(c, mkLoadInstr(dst, src))
}

func (c codeTable) xload(dst, src Index) codeTable {
	i := mkXloadInstr(dst, src)
	return append(c, uint32(i), uint32(i>>32))
}

func (c codeTable) jump(offset int, src Index) codeTable {
	return append(c, mkJumpInstr(offset, src))
}

func (c codeTable) push(sz int, src Index) codeTable {
	return append(c, mkPushPop(OpPush, sz, src))
}

func (c codeTable) pop(sz int, dst Index) codeTable {
	return append(c, mkPushPop(OpPop, sz, dst))
}

func (c codeTable) v() []uint32 {
	return []uint32(c)
}

func mkLoadInstr(dst, src Index) (instr uint32) {
	instr = opcodeBits(OpLoad)

	switch dst := dst.(type) {
	case RegisterIndex:
		instr |= registerOp(dst, opLoadDstOff)
	case StackIndex:
		if !canStore(int64(dst), opLoadDstLen) {
			panic(InvalidRegister(dst))
		}
		instr |= signedBits32(int32(dst), opLoadDstOff, opLoadDstLen) | uint32(opLoadDstStack)
	default:
		panic(fmt.Errorf("invalid index type %T; must be register or stack", dst))
	}

	switch src := src.(type) {
	case RegisterIndex:
		instr |= registerOp(src, opLoadSrcLen)
	case constIndex:
		if !canStoreUnsigned(uint64(src), opLoadSrcLen) {
			panic(InvalidConstIndex(src))
		}
		instr |= unsignedBits32(uint32(src), opLoadSrcOff, opLoadSrcLen) | uint32(opLoadSrcConst)
	case StackIndex:
		if !canStore(int64(src), opLoadSrcLen) {
			panic(InvalidStackIndex(src))
		}
		instr |= signedBits32(int32(src), opLoadSrcOff, opLoadSrcLen) | uint32(opLoadSrcStack)
	default:
		panic(fmt.Errorf("invalid index type %T; must be register, stack, or const", src))
	}

	return instr
}

func mkJumpInstr(offset int, src Index) (instr uint32) {
	if src != nil && offset != 0 {
		panic(fmt.Errorf("may not define an index (%v) and an offset (%d)", src, offset))
	}

	instr = opcodeBits(OpJump)

	if src == nil {
		if !canStore(int64(offset), opJumpLitLen) {
			panic(fmt.Errorf("offset exceeds 25-bit range: %d", offset))
		}
		return instr | signedBits32(int32(offset), opJumpLitOff, opJumpLitLen) | uint32(opJumpLiteral)
	}

	switch src := src.(type) {
	case RegisterIndex:
		instr |= registerOp(src, opJumpRelOff)
	case constIndex:
		if !canStoreUnsigned(uint64(src), opJumpRelLen) {
			panic(InvalidConstIndex(src))
		}
		instr |= unsignedBits32(uint32(src), opJumpRelOff, opJumpRelLen) | uint32(opJumpConst)
	case StackIndex:
		if !canStore(int64(src), opJumpStackLen) {
			panic(InvalidStackIndex(src))
		}
		instr |= signedBits32(int32(src), opJumpStackOff, opJumpStackLen) | uint32(opJumpStack)
	default:
		panic(fmt.Errorf("invalid index type %T; must be register, stack, or const", src))
	}

	return instr
}

func mkXloadInstr(dst, src Index) (instr uint64) {
	instr = uint64(instrExtendedBit) |
		xopcodeBits(OpLoad)

	switch dst := dst.(type) {
	case RegisterIndex:
		instr |= xregisterOp(dst, opXloadDstOff)
	case StackIndex:
		if !canStore(int64(dst), opXloadDstLen) {
			panic(InvalidRegister(dst))
		}
		instr |= signedBits64(int64(dst), opXloadDstOff, opXloadDstLen) | uint64(opXloadDstStack)
	default:
		panic(fmt.Errorf("invalid index type %T; must be register or stack", dst))
	}

	switch src := src.(type) {
	case RegisterIndex:
		instr |= xregisterOp(src, opXloadSrcLen)
	case constIndex:
		if !canStoreUnsigned(uint64(src), opXloadSrcLen) {
			panic(InvalidConstIndex(src))
		}
		instr |= unsignedBits64(uint64(src), opXloadSrcOff, opXloadSrcLen) | uint64(opXloadSrcConst)
	case StackIndex:
		if !canStore(int64(src), opXloadSrcLen) {
			panic(InvalidStackIndex(src))
		}
		instr |= signedBits64(int64(src), opXloadSrcOff, opXloadSrcLen) | uint64(opXloadSrcStack)
	default:
		panic(fmt.Errorf("invalid index type %T; must be register, stack, or const", src))
	}

	return instr
}

func mkBinaryInstr(op Opcode, out, argA, argB Index) (instr uint32) {
	instr = opcodeBits(op)

	switch out := out.(type) {
	case RegisterIndex:
		instr |= registerOp(out, opBinOutOff)
	case StackIndex:
		if !canStore(int64(out), opBinOutOff) {
			panic(InvalidRegister(out))
		}
		instr |= signedBits32(int32(out), opBinOutOff, opBinOutLen) | uint32(opBinOutStack)
	default:
		panic(fmt.Errorf("invalid index type %T; must be register or stack", out))
	}

	switch argA := argA.(type) {
	case RegisterIndex:
		instr |= registerOp(argA, opBinArgAOff)
	case StackIndex:
		if !canStore(int64(argA), opBinArgALen) {
			panic(InvalidRegister(argA))
		}
		instr |= signedBits32(int32(argA), opBinArgAOff, opBinArgALen) | uint32(opBinArgAStack)
	default:
		panic(fmt.Errorf("invalid index type %T; must be register or stack", argA))
	}

	switch argB := argB.(type) {
	case RegisterIndex:
		instr |= registerOp(argB, opBinArgBOff)
	case constIndex:
		if !canStoreUnsigned(uint64(argB), opBinArgBLen) {
			panic(InvalidConstIndex(argB))
		}
		instr |= unsignedBits32(uint32(argB), opBinArgBOff, opBinArgBLen) | uint32(opBinArgBConst)
	case StackIndex:
		if !canStore(int64(argB), opBinArgBLen) {
			panic(InvalidRegister(argB))
		}
		instr |= signedBits32(int32(argB), opBinArgBOff, opBinArgBStackLen) | uint32(opBinArgBStack)
	default:
		panic(fmt.Errorf("invalid index type %T; must be register, stack, or const", argB))
	}

	return instr
}

func mkTestInstr(oper compareOp, want bool, argA, argB Index) (instr uint32) {
	instr = opcodeBits(OpTest) |
		unsignedBits32(uint32(oper), opTestOperOff, opTestOperLen)

	if want {
		instr |= uint32(opCmpTestBit)
	}

	switch arg := argA.(type) {
	case RegisterIndex:
		instr |= registerOp(arg, opTestArgAOff)
	case constIndex:
		if !canStoreUnsigned(uint64(arg), opTestArgALen) {
			panic(InvalidConstIndex(arg))
		}
		instr |= unsignedBits32(uint32(arg), opTestArgAOff, opTestArgALen) | uint32(opCmpArgAConst)
	case StackIndex:
		if !canStore(int64(arg), opTestArgAStackLen) {
			panic(InvalidStackIndex(arg))
		}
		instr |= signedBits32(int32(arg), opTestArgAOff, opTestArgAStackLen) | uint32(opCmpArgAStack)
	default:
		panic(fmt.Errorf("invalid index type %T; must be register, stack, or const", arg))
	}

	switch arg := argB.(type) {
	case RegisterIndex:
		instr |= registerOp(arg, opTestArgBOff)
	case constIndex:
		if !canStoreUnsigned(uint64(arg), opTestArgBLen) {
			panic(InvalidConstIndex(arg))
		}
		instr |= unsignedBits32(uint32(arg), opTestArgBOff, opTestArgBLen) | uint32(opCmpArgBConst)
	case StackIndex:
		if !canStore(int64(arg), opTestArgBStackLen) {
			panic(InvalidStackIndex(arg))
		}
		instr |= signedBits32(int32(arg), opTestArgBOff, opTestArgBStackLen) | uint32(opCmpArgBStack)
	default:
		panic(fmt.Errorf("invalid index type %T; must be register, stack, or const", arg))
	}

	return instr
}

func mkPushPop(op Opcode, oprange int, arg Index) (instr uint32) {
	switch {
	case op != OpPush && op != OpPop:
		panic(fmt.Errorf("op is not push or pop: %v", op))
	case !canStoreUnsigned(uint64(oprange-1), opPushPopRangeLen):
		panic(fmt.Errorf("invalid push/pop range: %d not in 1..%d", oprange, (1 << opPushPopRangeLen)))
	}

	instr = opcodeBits(op) |
		unsignedBits32(uint32(oprange-1), opPushPopRangeOff, opPushPopRangeLen)

	switch arg := arg.(type) {
	case RegisterIndex:
		if arg+RegisterIndex(oprange) > registerCount {
			panic(InvalidRegister(arg))
		}
		instr |= registerOp(arg, opPushPopTargetOff)
	case StackIndex:
		if !canStore(int64(arg), opPushPopTargetLen) {
			panic(InvalidStackIndex(arg))
		}
		instr |= signedBits32(int32(arg), opPushPopTargetOff, opPushPopTargetLen) | uint32(opPushPopStack)
	case constIndex:
		if !canStoreUnsigned(uint64(arg), opPushPopTargetLen) {
			panic(InvalidConstIndex(arg))
		}
		instr |= unsignedBits32(uint32(arg), opPushPopTargetOff, opPushPopTargetLen) | uint32(opPushConst)
	default:
		req := "register, stack, or const"
		if op == OpPop {
			req = "register or stack"
		}
		panic(fmt.Errorf("invalid index type %T; must be %s", arg, req))
	}

	return instr
}

func opcodeBits(op Opcode) uint32 {
	return (uint32(op) & (1<<opBOpcodeLen - 1)) << opBOpcodeOff
}

func xopcodeBits(op Opcode) uint64 {
	return (uint64(op) & (1<<opXOpcodeLen - 1)) << opXOpcodeOff
}

func xregisterOp(r RegisterIndex, pos uint) uint64 {
	if r < 0 || r > registerCount {
		panic(InvalidRegister(r))
	}
	return uint64(r&opRegMask) << pos
}

func registerOp(r RegisterIndex, pos uint) uint32 {
	if r < 0 || r > registerCount {
		panic(InvalidRegister(r))
	}
	return uint32(r&opRegMask) << pos
}

func signedBits64(i int64, pos, length uint) uint64 {
	return (uint64(i<<(64-length)) >> (64 - (length + pos))) & ((1<<length - 1) << pos)
}

func signedBits32(i int32, pos, length uint) uint32 {
	return uint32(i<<(32-length)) >> (32 - (length + pos)) & ((1<<length - 1) << pos)
}

func unsignedBits64(i uint64, pos, length uint) uint64 {
	return uint64(i&(1<<length-1)) << pos
}

func unsignedBits32(i uint32, pos, length uint) uint32 {
	return uint32(i&(1<<length-1)) << pos
}

func canStore(i int64, bits uint) bool {
	bits-- // sign bit
	var (
		max int64 = 1<<bits - 1
		min int64 = -max - 1
	)
	return i >= min && i <= max
}

func canStoreUnsigned(i uint64, bits uint) bool {
	return i <= (^uint64(0) >> (64 - bits))
}
