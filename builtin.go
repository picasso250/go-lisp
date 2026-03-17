package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"math/big"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
)

// mustSameType 统一处理类型一致性检查
func mustSameType(name string, a, b interface{}) {
	if fmt.Sprintf("%T", a) != fmt.Sprintf("%T", b) {
		panic(fmt.Sprintf("%s: type mismatch", name))
	}
}

// binaryOp 算术运算工厂
func binaryOp(name string, opInt func(*big.Int, *big.Int) *big.Int, opFloat func(float64, float64) float64) func([]interface{}) interface{} {
	return func(args []interface{}) interface{} {
		a, b := args[0], args[1]
		mustSameType(name, a, b)
		switch av := a.(type) {
		case *big.Int:
			return opInt(av, b.(*big.Int))
		case float64:
			return opFloat(av, b.(float64))
		default:
			panic(fmt.Sprintf("%s: invalid types", name))
		}
	}
}

// compareOp 比较运算工厂
func compareOp(name string, opInt func(*big.Int, *big.Int) bool, opFloat func(float64, float64) bool, opStr func(string, string) bool) func([]interface{}) interface{} {
	return func(args []interface{}) interface{} {
		a, b := args[0], args[1]
		mustSameType(name, a, b)
		switch av := a.(type) {
		case *big.Int:
			return opInt(av, b.(*big.Int))
		case float64:
			return opFloat(av, b.(float64))
		case string:
			return opStr(av, b.(string))
		default:
			panic(fmt.Sprintf("%s: incomparable type", name))
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
	e.set("<", compareOp("<", func(a, b *big.Int) bool { return a.Cmp(b) < 0 }, func(a, b float64) bool { return a < b }, func(a, b string) bool { return a < b }))
	e.set("<=", compareOp("<=", func(a, b *big.Int) bool { return a.Cmp(b) <= 0 }, func(a, b float64) bool { return a <= b }, func(a, b string) bool { return a <= b }))
	e.set(">", compareOp(">", func(a, b *big.Int) bool { return a.Cmp(b) > 0 }, func(a, b float64) bool { return a > b }, func(a, b string) bool { return a > b }))
	e.set(">=", compareOp(">=", func(a, b *big.Int) bool { return a.Cmp(b) >= 0 }, func(a, b float64) bool { return a >= b }, func(a, b string) bool { return a >= b }))

	e.set("=", func(args []interface{}) interface{} {
		a, b := args[0], args[1]
		mustSameType("=", a, b)
		if bi, ok := a.(*big.Int); ok {
			return bi.Cmp(b.(*big.Int)) == 0
		}
		return a == b
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

	// Dictionary Ops
	e.set("dict", func(args []interface{}) interface{} {
		d := make(Dict)
		for i := 0; i < len(args); i += 2 {
			d[args[i].(string)] = args[i+1]
		}
		return d
	})
	e.set("dict-get", func(args []interface{}) interface{} {
		d, key := args[0].(Dict), args[1].(string)
		return d[key]
	})
	e.set("dict-set!", func(args []interface{}) interface{} {
		d, key, val := args[0].(Dict), args[1].(string), args[2]
		d[key] = val
		return val
	})
	e.set("dict-keys", func(args []interface{}) interface{} {
		d := args[0].(Dict)
		keys := make(List, 0, len(d))
		for k := range d {
			keys = append(keys, k)
		}
		return keys
	})
	e.set("dict-has?", func(args []interface{}) interface{} {
		d, key := args[0].(Dict), args[1].(string)
		_, ok := d[key]
		return ok
	})

	// TCP Ops
	e.set("tcp-listen", func(args []interface{}) interface{} {
		addr := args[0].(string)
		l, err := net.Listen("tcp", addr)
		if err != nil {
			panic(fmt.Sprintf("tcp-listen failed: %v", err))
		}
		return l
	})
	e.set("tcp-accept", func(args []interface{}) interface{} {
		l := args[0].(net.Listener)
		conn, err := l.Accept()
		if err != nil {
			panic(fmt.Sprintf("tcp-accept failed: %v", err))
		}
		return conn
	})
	e.set("tcp-connect", func(args []interface{}) interface{} {
		addr := args[0].(string)
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			panic(fmt.Sprintf("tcp-connect failed: %v", err))
		}
		return conn
	})
	e.set("tcp-send", func(args []interface{}) interface{} {
		conn, data := args[0].(net.Conn), args[1].(string)
		n, err := conn.Write([]byte(data))
		if err != nil {
			panic(fmt.Sprintf("tcp-send failed: %v", err))
		}
		return big.NewInt(int64(n))
	})
	e.set("tcp-read", func(args []interface{}) interface{} {
		conn := args[0].(net.Conn)
		max := 1024
		if len(args) > 1 {
			max = int(args[1].(*big.Int).Int64())
		}
		buf := make([]byte, max)
		n, err := conn.Read(buf)
		if err != nil && err != io.EOF {
			panic(fmt.Sprintf("tcp-read failed: %v", err))
		}
		return string(buf[:n])
	})
	e.set("tcp-close", func(args []interface{}) interface{} {
		if l, ok := args[0].(net.Listener); ok {
			l.Close()
		} else if c, ok := args[0].(net.Conn); ok {
			c.Close()
		}
		return nil
	})

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
		if len(l) <= 1 {
			return l
		}
		// Check for mixed types
		firstType := fmt.Sprintf("%T", l[0])
		for _, item := range l[1:] {
			if fmt.Sprintf("%T", item) != firstType {
				panic("sorted: mixed types")
			}
		}
		sort.Slice(l, func(i, j int) bool {
			switch av := l[i].(type) {
			case *big.Int:
				return av.Cmp(l[j].(*big.Int)) < 0
			case float64:
				return av < l[j].(float64)
			case string:
				return av < l[j].(string)
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
	e.set("input", func(args []interface{}) interface{} {
		if len(args) > 0 {
			fmt.Print(args[0])
		}
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			return scanner.Text()
		}
		return ""
	})
	e.set("type", func(args []interface{}) interface{} { return fmt.Sprintf("%T", args[0]) })
	e.set("bin", func(args []interface{}) interface{} { return "0b" + args[0].(*big.Int).Text(2) })
	e.set("hex", func(args []interface{}) interface{} { return "0x" + args[0].(*big.Int).Text(16) })

	return e
}
