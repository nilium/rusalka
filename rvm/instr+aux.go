package rvm

import "fmt"

func mkBinaryInstr(op Opcode, out, argA RegisterIndex, argB Index) Instruction {
	var instr uint32 = uint32(op&0x1F) |
		uint32(out)&0xFF<<5 |
		uint32(argA)&0xFF<<13

	if out < 0 || out >= registerCount {
		panic(InvalidRegister(out))
	} else if argA < 0 || argA >= registerCount {
		panic(InvalidRegister(argA))
	}

	switch argB := argB.(type) {
	case nil:
	case StackIndex:
		if argB < -512 || argB > 511 {
			panic(InvalidStackIndex(argB))
		}
		instr |= uint32(int32(argB)<<23)>>1 | opBinArgBStack
	case RegisterIndex:
		if argB < 0 || argB >= registerCount {
			panic(InvalidRegister(argB))
		}
		instr |= uint32(argB&0xFF) << 22
	case constIndex:
		if argB < 0 || argB > 1023 {
			panic(InvalidConstIndex(argB))
		}
		instr |= (uint32(argB&0x3FF))<<22 | opBinArgBConst
	default:
		panic(fmt.Errorf("invalid index type %T; must be register, stack, or const", argB))
	}
	return Instruction(instr)
}
