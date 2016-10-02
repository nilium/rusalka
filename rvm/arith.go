package rvm

import (
	"fmt"
	"math"
)

type InvalidRoundingMode RoundingMode

func (i InvalidRoundingMode) Error() string {
	return fmt.Sprintf("invalid rounding mode: %x", i)
}

type RoundingMode uint

const (
	RoundTruncate RoundingMode = iota
	RoundNearest
	RoundFloor
	RoundCeil
)

func round(v Value, mode RoundingMode) Value {
	if mode > RoundCeil {
		panic("invalid rounding mode")
	}

loop:
	switch vx := v.(type) {
	case Uint, Int:
		return vx
	case Float:
		return vx.Round(mode)
	case Rounder:
		return vx.Round(mode)
	case float64:
		return Float(vx).Round(mode)
	default:
		v = toarith(vx)
		goto loop
	}
}

type (
	Float float64
	Int   int64
	Uint  uint64

	Arith interface {
		Add(Arith) Arith
		Sub(Arith) Arith
		Neg() Arith
		Mul(Arith) Arith
		Div(Arith) Arith
		Mod(Arith) Arith
		Pow(Arith) Arith
		Sqrt() Arith
	}

	Bitwise interface {
		Arith
		Xor(Bitwise) Bitwise
		And(Bitwise) Bitwise
		Or(Bitwise) Bitwise
		Not() Bitwise
	}

	ArithmeticShifter interface {
		ArithShift(bits Value) Value
	}

	BitShifter interface {
		BitShift(bits Value) Value
	}

	Rounder interface {
		Round(RoundingMode) Value
	}

	FloatValuer interface {
		Float64() float64
	}

	IntValuer interface {
		Int64() int64
	}

	UintValuer interface {
		Uint64() uint64
	}
)

var (
	_ Arith = Float(0)
	_ Arith = Int(0)
	_ Arith = Uint(0)

	_ Bitwise = Int(0)
	_ Bitwise = Uint(0)

	_ Rounder = Float(0)
	_ Rounder = Int(0)
	_ Rounder = Uint(0)
)

// Float64

func (lhs Float) Float64() float64    { return float64(lhs) }
func (lhs Float) Int64() int64        { return int64(lhs) }
func (lhs Float) Uint64() uint64      { return uint64(lhs) }
func (lhs Float) Add(rhs Arith) Arith { return lhs + tofloat(rhs) }
func (lhs Float) Sub(rhs Arith) Arith { return lhs - tofloat(rhs) }
func (lhs Float) Mul(rhs Arith) Arith { return lhs * tofloat(rhs) }
func (lhs Float) Div(rhs Arith) Arith { return lhs / tofloat(rhs) }
func (lhs Float) Neg() Arith          { return -lhs }
func (lhs Float) Sqrt() Arith         { return Float(math.Sqrt(float64(lhs))) }

func (lhs Float) Round(mode RoundingMode) Value {
	switch x := float64(lhs); mode {
	case RoundTruncate:
		return math.Trunc(x)
	case RoundNearest:
		return math.Trunc(x + math.Copysign(0.5, x))
	case RoundFloor:
		return math.Floor(x)
	case RoundCeil:
		return math.Ceil(x)
	}
	panic("unreachable")
}

func (lhs Float) Pow(rhs Arith) Arith {
	return Float(math.Pow(float64(lhs), float64(tofloat(rhs))))
}

func (lhs Float) Mod(rhs Arith) Arith {
	return Float(math.Mod(float64(lhs), float64(tofloat(rhs))))
}

// Signed integer

func (lhs Int) Float64() float64 { return float64(lhs) }
func (lhs Int) Int64() int64     { return int64(lhs) }
func (lhs Int) Uint64() uint64   { return uint64(lhs) }
func (lhs Int) Neg() Arith       { return -lhs }

func (lhs Int) Round(RoundingMode) Value { return lhs }

func (lhs Int) ArithShift(bits Value) Value {
	if bits := toint(bits); bits < 0 {
		return lhs << uint(-bits)
	} else if bits > 0 {
		return lhs >> uint(bits)
	}
	return lhs
}

func (lhs Int) BitShift(bits Value) Value {
	if bits := toint(bits); bits < 0 {
		return Int(uint64(lhs) << uint(-bits))
	} else if bits > 0 {
		return Int(uint64(lhs) >> uint(bits))
	}
	return lhs
}

func (lhs Int) Add(rhs Arith) Arith {
	switch rhs := toarith(rhs).(type) {
	case Int:
		return Int(int64(lhs) + int64(rhs))
	case Uint:
		return Int(int64(lhs) + int64(rhs))
	case Float:
		return Float(float64(lhs) + float64(rhs))
	}
	panic("unreachable")
}

func (lhs Int) Sub(rhs Arith) Arith {
	switch rhs := toarith(rhs).(type) {
	case Int:
		return Int(int64(lhs) - int64(rhs))
	case Uint:
		return Int(int64(lhs) - int64(rhs))
	case Float:
		return Float(float64(lhs) - float64(rhs))
	}
	panic("unreachable")
}

func (lhs Int) Mul(rhs Arith) Arith {
	switch rhs := toarith(rhs).(type) {
	case Int:
		return Int(int64(lhs) * int64(rhs))
	case Uint:
		return Int(int64(lhs) * int64(rhs))
	case Float:
		return Float(float64(lhs) * float64(rhs))
	}
	panic("unreachable")
}

func (lhs Int) Div(rhs Arith) Arith {
	switch rhs := toarith(rhs).(type) {
	case Int:
		return Int(int64(lhs) / int64(rhs))
	case Uint:
		return Int(int64(lhs) / int64(rhs))
	case Float:
		return Float(float64(lhs) / float64(rhs))
	}
	panic("unreachable")
}

func (lhs Int) Mod(rhs Arith) Arith {
	switch rhs := toarith(rhs).(type) {
	case Int:
		return Int(int64(lhs) % int64(rhs))
	case Uint:
		return Int(int64(lhs) % int64(rhs))
	case Float:
		return Float(math.Mod(float64(lhs), float64(rhs)))
	}
	panic("unreachable")
}

func (lhs Int) Sqrt() Arith { return Int(math.Sqrt(float64(lhs))) }

func (lhs Int) Pow(rhs Arith) Arith {
	switch rhs := toarith(rhs).(type) {
	case Int:
		if rhs == 0 {
			return Uint(1)
		} else if rhs < 0 {
			return Float(math.Pow(float64(lhs), float64(rhs)))
		}
		for q, i := lhs, Int(0); i < rhs; i++ {
			lhs = lhs * q
		}
		return lhs
	case Uint:
		if rhs == 0 {
			return Uint(1)
		}
		for q, i := lhs, Uint(0); i < rhs; i++ {
			lhs = lhs * q
		}
		return lhs
	case Float:
		return Float(math.Pow(float64(lhs), float64(rhs)))
	}
	panic("unreachable")
}

func (lhs Int) Xor(rhs Bitwise) Bitwise { return Int(uint64(lhs) ^ uint64(touint(rhs))) }
func (lhs Int) And(rhs Bitwise) Bitwise { return Int(uint64(lhs) & uint64(touint(rhs))) }
func (lhs Int) Or(rhs Bitwise) Bitwise  { return Int(uint64(lhs) | uint64(touint(rhs))) }
func (lhs Int) Not() Bitwise            { return Int(^uint64(lhs)) }

// Unsigned integer

func (lhs Uint) Float64() float64 { return float64(lhs) }
func (lhs Uint) Int64() int64     { return int64(lhs) }
func (lhs Uint) Uint64() uint64   { return uint64(lhs) }
func (lhs Uint) Neg() Arith       { return -lhs }

func (lhs Uint) Round(RoundingMode) Value { return lhs }

func (lhs Uint) ArithShift(bits Value) Value {
	if bits := toint(bits); bits < 0 {
		return Uint(int64(lhs) << uint(-bits))
	} else if bits > 0 {
		return Uint(int64(lhs) >> uint(bits))
	}
	return lhs
}

func (lhs Uint) BitShift(bits Value) Value {
	if bits := toint(bits); bits < 0 {
		return lhs << uint(-bits)
	} else if bits > 0 {
		return lhs >> uint(bits)
	}
	return lhs
}

func (lhs Uint) Add(rhs Arith) Arith {
	switch rhs := toarith(rhs).(type) {
	case Uint:
		return Uint(uint64(lhs) + uint64(rhs))
	case Int:
		return Uint(int64(lhs) + int64(rhs))
	case Float:
		return Float(float64(lhs) + float64(rhs))
	}
	panic("unreachable")
}

func (lhs Uint) Sub(rhs Arith) Arith {
	switch rhs := toarith(rhs).(type) {
	case Uint:
		return Uint(uint64(lhs) - uint64(rhs))
	case Int:
		return Uint(int64(lhs) - int64(rhs))
	case Float:
		return Float(float64(lhs) - float64(rhs))
	}
	panic("unreachable")
}

func (lhs Uint) Mul(rhs Arith) Arith {
	switch rhs := toarith(rhs).(type) {
	case Uint:
		return Int(uint64(lhs) * uint64(rhs))
	case Int:
		return Uint(int64(lhs) * int64(rhs))
	case Float:
		return Float(float64(lhs) * float64(rhs))
	}
	panic("unreachable")
}

func (lhs Uint) Div(rhs Arith) Arith {
	switch rhs := toarith(rhs).(type) {
	case Uint:
		return Int(uint64(lhs) / uint64(rhs))
	case Int:
		return Uint(int64(lhs) / int64(rhs))
	case Float:
		return Float(float64(lhs) / float64(rhs))
	}
	panic("unreachable")
}

func (lhs Uint) Mod(rhs Arith) Arith {
	switch rhs := toarith(rhs).(type) {
	case Uint:
		return Int(uint64(lhs) % uint64(rhs))
	case Int:
		return Uint(int64(lhs) % int64(rhs))
	case Float:
		return Float(math.Mod(float64(lhs), float64(rhs)))
	}
	panic("unreachable")
}

func (lhs Uint) Sqrt() Arith { return Uint(math.Sqrt(float64(lhs))) }

func (lhs Uint) Pow(rhs Arith) Arith {
	switch rhs := toarith(rhs).(type) {
	case Uint:
		if rhs == 0 {
			return Uint(1)
		}
		for q, i := lhs, Uint(0); i < rhs; i++ {
			lhs = lhs * q
		}
		return lhs
	case Int:
		if rhs == 0 {
			return Uint(1)
		} else if rhs < 0 {
			return Float(math.Pow(float64(lhs), float64(rhs)))
		}
		for q, i := lhs, Int(0); i < rhs; i++ {
			lhs = lhs * q
		}
		return lhs
	case Float:
		return Float(math.Pow(float64(lhs), float64(rhs)))
	}
	panic("unreachable")
}

func (lhs Uint) Xor(rhs Bitwise) Bitwise { return lhs ^ touint(rhs) }
func (lhs Uint) And(rhs Bitwise) Bitwise { return lhs & touint(rhs) }
func (lhs Uint) Or(rhs Bitwise) Bitwise  { return lhs | touint(rhs) }
func (lhs Uint) Not() Bitwise            { return ^lhs }

func toarith(v Value) (r Arith) {
	switch v := v.(type) {
	case Arith:
		return v
	case FloatValuer:
		return Float(v.Float64())
	case IntValuer:
		return Int(v.Int64())
	case UintValuer:
		return Uint(v.Uint64())
	case int:
		return Int(v)
	case int64:
		return Int(v)
	case int32:
		return Int(v)
	case int16:
		return Int(v)
	case float64:
		return Float(v)
	case float32:
		return Float(v)
	case uint:
		return Uint(v)
	case uint64:
		return Uint(v)
	case uint32:
		return Uint(v)
	case uint16:
		return Uint(v)
	case uint8:
		return Uint(v)
	default:
		panic(fmt.Errorf("unable to convert %T to arithmetic type", v))
	}
}

func tobitwise(v Value) (r Bitwise) {
	switch v := v.(type) {
	case Bitwise:
		return v
	case IntValuer:
		return Int(v.Int64())
	case UintValuer:
		return Uint(v.Uint64())
	case float64:
		return Int(v)
	case float32:
		return Int(v)
	case int:
		return Int(v)
	case int64:
		return Int(v)
	case int32:
		return Int(v)
	case int16:
		return Int(v)
	case uint:
		return Uint(v)
	case uint64:
		return Uint(v)
	case uint32:
		return Uint(v)
	case uint16:
		return Uint(v)
	case uint8:
		return Uint(v)
	default:
		panic(fmt.Errorf("unable to convert %T to bitwise type", v))
	}
}

func tofloat(v Value) Float {
	switch v := toarith(v).(type) {
	case Float:
		return v
	case Int:
		return Float(v)
	case Uint:
		return Float(v)
	}
	panic("unreachable")
}

func toint(v Value) Int {
	switch v := toarith(v).(type) {
	case Int:
		return v
	case Float:
		return Int(v)
	case Uint:
		return Int(v)
	}
	panic("unreachable")
}

func touint(v Value) Uint {
	switch v := toarith(v).(type) {
	case Int:
		return Uint(v)
	case Float:
		return Uint(v)
	case Uint:
		return v
	}
	panic("unreachable")
}

func arithShift(v, bits Value) Value {
	var (
		ov  = v
		try bool
	)
loop:
	switch vx := v.(type) {
	case Uint:
		return vx.ArithShift(bits)
	case Int:
		return vx.ArithShift(bits)
	case ArithmeticShifter:
		return vx.ArithShift(bits)
	default:
		if try {
			panic(fmt.Errorf("invalid type for arithmetic shift: %T", ov))
		}
		try = true
		v = tobitwise(v)
		goto loop
	}
}

func bitwiseShift(v, bits Value) Value {
	var (
		ov  = v
		try bool
	)
loop:
	switch vx := v.(type) {
	case Uint:
		return vx.BitShift(bits)
	case Int:
		return vx.BitShift(bits)
	case BitShifter:
		return vx.BitShift(bits)
	default:
		if try {
			panic(fmt.Errorf("invalid type for bitwise shift: %T", ov))
		}
		try = true
		v = tobitwise(v)
		goto loop
	}
}
