package main

import (
	"fmt"
	"math"
	"math/big"
	"os"
	"sort"
	"strconv"
	"strings"
)

func binaryOp(name string, opInt func(*big.Int, *big.Int) *big.Int, opFloat func(float64, float64) float64) func([]interface{}) interface{} {
	return func(args []interface{}) interface{} {
		a, b := args[0], args[1]
		switch av := a.(type) {
		case *big.Int:
			bv, ok := b.(*big.Int)
			if !ok {
				panic(fmt.Sprintf("%s: type mismatch, expected integers", name))
			}
			return opInt(av, bv)
		case float64:
			bv, ok := b.(float64)
			if !ok {
				panic(fmt.Sprintf("%s: type mismatch, expected floats", name))
			}
			return opFloat(av, bv)
		default:
			panic(fmt.Sprintf("%s: invalid types", name))
		}
	}
}

func standardEnv() *Env {
	e := &Env{vars: make(map[Symbol]interface{}), outer: nil}

	// Basic Arithmetic
	e.set("+", binaryOp("+", func(a, b *big.Int) *big.Int { return new(big.Int).Add(a, b) }, func(a, b float64) float64 { return a + b }))
	e.set("-", binaryOp("-", func(a, b *big.Int) *big.Int { return new(big.Int).Sub(a, b) }, func(a, b float64) float64 { return a - b }))
	e.set("*", binaryOp("*", func(a, b *big.Int) *big.Int { return new(big.Int).Mul(a, b) }, func(a, b float64) float64 { return a * b }))
	e.set("/", binaryOp("/", func(a, b *big.Int) *big.Int { return new(big.Int).Div(a, b) }, func(a, b float64) float64 { return a / b }))
	e.set("%", binaryOp("%", func(a, b *big.Int) *big.Int { return new(big.Int).Mod(a, b) }, func(a, b float64) float64 { return math.Mod(a, b) }))

	// Comparisons
	e.set("<", func(args []interface{}) interface{} {
		a, b := args[0], args[1]
		switch av := a.(type) {
		case *big.Int:
			bv, ok := b.(*big.Int)
			if !ok { panic("<: type mismatch") }
			return av.Cmp(bv) < 0
		case float64:
			bv, ok := b.(float64)
			if !ok { panic("<: type mismatch") }
			return av < bv
		case string:
			bv, ok := b.(string)
			if !ok { panic("<: type mismatch") }
			return av < bv
		default:
			panic("<: incomparable type")
		}
	})
	e.set("<=", func(args []interface{}) interface{} {
		a, b := args[0], args[1]
		switch av := a.(type) {
		case *big.Int:
			bv, ok := b.(*big.Int)
			if !ok { panic("<=: type mismatch") }
			return av.Cmp(bv) <= 0
		case float64:
			bv, ok := b.(float64)
			if !ok { panic("<=: type mismatch") }
			return av <= bv
		case string:
			bv, ok := b.(string)
			if !ok { panic("<=: type mismatch") }
			return av <= bv
		default:
			panic("<=: incomparable type")
		}
	})
	e.set(">", func(args []interface{}) interface{} {
		a, b := args[0], args[1]
		switch av := a.(type) {
		case *big.Int:
			bv, ok := b.(*big.Int)
			if !ok { panic(">: type mismatch") }
			return av.Cmp(bv) > 0
		case float64:
			bv, ok := b.(float64)
			if !ok { panic(">: type mismatch") }
			return av > bv
		case string:
			bv, ok := b.(string)
			if !ok { panic(">: type mismatch") }
			return av > bv
		default:
			panic(">: incomparable type")
		}
	})
	e.set(">=", func(args []interface{}) interface{} {
		a, b := args[0], args[1]
		switch av := a.(type) {
		case *big.Int:
			bv, ok := b.(*big.Int)
			if !ok { panic(">=: type mismatch") }
			return av.Cmp(bv) >= 0
		case float64:
			bv, ok := b.(float64)
			if !ok { panic(">=: type mismatch") }
			return av >= bv
		case string:
			bv, ok := b.(string)
			if !ok { panic(">=: type mismatch") }
			return av >= bv
		default:
			panic(">=: incomparable type")
		}
	})
	e.set("=", func(args []interface{}) interface{} {
		a, b := args[0], args[1]
		if fmt.Sprintf("%T", a) != fmt.Sprintf("%T", b) {
			panic("=: type mismatch")
		}
		switch av := a.(type) {
		case *big.Int:
			return av.Cmp(b.(*big.Int)) == 0
		default:
			return a == b
		}
	})
	e.set("!=", func(args []interface{}) interface{} {
		return !e.get("=").(func([]interface{}) interface{})(args).(bool)
	})

	// Python-like Math
	e.set("abs", func(args []interface{}) interface{} {
		switch v := args[0].(type) {
		case *big.Int: return new(big.Int).Abs(v)
		case float64: return math.Abs(v)
		default: panic("abs requires number")
		}
	})
	e.set("pow", func(args []interface{}) interface{} {
		a, b := args[0], args[1]
		if ai, ok := a.(*big.Int); ok { return new(big.Int).Exp(ai, b.(*big.Int), nil) }
		return math.Pow(a.(float64), b.(float64))
	})
	e.set("divmod", func(args []interface{}) interface{} {
		a, b := args[0].(*big.Int), args[1].(*big.Int)
		q, r := new(big.Int).DivMod(a, b, new(big.Int))
		return List{q, r}
	})
	e.set("round", func(args []interface{}) interface{} { return math.Round(args[0].(float64)) })

	// Conversions & Predicates
	e.set("int", func(args []interface{}) interface{} {
		switch v := args[0].(type) {
		case float64: return big.NewInt(int64(v))
		case string: bi := new(big.Int); bi.SetString(v, 10); return bi
		default: return v.(*big.Int)
		}
	})
	e.set("integer", e.get("int")) // alias
	e.set("float", func(args []interface{}) interface{} {
		switch v := args[0].(type) {
		case *big.Int: f, _ := new(big.Float).SetInt(v).Float64(); return f
		case string: f, _ := strconv.ParseFloat(v, 64); return f
		case float64: return v
		default: panic(fmt.Sprintf("float conversion failed: unsupported type %T", v))
		}
	})
	e.set("str", func(args []interface{}) interface{} {
		if bi, ok := args[0].(*big.Int); ok { return bi.String() }
		return fmt.Sprintf("%v", args[0])
	})
	e.set("bool", func(args []interface{}) interface{} {
		if args[0] == nil || args[0] == false { return false }
		if bi, ok := args[0].(*big.Int); ok && bi.Sign() == 0 { return false }
		if f, ok := args[0].(float64); ok && f == 0 { return false }
		if s, ok := args[0].(string); ok && s == "" { return false }
		if l, ok := args[0].(List); ok && len(l) == 0 { return false }
		return true
	})
	e.set("integer?", func(args []interface{}) interface{} { _, ok := args[0].(*big.Int); return ok })
	e.set("float?", func(args []interface{}) interface{} { _, ok := args[0].(float64); return ok })
	e.set("string?", func(args []interface{}) interface{} { _, ok := args[0].(string); return ok })

	// List Ops
	e.set("car", func(args []interface{}) interface{} { return args[0].(List)[0] })
	e.set("cdr", func(args []interface{}) interface{} { return args[0].(List)[1:] })
	e.set("cons", func(args []interface{}) interface{} { return append(List{args[0]}, args[1].(List)...) })
	e.set("list", func(args []interface{}) interface{} { return List(args) })
	e.set("null?", func(args []interface{}) interface{} { return len(args[0].(List)) == 0 })
	e.set("length", func(args []interface{}) interface{} { return big.NewInt(int64(len(args[0].(List)))) })
	e.set("append", func(args []interface{}) interface{} {
		res := List{}
		for _, arg := range args { res = append(res, arg.(List)...) }
		return res
	})
	e.set("nth", func(args []interface{}) interface{} {
		idx := args[1].(*big.Int).Int64()
		return args[0].(List)[idx]
	})
	e.set("reverse", func(args []interface{}) interface{} {
		l := args[0].(List)
		res := make(List, len(l))
		for i, v := range l { res[len(l)-1-i] = v }
		return res
	})
	e.set("reversed", e.get("reverse"))

	// String Ops
	e.set("concat", func(args []interface{}) interface{} {
		res := ""
		for _, arg := range args {
			if bi, ok := arg.(*big.Int); ok { res += bi.String() } else { res += fmt.Sprintf("%v", arg) }
		}
		return res
	})
	e.set("string-split", func(args []interface{}) interface{} {
		s, sep := args[0].(string), args[1].(string)
		parts := strings.Split(s, sep)
		res := make(List, len(parts))
		for i, p := range parts { res[i] = p }
		return res
	})
	e.set("string-trim", func(args []interface{}) interface{} { return strings.TrimSpace(args[0].(string)) })
	e.set("string-length", func(args []interface{}) interface{} { return big.NewInt(int64(len(args[0].(string)))) })
	e.set("string-contains?", func(args []interface{}) interface{} { return strings.Contains(args[0].(string), args[1].(string)) })
	e.set("string-replace", func(args []interface{}) interface{} {
		return strings.ReplaceAll(args[0].(string), args[1].(string), args[2].(string))
	})

	// Files
	e.set("read-file", func(args []interface{}) interface{} {
		content, _ := os.ReadFile(args[0].(string)); return string(content)
	})
	e.set("write-file", func(args []interface{}) interface{} {
		os.WriteFile(args[0].(string), []byte(args[1].(string)), 0644); return true
	})

	// Iterables
	e.set("range", func(args []interface{}) interface{} {
		start := int64(0); stop := args[0].(*big.Int).Int64()
		if len(args) > 1 { start = stop; stop = args[1].(*big.Int).Int64() }
		res := List{}
		for i := start; i < stop; i++ { res = append(res, big.NewInt(i)) }
		return res
	})
	e.set("sorted", func(args []interface{}) interface{} {
		l := append(List{}, args[0].(List)...)
		sort.Slice(l, func(i, j int) bool {
			switch av := l[i].(type) {
			case *big.Int:
				bv, ok := l[j].(*big.Int)
				return ok && av.Cmp(bv) < 0
			case float64:
				bv, ok := l[j].(float64)
				return ok && av < bv
			case string:
				bv, ok := l[j].(string)
				return ok && av < bv
			default:
				return false
			}
		})
		return l
	})

	// Higher Order
	e.set("map", func(args []interface{}) interface{} {
		fn := args[0].(func([]interface{}) interface{})
		l := args[1].(List)
		res := make(List, len(l))
		for i, v := range l { res[i] = fn([]interface{}{v}) }
		return res
	})
	e.set("filter", func(args []interface{}) interface{} {
		fn := args[0].(func([]interface{}) interface{})
		l := args[1].(List)
		res := List{}
		for _, v := range l { if test := fn([]interface{}{v}); test.(bool) { res = append(res, v) } }
		return res
	})
	e.set("reduce", func(args []interface{}) interface{} {
		fn := args[0].(func([]interface{}) interface{})
		l := args[1].(List)
		acc := args[2]
		for _, v := range l { acc = fn([]interface{}{acc, v}) }
		return acc
	})

	// System & Debug
	e.set("print", func(args []interface{}) interface{} {
		for i, arg := range args {
			if i > 0 { fmt.Print(" ") }
			if bi, ok := arg.(*big.Int); ok { fmt.Print(bi.String()) } else { fmt.Print(arg) }
		}
		fmt.Println(); return nil
	})
	e.set("type", func(args []interface{}) interface{} { return fmt.Sprintf("%T", args[0]) })
	e.set("bin", func(args []interface{}) interface{} { return "0b" + args[0].(*big.Int).Text(2) })
	e.set("hex", func(args []interface{}) interface{} { return "0x" + args[0].(*big.Int).Text(16) })

	return e
}
