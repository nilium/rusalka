package rvm

import (
	"fmt"
	"math"
)

type roundingMode int

const (
	rndTrunc roundingMode = iota
	rndNearest
	rndFloor
	rndCeil
)

func round(x float64, mode roundingMode) float64 {
	switch mode {
	case rndTrunc:
		return math.Trunc(x)
	case rndNearest:
		return math.Trunc(x + math.Copysign(0.5, x))
	case rndFloor:
		return math.Floor(x)
	case rndCeil:
		return math.Ceil(x)
	}
	// Invalid result:
	return math.Copysign(math.NaN(), x)
}

type (
	vnum  float64
	vint  int64
	vuint uint64

	arithValue interface {
		add(arithValue) arithValue
		neg() arithValue
		mul(arithValue) arithValue
		div(arithValue) arithValue
		mod(arithValue) arithValue
		pow(arithValue) arithValue
		sqrt() arithValue
	}

	bitsValue interface {
		arithValue
		xor(bitsValue) bitsValue
		and(bitsValue) bitsValue
		or(bitsValue) bitsValue
		not() bitsValue
	}

	floatValuer interface {
		Float64() float64
	}

	intValuer interface {
		Int64() int64
	}
)

var (
	_ arithValue = vnum(0)
	_ arithValue = vint(0)
	_ arithValue = vuint(0)
)

// Float64

func (lhs vnum) Float64() float64              { return float64(lhs) }
func (lhs vnum) Int64() int64                  { return int64(lhs) }
func (lhs vnum) add(rhs arithValue) arithValue { return lhs + tovnum(rhs) }
func (lhs vnum) mul(rhs arithValue) arithValue { return lhs * tovnum(rhs) }
func (lhs vnum) div(rhs arithValue) arithValue { return lhs / tovnum(rhs) }
func (lhs vnum) neg() arithValue               { return -lhs }
func (lhs vnum) sqrt() arithValue              { return vnum(math.Sqrt(float64(lhs))) }

func (lhs vnum) pow(rhs arithValue) arithValue {
	return vnum(math.Pow(float64(lhs), float64(tovnum(rhs))))
}

func (lhs vnum) mod(rhs arithValue) arithValue {
	return vnum(math.Mod(float64(lhs), float64(tovnum(rhs))))
}

// Signed integer

func (lhs vint) Float64() float64 { return float64(lhs) }
func (lhs vint) Int64() int64     { return int64(lhs) }
func (lhs vint) neg() arithValue  { return -lhs }

func (lhs vint) add(rhs arithValue) arithValue {
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

func (lhs vint) mul(rhs arithValue) arithValue {
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

func (lhs vint) div(rhs arithValue) arithValue {
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

func (lhs vint) mod(rhs arithValue) arithValue {
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

func (lhs vint) sqrt() arithValue { return vint(math.Sqrt(float64(lhs))) }

func (lhs vint) pow(rhs arithValue) arithValue {
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

// Unsigned integer

func (lhs vuint) Float64() float64 { return float64(lhs) }
func (lhs vuint) Int64() int64     { return int64(lhs) }
func (lhs vuint) neg() arithValue  { return -lhs }

func (lhs vuint) add(rhs arithValue) arithValue {
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

func (lhs vuint) mul(rhs arithValue) arithValue {
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

func (lhs vuint) div(rhs arithValue) arithValue {
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

func (lhs vuint) mod(rhs arithValue) arithValue {
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

func (lhs vuint) sqrt() arithValue { return vuint(math.Sqrt(float64(lhs))) }

func (lhs vuint) pow(rhs arithValue) arithValue {
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

func toarith(v interface{}) (r arithValue) {
	switch v := v.(type) {
	case vnum:
		f := float64(v)
		return vnum(f)
	case vint:
		return v
	case vuint:
		return v
	case floatValuer:
		return vnum(v.Float64())
	case intValuer:
		return vint(v.Int64())
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
		panic(fmt.Errorf("unable to convert %T to float64", v))
	}
}

func tovnum(v interface{}) vnum {
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

func tovint(v interface{}) vint {
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

func tovuint(v interface{}) vuint {
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
