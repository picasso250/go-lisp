package main

import (
	"fmt"
	"math/big"
	"os"
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

	// Arithmetic
	e.set("+", binaryOp("+", func(a, b *big.Int) *big.Int { return new(big.Int).Add(a, b) }, func(a, b float64) float64 { return a + b }))
	e.set("-", binaryOp("-", func(a, b *big.Int) *big.Int { return new(big.Int).Sub(a, b) }, func(a, b float64) float64 { return a - b }))
	e.set("*", binaryOp("*", func(a, b *big.Int) *big.Int { return new(big.Int).Mul(a, b) }, func(a, b float64) float64 { return a * b }))
	e.set("/", binaryOp("/", func(a, b *big.Int) *big.Int { return new(big.Int).Div(a, b) }, func(a, b float64) float64 { return a / b }))
	e.set("%", binaryOp("%", func(a, b *big.Int) *big.Int { return new(big.Int).Mod(a, b) }, func(a, b float64) float64 { panic("modulo not supported for floats") }))

	// Comparisons
	e.set("<", func(args []interface{}) interface{} {
		a, b := args[0], args[1]
		switch av := a.(type) {
		case *big.Int:
			return av.Cmp(b.(*big.Int)) < 0
		case float64:
			return av < b.(float64)
		default:
			return false
		}
	})
	e.set(">", func(args []interface{}) interface{} {
		a, b := args[0], args[1]
		switch av := a.(type) {
		case *big.Int:
			return av.Cmp(b.(*big.Int)) > 0
		case float64:
			return av > b.(float64)
		default:
			return false
		}
	})
	e.set("<=", func(args []interface{}) interface{} {
		a, b := args[0], args[1]
		switch av := a.(type) {
		case *big.Int:
			return av.Cmp(b.(*big.Int)) <= 0
		case float64:
			return av <= b.(float64)
		default:
			return false
		}
	})
	e.set(">=", func(args []interface{}) interface{} {
		a, b := args[0], args[1]
		switch av := a.(type) {
		case *big.Int:
			return av.Cmp(b.(*big.Int)) >= 0
		case float64:
			return av >= b.(float64)
		default:
			return false
		}
	})
	e.set("=", func(args []interface{}) interface{} {
		a, b := args[0], args[1]
		switch av := a.(type) {
		case *big.Int:
			bv, ok := b.(*big.Int)
			return ok && av.Cmp(bv) == 0
		default:
			return a == b
		}
	})
	e.set("!=", func(args []interface{}) interface{} {
		a, b := args[0], args[1]
		switch av := a.(type) {
		case *big.Int:
			bv, ok := b.(*big.Int)
			return !ok || av.Cmp(bv) != 0
		default:
			return a != b
		}
	})

	// Type Conversions
	e.set("float", func(args []interface{}) interface{} {
		switch v := args[0].(type) {
		case *big.Int:
			f, _ := new(big.Float).SetInt(v).Float64()
			return f
		case float64:
			return v
		case string:
			f, _ := strconv.ParseFloat(v, 64)
			return f
		default:
			panic("cannot convert to float")
		}
	})
	e.set("integer", func(args []interface{}) interface{} {
		switch v := args[0].(type) {
		case float64:
			return big.NewInt(int64(v))
		case *big.Int:
			return v
		case string:
			bi := new(big.Int)
			bi.SetString(v, 10)
			return bi
		default:
			panic("cannot convert to integer")
		}
	})

	// Type Predicates
	e.set("float?", func(args []interface{}) interface{} { _, ok := args[0].(float64); return ok })
	e.set("integer?", func(args []interface{}) interface{} { _, ok := args[0].(*big.Int); return ok })
	e.set("string?", func(args []interface{}) interface{} { _, ok := args[0].(string); return ok })
	e.set("atom?", func(args []interface{}) interface{} { _, ok := args[0].(List); return !ok })

	// List Operations
	e.set("car", func(args []interface{}) interface{} { return args[0].(List)[0] })
	e.set("cdr", func(args []interface{}) interface{} { return args[0].(List)[1:] })
	e.set("cons", func(args []interface{}) interface{} { return append(List{args[0]}, args[1].(List)...) })
	e.set("list", func(args []interface{}) interface{} { return List(args) })
	e.set("null?", func(args []interface{}) interface{} { return len(args[0].(List)) == 0 })
	e.set("length", func(args []interface{}) interface{} { return big.NewInt(int64(len(args[0].(List)))) })
	e.set("append", func(args []interface{}) interface{} {
		res := List{}
		for _, arg := range args {
			res = append(res, arg.(List)...)
		}
		return res
	})
	e.set("nth", func(args []interface{}) interface{} {
		idx := args[1].(*big.Int).Int64()
		return args[0].(List)[idx]
	})
	e.set("reverse", func(args []interface{}) interface{} {
		l := args[0].(List)
		res := make(List, len(l))
		for i, v := range l {
			res[len(l)-1-i] = v
		}
		return res
	})
	e.set("map", func(args []interface{}) interface{} {
		fn := args[0].(func([]interface{}) interface{})
		l := args[1].(List)
		res := make(List, len(l))
		for i, v := range l {
			res[i] = fn([]interface{}{v})
		}
		return res
	})
	e.set("filter", func(args []interface{}) interface{} {
		fn := args[0].(func([]interface{}) interface{})
		l := args[1].(List)
		res := List{}
		for _, v := range l {
			if test := fn([]interface{}{v}); test.(bool) {
				res = append(res, v)
			}
		}
		return res
	})
	e.set("reduce", func(args []interface{}) interface{} {
		fn := args[0].(func([]interface{}) interface{})
		l := args[1].(List)
		acc := args[2]
		for _, v := range l {
			acc = fn([]interface{}{acc, v})
		}
		return acc
	})

	// String Operations
	e.set("concat", func(args []interface{}) interface{} {
		res := ""
		for _, arg := range args {
			if bi, ok := arg.(*big.Int); ok {
				res += bi.String()
			} else {
				res += fmt.Sprintf("%v", arg)
			}
		}
		return res
	})
	e.set("string-split", func(args []interface{}) interface{} {
		s, sep := args[0].(string), args[1].(string)
		parts := strings.Split(s, sep)
		res := make(List, len(parts))
		for i, p := range parts {
			res[i] = p
		}
		return res
	})
	e.set("string-trim", func(args []interface{}) interface{} { return strings.TrimSpace(args[0].(string)) })
	e.set("string-length", func(args []interface{}) interface{} { return big.NewInt(int64(len(args[0].(string)))) })
	e.set("string-contains?", func(args []interface{}) interface{} { return strings.Contains(args[0].(string), args[1].(string)) })
	e.set("string-replace", func(args []interface{}) interface{} {
		return strings.ReplaceAll(args[0].(string), args[1].(string), args[2].(string))
	})

	// File Operations
	e.set("read-file", func(args []interface{}) interface{} {
		path := args[0].(string)
		content, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		return string(content)
	})
	e.set("write-file", func(args []interface{}) interface{} {
		path, content := args[0].(string), args[1].(string)
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			panic(err)
		}
		return true
	})

	return e
}
