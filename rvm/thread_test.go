package rvm

import "testing"

func TestOpAdd(t *testing.T) {
	th := NewThread()

	fn := funcData{
		consts: []Value{10.3},
	}
	th.pushFrame(0, fn)
	th.Push(5)

	instr := mkBinaryInstr(OpAdd, RegisterIndex(2), StackIndex(-1), constIndex(0))
	t.Log(instr)

	instr.execer()(instr, th)
	if got := th.At(RegisterIndex(2)); got != vnum(15.3) {
		t.Fatalf("th.At(RegisterIndex(2)) = %q; want %q", got, vnum(15.3))
	}
}
