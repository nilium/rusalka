package rvm

import "testing"

func TestOpAdd(t *testing.T) {
	th := NewThread()

	fn := funcData{
		code: []Instruction{
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
		consts: []Value{vnum(0), vnum(4), vnum(10.3)},
	}

	for pc, instr := range fn.code {
		t.Logf("%2d %v", pc, instr)
	}

	th.pushFrame(0, fn)
	th.Push(643.219) // 0 -4
	th.Push(5)       // 1 -3
	th.Push(-123.45) // 2 -2
	th.Push(1)       // 3 -1
	th.Run()

	want := vnum(26.6)
	if got := th.At(RegisterIndex(4)); got != want {
		t.Fatalf("th.At(RegisterIndex(0)) = %#v; want %#v", got, want)
	}

	th.popFrame(0)
}
