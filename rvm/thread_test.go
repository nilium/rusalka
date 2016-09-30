package rvm

import "testing"

func TestOpAdd(t *testing.T) {
	th := NewThread()

	fn := funcData{
		code: []Instruction{
			// r[3] = 4
			mkBinaryInstr(OpLoad, RegisterIndex(3), nil, constIndex(1)),
			// r[2] = s[-1] + 10.3
			mkBinaryInstr(OpAdd, RegisterIndex(2), StackIndex(-1), constIndex(2)),
			// r[2] += r[3]
			mkBinaryInstr(OpAdd, RegisterIndex(2), RegisterIndex(2), RegisterIndex(3)),
			// r[2] += 10.3
			mkBinaryInstr(OpAdd, RegisterIndex(2), RegisterIndex(2), constIndex(2)),
			// r[0] = r[2] - 4
			mkBinaryInstr(OpSub, RegisterIndex(0), RegisterIndex(2), constIndex(1)),
		},
		consts: []Value{vnum(0), vnum(4), vnum(10.3)},
	}
	th.pushFrame(0, fn)
	th.Push(5)
	th.Run()

	for pc, instr := range fn.code {
		t.Logf("%2d %v", pc, instr)
	}

	want := vnum(25.6)
	if got := th.At(RegisterIndex(0)); got != want {
		t.Fatalf("th.At(RegisterIndex(0)) = %#v; want %#v", got, want)
	}

	th.popFrame(0)
}
