package rvm

import "fmt"

func mkBinaryInstr(op Opcode, out RegisterIndex, argA, argB Index) Instruction {
	var instr uint32 = uint32(op&0x7F) |
		uint32(out&0x1F)<<8

	switch argA := argA.(type) {
	case nil:
	case StackIndex:
		if argA < -16 || argA > 15 {
			panic(InvalidStackIndex(argA))
		}
		instr |= (uint32(int32(argA)<<27)&0xF8000000)>>14 | opBinArgAStack
	case RegisterIndex:
		instr |= (uint32(argA) & 0x1F) << 13
	default:
		panic(fmt.Errorf("invalid index type %T; must be register or stack", argB))
	}

	switch argB := argB.(type) {
	case nil:
	case StackIndex:
		if argB < -1024 || argB > 1023 {
			panic(InvalidStackIndex(argB))
		}
		instr |= (uint32(int32(argB)<<21)&0xFFE00000)>>3 | opBinArgBStack
	case RegisterIndex:
		instr |= uint32(argB&0x1F) << 18
	case constIndex:
		if argB < 0 || (argB & ^0x7FF) != 0 {
			panic(InvalidConstIndex(argB))
		}
		instr |= uint32(argB&0x7FF)<<18 | opBinArgBConst
	default:
		panic(fmt.Errorf("invalid index type %T; must be register, stack, or const", argB))
	}
	return Instruction(instr)
}
