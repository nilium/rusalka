package rvm

import "fmt"

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
		instr |= uint32(out&0x3F) << 7
	case StackIndex:
		if out < -32 || out > 31 {
			panic(InvalidRegister(out))
		}
		instr |= uint32(int32(out<<27))>>19 | uint32(opBinOutStack)
	default:
		panic(fmt.Errorf("invalid index type %T; must be register or stack", out))
	}

	switch argA := argA.(type) {
	case nil:
	case RegisterIndex:
		if argA < 0 || argA >= registerCount {
			panic(InvalidRegister(argA))
		}
		instr |= uint32(argA&0x3F) << 14
	case StackIndex:
		if argA < -32 || argA > 31 {
			panic(InvalidRegister(argA))
		}
		instr |= uint32(int32(argA<<27))>>12 | uint32(opBinArgAStack)
	default:
		panic(fmt.Errorf("invalid index type %T; must be register or stack", argA))
	}

	switch argB := argB.(type) {
	case nil:
	case RegisterIndex:
		if argB < 0 || argB >= registerCount {
			panic(InvalidRegister(argB))
		}
		instr |= uint32(argB&0x3F) << 21
	case constIndex:
		if argB < 0 || argB > 2047 {
			panic(InvalidConstIndex(argB))
		}
		instr |= (uint32(argB&0x7FF))<<21 | uint32(opBinArgBConst)
	case StackIndex:
		if argB < -512 || argB > 511 {
			panic(InvalidStackIndex(argB))
		}
		instr |= uint32(int32(argB<<23))>>1 | uint32(opBinArgBStack)
	default:
		panic(fmt.Errorf("invalid index type %T; must be register, stack, or const", argB))
	}

	return instr
}

func mkTestInstr(oper compareOp, argA, argB Index) (instr uint32) {
	instr = uint32(OpTest&0x1F)<<1 |
		uint32(oper&0x7)<<6

	switch arg := argA.(type) {
	case nil:
	case RegisterIndex:
		if arg < 0 || arg >= registerCount {
			panic(InvalidRegister(arg))
		}
		instr |= uint32(arg&0x3F) << 10
	case constIndex:
		if arg < 0 || arg > 1023 {
			panic(InvalidConstIndex(arg))
		}
		instr |= (uint32(arg&0x3FF))<<10 | uint32(opCmpArgAConst)
	case StackIndex:
		if arg < -256 || arg > 255 {
			panic(InvalidStackIndex(arg))
		}
		instr |= uint32(int32(arg<<24))>>13 | uint32(opCmpArgAStack)
	default:
		panic(fmt.Errorf("invalid index type %T; must be register, stack, or const", arg))
	}

	switch arg := argB.(type) {
	case nil:
	case RegisterIndex:
		if arg < 0 || arg >= registerCount {
			panic(InvalidRegister(arg))
		}
		instr |= uint32(arg&0x3F) << 21
	case constIndex:
		if arg < 0 || arg > 1023 {
			panic(InvalidConstIndex(arg))
		}
		instr |= (uint32(arg&0x3FF))<<21 | uint32(opCmpArgBConst)
	case StackIndex:
		if arg < -256 || arg > 255 {
			panic(InvalidStackIndex(arg))
		}
		instr |= uint32(int32(arg<<24))>>2 | uint32(opCmpArgBStack)
	default:
		panic(fmt.Errorf("invalid index type %T; must be register, stack, or const", arg))
	}

	return instr
}
