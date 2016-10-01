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
	OpGrow
	OpShrink
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
	OpGrow:       `grow`,
	OpShrink:     `shrink`,
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
		out.store(vm, lhs.Add(rhs.Neg()))
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
			out    = instr.regOut()
			lhs    = tobitwise(instr.argA().load(vm))
			lhsi   = int64(tovuint(lhs))
			rhs    = tovint(instr.argB().load(vm))
			result = lhsi
		)

		if rhs < 0 {
			result = lhsi << uint(-rhs)
		} else if rhs > 0 {
			result = lhsi >> uint(rhs)
		}

		var final Value
		switch lhs.(type) {
		case vint:
			final = vint(result)
		case vuint:
			final = vuint(result)
		default:
			panic("unreachable")
		}

		out.store(vm, final)
	},

	OpBitshift: func(instr Instruction, vm *Thread) {
		var (
			out    = instr.regOut()
			lhs    = tobitwise(instr.argA().load(vm))
			lhsu   = uint64(tovuint(lhs))
			rhs    = tovint(instr.argB().load(vm))
			result = lhsu
		)

		if rhs < 0 {
			result = lhsu << uint(-rhs)
		} else if rhs > 0 {
			result = lhsu >> uint(rhs)
		}

		var final Value
		switch lhs.(type) {
		case vint:
			final = vint(result)
		case vuint:
			final = vuint(result)
		default:
			panic("unreachable")
		}

		out.store(vm, final)
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

	OpGrow: func(instr Instruction, vm *Thread) {
		sz := int(tovint(instr.argB().load(vm)))
		vm.growStack(sz)
	},

	OpShrink: func(instr Instruction, vm *Thread) {
		var (
			sz  = int(tovint(instr.argB().load(vm)))
			ebp = vm.ebp
			nsz int
		)
		if sz < 0 {
			nsz = len(vm.stack) + sz
		} else {
			nsz = ebp + sz
		}

		if nsz < ebp {
			panic(ErrUnderflow)
		}
		vm.resizeStack(nsz)
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
