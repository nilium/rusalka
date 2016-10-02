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
	OpRound
	OpTest
	OpJump
	OpPush
	OpPop
	OpReserve
	OpLoad
	OpCall
	OpReturn
	OpDefer
	OpFork
	OpJoin
	opCount
)

const OpExtended Opcode = 0x3F

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
	OpRound:      `round`,
	OpTest:       `test`,
	OpJump:       `jump`,
	OpPush:       `push`,
	OpPop:        `pop`,
	OpReserve:    `reserve`,
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
		out.store(vm, lhs.Add(rhs))
	},

	OpSub: func(instr Instruction, vm *Thread) {
		var (
			out = instr.regOut()
			lhs = toarith(instr.argA().load(vm))
			rhs = toarith(instr.argB().load(vm))
		)
		out.store(vm, lhs.Sub(rhs))
	},

	OpDiv: func(instr Instruction, vm *Thread) {
		var (
			out = instr.regOut()
			lhs = toarith(instr.argA().load(vm))
			rhs = toarith(instr.argB().load(vm))
		)
		out.store(vm, lhs.Div(rhs))
	},

	OpMul: func(instr Instruction, vm *Thread) {
		var (
			out = instr.regOut()
			lhs = toarith(instr.argA().load(vm))
			rhs = toarith(instr.argB().load(vm))
		)
		out.store(vm, lhs.Mul(rhs))
	},

	OpPow: func(instr Instruction, vm *Thread) {
		var (
			out = instr.regOut()
			lhs = toarith(instr.argA().load(vm))
			rhs = toarith(instr.argB().load(vm))
		)
		out.store(vm, lhs.Pow(rhs))
	},

	OpMod: func(instr Instruction, vm *Thread) {
		var (
			out = instr.regOut()
			lhs = toarith(instr.argA().load(vm))
			rhs = toarith(instr.argB().load(vm))
		)
		out.store(vm, lhs.Mod(rhs))
	},

	OpNeg: func(instr Instruction, vm *Thread) {
		var (
			out  = instr.regOut()
			recv = toarith(instr.argA().load(vm))
		)
		out.store(vm, recv.Neg())
	},

	OpNot: func(instr Instruction, vm *Thread) {
		var (
			out  = instr.regOut()
			recv = tobitwise(instr.argA().load(vm))
		)
		out.store(vm, recv.Not())
	},

	OpOr: func(instr Instruction, vm *Thread) {
		var (
			out = instr.regOut()
			lhs = tobitwise(instr.argA().load(vm))
			rhs = tobitwise(instr.argB().load(vm))
		)
		out.store(vm, lhs.Or(rhs))
	},

	OpAnd: func(instr Instruction, vm *Thread) {
		var (
			out = instr.regOut()
			lhs = tobitwise(instr.argA().load(vm))
			rhs = tobitwise(instr.argB().load(vm))
		)
		out.store(vm, lhs.And(rhs))
	},

	OpXor: func(instr Instruction, vm *Thread) {
		var (
			out = instr.regOut()
			lhs = tobitwise(instr.argA().load(vm))
			rhs = tobitwise(instr.argB().load(vm))
		)
		out.store(vm, lhs.Xor(rhs))
	},

	OpArithshift: func(instr Instruction, vm *Thread) {
		var (
			out = instr.regOut()
			lhs = instr.argA().load(vm)
			rhs = instr.argB().load(vm)
		)

		out.store(vm, arithShift(lhs, rhs))
	},

	OpBitshift: func(instr Instruction, vm *Thread) {
		var (
			out = instr.regOut()
			lhs = instr.argA().load(vm)
			rhs = instr.argB().load(vm)
		)

		out.store(vm, bitwiseShift(lhs, rhs))
	},

	OpRound: func(instr Instruction, vm *Thread) {
		var (
			out  = instr.regOut()
			mode = roundingMode(instr.argAU())
			val  = round(instr.argA().load(vm), mode)
		)
		out.store(vm, val)
	},

	OpTest: func(instr Instruction, vm *Thread) {
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

	OpReserve: func(instr Instruction, vm *Thread) {
		sz := int(tovint(instr.argB().load(vm)))
		vm.growStack(sz)
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
