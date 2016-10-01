package rvm

import "testing"

func TestOpAdd(t *testing.T) {
	th := NewThread()

	fn := funcData{
		code: []Instruction{
			// grow 4
			mkBinaryInstr(OpGrow, 0, nil, constIndex(1)),
			mkBinaryInstr(OpPush, 0, nil, constIndex(2)),
			mkBinaryInstr(OpShrink, 0, nil, constIndex(3)),
			// r[3] = 4
			mkBinaryInstr(OpLoad, RegisterIndex(31), nil, constIndex(1)),
			// r[2] = s[-3] + 10.3
			mkBinaryInstr(OpAdd, RegisterIndex(11), StackIndex(-3), constIndex(2)),
			// r[2] += s[3]
			mkBinaryInstr(OpAdd, RegisterIndex(11), RegisterIndex(11), StackIndex(3)),
			// r[2] += r[3]
			mkBinaryInstr(OpAdd, RegisterIndex(11), RegisterIndex(11), RegisterIndex(31)),
			// r[2] += 10.3
			mkBinaryInstr(OpAdd, RegisterIndex(11), RegisterIndex(11), constIndex(2)),
			// r[0] = r[2] - 4
			mkBinaryInstr(OpSub, RegisterIndex(4), RegisterIndex(11), constIndex(1)),
		},
		consts: []Value{vnum(0), vnum(4), vnum(10.3), vint(-1)},
	}

	th.pushFrame(0, fn)
	th.Push(643.219) // 0 -4
	th.Push(5)       // 1 -3
	th.Push(-123.45) // 2 -2
	th.Push(1)       // 3 -1

	testRunThread(t, th)
	testThreadState(t, th, []threadStateTest{
		{RegisterIndex(4), vnum(26.6)},
	})
}

func TestOpBitwiseShift(t *testing.T) {
	th := NewThread()

	fn := funcData{
		code: []Instruction{
			// r[3], r[6] = 1003, -1003
			mkBinaryInstr(OpLoad, RegisterIndex(3), nil, constIndex(0)),
			mkBinaryInstr(OpLoad, RegisterIndex(6), nil, constIndex(1)),
			mkBinaryInstr(OpBitshift, RegisterIndex(4), RegisterIndex(3), constIndex(2)),
			mkBinaryInstr(OpBitshift, RegisterIndex(5), RegisterIndex(3), constIndex(3)),
			mkBinaryInstr(OpBitshift, RegisterIndex(7), RegisterIndex(6), constIndex(2)),
			mkBinaryInstr(OpBitshift, RegisterIndex(8), RegisterIndex(6), constIndex(3)),
		},
		consts: []Value{vuint(1003), vnum(-1003), vnum(4), vnum(-4)},
	}

	th.pushFrame(0, fn)
	th.Push(4)
	th.Push(-4)

	testRunThread(t, th)
	testThreadState(t, th, []threadStateTest{
		{RegisterIndex(4), vuint(62)},
		{RegisterIndex(5), vuint(16048)},
		{RegisterIndex(7), vint(1152921504606846913)},
		{RegisterIndex(8), vint(-16048)},
	})
}

func TestOpArithShift(t *testing.T) {
	th := NewThread()

	fn := funcData{
		code: []Instruction{
			// r[3], r[6] = 1003, -1003
			mkBinaryInstr(OpLoad, RegisterIndex(3), nil, constIndex(0)),
			mkBinaryInstr(OpLoad, RegisterIndex(6), nil, constIndex(1)),
			mkBinaryInstr(OpArithshift, RegisterIndex(4), RegisterIndex(3), constIndex(2)),
			mkBinaryInstr(OpArithshift, RegisterIndex(5), RegisterIndex(3), constIndex(3)),
			mkBinaryInstr(OpArithshift, RegisterIndex(7), RegisterIndex(6), constIndex(2)),
			mkBinaryInstr(OpArithshift, RegisterIndex(8), RegisterIndex(6), constIndex(3)),
		},
		// Test with float64 for negative side just to ensure conversion works
		consts: []Value{vuint(1003), vnum(-1003), vnum(4), vnum(-4)},
	}

	th.pushFrame(0, fn)
	th.Push(4)
	th.Push(-4)

	testRunThread(t, th)
	testThreadState(t, th, []threadStateTest{
		{RegisterIndex(4), vuint(62)},
		{RegisterIndex(5), vuint(16048)},
		{RegisterIndex(7), vint(-63)},
		{RegisterIndex(8), vint(-16048)},
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
	for pc, instr := range th.code {
		t.Logf("%2d %v", pc, instr)
	}

	t.Log("Stack (before):")
	for i, e := range th.stack {
		t.Logf("%2d %#+v", i, e)
	}

	th.Run()

	t.Log("Stack (after):")
	for i, e := range th.stack {
		t.Logf("%2d %#+v", i, e)
	}
}
