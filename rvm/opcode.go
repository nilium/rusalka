package rvm

import "strconv"

type InvalidOpcode uint32

func (o InvalidOpcode) Error() string {
	return "invalid opcode " + strconv.FormatUint(uint64(o), 10)
}

type Opcode uint32

func (o Opcode) String() string {
	i := int(o)
	if i < 0 || i >= len(opNames) {
		return "INVALID"
	}
	return opNames[i]
}

const (
	opBinArgMask   uint32 = 0xE0000000
	opBinArgBConst uint32 = 0x1 << 29
	opBinArgBStack uint32 = 0x2 << 29
	opBinArgAStack uint32 = 0x4 << 29
)

const (
	OpAdd Opcode = iota
	OpSub
	OpDiv
	OpMul
	OpPow
	OpMod
	OpNeg
	OpNot
	OpOr
	OpAnd
	OpXor
	OpArithshift
	OpBitshift
	OpFloor
	OpCeil
	OpRound
	OpRint
	OpEq
	OpLe
	OpLt
	OpJump
	OpPush
	OpPop
	OpLoad
	OpCall
	OpReturn
	OpDefer
	OpFork
	OpJoin
)

var opNames = [...]string{
	OpAdd:        `add`,
	OpSub:        `sub`,
	OpDiv:        `div`,
	OpMul:        `mul`,
	OpPow:        `pow`,
	OpMod:        `mod`,
	OpNeg:        `neg`,
	OpNot:        `not`,
	OpOr:         `or`,
	OpAnd:        `and`,
	OpXor:        `xor`,
	OpArithshift: `ashift`,
	OpBitshift:   `bshift`,
	OpFloor:      `floor`,
	OpCeil:       `ceil`,
	OpRound:      `round`,
	OpRint:       `rint`,
	OpEq:         `eq`,
	OpLe:         `le`,
	OpLt:         `lt`,
	OpJump:       `jump`,
	OpPush:       `push`,
	OpPop:        `pop`,
	OpLoad:       `load`,
	OpCall:       `call`,
	OpReturn:     `return`,
	OpDefer:      `defer`,
	OpFork:       `fork`,
	OpJoin:       `join`,
}

type opFunc func(instr Instruction, vm *Thread)

var opFuncTable = [...]opFunc{

	OpAdd: func(instr Instruction, vm *Thread) {
		var (
			out = instr.regOut()
			lhs = toarith(instr.argA().load(vm))
			rhs = toarith(instr.argB().load(vm))
		)
		out.store(vm, lhs.add(rhs))
	},

	OpSub: func(instr Instruction, vm *Thread) {
		var (
			out = instr.regOut()
			lhs = toarith(instr.argA().load(vm))
			rhs = toarith(instr.argB().load(vm))
		)
		out.store(vm, lhs.add(rhs.neg()))
	},

	OpDiv: func(instr Instruction, vm *Thread) {
		var (
			out = instr.regOut()
			lhs = toarith(instr.argA().load(vm))
			rhs = toarith(instr.argB().load(vm))
		)
		out.store(vm, lhs.div(rhs))
	},

	OpMul: func(instr Instruction, vm *Thread) {
		var (
			out = instr.regOut()
			lhs = toarith(instr.argA().load(vm))
			rhs = toarith(instr.argB().load(vm))
		)
		out.store(vm, lhs.mul(rhs))
	},

	OpPow: func(instr Instruction, vm *Thread) {
		var (
			out = instr.regOut()
			lhs = toarith(instr.argA().load(vm))
			rhs = toarith(instr.argB().load(vm))
		)
		out.store(vm, lhs.pow(rhs))
	},

	OpMod: func(instr Instruction, vm *Thread) {
		var (
			out = instr.regOut()
			lhs = toarith(instr.argA().load(vm))
			rhs = toarith(instr.argB().load(vm))
		)
		out.store(vm, lhs.mod(rhs))
	},

	OpNeg: func(instr Instruction, vm *Thread) {
		var (
			out  = instr.regOut()
			recv = toarith(instr.argA().load(vm))
		)
		out.store(vm, recv.neg())
	},

	OpNot: func(instr Instruction, vm *Thread) {
		panic("unimplemented")
	},

	OpOr: func(instr Instruction, vm *Thread) {
		panic("unimplemented")
	},

	OpAnd: func(instr Instruction, vm *Thread) {
		panic("unimplemented")
	},

	OpXor: func(instr Instruction, vm *Thread) {
		panic("unimplemented")
	},

	OpArithshift: func(instr Instruction, vm *Thread) {
		panic("unimplemented")
	},

	OpBitshift: func(instr Instruction, vm *Thread) {
		panic("unimplemented")
	},

	OpFloor: func(instr Instruction, vm *Thread) {
		var (
			out = instr.regOut()
			val = toarith(instr.argA().load(vm))
		)
		if f, ok := val.(vnum); ok {
			val = vnum(round(float64(f), rndCeil))
		}
		out.store(vm, val)
	},

	OpCeil: func(instr Instruction, vm *Thread) {
		var (
			out = instr.regOut()
			val = toarith(instr.argA().load(vm))
		)
		if f, ok := val.(vnum); ok {
			val = vnum(round(float64(f), rndFloor))
		}
		out.store(vm, val)
	},

	OpRound: func(instr Instruction, vm *Thread) {
		var (
			out = instr.regOut()
			val = toarith(instr.argA().load(vm))
		)
		if f, ok := val.(vnum); ok {
			val = vnum(round(float64(f), rndNearest))
		}
		out.store(vm, val)
	},

	OpRint: func(instr Instruction, vm *Thread) {
		var (
			out = instr.regOut()
			val = toarith(instr.argA().load(vm))
		)
		if f, ok := val.(vnum); ok {
			val = vnum(round(float64(f), rndTrunc))
		}
		out.store(vm, val)
	},

	OpEq: func(instr Instruction, vm *Thread) {
		panic("unimplemented")
	},

	OpLe: func(instr Instruction, vm *Thread) {
		panic("unimplemented")
	},

	OpLt: func(instr Instruction, vm *Thread) {
		panic("unimplemented")
	},

	OpJump: func(instr Instruction, vm *Thread) {
		vm.pc += int64(tovint(instr.argB().load(vm)))
	},

	// push - - {reg|const|stack}
	OpPush: func(instr Instruction, vm *Thread) {
		vm.Push(instr.argB().load(vm))
	},

	// pop reg
	OpPop: func(instr Instruction, vm *Thread) {
		instr.regOut().store(vm, vm.Pop())
	},

	// load out - {reg|const|stack}
	OpLoad: func(instr Instruction, vm *Thread) {
		instr.regOut().store(vm, instr.argB().load(vm))
	},

	OpCall: func(instr Instruction, vm *Thread) {
		panic("unimplemented")
	},

	OpReturn: func(instr Instruction, vm *Thread) {
		panic("unimplemented")
	},

	OpDefer: func(instr Instruction, vm *Thread) {
		panic("unimplemented")
	},

	OpFork: func(instr Instruction, vm *Thread) {
		panic("unimplemented")
	},

	OpJoin: func(instr Instruction, vm *Thread) {
		panic("unimplemented")
	},
}
