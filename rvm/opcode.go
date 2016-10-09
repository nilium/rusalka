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
			mode = RoundingMode(instr.argAU())
			val  = round(instr.argA().load(vm), mode)
		)
		out.store(vm, val)
	},

	OpTest: func(instr Instruction, vm *Thread) {
		var (
			op       = instr.cmpOp()
			want, fn = op.comparator()
			lhs      = instr.cmpArgA()
			rhs      = instr.cmpArgB()
		)

		if (fn(lhs, rhs) == want) != instr.cmpWant() {
			// test failed
			vm.step(true)
			return
		}

		// If the next instruction is a jump, execute it immediately
		if sz, ji, ok := vm.step(false); ok && ji.Opcode() == OpJump {
			if off, ix := ji.jumpOffset(); ix == nil {
				vm.pc += sz + off
			} else {
				vm.pc += sz + int64(toint(ix.load(vm)))
			}
		}
	},

	OpJump: func(instr Instruction, vm *Thread) {
		if off, ix := instr.jumpOffset(); ix == nil {
			vm.pc += off
		} else {
			vm.pc += int64(toint(ix.load(vm)))
		}
	},

	// push n src
	OpPush: func(instr Instruction, vm *Thread) {
		n := instr.pushPopRange()
		switch src := instr.pushArg().(type) {
		case StackIndex:
			var incr StackIndex = 1
			if src < 0 {
				incr = -1
			}

			for i, top := src, src+StackIndex(n)*incr; i != top; i = i + incr {
				vm.Push(i.load(vm))
			}
		case RegisterIndex:
			for i, top := src, src+RegisterIndex(n); i < top; i++ {
				vm.Push(i.load(vm))
			}
		case constIndex:
			for i, top := src, src+constIndex(n); i < top; i++ {
				vm.Push(i.load(vm))
			}
		}
	},

	// pop n dst
	OpPop: func(instr Instruction, vm *Thread) {
		n := instr.pushPopRange()
		switch src := instr.popArg().(type) {
		case StackIndex:
			if src < 0 {
				for i := src + StackIndex(n-1); i >= src; i-- {
					i.store(vm, vm.Pop())
				}
				return
			}

			for i := src - StackIndex(n-1); i <= src; i++ {
				i.store(vm, vm.Pop())
			}
		case RegisterIndex:
			for i := src + RegisterIndex(n-1); i >= src; i-- {
				i.store(vm, vm.Pop())
			}
		}
	},

	OpReserve: func(instr Instruction, vm *Thread) {
		sz := int(toint(instr.argB().load(vm)))
		vm.growStack(sz)
	},

	OpLoad: func(instr Instruction, vm *Thread) {
		instr.loadDst().store(vm, instr.loadSrc().load(vm))
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
