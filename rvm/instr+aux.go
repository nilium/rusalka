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

func mkLoadInstr(dst, src Index) (instr uint32) {
	instr = uint32(OpLoad) << 1

	switch dst := dst.(type) {
	case RegisterIndex:
		if dst < 0 || dst >= registerCount {
			panic(InvalidRegister(dst))
		}
		instr |= uint32(dst&opRegMask) << 7
	case StackIndex:
		if dst < -64 || dst > 63 {
			panic(InvalidRegister(dst))
		}
		instr |= uint32(int32(dst<<25))>>18 | uint32(opLoadDstStack)
	default:
		panic(fmt.Errorf("invalid index type %T; must be register or stack", dst))
	}

	switch src := src.(type) {
	case RegisterIndex:
		if src < 0 || src >= registerCount {
			panic(InvalidRegister(src))
		}
		instr |= uint32(src&opRegMask) << 16
	case constIndex:
		if src < 0 || src > 65535 {
			panic(InvalidConstIndex(src))
		}
		instr |= uint32(src)<<16 | uint32(opLoadSrcConst)
	case StackIndex:
		if src < -32768 || src > 32767 {
			panic(InvalidStackIndex(src))
		}
		instr |= uint32(int32(src<<16)) | uint32(opLoadSrcStack)
	default:
		panic(fmt.Errorf("invalid index type %T; must be register, stack, or const", src))
	}

	return instr
}

func mkJumpInstr(offset int, src Index) (instr uint32) {
	if src != nil && offset != 0 {
		panic(fmt.Errorf("may not define an index (%v) and an offset (%d)", src, offset))
	}

	instr = uint32(OpJump) << 1

	if src == nil {
		if offset < -16777216 || offset > 16777215 {
			panic(fmt.Errorf("offset exceeds 25-bit range: %d", offset))
		}
		return instr | uint32(int32(offset<<7)) | uint32(opJumpLiteral)
	}

	switch src := src.(type) {
	case RegisterIndex:
		if src < 0 || src >= registerCount {
			panic(InvalidRegister(src))
		}
		instr |= uint32(src&opRegMask) << 8
	case constIndex:
		if src < 0 || src > 16777215 {
			panic(InvalidConstIndex(src))
		}
		instr |= uint32(src)<<8 | uint32(opJumpConst)
	case StackIndex:
		if src < -4194304 || src > 4194303 {
			panic(InvalidStackIndex(src))
		}
		instr |= uint32(int32(src<<9))>>1 | uint32(opJumpStack)
	default:
		panic(fmt.Errorf("invalid index type %T; must be register, stack, or const", src))
	}

	return instr
}

func mkXloadInstr(dst, src Index) (instr uint64) {
	instr = uint64(instrExtendedBit) |
		uint64(OpLoad)<<1

	switch dst := dst.(type) {
	case RegisterIndex:
		if dst < 0 || dst >= registerCount {
			panic(InvalidRegister(dst))
		}
		instr |= uint64(dst&opRegMask) << 14
	case StackIndex:
		if dst < -32768 || dst > 32767 {
			panic(InvalidRegister(dst))
		}
		instr |= uint64(int64(dst<<48))>>34 | uint64(opXloadDstStack)
	default:
		panic(fmt.Errorf("invalid index type %T; must be register or stack", dst))
	}

	switch src := src.(type) {
	case RegisterIndex:
		if src < 0 || src >= registerCount {
			panic(InvalidRegister(src))
		}
		instr |= uint64(src&opRegMask) << 32
	case constIndex:
		if src < 0 || src > 4294967295 {
			panic(InvalidConstIndex(src))
		}
		instr |= uint64(src)<<32 | uint64(opXloadSrcConst)
	case StackIndex:
		if src < -2147483648 || src > 2147483647 {
			panic(InvalidStackIndex(src))
		}
		instr |= uint64(int64(src<<32)) | uint64(opXloadSrcStack)
	default:
		panic(fmt.Errorf("invalid index type %T; must be register, stack, or const", src))
	}

	return instr
}

func mkBinaryInstr(op Opcode, out, argA, argB Index) (instr uint32) {
	instr = uint32(op&0x1F) << 1

loop:
	switch out := out.(type) {
	case nil:
		out = RegisterIndex(63)
		goto loop
	case RegisterIndex:
		if out < 0 || out >= registerCount {
			panic(InvalidRegister(out))
		}
		instr |= uint32(out&opRegMask) << 7
	case StackIndex:
		if out < -32 || out > 31 {
			panic(InvalidRegister(out))
		}
		instr |= uint32(int32(out<<26))>>19 | uint32(opBinOutStack)
	default:
		panic(fmt.Errorf("invalid index type %T; must be register or stack", out))
	}

	switch argA := argA.(type) {
	case nil:
	case RegisterIndex:
		if argA < 0 || argA >= registerCount {
			panic(InvalidRegister(argA))
		}
		instr |= uint32(argA&opRegMask) << 14
	case StackIndex:
		if argA < -32 || argA > 31 {
			panic(InvalidRegister(argA))
		}
		instr |= uint32(int32(argA<<26))>>12 | uint32(opBinArgAStack)
	default:
		panic(fmt.Errorf("invalid index type %T; must be register or stack", argA))
	}

	switch argB := argB.(type) {
	case nil:
	case RegisterIndex:
		if argB < 0 || argB >= registerCount {
			panic(InvalidRegister(argB))
		}
		instr |= uint32(argB&opRegMask) << 21
	case constIndex:
		if argB < 0 || argB > 2047 {
			panic(InvalidConstIndex(argB))
		}
		instr |= uint32(argB&0x7FF)<<21 | uint32(opBinArgBConst)
	case StackIndex:
		if argB < -512 || argB > 511 {
			panic(InvalidStackIndex(argB))
		}
		instr |= uint32(int32(argB<<22))>>1 | uint32(opBinArgBStack)
	default:
		panic(fmt.Errorf("invalid index type %T; must be register, stack, or const", argB))
	}

	return instr
}

func mkTestInstr(oper compareOp, want bool, argA, argB Index) (instr uint32) {
	instr = uint32(OpTest&0x1F)<<1 |
		uint32(oper&0x7)<<6

	if want {
		instr |= uint32(opCmpTestBit)
	}

	switch arg := argA.(type) {
	case nil:
	case RegisterIndex:
		if arg < 0 || arg >= registerCount {
			panic(InvalidRegister(arg))
		}
		instr |= uint32(arg&opRegMask) << 11
	case constIndex:
		if arg < 0 || arg > 1023 {
			panic(InvalidConstIndex(arg))
		}
		instr |= uint32(arg&0x3FF)<<11 | uint32(opCmpArgAConst)
	case StackIndex:
		if arg < -256 || arg > 255 {
			panic(InvalidStackIndex(arg))
		}
		instr |= uint32(int32(arg<<23))>>12 | uint32(opCmpArgAStack)
	default:
		panic(fmt.Errorf("invalid index type %T; must be register, stack, or const", arg))
	}

	switch arg := argB.(type) {
	case nil:
	case RegisterIndex:
		if arg < 0 || arg >= registerCount {
			panic(InvalidRegister(arg))
		}
		instr |= uint32(arg&opRegMask) << 22
	case constIndex:
		if arg < 0 || arg > 1023 {
			panic(InvalidConstIndex(arg))
		}
		instr |= uint32(arg&0x3FF)<<22 | uint32(opCmpArgBConst)
	case StackIndex:
		if arg < -256 || arg > 255 {
			panic(InvalidStackIndex(arg))
		}
		instr |= uint32(int32(arg<<23))>>1 | uint32(opCmpArgBStack)
	default:
		panic(fmt.Errorf("invalid index type %T; must be register, stack, or const", arg))
	}

	return instr
}

func mkPushPop(op Opcode, oprange int, arg Index) (instr uint32) {
	switch {
	case op != OpPush && op != OpPop:
		panic(fmt.Errorf("op is not push or pop: %v", op))
	case oprange < 1 || oprange > 64:
		panic(fmt.Errorf("invalid push/pop range: %d not in 1..64", oprange))
	}

	instr = uint32(op<<1)&opRegMask |
		(uint32(oprange-1)&0x3F)<<6

	switch arg := arg.(type) {
	case RegisterIndex:
		if arg < 0 || arg+RegisterIndex(oprange) > registerCount {
			panic(InvalidRegister(arg))
		}
		instr |= uint32(arg&opRegMask) << 14
	case StackIndex:
		if arg < -131072 || arg > 131071 {
			panic(InvalidStackIndex(arg))
		}
		instr |= uint32(int32(arg<<14)) | uint32(opPushPopStack)
	case constIndex:
		if arg < 0 || arg > 262143 {
			panic(InvalidConstIndex(arg))
		}
		instr |= uint32(arg&0x3FFFF)<<14 | uint32(opPushConst)
	default:
		req := "register, stack, or const"
		if op == OpPop {
			req = "register or stack"
		}
		panic(fmt.Errorf("invalid index type %T; must be %s", arg, req))
	}

	return instr
}
