package rvm

import (
	"fmt"
	"testing"
)

func ExampleInstruction_String() {
	fmt.Println(Instruction(mkTestInstr(cmpLess, true, RegisterIndex(5), constIndex(1023))))
	fmt.Println(Instruction(mkTestInstr(cmpLequal, true, RegisterIndex(5), constIndex(10))))
	fmt.Println(Instruction(mkTestInstr(cmpEqual, true, RegisterIndex(5), constIndex(10))))
	fmt.Println(Instruction(mkTestInstr(cmpNotEqual, true, RegisterIndex(5), constIndex(10))))
	fmt.Println(Instruction(mkTestInstr(cmpGreater, true, RegisterIndex(5), constIndex(10))))
	fmt.Println(Instruction(mkTestInstr(cmpGequal, true, StackIndex(233), StackIndex(-233))))
	fmt.Println(Instruction(mkTestInstr(cmpIncludes, true, StackIndex(255), StackIndex(-254))))
	fmt.Println(Instruction(mkTestInstr(cmpExcludes, true, RegisterIndex(5), constIndex(10))))
	fmt.Println(Instruction(mkLoadInstr(StackIndex(-64), constIndex(65535))))
	fmt.Println(Instruction(mkLoadInstr(RegisterIndex(63), StackIndex(-32768))))
	fmt.Println(Instruction(mkLoadInstr(RegisterIndex(2), RegisterIndex(1))))
	fmt.Println(Instruction(mkXloadInstr(StackIndex(-32768), constIndex(4294967295))))
	fmt.Println(Instruction(mkXloadInstr(RegisterIndex(63), StackIndex(-2147483648))))
	fmt.Println(Instruction(mkXloadInstr(RegisterIndex(2), RegisterIndex(1))))
	fmt.Println(Instruction(mkJumpInstr(50, nil)))
	fmt.Println(Instruction(mkJumpInstr(-16777216, nil)))
	fmt.Println(Instruction(mkJumpInstr(16777215, nil)))
	fmt.Println(Instruction(mkJumpInstr(0, RegisterIndex(63))))
	fmt.Println(Instruction(mkJumpInstr(0, RegisterIndex(0))))
	fmt.Println(Instruction(mkJumpInstr(0, StackIndex(-4194304))))
	fmt.Println(Instruction(mkJumpInstr(0, constIndex(16777215))))
	fmt.Println(Instruction(mkBinaryInstr(OpAdd, StackIndex(-32), StackIndex(-32), StackIndex(-512))))

	// Output:
	// test (%5 < const[1023]) == true
	// test (%5 <= const[10]) == true
	// test (%5 == const[10]) == true
	// test (%5 <> const[10]) == true
	// test (%5 > const[10]) == true
	// test (stack[233] >= stack[-233]) == true
	// test (stack[255] includes stack[-254]) == true
	// test (%5 excludes const[10]) == true
	// load stack[-64] const[65535]
	// load %63 stack[-32768]
	// load %esp %ebp
	// xload stack[-32768] const[4294967295]
	// xload %63 stack[-2147483648]
	// xload %esp %ebp
	// jump 50
	// jump -16777216
	// jump 16777215
	// jump %63
	// jump %pc
	// jump stack[-4194304]
	// jump const[16777215]
	// add stack[-32] stack[-32] stack[-512]
}

func TestOpAdd(t *testing.T) {
	th := NewThread()

	fn := funcData{
		code: []uint32{
			// grow 4
			mkBinaryInstr(OpReserve, RegisterIndex(0), RegisterIndex(0), constIndex(1)),
			mkBinaryInstr(OpPush, RegisterIndex(0), RegisterIndex(0), constIndex(2)),
			mkBinaryInstr(OpAdd, RegisterIndex(2), RegisterIndex(2), constIndex(3)),
			// r[3] = 4
			mkBinaryInstr(OpLoad, RegisterIndex(31), RegisterIndex(0), constIndex(1)),
			// r[3] = 4
			mkBinaryInstr(OpLoad, RegisterIndex(11), RegisterIndex(0), StackIndex(-3)),
			// r[2] = s[-3] + 10.3
			mkBinaryInstr(OpAdd, RegisterIndex(11), RegisterIndex(11), constIndex(2)),
			// r[2] += s[3]
			mkBinaryInstr(OpAdd, RegisterIndex(11), RegisterIndex(11), StackIndex(3)),
			// r[2] += r[3]
			mkBinaryInstr(OpAdd, RegisterIndex(11), RegisterIndex(11), RegisterIndex(31)),
			// r[2] += 10.3
			mkBinaryInstr(OpAdd, RegisterIndex(11), RegisterIndex(11), constIndex(2)),
			// r[0] = r[2] - 4
			mkBinaryInstr(OpSub, RegisterIndex(4), RegisterIndex(11), constIndex(1)),
		},
		consts: []Value{Float(0), Float(4), Float(10.3), Int(-1)},
	}

	th.pushFrame(0, fn)
	th.Push(643.219) // 0 -4
	th.Push(5)       // 1 -3
	th.Push(-123.45) // 2 -2
	th.Push(1)       // 3 -1

	testRunThread(t, th)
	testThreadState(t, th, []threadStateTest{
		{RegisterIndex(4), Float(26.6)},
	})
}

func TestOpBitwiseShift(t *testing.T) {
	th := NewThread()

	fn := funcData{
		code: []uint32{
			// r[3], r[6] = 1003, -1003
			mkBinaryInstr(OpLoad, RegisterIndex(3), RegisterIndex(0), constIndex(0)),
			mkBinaryInstr(OpLoad, RegisterIndex(6), RegisterIndex(0), constIndex(1)),
			mkBinaryInstr(OpBitshift, RegisterIndex(4), RegisterIndex(3), constIndex(2)),
			mkBinaryInstr(OpBitshift, RegisterIndex(5), RegisterIndex(3), constIndex(3)),
			mkBinaryInstr(OpBitshift, RegisterIndex(7), RegisterIndex(6), constIndex(2)),
			mkBinaryInstr(OpBitshift, RegisterIndex(8), RegisterIndex(6), constIndex(3)),
		},
		consts: []Value{Uint(1003), Float(-1003), Float(4), Float(-4)},
	}

	th.pushFrame(0, fn)
	th.Push(4)
	th.Push(-4)

	testRunThread(t, th)
	testThreadState(t, th, []threadStateTest{
		{RegisterIndex(4), Uint(62)},
		{RegisterIndex(5), Uint(16048)},
		{RegisterIndex(7), Int(1152921504606846913)},
		{RegisterIndex(8), Int(-16048)},
	})
}

func TestOpArithShift(t *testing.T) {
	th := NewThread()

	fn := funcData{
		code: []uint32{
			// r[3], r[6] = 1003, -1003
			mkBinaryInstr(OpLoad, RegisterIndex(3), RegisterIndex(0), constIndex(0)),
			mkBinaryInstr(OpLoad, RegisterIndex(6), RegisterIndex(0), constIndex(1)),
			mkBinaryInstr(OpArithshift, RegisterIndex(4), RegisterIndex(3), constIndex(2)),
			mkBinaryInstr(OpArithshift, RegisterIndex(5), RegisterIndex(3), constIndex(3)),
			mkBinaryInstr(OpArithshift, RegisterIndex(7), RegisterIndex(6), constIndex(2)),
			mkBinaryInstr(OpArithshift, RegisterIndex(8), RegisterIndex(6), constIndex(3)),
		},
		// Test with float64 for negative side just to ensure conversion works
		consts: []Value{Uint(1003), Float(-1003), Float(4), Float(-4)},
	}

	th.pushFrame(0, fn)
	th.Push(4)
	th.Push(-4)

	testRunThread(t, th)
	testThreadState(t, th, []threadStateTest{
		{RegisterIndex(4), Uint(62)},
		{RegisterIndex(5), Uint(16048)},
		{RegisterIndex(7), Int(-63)},
		{RegisterIndex(8), Int(-16048)},
	})
}

type threadStateTest struct {
	index Index
	want  Value
}

func testThreadState(t *testing.T, th *Thread, tests []threadStateTest) {
	for i, test := range tests {
		if got := th.At(test.index); got != test.want {
			t.Errorf("(%d) th.At(%v) = %v (%T); want %#v (%T)", i+1, test.index, got, got, test.want, test.want)
		} else {
			t.Logf("(%d) th.At(%v) = %v (%T)", i+1, test.index, got, got)
		}
	}
}

func testRunThread(t *testing.T, th *Thread) {
	t.Log("Code:")
	for pc, i := 0, 0; i < len(th.code); pc, i = pc+1, i+1 {
		instr := Instruction(th.code[i])
		fx := "%08x"
		if pc < len(th.code) && instr.isExt() {
			i++
			instr |= Instruction(th.code[i]) << 32
			fx = "%016x"
		}
		t.Logf("%-6d %-6d %-48v ["+fx+"]", pc, i, instr, uint64(instr))
	}

	t.Log("Stack (before):")
	for i, e := range th.stack {
		t.Logf("%2d %#+v", i, e)
	}

	done := false
	defer func() {
		pc := th.pc
		if done || pc <= 0 {
			return
		}

		rc := recover()
		if rc != nil {
			instr := Instruction(th.code[pc-1])
			fx := "%08x"
			if pc > 1 && Instruction(th.code[pc-2]).isExt() {
				instr = instr<<32 | Instruction(th.code[pc-2])
				pc--
				fx = "%016x"
			}
			t.Logf("last instruction: %d %-48v ["+fx+"]", pc, instr, uint64(instr))
		}
		panic(rc)
	}()
	th.Run()
	done = true

	t.Log("Stack (after):")
	for i, e := range th.stack {
		t.Logf("%2d %#+v", i, e)
	}
}
