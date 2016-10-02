package rvm

import (
	"fmt"
	"math"
)

type InvalidRoundingMode roundingMode

func (i InvalidRoundingMode) Error() string {
	return fmt.Sprintf("invalid rounding mode: %x", i)
}

type roundingMode uint

const (
	rndTrunc roundingMode = iota
	rndNearest
	rndFloor
	rndCeil
)

func arithShift(v, bits Value) Value {
	var (
		ov  = v
		try bool
	)
loop:
	switch vx := v.(type) {
	case vuint:
		return vx.ArithShift(bits)
	case vint:
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
	case vuint:
		return vx.BitShift(bits)
	case vint:
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

func round(v Value, mode roundingMode) Value {
	if mode > rndCeil {
		panic("invalid rounding mode")
	}

loop:
	switch vx := v.(type) {
	case vuint, vint:
		return vx
	case vnum:
		return vx.Round(mode)
	case Rounder:
		return vx.Round(mode)
	case float64:
		return vnum(vx).Round(mode)
	default:
		v = toarith(vx)
		goto loop
	}
}

type (
	vnum  float64
	vint  int64
	vuint uint64

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
		Round(roundingMode) Value
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
	_ Arith = vnum(0)
	_ Arith = vint(0)
	_ Arith = vuint(0)

	_ Bitwise = vint(0)
	_ Bitwise = vuint(0)

	_ Rounder = vnum(0)
	_ Rounder = vint(0)
	_ Rounder = vuint(0)
)

// Float64

func (lhs vnum) Float64() float64    { return float64(lhs) }
func (lhs vnum) Int64() int64        { return int64(lhs) }
func (lhs vnum) Uint64() uint64      { return uint64(lhs) }
func (lhs vnum) Add(rhs Arith) Arith { return lhs + tovnum(rhs) }
func (lhs vnum) Sub(rhs Arith) Arith { return lhs - tovnum(rhs) }
func (lhs vnum) Mul(rhs Arith) Arith { return lhs * tovnum(rhs) }
func (lhs vnum) Div(rhs Arith) Arith { return lhs / tovnum(rhs) }
func (lhs vnum) Neg() Arith          { return -lhs }
func (lhs vnum) Sqrt() Arith         { return vnum(math.Sqrt(float64(lhs))) }

func (lhs vnum) Round(mode roundingMode) Value {
	switch x := float64(lhs); mode {
	case rndTrunc:
		return math.Trunc(x)
	case rndNearest:
		return math.Trunc(x + math.Copysign(0.5, x))
	case rndFloor:
		return math.Floor(x)
	case rndCeil:
		return math.Ceil(x)
	}
	panic("unreachable")
}

func (lhs vnum) Pow(rhs Arith) Arith {
	return vnum(math.Pow(float64(lhs), float64(tovnum(rhs))))
}

func (lhs vnum) Mod(rhs Arith) Arith {
	return vnum(math.Mod(float64(lhs), float64(tovnum(rhs))))
}

// Signed integer

func (lhs vint) Float64() float64 { return float64(lhs) }
func (lhs vint) Int64() int64     { return int64(lhs) }
func (lhs vint) Uint64() uint64   { return uint64(lhs) }
func (lhs vint) Neg() Arith       { return -lhs }

func (lhs vint) Round(roundingMode) Value { return lhs }

func (lhs vint) ArithShift(bits Value) Value {
	if bits := tovint(bits); bits < 0 {
		return lhs << uint(-bits)
	} else if bits > 0 {
		return lhs >> uint(bits)
	}
	return lhs
}

func (lhs vint) BitShift(bits Value) Value {
	if bits := tovint(bits); bits < 0 {
		return vint(uint64(lhs) << uint(-bits))
	} else if bits > 0 {
		return vint(uint64(lhs) >> uint(bits))
	}
	return lhs
}

func (lhs vint) Add(rhs Arith) Arith {
	switch rhs := toarith(rhs).(type) {
	case vint:
		return vint(int64(lhs) + int64(rhs))
	case vuint:
		return vint(int64(lhs) + int64(rhs))
	case vnum:
		return vnum(float64(lhs) + float64(rhs))
	}
	panic("unreachable")
}

func (lhs vint) Sub(rhs Arith) Arith {
	switch rhs := toarith(rhs).(type) {
	case vint:
		return vint(int64(lhs) - int64(rhs))
	case vuint:
		return vint(int64(lhs) - int64(rhs))
	case vnum:
		return vnum(float64(lhs) - float64(rhs))
	}
	panic("unreachable")
}

func (lhs vint) Mul(rhs Arith) Arith {
	switch rhs := toarith(rhs).(type) {
	case vint:
		return vint(int64(lhs) * int64(rhs))
	case vuint:
		return vint(int64(lhs) * int64(rhs))
	case vnum:
		return vnum(float64(lhs) * float64(rhs))
	}
	panic("unreachable")
}

func (lhs vint) Div(rhs Arith) Arith {
	switch rhs := toarith(rhs).(type) {
	case vint:
		return vint(int64(lhs) / int64(rhs))
	case vuint:
		return vint(int64(lhs) / int64(rhs))
	case vnum:
		return vnum(float64(lhs) / float64(rhs))
	}
	panic("unreachable")
}

func (lhs vint) Mod(rhs Arith) Arith {
	switch rhs := toarith(rhs).(type) {
	case vint:
		return vint(int64(lhs) % int64(rhs))
	case vuint:
		return vint(int64(lhs) % int64(rhs))
	case vnum:
		return vnum(math.Mod(float64(lhs), float64(rhs)))
	}
	panic("unreachable")
}

func (lhs vint) Sqrt() Arith { return vint(math.Sqrt(float64(lhs))) }

func (lhs vint) Pow(rhs Arith) Arith {
	switch rhs := toarith(rhs).(type) {
	case vint:
		if rhs == 0 {
			return vuint(1)
		} else if rhs < 0 {
			return vnum(math.Pow(float64(lhs), float64(rhs)))
		}
		for q, i := lhs, vint(0); i < rhs; i++ {
			lhs = lhs * q
		}
		return lhs
	case vuint:
		if rhs == 0 {
			return vuint(1)
		}
		for q, i := lhs, vuint(0); i < rhs; i++ {
			lhs = lhs * q
		}
		return lhs
	case vnum:
		return vnum(math.Pow(float64(lhs), float64(rhs)))
	}
	panic("unreachable")
}

func (lhs vint) Xor(rhs Bitwise) Bitwise { return vint(uint64(lhs) ^ uint64(tovuint(rhs))) }
func (lhs vint) And(rhs Bitwise) Bitwise { return vint(uint64(lhs) & uint64(tovuint(rhs))) }
func (lhs vint) Or(rhs Bitwise) Bitwise  { return vint(uint64(lhs) | uint64(tovuint(rhs))) }
func (lhs vint) Not() Bitwise            { return vint(^uint64(lhs)) }

// Unsigned integer

func (lhs vuint) Float64() float64 { return float64(lhs) }
func (lhs vuint) Int64() int64     { return int64(lhs) }
func (lhs vuint) Uint64() uint64   { return uint64(lhs) }
func (lhs vuint) Neg() Arith       { return -lhs }

func (lhs vuint) Round(roundingMode) Value { return lhs }

func (lhs vuint) ArithShift(bits Value) Value {
	if bits := tovint(bits); bits < 0 {
		return vuint(int64(lhs) << uint(-bits))
	} else if bits > 0 {
		return vuint(int64(lhs) >> uint(bits))
	}
	return lhs
}

func (lhs vuint) BitShift(bits Value) Value {
	if bits := tovint(bits); bits < 0 {
		return lhs << uint(-bits)
	} else if bits > 0 {
		return lhs >> uint(bits)
	}
	return lhs
}

func (lhs vuint) Add(rhs Arith) Arith {
	switch rhs := toarith(rhs).(type) {
	case vuint:
		return vuint(uint64(lhs) + uint64(rhs))
	case vint:
		return vuint(int64(lhs) + int64(rhs))
	case vnum:
		return vnum(float64(lhs) + float64(rhs))
	}
	panic("unreachable")
}

func (lhs vuint) Sub(rhs Arith) Arith {
	switch rhs := toarith(rhs).(type) {
	case vuint:
		return vuint(uint64(lhs) - uint64(rhs))
	case vint:
		return vuint(int64(lhs) - int64(rhs))
	case vnum:
		return vnum(float64(lhs) - float64(rhs))
	}
	panic("unreachable")
}

func (lhs vuint) Mul(rhs Arith) Arith {
	switch rhs := toarith(rhs).(type) {
	case vuint:
		return vint(uint64(lhs) * uint64(rhs))
	case vint:
		return vuint(int64(lhs) * int64(rhs))
	case vnum:
		return vnum(float64(lhs) * float64(rhs))
	}
	panic("unreachable")
}

func (lhs vuint) Div(rhs Arith) Arith {
	switch rhs := toarith(rhs).(type) {
	case vuint:
		return vint(uint64(lhs) / uint64(rhs))
	case vint:
		return vuint(int64(lhs) / int64(rhs))
	case vnum:
		return vnum(float64(lhs) / float64(rhs))
	}
	panic("unreachable")
}

func (lhs vuint) Mod(rhs Arith) Arith {
	switch rhs := toarith(rhs).(type) {
	case vuint:
		return vint(uint64(lhs) % uint64(rhs))
	case vint:
		return vuint(int64(lhs) % int64(rhs))
	case vnum:
		return vnum(math.Mod(float64(lhs), float64(rhs)))
	}
	panic("unreachable")
}

func (lhs vuint) Sqrt() Arith { return vuint(math.Sqrt(float64(lhs))) }

func (lhs vuint) Pow(rhs Arith) Arith {
	switch rhs := toarith(rhs).(type) {
	case vuint:
		if rhs == 0 {
			return vuint(1)
		}
		for q, i := lhs, vuint(0); i < rhs; i++ {
			lhs = lhs * q
		}
		return lhs
	case vint:
		if rhs == 0 {
			return vuint(1)
		} else if rhs < 0 {
			return vnum(math.Pow(float64(lhs), float64(rhs)))
		}
		for q, i := lhs, vint(0); i < rhs; i++ {
			lhs = lhs * q
		}
		return lhs
	case vnum:
		return vnum(math.Pow(float64(lhs), float64(rhs)))
	}
	panic("unreachable")
}

func (lhs vuint) Xor(rhs Bitwise) Bitwise { return lhs ^ tovuint(rhs) }
func (lhs vuint) And(rhs Bitwise) Bitwise { return lhs & tovuint(rhs) }
func (lhs vuint) Or(rhs Bitwise) Bitwise  { return lhs | tovuint(rhs) }
func (lhs vuint) Not() Bitwise            { return ^lhs }

func toarith(v Value) (r Arith) {
	switch v := v.(type) {
	case Arith:
		return v
	case FloatValuer:
		return vnum(v.Float64())
	case IntValuer:
		return vint(v.Int64())
	case UintValuer:
		return vuint(v.Uint64())
	case int:
		return vint(v)
	case int64:
		return vint(v)
	case int32:
		return vint(v)
	case int16:
		return vint(v)
	case float64:
		return vnum(v)
	case float32:
		return vnum(v)
	case uint:
		return vuint(v)
	case uint64:
		return vuint(v)
	case uint32:
		return vuint(v)
	case uint16:
		return vuint(v)
	case uint8:
		return vuint(v)
	default:
		panic(fmt.Errorf("unable to convert %T to arithmetic type", v))
	}
}

func tobitwise(v Value) (r Bitwise) {
	switch v := v.(type) {
	case Bitwise:
		return v
	case IntValuer:
		return vint(v.Int64())
	case UintValuer:
		return vuint(v.Uint64())
	case float64:
		return vint(v)
	case float32:
		return vint(v)
	case int:
		return vint(v)
	case int64:
		return vint(v)
	case int32:
		return vint(v)
	case int16:
		return vint(v)
	case uint:
		return vuint(v)
	case uint64:
		return vuint(v)
	case uint32:
		return vuint(v)
	case uint16:
		return vuint(v)
	case uint8:
		return vuint(v)
	default:
		panic(fmt.Errorf("unable to convert %T to bitwise type", v))
	}
}

func tovnum(v Value) vnum {
	switch v := toarith(v).(type) {
	case vnum:
		return v
	case vint:
		return vnum(v)
	case vuint:
		return vnum(v)
	}
	panic("unreachable")
}

func tovint(v Value) vint {
	switch v := toarith(v).(type) {
	case vint:
		return v
	case vnum:
		return vint(v)
	case vuint:
		return vint(v)
	}
	panic("unreachable")
}

func tovuint(v Value) vuint {
	switch v := toarith(v).(type) {
	case vint:
		return vuint(v)
	case vnum:
		return vuint(v)
	case vuint:
		return v
	}
	panic("unreachable")
}
